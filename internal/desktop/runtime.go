// main.go - Wails + 原生 Win32 托盘 集成入口（需求文档 §55/§56/§68）
// 1. 自研 Shell_NotifyIcon 托盘：动态数字图标 + hover 悬浮卡 + 点击面板 + 右键菜单
// 2. Wails 单窗口双形态：hover 卡片 / 详细面板（Vue 3）
// 3. 数据更新 → 托盘图标 + 推送 WebView
//

package desktop

import (
	"codex-monitor/internal/usage"
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	quitting    atomic.Bool
	appCtx      context.Context
	appInstance *App
	providerReg *usage.ProviderRegistry
	snapshotsMu sync.RWMutex
	snapshots   = make(map[string]*usage.Snapshot)
)

func Run(assets fs.FS) {
	// §55 启动流程：单实例 → 配置 → provider → 托盘 → Wails → 首刷 → 调度
	if !acquireSingleInstance() {
		return
	}
	cfg := loadConfig()
	applyAutostart(cfg.StartWithWindows)

	appInstance = NewApp()
	providerReg = usage.NewProviderRegistry()
	providerReg.Register(usage.NewCodexProvider())

	// 启动原生托盘（内部 goroutine 消息泵）
	startTray()

	err := wails.Run(&options.App{
		Title:            windowTitle,
		Width:            dashW,
		Height:           dashH,
		MinWidth:         hoverCardW,
		MinHeight:        hoverCardH,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 14, G: 26, B: 48, A: 255},
		OnStartup: func(ctx context.Context) {
			appCtx = ctx
			appInstance.startup(ctx)
			providerReg.StartAll(onSnapshotUpdate)
			startScheduler()
			// 首刷（§28：启动立即刷新）
			go func() {
				time.Sleep(300 * time.Millisecond)
				forceRefresh()
				if !cfg.HideOnStartup {
					showDashboard()
				}
			}()
		},
		OnBeforeClose: func(ctx context.Context) bool {
			if quitting.Load() {
				return false
			}
			// 点 × 隐藏到托盘（§13 生命周期：常驻）
			hideMainWindow()
			return true
		},
		Bind: []interface{}{appInstance},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			DisableWindowIcon:    false,
		},
		Frameless:   true,
		StartHidden: true,
	})

	if err != nil {
		log.Fatal("wails.Run: ", err)
	}

	stopScheduler()
	stopHoverWatch()
	providerReg.StopAll()
	removeTrayIcon()
	trayMu.Lock()
	if curHIcon != 0 {
		procDestroyIcon.Call(curHIcon)
		curHIcon = 0
	}
	trayMu.Unlock()
}

// ----- 数据流（§56） -----

func onSnapshotUpdate(snap *usage.Snapshot) {
	if snap == nil {
		return
	}
	snapshotsMu.Lock()
	prev := snapshots[snap.ID]
	snapshots[snap.ID] = snap
	snapshotsMu.Unlock()

	updateTrayIcon(snap)
	checkThresholds(snap)

	if appCtx != nil {
		wailsruntime.EventsEmit(appCtx, "snapshots:update", getAllSnapshots())
	}
	_ = prev
}

func getAllSnapshots() []*usage.Snapshot {
	snapshotsMu.RLock()
	defer snapshotsMu.RUnlock()
	out := make([]*usage.Snapshot, 0, len(providerReg.List()))
	for _, p := range providerReg.List() {
		if s, ok := snapshots[p.ID()]; ok && s != nil {
			out = append(out, s)
		}
	}
	return out
}

// ----- 通用动作 -----

func openCodexHome() {
	home := getCodexHome()
	// 用资源管理器打开（比 BrowserOpenURL 更自然）
	c := exec.Command("explorer.exe", home)
	_ = c.Start()
}

func quitApp() {
	if !quitting.CompareAndSwap(false, true) {
		return
	}
	if appCtx != nil {
		wailsruntime.Quit(appCtx)
	}
}

func getCodexHome() string {
	home := os.Getenv("CODEX_HOME")
	if home == "" {
		if h, err := os.UserHomeDir(); err == nil {
			home = filepath.Join(h, ".codex")
		}
	}
	return home
}

// formatReset unix 秒 → "2h13m"/"13m"/"已重置"
func formatReset(epochSec *int64) string {
	if epochSec == nil || *epochSec == 0 {
		return "—"
	}
	d := time.Unix(*epochSec, 0)
	ms := time.Until(d)
	if ms <= 0 {
		return "已重置"
	}
	m := int(ms.Minutes())
	if m < 1 {
		return "<1m"
	}
	if m < 60 {
		return fmt.Sprintf("%dm", m)
	}
	if m < 1440 {
		return fmt.Sprintf("%dh%dm", m/60, m%60)
	}
	return fmt.Sprintf("%dd%dh", m/1440, (m%1440)/60)
}
