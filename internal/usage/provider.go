// Codex usage provider.
// 监听 ~/.codex/sessions/**/*.jsonl，chokidar → fsnotify，加 3s 轮询兜底。
// 内存聚合：今日 Token / 累计 Token / 本任务 / 上次对话。

package usage

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// UsageProvider 统一接口
type UsageProvider interface {
	ID() string
	Name() string
	ShortName() string
	Color() string
	Emoji() string
	Poll() *Snapshot
	Refresh() *Snapshot
	Start(onUpdate func(*Snapshot)) error
	Stop() error
}

type CodexProvider struct {
	home         string
	sessionsDir  string
	watcher      *fsnotify.Watcher
	pollTimer    *time.Ticker
	fileOffsets  map[string]int64
	lastMtimes   map[string]int64
	tokenEvents  []*Event
	fileTotals   map[string]int64 // 文件路径 -> 该会话终身 Token（格式A累计快照 + 格式B增量）
	latestQuota  *Event           // JSONL 里的最新额度快照（响应间隔期间会过期）
	liveQuota    *Event           // wham/usage 实时额度（在线快照，由统一调度器更新）
	stop         chan struct{}
	stopOnce     sync.Once
	workers      sync.WaitGroup
	refreshMu    sync.Mutex
	fetchQuota   func() (*Event, error)
	onUpdate     func(*Snapshot)
	mu           sync.RWMutex
	ingestMu     sync.Mutex
	refreshError string
}

func NewCodexProvider() *CodexProvider {
	home := os.Getenv("CODEX_HOME")
	if home == "" {
		if h, err := os.UserHomeDir(); err == nil {
			home = filepath.Join(h, ".codex")
		}
	}
	return &CodexProvider{
		home:        home,
		sessionsDir: filepath.Join(home, "sessions"),
		fileOffsets: make(map[string]int64),
		lastMtimes:  make(map[string]int64),
		stop:        make(chan struct{}),
		fetchQuota:  fetchWhamUsage,
		fileTotals:  make(map[string]int64),
	}
}

func (p *CodexProvider) ID() string        { return "codex" }
func (p *CodexProvider) Name() string      { return "Codex" }
func (p *CodexProvider) ShortName() string { return "CODEX" }
func (p *CodexProvider) Color() string     { return "#3FA8C7" }
func (p *CodexProvider) Emoji() string     { return "◉" }

func (p *CodexProvider) Start(onUpdate func(*Snapshot)) error {
	p.onUpdate = onUpdate

	// 冷启动：全量扫描（sessions + archived_sessions，逐文件 ingest 以构建终身累计）
	if _, err := os.Stat(p.sessionsDir); err == nil {
		for _, f := range walkAll(p.sessionsDir) {
			p.ingestFile(f)
		}
		// 归档会话是只读历史，一次性 ingest，不纳入监听
		archived := filepath.Join(p.home, "archived_sessions")
		if _, err := os.Stat(archived); err == nil {
			for _, f := range walkAll(archived) {
				p.ingestFile(f)
			}
		}
		p.emit()
	}

	// 启动 fsnotify
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	p.watcher = w
	_ = p.addAllDirs()

	p.workers.Add(2)
	go func() { defer p.workers.Done(); p.watchLoop() }()
	// Local file detection stays responsive; online requests belong to the scheduler.
	p.pollTimer = time.NewTicker(3 * time.Second)
	go func() {
		defer p.workers.Done()
		for {
			select {
			case <-p.stop:
				return
			case <-p.pollTimer.C:
				p.pollOnce()
			}
		}
	}()
	return nil
}

// Refresh performs real online collection. Poll only reads the current snapshot.
func (p *CodexProvider) Refresh() *Snapshot {
	p.refreshMu.Lock()
	defer p.refreshMu.Unlock()
	p.pollOnce()
	// Publish local data first; network latency must not delay local sync.
	p.emit()
	ev, err := p.fetchQuota()
	p.mu.Lock()
	defer p.mu.Unlock()
	if err == nil && (ev == nil || (ev.Primary == nil && ev.Secondary == nil)) {
		err = fmt.Errorf("missing quota windows")
	}
	if err == nil {
		p.liveQuota = ev
	}
	p.refreshError = ""
	if err != nil {
		p.refreshError = refreshErrorMessage(err)
	}
	return p.buildSnapshot()
}

func (p *CodexProvider) addAllDirs() error {
	if _, err := os.Stat(p.sessionsDir); err != nil {
		return nil
	}
	return filepath.WalkDir(p.sessionsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return p.watcher.Add(path)
		}
		return nil
	})
}

