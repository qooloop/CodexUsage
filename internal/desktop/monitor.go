// monitor.go - 刷新调度器（需求文档 §26/§27/§28/§43）
// 默认 60s 兜底刷新（数据主体由 provider 的 fsnotify + 3s 轮询驱动）
// 连续失败时退避：60s → 120s → 300s 封顶

package desktop

import (
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	refreshMu      sync.Mutex
	refreshDone    chan struct{}
	schedulerStop  = make(chan struct{})
	schedulerMu    sync.Mutex
	failureBackoff atomic.Int32
	refreshRunning atomic.Bool // singleflight（§29）
)

// startScheduler 启动周期刷新
func startScheduler() {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()
	startSchedulerLocked()
}
func startSchedulerLocked() {
	stop := schedulerStop
	interval := time.Duration(getConfig().RefreshIntervalSec) * time.Second
	if interval < 5*time.Second {
		interval = 60 * time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				forceRefresh()
				// 连续失败退避（§43）
				fails := failureBackoff.Load()
				if fails > 0 {
					backoff := time.Duration(1<<min(fails, 6)) * interval
					if backoff > 5*time.Minute {
						backoff = 5 * time.Minute
					}
					ticker.Reset(backoff)
				} else {
					ticker.Reset(interval)
				}
			}
		}
	}()
}

// restartScheduler 配置变更后按新间隔重启
func restartScheduler() {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()
	close(schedulerStop)
	schedulerStop = make(chan struct{})
	failureBackoff.Store(0)
	startSchedulerLocked()
}
func stopScheduler() {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()
	close(schedulerStop)
}

// forceRefresh 立即刷新全部 provider（防重入，§29）
func forceRefresh() {
	if providerReg == nil {
		return
	}
	refreshMu.Lock()
	if refreshDone != nil {
		done := refreshDone
		refreshMu.Unlock()
		<-done
		return
	}
	refreshDone = make(chan struct{})
	refreshRunning.Store(true)
	refreshMu.Unlock()
	if appCtx != nil {
		wailsruntime.EventsEmit(appCtx, "refresh:state", true)
	}
	defer func() {
		if appCtx != nil {
			wailsruntime.EventsEmit(appCtx, "refresh:state", false)
		}
		refreshMu.Lock()
		refreshRunning.Store(false)
		close(refreshDone)
		refreshDone = nil
		refreshMu.Unlock()
	}()
	for _, p := range providerReg.List() {
		snap := p.Refresh()
		if snap != nil && snap.Error != "" {
			f := failureBackoff.Add(1)
			if f > 6 {
				failureBackoff.Store(6)
			}
		} else {
			failureBackoff.Store(0)
		}
		onSnapshotUpdate(snap)
	}
}
