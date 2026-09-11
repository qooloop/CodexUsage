// app.go - Wails 绑定：暴露给前端调用的方法
// 前端通过 window.go.desktop.App.GetSnapshots() 等调用

package desktop

import (
	"codex-monitor/internal/config"
	"codex-monitor/internal/usage"
	"context"
	"path/filepath"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetSnapshots 暴露给前端：返回所有 provider 的最新快照
func (a *App) GetSnapshots() []*usage.Snapshot {
	return getAllSnapshots()
}

// Refresh 暴露给前端：强制刷新所有 provider
func (a *App) Refresh() []*usage.Snapshot {
	forceRefresh()
	return getAllSnapshots()
}

// OpenCodexHome 暴露给前端：用资源管理器打开数据目录
func (a *App) OpenCodexHome() string {
	home := getCodexHome()
	abs, _ := filepath.Abs(home)
	openCodexHome()
	return abs
}

// HidePopup 暴露给前端：最小化/关闭 → 隐藏到托盘
func (a *App) HidePopup() {
	hideMainWindow()
}

// Quit 暴露给前端：退出应用
func (a *App) Quit() {
	quitApp()
}

// ToggleMaximize 前端窗口最大化切换
func (a *App) ToggleMaximize() {
	if a.ctx != nil {
		wailsruntime.WindowToggleMaximise(a.ctx)
	}
}

// GetConfig 读取配置
func (a *App) GetConfig() config.Config {
	return getConfig()
}

// SaveConfig 保存配置（自动应用开机启动/刷新间隔），返回落盘后的配置
func (a *App) SaveConfig(c config.Config) config.Config {
	if err := saveConfig(c); err != nil {
		return getConfig()
	}
	restartScheduler()
	refreshTrayFromSnapshot()
	return getConfig()
}

// OpenDashboard 从前端请求打开面板（hover 卡片里的跳转按钮）
func (a *App) OpenDashboard(route string) {
	showDashboardAt(route)
}

// OpenOfficialPanel opens the subscription usage page, rather than a local folder.
func (a *App) OpenOfficialPanel() {
	if a.ctx != nil {
		wailsruntime.BrowserOpenURL(a.ctx, "https://chatgpt.com/codex/settings/usage")
	}
}