func (p *CodexProvider) watchLoop() {
	for {
		select {
		case ev, ok := <-p.watcher.Events:
			if !ok {
				return
			}
			p.handleFsEvent(ev)
		case _, ok := <-p.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (p *CodexProvider) handleFsEvent(ev fsnotify.Event) {
	// 新建目录 → 添加监听
	if ev.Has(fsnotify.Create) {
		if fi, err := os.Stat(ev.Name); err == nil && fi.IsDir() {
			_ = p.watcher.Add(ev.Name)
			return
		}
	}
	if !strings.HasSuffix(ev.Name, ".jsonl") {
		return
	}
	p.ingestFile(ev.Name)
}

func (p *CodexProvider) ingestFile(path string) {
	p.ingestMu.Lock()
	defer p.ingestMu.Unlock()
	p.mu.RLock()
	prev := p.fileOffsets[path]
	p.mu.RUnlock()

	events, newOffset := ReadNewEvents(path, prev)

	p.mu.Lock()
	p.fileOffsets[path] = newOffset
	if fi, err := os.Stat(path); err == nil {
		p.lastMtimes[path] = fi.ModTime().UnixMilli()
	}
	p.applyEvents(path, events)
	p.mu.Unlock()
}

func (p *CodexProvider) applyEvents(path string, events []*Event) {
	changed := false
	for _, e := range events {
		switch e.Kind {
		case "quota":
			// 只有带窗口信息的才是额度快照（部分 token_count 只有会话累计）
			if e.Primary != nil || e.Secondary != nil {
				if p.latestQuota == nil || e.Ts > p.latestQuota.Ts {
					p.latestQuota = e
					changed = true
				}
			}
			// 格式 A：会话累计快照（累计值非增量），单调递增取 max
			if e.SessionTotal != nil {
				st := int64(e.SessionTotal.TotalTokens)
				if st > p.fileTotals[path] {
					p.fileTotals[path] = st
					changed = true
				}
			}
		case "token":
			dup := false
			if e.ResponseId != nil {
				for _, t := range p.tokenEvents {
					if t.ResponseId != nil && *t.ResponseId == *e.ResponseId && t.Ts == e.Ts {
						dup = true
						break
					}
				}
			}
			if !dup {
				p.tokenEvents = append(p.tokenEvents, e)
				if e.Usage != nil {
					// 格式 B：per-response 增量累加
					p.fileTotals[path] += int64(e.Usage.TotalTokens)
				}
				changed = true
			}
		}
	}
	// 注意：不做 GC —— "累计 Token"是终身口径，丢事件就少算。
	// 个人工具一年也就十几 MB 内存，可接受。
	if changed {
		p.emitLocked()
	}
}

func (p *CodexProvider) pollOnce() {
	// 已知文件检查 mtime
	p.mu.RLock()
	known := make(map[string]int64, len(p.fileOffsets))
	for f, off := range p.fileOffsets {
		known[f] = off
	}
	p.mu.RUnlock()

	for f, offset := range known {
		fi, err := os.Stat(f)
		if err != nil {
			p.mu.Lock()
			delete(p.fileOffsets, f)
			delete(p.lastMtimes, f)
			p.mu.Unlock()
			continue
		}
		mtime := fi.ModTime().UnixMilli()
		p.mu.RLock()
		prev := p.lastMtimes[f]
		p.mu.RUnlock()
		if mtime != prev || fi.Size() != offset {
			p.ingestFile(f)
		}
	}

	// 发现新文件
	for _, f := range walkAll(p.sessionsDir) {
		p.mu.RLock()
		_, exists := p.fileOffsets[f]
		p.mu.RUnlock()
		if !exists {
			p.ingestFile(f)
		}
	}
}

func (p *CodexProvider) emit() {
	p.mu.Lock()
	p.emitLocked()
	p.mu.Unlock()
}

func (p *CodexProvider) emitLocked() {
	if p.onUpdate == nil {
		return
	}
	p.onUpdate(p.buildSnapshot())
}

func (p *CodexProvider) buildSnapshot() *Snapshot {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	last7d := now.UnixMilli() - 7*86400*1000

	var today, near7d int64
	var total int64
	for _, v := range p.fileTotals {
		total += v
	}
	var todayInput, todayCached int64
	var lastEventTs int64
	var lastTurnTs, latestThreadTs int64
	var lastTurn, currentTask int64
	threads := make(map[string]struct {
		ts    int64
		total int64
	})

	for _, t := range p.tokenEvents {
		if t.Usage != nil {
			if t.Ts >= midnight {
				today += int64(t.Usage.TotalTokens)
				todayInput += int64(t.Usage.InputTokens)
				todayCached += int64(t.Usage.CachedInputTokens)
			}
			if t.Ts >= last7d {
				near7d += int64(t.Usage.TotalTokens)
			}
		}
		if t.Ts > lastEventTs {
			lastEventTs = t.Ts
		}
		if t.ThreadId != nil && t.ThreadTokenUsage != nil {
			cur, ok := threads[*t.ThreadId]
			if !ok || t.Ts > cur.ts {
				threads[*t.ThreadId] = struct {
					ts    int64
					total int64
				}{t.Ts, int64(t.ThreadTokenUsage.TotalTokens)}
			}
		}
		if t.TurnTokenUsage != nil && t.Ts > lastTurnTs {
			lastTurnTs = t.Ts
			lastTurn = int64(t.TurnTokenUsage.TotalTokens)
		}
	}
	for _, v := range threads {
		if v.ts > latestThreadTs {
			latestThreadTs = v.ts
			currentTask = v.total
		}
	}

	// 缓存命中率：今日 cached / input
	var cacheHitRate float64
	if todayInput > 0 {
		cacheHitRate = float64(todayCached) / float64(todayInput) * 100
	}

	// 状态：最近 2 分钟有事件 → 执行中
	status := "空闲"
	if lastEventTs > 0 && now.UnixMilli()-lastEventTs < 120*1000 {
		status = "执行中"
	}

	// A failed request must not pin an old online value above newer local events.
	quota := p.latestQuota
	quotaSource := "local"
	if p.liveQuota != nil && (quota == nil || p.liveQuota.Ts >= quota.Ts) {
		quota = p.liveQuota
		quotaSource = "online"
	}
	if quota == nil {
		quotaSource = "none"
	}
	nowSec := now.Unix()

	src := "fresh"
	var snapTs int64
	if quota != nil {
		snapTs = quota.Ts
		if now.UnixMilli()-snapTs > 30*60*1000 {
			src = "stale"
		}
		for _, w := range []*Window{quota.Primary, quota.Secondary} {
			if w != nil && w.ResetsAt != nil && *w.ResetsAt <= nowSec {
				src = "stale"
			}
		}
	} else {
		src = "stale"
	}

	snap := &Snapshot{
		ID:            p.ID(),
		Name:          p.Name(),
		ShortName:     p.ShortName(),
		Color:         p.Color(),
		Emoji:         p.Emoji(),
		Source:        src,
		QuotaSource:   quotaSource,
		Ts:            snapTs,
		LastRequestAt: lastEventTs,
		Status:        status,
	}
	if quota != nil {
		if quota.Primary != nil {
			snap.Primary = windowToQuotaWindow(quota.Primary, nowSec)
		}
		if quota.Secondary != nil {
			snap.Secondary = windowToQuotaWindow(quota.Secondary, nowSec)
		}
		if c := quota.Credits; c != nil {
			snap.Credits = &CreditsInfo{
				HasCredits: c.HasCredits,
				Unlimited:  c.Unlimited,
				Balance:    c.Balance,
			}
		}
	}
	snap.Token = TokenStats{
		Today:        today,
		Near7d:       near7d,
		Total:        total,
		CurrentTask:  currentTask,
		LastTurn:     lastTurn,
		CacheHitRate: cacheHitRate,
	}
	if quota == nil {
		snap.Error = "no data yet"
	}
	if p.refreshError != "" {
		snap.Error = p.refreshError
		if quotaSource != "local" {
			snap.Source = "stale"
		}
	}
	return snap
}

func (p *CodexProvider) Poll() *Snapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.buildSnapshot()
}

func (p *CodexProvider) Stop() error {
	p.stopOnce.Do(func() {
		close(p.stop)
		if p.pollTimer != nil {
			p.pollTimer.Stop()
		}
		if p.watcher != nil {
			_ = p.watcher.Close()
		}
	})
	p.workers.Wait()
	return nil
}

func windowToQuotaWindow(w *Window, _ int64) *QuotaWindow {
	if w == nil {
		return nil
	}
	qw := &QuotaWindow{
		UsedPct:       w.UsedPercent,
		RemainingPct:  100 - w.UsedPercent,
		WindowMinutes: w.WindowMinutes,
	}
	if w.ResetsAt != nil {
		sec := *w.ResetsAt
		qw.ResetsAt = &sec
		// Keep the last observed value until the server confirms a reset.
	}
	return qw
}

func refreshErrorMessage(err error) string {
	var networkError net.Error
	if errors.As(err, &networkError) && networkError.Timeout() {
		return "在线刷新超时，请检查网络和系统代理；当前显示最近记录"
	}
	// Error strings must not include credentials or raw response bodies.
	return "在线刷新失败：" + err.Error() + "；当前显示最近记录"
}

// 注册表
type ProviderRegistry struct {
	providers []UsageProvider
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{}
}

func (r *ProviderRegistry) Register(p UsageProvider) {
	r.providers = append(r.providers, p)
}

func (r *ProviderRegistry) List() []UsageProvider {
	return r.providers
}

func (r *ProviderRegistry) StartAll(onUpdate func(*Snapshot)) {
	for _, p := range r.providers {
		_ = p.Start(onUpdate)
	}
}

func (r *ProviderRegistry) StopAll() {
	for _, p := range r.providers {
		_ = p.Stop()
	}
}
