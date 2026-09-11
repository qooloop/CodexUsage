// window.go - 单 Wails 窗口双模式控制器（需求文档 §10/§12/§31/§34）
// Wails v2 不支持多窗口，用同一个窗口的两种形态：
//   hover 模式：344x288 逻辑像素，置顶、不抢焦点、鼠标离开自动隐藏（250ms+ 延迟）
//   dashboard 模式：780x520 逻辑像素，普通置前面板
// 尺寸/定位在物理像素层完成（SetWindowPos），逻辑尺寸 = 物理 / DPI scale

package desktop

import (
	"sync"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	hoverCardW = 344
	hoverCardH = 288
	dashW      = 780
	dashH      = 520
)

var (
	winMu               sync.Mutex
	wailsHwndCached     uintptr
	desiredRoute        string
	pendingPresentation *windowPresentation
	currentMode         string // "" | "hover" | "dashboard"
	hoverWatchStop      = make(chan struct{})
)

func ensureWailsHwnd() uintptr {
	winMu.Lock()
	h := wailsHwndCached
	winMu.Unlock()
	if h != 0 {
		return h
	}
	h = findWindowByTitle(windowTitle)
	if h != 0 {
		winMu.Lock()
		wailsHwndCached = h
		winMu.Unlock()
	}
	return h
}

func hideMainWindow() {
	winMu.Lock()
	currentMode = ""
	desiredRoute = "hidden"
	pendingPresentation = nil
	winMu.Unlock()
	stopHoverWatch()
	if h := ensureWailsHwnd(); h != 0 {
		procShowWindow.Call(h, uintptr(SW_HIDE))
	}
	if appCtx != nil {
		wailsruntime.EventsEmit(appCtx, "ui:mode", "hidden")
	}
}

// showHoverCard 在托盘/光标附近弹出悬浮卡片（不抢焦点，§12/§31/§32）
func showHoverCard(cursor pointW) {
	if appCtx == nil || isHoverUp() || isDashboardUp() || trayMenuOpen.Load() {
		return
	}
	scale := getMonitorDpiScale(cursor)
	work := getMonitorWorkArea(cursor)
	if work.right <= work.left {
		return
	}

	phyW := int(float64(hoverCardW) * scale)
	phyH := int(float64(hoverCardH) * scale)

	anchor, ok := getTrayRect()
	if !ok {
		anchor = rectW{cursor.x - 12, cursor.y - 12, cursor.x + 12, cursor.y + 12}
	}
	bounds := popupBounds(work, anchor, phyW, phyH, int(8*scale))
	x, y := int(bounds.left), int(bounds.top)

	winMu.Lock()
	currentMode = "hover"
	winMu.Unlock()

	wailsruntime.WindowSetMinSize(appCtx, hoverCardW, hoverCardH)
	wailsruntime.WindowSetMaxSize(appCtx, hoverCardW, hoverCardH)
	prepareNativeWindow(true)
	queuePresentation("hover", x, y, phyW, phyH, true)

	// 启动 hover 离开监视（§12 隐藏延迟）
	startHoverWatch(anchor)
}

// showDashboard 打开详细面板（点击托盘 / 菜单）
func showDashboard() {
	showDashboardAt("overview")
}

func showDashboardAt(route string) {
	if appCtx == nil {
		return
	}
	stopHoverWatch()

	// 以主光标所在显示器的工作区定位（跟随托盘所在屏，§35）
	var pt pointW
	getCursorPos(&pt)
	scale := getMonitorDpiScale(pt)
	work := getMonitorWorkArea(pt)
	if work.right <= work.left {
		// 兜底：主屏
		work = getMonitorWorkArea(pointW{x: 1, y: 1})
	}
	if work.right <= work.left || work.bottom <= work.top {
		return
	}
	phyW := int(float64(dashW) * scale)
	phyH := int(float64(dashH) * scale)
	if phyW > int(work.right-work.left) {
		phyW = int(work.right - work.left)
	}
	if phyH > int(work.bottom-work.top) {
		phyH = int(work.bottom - work.top)
	}
	x := int(work.left) + (int(work.right-work.left)-phyW)/2
	y := int(work.top) + (int(work.bottom-work.top)-phyH)/2

	winMu.Lock()
	currentMode = "dashboard"
	winMu.Unlock()

	wailsruntime.WindowSetMaxSize(appCtx, 0, 0)
	wailsruntime.WindowSetMinSize(appCtx, 720, 480)
	prepareNativeWindow(false)
	queuePresentation(route, x, y, phyW, phyH, false)
}

