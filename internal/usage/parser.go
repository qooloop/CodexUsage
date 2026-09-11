// Session JSONL parser.
// 双格式兼容：event_msg/token_count（带 rate_limits）+ token_usage_record（不带）。
// 关键：primary/secondary 按 window_minutes 判定（300=5h, 10080=7d）。

package usage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	Window5H = 300
	Window7D = 10080
)

type Credits struct {
	HasCredits bool   `json:"has_credits"`
	Unlimited  bool   `json:"unlimited"`
	Balance    string `json:"balance"`
}

type Window struct {
	UsedPercent   float64 `json:"used_percent"`
	WindowMinutes int     `json:"window_minutes"`
	ResetsAt      *int64  `json:"resets_at"`
}

type TokenUsage struct {
	InputTokens           int `json:"input_tokens"`
	CachedInputTokens     int `json:"cached_input_tokens"`
	CacheWriteInputTokens int `json:"cache_write_input_tokens"`
	OutputTokens          int `json:"output_tokens"`
	ReasoningOutputTokens int `json:"reasoning_output_tokens"`
	TotalTokens           int `json:"total_tokens"`
}

type Event struct {
	Kind string // "quota" | "token"

	// quota 字段
	Ts          int64
	LimitId     *string
	PlanType    *string
	ReachedType *string
	Credits     *Credits
	Primary     *Window
	Secondary   *Window

	// token 字段
	SessionId        *string
	ThreadId         *string
	TurnId           *string
	ResponseId       *string
	Usage            *TokenUsage
	TurnTokenUsage   *TokenUsage
	ThreadTokenUsage *TokenUsage

	// 格式 A（event_msg/token_count）的 info.total_token_usage
	// —— 会话内累计快照（不是增量），是老会话"终身累计"的唯一来源
	SessionTotal *TokenUsage
}

// ---------- 解析 ----------

func parseTimestamp(raw map[string]interface{}) int64 {
	if v, ok := raw["timestamp"].(string); ok {
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
			if t, err := time.Parse(layout, v); err == nil {
				return t.UnixMilli()
			}
		}
	}
	return time.Now().UnixMilli()
}

func parseWindow(v interface{}) *Window {
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	w := &Window{}
	if x, ok := m["used_percent"].(float64); ok {
		w.UsedPercent = x
	}
	if x, ok := m["window_minutes"].(float64); ok {
		w.WindowMinutes = int(x)
	}
	if x, ok := m["resets_at"].(float64); ok {
		r := int64(x)
		w.ResetsAt = &r
	}
	return w
}

func parseCredits(v interface{}) *Credits {
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	c := &Credits{}
	if x, ok := m["has_credits"].(bool); ok {
		c.HasCredits = x
	}
	if x, ok := m["unlimited"].(bool); ok {
		c.Unlimited = x
	}
	if x, ok := m["balance"].(string); ok {
		c.Balance = x
	}
	return c
}

func parseTokenUsage(v interface{}) *TokenUsage {
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	u := &TokenUsage{}
	if x, ok := m["input_tokens"].(float64); ok {
		u.InputTokens = int(x)
	}
	if x, ok := m["cached_input_tokens"].(float64); ok {
		u.CachedInputTokens = int(x)
	}
	if x, ok := m["cache_write_input_tokens"].(float64); ok {
		u.CacheWriteInputTokens = int(x)
	}
	if x, ok := m["output_tokens"].(float64); ok {
		u.OutputTokens = int(x)
	}
	if x, ok := m["reasoning_output_tokens"].(float64); ok {
		u.ReasoningOutputTokens = int(x)
	}
	if x, ok := m["total_tokens"].(float64); ok {
		u.TotalTokens = int(x)
	}
	return u
}

func pickWindow(primary, secondary *Window, wantMin int) *Window {
	for _, w := range []*Window{primary, secondary} {
		if w != nil && w.WindowMinutes == wantMin {
			return w
		}
	}
	return nil
}

func ParseLine(line string) *Event {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return nil
	}

	eventType, _ := raw["type"].(string)
	payload, _ := raw["payload"].(map[string]interface{})
	if payload == nil {
		return nil
	}

	switch eventType {
	case "event_msg":
		if pmType, _ := payload["type"].(string); pmType == "token_count" {
			return parseTokenCount(raw, payload)
		}
	case "token_usage_record":
		return parseTokenUsageRecord(raw, payload)
	}
	return nil
}

func parseTokenCount(raw, payload map[string]interface{}) *Event {
	rl, _ := payload["rate_limits"].(map[string]interface{})
	primary, secondary := (*Window)(nil), (*Window)(nil)
	if rl != nil {
		primary = parseWindow(rl["primary"])
		secondary = parseWindow(rl["secondary"])
	}

	// 格式 A 的会话累计 Token（info.total_token_usage，累计值非增量）
	var sessionTotal *TokenUsage
	if info, ok := payload["info"].(map[string]interface{}); ok {
		if ttu, ok := info["total_token_usage"].(map[string]interface{}); ok {
			u := parseTokenUsage(ttu)
			if u != nil && u.TotalTokens > 0 {
				sessionTotal = u
			}
		}
	}

	if primary == nil && secondary == nil && sessionTotal == nil {
		return nil
	}

	ev := &Event{Kind: "quota", SessionTotal: sessionTotal}
	ev.Primary = pickWindow(primary, secondary, Window5H)
	ev.Secondary = pickWindow(primary, secondary, Window7D)
	ev.Ts = parseTimestamp(raw)

	if v, ok := rl["limit_id"].(string); ok {
		s := v
		ev.LimitId = &s
	}
	if v, ok := rl["plan_type"].(string); ok {
		s := v
		ev.PlanType = &s
	}
	if v, ok := rl["rate_limit_reached_type"].(string); ok {
		s := v
		ev.ReachedType = &s
	}
	ev.Credits = parseCredits(rl["credits"])
	return ev
}

func parseTokenUsageRecord(raw, payload map[string]interface{}) *Event {
	usage := parseTokenUsage(payload["usage"])
	if usage == nil {
		return nil
	}
	ev := &Event{Kind: "token"}
	ev.Usage = usage
	if v, ok := payload["turn_token_usage"]; ok {
		ev.TurnTokenUsage = parseTokenUsage(v)
	}
	if v, ok := payload["thread_token_usage"]; ok {
		ev.ThreadTokenUsage = parseTokenUsage(v)
	}
	ev.Ts = parseTimestamp(raw)

	if v, ok := payload["session_id"].(string); ok {
		s := v
		ev.SessionId = &s
	}
	if v, ok := payload["thread_id"].(string); ok {
		s := v
		ev.ThreadId = &s
	}
	if v, ok := payload["turn_id"].(string); ok {
		s := v
		ev.TurnId = &s
	}
	if v, ok := payload["response_id"].(string); ok {
		s := v
		ev.ResponseId = &s
	}
	return ev
}

// ---------- 文件读取 ----------

// 增量读一个 JSONL 文件，只解析 [startOffset, end] 的内容。
// 返回新事件数组和新的 offset。损坏行跳过不抛。
func ReadNewEvents(filePath string, startOffset int64) (events []*Event, newOffset int64) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, startOffset
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, startOffset
	}
	if startOffset < 0 || stat.Size() < startOffset {
		startOffset = 0
	}
	if stat.Size() == startOffset {
		return nil, startOffset
	}

	size := stat.Size() - startOffset
	buf := make([]byte, size)
	if _, err := f.ReadAt(buf, startOffset); err != nil && err != io.EOF {
		return nil, startOffset
	}
	newOffset = stat.Size()

	text := string(buf)
	lines := strings.Split(text, "\n")
	// 最后一段可能是半行（文件正在被写入），回退 offset
	if len(lines) > 0 {
		tail := lines[len(lines)-1]
		if tail != "" {
			newOffset -= int64(len(tail))
			lines = lines[:len(lines)-1]
		}
	}

	for _, line := range lines {
		if e := ParseLine(line); e != nil {
			events = append(events, e)
		}
	}
	return events, newOffset
}

// 全量扫描目录下所有 rollout-*.jsonl
func ScanAll(rootDir string) (latestQuota *Event, tokenEvents []*Event, filesScanned int) {
	files := walkAll(rootDir)
	for _, f := range files {
		fp, err := os.Open(f)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(fp)
		scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024) // 大行支持到 16MB
		for scanner.Scan() {
			if e := ParseLine(scanner.Text()); e != nil {
				switch e.Kind {
				case "quota":
					if latestQuota == nil || e.Ts > latestQuota.Ts {
						latestQuota = e
					}
				case "token":
					tokenEvents = append(tokenEvents, e)
				}
			}
		}
		fp.Close()
		filesScanned++
	}
	return
}

func walkAll(root string) []string {
	var out []string
	if root == "" {
		return out
	}
	// 递归遍历：sessions 是 年/月/日 嵌套，archived_sessions 是扁平目录，两种都要兼容
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if strings.HasPrefix(name, "rollout-") && strings.HasSuffix(name, ".jsonl") {
			out = append(out, path)
		}
		return nil
	})
	// 按 mtime 倒序（最新优先）
	sort.Slice(out, func(i, j int) bool {
		fi, _ := os.Stat(out[i])
		fj, _ := os.Stat(out[j])
		if fi == nil || fj == nil {
			return false
		}
		return fi.ModTime().After(fj.ModTime())
	})
	return out
}

// 格式化时间戳为可读字符串
func FormatTs(ts int64) string {
	if ts == 0 {
		return "—"
	}
	return time.UnixMilli(ts).Format("2006-01-02 15:04:05")
}

// 毫秒时间戳转 Unix 秒（用于 resets_at 显示）
func TsToUnixSec(ts int64) int64 { return ts / 1000 }

// 辅助：从 int64 指针解引用
func DerefInt64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func DerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// 静默忽略 strconv 错误
func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

var _ = fmt.Sprintf // 防止 fmt 未使用警告