// showDashboardFromTray 由 systray/hover 事件调（带屏幕定位）
func showPopup() {
	showDashboard()
}

// emitUIMode 通知前端切换视图
func emitUIMode(route string) {
	if appCtx != nil {
		wailsruntime.EventsEmit(appCtx, "ui:mode", route)
	}
}

// ---- hover 离开监视（§12：离开托盘+卡片 250ms 后隐藏） ----

func startHoverWatch(anchor rectW) {
	stopHoverWatch()
	winMu.Lock()
	currentMode = "hover"
	winMu.Unlock()

	stop := make(chan struct{})
	winMu.Lock()
	hoverWatchStop = stop
	winMu.Unlock()

	go func() {
		lastInside := time.Now()
		for {
			select {
			case <-stop:
				return
			case <-time.After(50 * time.Millisecond):
			}
			winMu.Lock()
			mode := currentMode
			winMu.Unlock()
			if mode != "hover" {
				return
			}
			hwnd := ensureWailsHwnd()
			if hwnd == 0 {
				return
			}
			var pt pointW
			if !getCursorPos(&pt) {
				continue
			}
			insideWin := false
			if r, ok := getWindowRectPhy(hwnd); ok {
				insideWin = pt.x >= r.left && pt.x < r.right && pt.y >= r.top && pt.y < r.bottom
			}
			// 托盘附近 48 物理像素内也算 inside（鼠标在托盘上停留）
			nearTray := pt.x >= anchor.left && pt.x < anchor.right && pt.y >= anchor.top && pt.y < anchor.bottom
			if insideWin || nearTray {
				lastInside = time.Now()
				continue
			}
			if time.Since(lastInside) >= 250*time.Millisecond {
				hideMainWindow()
				return
			}
		}
	}()
}

func stopHoverWatch() {
	winMu.Lock()
	close(hoverWatchStop)
	hoverWatchStop = make(chan struct{})
	winMu.Unlock()
}

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// isDashboardUp 详细面板当前是否打开
func isDashboardUp() bool {
	winMu.Lock()
	defer winMu.Unlock()
	return currentMode == "dashboard"
}

// isHoverUp 悬浮卡片当前是否打开
func isHoverUp() bool {
	winMu.Lock()
	defer winMu.Unlock()
	return currentMode == "hover"
}

// Pure physical-coordinate placement, shared with regression tests.
func popupBounds(work, anchor rectW, width, height, gap int) rectW {
	width = min(width, int(work.right-work.left))
	height = min(height, int(work.bottom-work.top))
	x := int(anchor.left+anchor.right)/2 - width/2
	y := int(anchor.top) - height - gap
	if anchor.bottom <= work.top {
		y = int(anchor.bottom) + gap
	} else if anchor.right <= work.left {
		x = int(anchor.right) + gap
		y = int(anchor.top)
	} else if anchor.left >= work.right {
		x = int(anchor.left) - width - gap
		y = int(anchor.top)
	} else if y < int(work.top) {
		y = int(anchor.bottom) + gap
	}
	x = max(int(work.left), min(x, int(work.right)-width))
	y = max(int(work.top), min(y, int(work.bottom)-height))
	return rectW{int32(x), int32(y), int32(x + width), int32(y + height)}
}

// Only show after Vue has mounted the requested route, avoiding a dashboard flash in the small card.
type windowPresentation struct {
	x, y, w, h int
	hover      bool
}

func queuePresentation(route string, x, y, w, h int, hover bool) {
	winMu.Lock()
	desiredRoute = route
	pendingPresentation = &windowPresentation{x, y, w, h, hover}
	winMu.Unlock()
	emitUIMode(route)
}
func (a *App) GetUIMode() string {
	winMu.Lock()
	defer winMu.Unlock()
	if desiredRoute == "" {
		return "hidden"
	}
	return desiredRoute
}
func (a *App) ConfirmUIMode(route string) {
	winMu.Lock()
	defer winMu.Unlock()
	if route != desiredRoute || pendingPresentation == nil {
		return
	}
	p := pendingPresentation
	pendingPresentation = nil
	// HWND was resolved while preparing the mode; do not reacquire winMu here.
	hwnd := wailsHwndCached
	if hwnd == 0 {
		return
	}
	z := uintptr(HWND_NOTOPMOST)
	if p.hover {
		z = HWND_TOPMOST
	}
	procSetWindowPos.Call(hwnd, z, uintptr(p.x), uintptr(p.y), uintptr(p.w), uintptr(p.h), SWP_NOACTIVATE|SWP_SHOWWINDOW)
	if !p.hover {
		procSetForegroundWindow.Call(hwnd)
	}
}
