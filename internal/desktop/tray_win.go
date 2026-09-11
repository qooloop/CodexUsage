// tray_win.go - 原生 Win32 系统托盘（需求文档 §7/§10/§12/§13/§48/§68）
// 不依赖 getlantern/systray：自己维护 Shell_NotifyIcon + 隐藏消息窗口
// - NOTIFYICON_VERSION_4：WM_MOUSEMOVE→Hover、WM_LBUTTONUP→点击、WM_CONTEXTMENU→右键菜单
// - TaskbarCreated 消息 → Explorer 重启后自动恢复托盘（§48）
// - wndProc 只投递事件到 channel，业务在独立 goroutine 处理，避免卡消息泵

package desktop

import (
	"codex-monitor/internal/usage"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

const windowTitle = "Codex 用量监控"

var (
	modUser32   = windows.NewLazySystemDLL("user32.dll")
	modShell32  = windows.NewLazySystemDLL("shell32.dll")
	modKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procGetModuleHandleW = modKernel32.NewProc("GetModuleHandleW")

	procRegisterClassExW         = modUser32.NewProc("RegisterClassExW")
	procCreateWindowExW          = modUser32.NewProc("CreateWindowExW")
	procDefWindowProcW           = modUser32.NewProc("DefWindowProcW")
	procDispatchMessageW         = modUser32.NewProc("DispatchMessageW")
	procGetMessageW              = modUser32.NewProc("GetMessageW")
	procPostQuitMessage          = modUser32.NewProc("PostQuitMessage")
	procFindWindowW              = modUser32.NewProc("FindWindowW")
	procShellNotifyIconW         = modShell32.NewProc("Shell_NotifyIconW")
	procRegisterWindowMessageW   = modUser32.NewProc("RegisterWindowMessageW")
	procPostMessageW             = modUser32.NewProc("PostMessageW")
	procShellNotifyIconGetRect   = modShell32.NewProc("Shell_NotifyIconGetRect")
	procGetCursorPos             = modUser32.NewProc("GetCursorPos")
	procLoadCursorW              = modUser32.NewProc("LoadCursorW")
	procGetSystemMetrics         = modUser32.NewProc("GetSystemMetrics")
	procCreatePopupMenu          = modUser32.NewProc("CreatePopupMenu")
	procAppendMenuW              = modUser32.NewProc("AppendMenuW")
	procTrackPopupMenuEx         = modUser32.NewProc("TrackPopupMenuEx")
	procDestroyMenu              = modUser32.NewProc("DestroyMenu")
	procSetForegroundWindow      = modUser32.NewProc("SetForegroundWindow")
	procShowWindow               = modUser32.NewProc("ShowWindow")
	procSetWindowPos             = modUser32.NewProc("SetWindowPos")
	procGetWindowRect            = modUser32.NewProc("GetWindowRect")
	procIsWindowVisible          = modUser32.NewProc("IsWindowVisible")
	procGetDpiForWindow          = modUser32.NewProc("GetDpiForWindow")
	procMonitorFromPoint         = modUser32.NewProc("MonitorFromPoint")
	procGetMonitorInfoW          = modUser32.NewProc("GetMonitorInfoW")
	procCreateIconFromResourceEx = modUser32.NewProc("CreateIconFromResourceEx")
	procDestroyIcon              = modUser32.NewProc("DestroyIcon")
)

// ---- Win32 常量 ----

const (
	NIM_ADD        = 0x00000000
	NIM_MODIFY     = 0x00000001
	NIM_DELETE     = 0x00000002
	NIM_SETVERSION = 0x00000004

	NOTIFYICON_VERSION_4 = 4

	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004
	NIF_SHOWTIP = 0x00000080

	WM_MOUSEMOVE   = 0x0200
	WM_LBUTTONUP   = 0x0202
	WM_RBUTTONUP   = 0x0205
	WM_CONTEXTMENU = 0x007B
	WM_DESTROY     = 0x0002
	WM_QUIT        = 0x0012

	WM_APP             = 0x8000
	appWM_TRAYCALLBACK = WM_APP + 0x100
	appWM_SHOWMENU     = WM_APP + 0x101
	appWM_OPENPANEL    = WM_APP + 0x102
	NIN_SELECT         = 0x400
	NIN_KEYSELECT      = 0x401
	NIN_POPUPOPEN      = 0x406

	MF_STRING    = 0x00000000
	MF_SEPARATOR = 0x00000800
	MF_CHECKED   = 0x00000008
	MF_GRAYED    = 0x00000001

	TPM_LEFTALIGN   = 0x0000
	TPM_RIGHTBUTTON = 0x0002
	TPM_RETURNCMD   = 0x0100
	TPM_NONOTIFY    = 0x0080

	SW_HIDE           = 0
	SW_SHOWNOACTIVATE = 4
	SW_SHOW           = 5
	SW_RESTORE        = 9

	HWND_TOPMOST   = ^uintptr(0) // -1
	HWND_NOTOPMOST = ^uintptr(1) // -2

	SWP_NOSIZE     = 0x0001
	SWP_NOMOVE     = 0x0002
	SWP_NOACTIVATE = 0x0010
	SWP_SHOWWINDOW = 0x0040

	MONITOR_DEFAULTTONEAREST = 2

	SM_CXSMICON = 49

	CS_HREDRAW = 0x0002
	CS_VREDRAW = 0x0001

	IDC_ARROW = 32512

	IMAGE_ICON      = 1
	LR_DEFAULTCOLOR = 0
)

// 菜单命令 ID
const (
	menuIDHeader    = 1
	menuIDHeader5h  = 2
	menuIDHeader7d  = 3
	menuIDRefresh   = 10
	menuIDPanel     = 11
	menuIDAutostart = 12
	menuIDSettings  = 13
	menuIDQuit      = 14
)

// ---- Win32 结构 ----

type pointW struct{ x, y int32 }

type rectW struct{ left, top, right, bottom int32 }

type wndClassExW struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  uintptr
	lpszClassName uintptr
	hIconSm       uintptr
}

type notifyIconDataW struct {
	cbSize           uint32
	_                uint32
	hWnd             uintptr
	uID              uint32
	uFlags           uint32
	uCallbackMessage uint32
	_                uint32
	hIcon            uintptr
	szTip            [128]uint16
	dwState          uint32
	dwStateMask      uint32
	szInfo           [256]uint16
	uVersion         uint32
	szInfoTitle      [64]uint16
	dwInfoFlags      uint32
	guidItem         [16]byte
	hBalloonIcon     uintptr
}

type monitorInfoW struct {
	cbSize    uint32
	rcMonitor rectW
	rcWork    rectW
	dwFlags   uint32
}

// ---- 托盘状态 ----

var (
	trayMenuOpen  atomic.Bool
	trayVersion4  bool
	trayHwnd      uintptr
	trayIconAdded bool
	taskbarMsg    uint32

	curHIcon   uintptr // 当前托盘 HICON（更新成功后销毁旧的，防 GDI 泄漏 §47）
	curIconKey string

	trayMu     sync.Mutex
	trayLatest *usage.Snapshot // 最近快照（菜单/tooltip 用）

	trayEvents chan trayEvent
)

type trayEvent struct {
	kind string // "hover" | "click" | "menu" | "taskbar"
	x, y int
}

// startTray 启动托盘（独立 OS 线程跑消息泵）
func startTray() {
	trayEvents = make(chan trayEvent, 32)
	go func() {
		runtime.LockOSThread()
		runTrayMessagePump()
	}()
	go trayEventsLoop()
}

func runTrayMessagePump() {
	hInst, _, _ := procGetModuleHandleW.Call(0)

	className, _ := windows.UTF16PtrFromString("CodexMonitorTrayWnd")
	wc := wndClassExW{
		cbSize:        uint32(unsafe.Sizeof(wndClassExW{})),
		style:         CS_HREDRAW | CS_VREDRAW,
		lpfnWndProc:   windows.NewCallback(trayWndProc),
		hInstance:     uintptr(hInst),
		hCursor:       mustLoadCursor(),
		lpszClassName: uintptr(unsafe.Pointer(className)),
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	// 隐藏顶层窗口接收 Explorer 的 TaskbarCreated 广播；message-only 窗口收不到。
	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		0,
		0, 0, 0, 0, 0,
		0, 0, uintptr(hInst), 0,
	)
	trayHwnd = hwnd

	taskbarMsg = uint32(mustRegisterWindowMessage("TaskbarCreated"))

	addTrayIcon()

	// 消息泵：必须 DispatchMessage 才能把消息路由到窗口类注册的 trayWndProc。
	// 之前误用 DefWindowProc（默认处理）直调，托盘 hover/click/menu 事件全部丢失。
	var m [12]uintptr // >= sizeof(MSG)
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m[0])), 0, 0, 0)
		if int32(r) <= 0 {
			return
		}
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m[0])))
	}
}

func trayWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case appWM_TRAYCALLBACK:
		mouseMsg := uint32(lParam & 0xFFFF)
		switch mouseMsg {
		case WM_MOUSEMOVE, NIN_POPUPOPEN:
			postTrayEvent(trayEvent{kind: "hover"})
		case WM_LBUTTONUP:
			if trayVersion4 {
				break
			}
			postTrayEvent(trayEvent{kind: "click"})
		case NIN_SELECT, NIN_KEYSELECT:
			postTrayEvent(trayEvent{kind: "click"})
		case WM_CONTEXTMENU:
			x := int32(int16(wParam & 0xFFFF))
			y := int32(int16((wParam >> 16) & 0xFFFF))
			handleContextMenu(int(x), int(y))
		case WM_RBUTTONUP:
			if !trayVersion4 {
				handleContextMenu(0, 0)
			}
		}
		return 0
	case appWM_SHOWMENU:
		handleContextMenu(0, 0)
		return 0
	case appWM_OPENPANEL:
		postTrayEvent(trayEvent{kind: "open"})
		return 0
	case taskbarMsg:
		// Explorer 重启 → 托盘消失，重新注册（§48）
		postTrayEvent(trayEvent{kind: "taskbar"})
		return 0
	case WM_DESTROY:
		removeTrayIcon()
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func postTrayEvent(ev trayEvent) {
	select {
	case trayEvents <- ev:
	default:
	}
}

// trayEventsLoop 托盘事件处理（独立 goroutine，可安全调 Wails runtime）
func trayEventsLoop() {
	for ev := range trayEvents {
		switch ev.kind {
		case "hover":
			if isDashboardUp() || isHoverUp() || trayMenuOpen.Load() {
				continue // 面板已开，不再弹 hover 卡
			}
			cfg := getConfig()
			if cfg.ShowHoverPopup {
				var pt pointW
				getCursorPos(&pt)
				showHoverCard(pt)
			}
		case "click":
			if isDashboardUp() {
				hideMainWindow()
			} else {
				showDashboard()
			}
		case "menu":
			procPostMessageW.Call(trayHwnd, appWM_SHOWMENU, 0, 0)
		case "open":
			showDashboard()
		case "command":
			handleMenuCommand(uintptr(ev.x))
		case "taskbar":
			addTrayIcon()
			refreshTrayFromSnapshot()
		}
	}
}

// ---- Shell_NotifyIcon ----

func addTrayIcon() {
	if trayHwnd == 0 {
		return
	}
	// 先删再加（Explorer 重启场景下原句柄已失效）
	removeTrayIcon()

	nid := notifyIconDataW{
		cbSize:           uint32(unsafe.Sizeof(notifyIconDataW{})),
		hWnd:             trayHwnd,
		uID:              1,
		uFlags:           NIF_MESSAGE | NIF_ICON | NIF_TIP,
		uCallbackMessage: appWM_TRAYCALLBACK,
	}
	nid.hIcon = mustCreateHIconFromPNG(renderTrayIconPNGCached(TrayIconState{ShortQuotaPercent: -1, LongQuotaPercent: -1}, trayIconSourceSize()))
	copyTip(nid.szTip[:], "Codex 用量监控\n正在加载数据…")

	added, _, _ := procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
	if added == 0 {
		procDestroyIcon.Call(nid.hIcon)
		return
	}
	trayMu.Lock()
	old := curHIcon
	curHIcon = nid.hIcon
	trayMu.Unlock()
	if old != 0 {
		procDestroyIcon.Call(old)
	}

	ver := notifyIconDataW{
		cbSize:   uint32(unsafe.Sizeof(notifyIconDataW{})),
		hWnd:     trayHwnd,
		uID:      1,
		uVersion: NOTIFYICON_VERSION_4,
	}
	versionOK, _, _ := procShellNotifyIconW.Call(NIM_SETVERSION, uintptr(unsafe.Pointer(&ver)))
	trayVersion4 = versionOK != 0

	trayMu.Lock()
	trayIconAdded = true
	trayMu.Unlock()
}

func removeTrayIcon() {
	trayMu.Lock()
	added := trayIconAdded
	trayIconAdded = false
	trayMu.Unlock()
	if !added || trayHwnd == 0 {
		return
	}
	nid := notifyIconDataW{cbSize: uint32(unsafe.Sizeof(notifyIconDataW{})), hWnd: trayHwnd, uID: 1}
	procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
}

// updateTrayIcon 数据更新 → 重绘托盘数字（入口由 main.onSnapshotUpdate 调）
func updateTrayIcon(snap *usage.Snapshot) {
	trayMu.Lock()
	trayLatest = snap
	trayMu.Unlock()

	st := trayStateFromSnapshot(snap)
	applyTrayState(st, buildTooltip(snap))
}

func refreshTrayFromSnapshot() {
	trayMu.Lock()
	s := trayLatest
	trayMu.Unlock()
	if s != nil {
		st := trayStateFromSnapshot(s)
		applyTrayState(st, buildTooltip(s))
	}
}

func trayStateFromSnapshot(snap *usage.Snapshot) TrayIconState {
	cfg := getConfig()
	st := TrayIconState{ShortQuotaPercent: -1, LongQuotaPercent: -1}
	if snap == nil {
		return st
	}
	shortPct := -1
	longPct := -1
	if snap.Primary != nil {
		shortPct = usage.DisplayPercent(snap.Primary.RemainingPct)
	}
	if snap.Secondary != nil {
		longPct = usage.DisplayPercent(snap.Secondary.RemainingPct)
	}
	if cfg.TrayMetric == "longQuota" {
		st.ShortQuotaPercent, st.LongQuotaPercent = longPct, shortPct
	} else {
		st.ShortQuotaPercent, st.LongQuotaPercent = shortPct, longPct
	}
	return st
}

func applyTrayState(st TrayIconState, tip string) {
	if trayHwnd == 0 {
		return
	}
	size := trayIconSourceSize()
	png := renderTrayIconPNGCached(st, size)
	hicon := mustCreateHIconFromPNG(png)
	if hicon == 0 {
		return
	}

	nid := notifyIconDataW{
		cbSize: uint32(unsafe.Sizeof(notifyIconDataW{})),
		hWnd:   trayHwnd,
		uID:    1,
		uFlags: NIF_ICON | NIF_TIP,
		hIcon:  hicon,
	}
	copyTip(nid.szTip[:], tip)

	trayMu.Lock()
	added := trayIconAdded
	trayMu.Unlock()
	if !added {
		addTrayIcon()
	}
	procShellNotifyIconW.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&nid)))

	// 销毁旧 HICON，防 GDI 泄漏（§47）
	trayMu.Lock()
	old := curHIcon
	curHIcon = hicon
	trayMu.Unlock()
	if old != 0 {
		procDestroyIcon.Call(old)
	}
}

func copyTip(dst []uint16, s string) {
	u16, err := windows.UTF16FromString(s)
	if err != nil {
		return
	}
	n := copy(dst, u16)
	if n < len(dst) {
		dst[n] = 0
	} else {
		dst[len(dst)-1] = 0
	}
}

func buildTooltip(snap *usage.Snapshot) string {
	tip := "Codex 用量监控"
	if snap == nil {
		return tip
	}
	if snap.Primary != nil {
		tip += fmt.Sprintf("\n5小时剩余 %d%%（%s 重置）", usage.DisplayPercent(snap.Primary.RemainingPct), formatReset(snap.Primary.ResetsAt))
	}
	if snap.Secondary != nil {
		tip += fmt.Sprintf("\n7天剩余 %d%%（%s 重置）", usage.DisplayPercent(snap.Secondary.RemainingPct), formatReset(snap.Secondary.ResetsAt))
	}
	if snap.Source == "stale" {
		tip += "\n（数据可能已过期）"
	}
	return tip
}

// ---- 右键菜单（需求文档 §49） ----

func handleContextMenu(x, y int) {
	if !trayMenuOpen.CompareAndSwap(false, true) {
		return
	}
	defer trayMenuOpen.Store(false)
	if isHoverUp() {
		hideMainWindow()
	}
	if (x == 0 && y == 0) || (x == -1 && y == -1) {
		var pt pointW
		if getCursorPos(&pt) {
			x, y = int(pt.x), int(pt.y)
		}
	}
	hmenu, _, _ := procCreatePopupMenu.Call()
	if hmenu == 0 {
		return
	}
	defer procDestroyMenu.Call(hmenu)

	trayMu.Lock()
	s := trayLatest
	trayMu.Unlock()
	autostartOn := isAutostartEnabled()
	dashOn := isDashboardUp()

	addMenuItem := func(flags uintptr, id uintptr, text string) {
		p, _ := windows.UTF16PtrFromString(text)
		procAppendMenuW.Call(hmenu, flags, id, uintptr(unsafe.Pointer(p)))
	}
	sep := func() {
		procAppendMenuW.Call(hmenu, MF_SEPARATOR, 0, 0)
	}

	addMenuItem(MF_STRING|MF_GRAYED, 0, "Codex Usage Monitor")
	if s != nil {
		if s.Primary != nil {
			addMenuItem(MF_STRING|MF_GRAYED, 0, fmt.Sprintf("5小时剩余：%d%%", usage.DisplayPercent(s.Primary.RemainingPct)))
		}
		if s.Secondary != nil {
			addMenuItem(MF_STRING|MF_GRAYED, 0, fmt.Sprintf("7天剩余：%d%%", usage.DisplayPercent(s.Secondary.RemainingPct)))
		}
	}
	sep()
	autostartFlag := uintptr(MF_STRING)
	if autostartOn {
		autostartFlag |= MF_CHECKED
	}
	refreshFlag := uintptr(MF_STRING)
	refreshLabel := "刷新数据"
	if refreshRunning.Load() {
		refreshFlag |= MF_GRAYED
		refreshLabel = "正在刷新…"
	}
	addMenuItem(refreshFlag, menuIDRefresh, refreshLabel)
	dashLabel := "打开面板"
	if dashOn {
		dashLabel = "隐藏面板"
	}
	addMenuItem(MF_STRING, menuIDPanel, dashLabel)
	sep()
	addMenuItem(autostartFlag, menuIDAutostart, "开机启动")
	addMenuItem(MF_STRING, menuIDSettings, "设置")
	sep()
	addMenuItem(MF_STRING, menuIDQuit, "退出")

	// 经典 quirk：TrackPopupMenu 前必须 SetForegroundWindow，否则点外部不关闭
	procSetForegroundWindow.Call(trayHwnd)
	cmd, _, _ := procTrackPopupMenuEx.Call(
		hmenu,
		TPM_LEFTALIGN|TPM_RIGHTBUTTON|TPM_RETURNCMD|TPM_NONOTIFY,
		uintptr(x), uintptr(y),
		trayHwnd, 0,
	)

	procPostMessageW.Call(trayHwnd, 0, 0, 0)
	postTrayEvent(trayEvent{kind: "command", x: int(cmd)})
}

func handleMenuCommand(cmd uintptr) {
	cfg := getConfig()
	switch cmd {
	case menuIDRefresh:
		go forceRefresh()
	case menuIDPanel:
		if isDashboardUp() {
			hideMainWindow()
		} else {
			showDashboard()
		}
	case menuIDAutostart:
		cfg.StartWithWindows = !isAutostartEnabled()
		_ = saveConfig(cfg)
	case menuIDSettings:
		showDashboardAt("settings")
	case menuIDQuit:
		quitApp()
	}
}

// ---- Win32 小工具 ----

func mustLoadCursor() uintptr {
	c, _, _ := procLoadCursorW.Call(0, uintptr(IDC_ARROW))
	return c
}

func mustRegisterWindowMessage(name string) uintptr {
	p, _ := windows.UTF16PtrFromString(name)
	r, _, _ := procRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(p)))
	return r
}

func getCursorPos(pt *pointW) bool {
	r, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(pt)))
	return r != 0
}

func getSystemMetrics(index int) int32 {
	r, _, _ := procGetSystemMetrics.Call(uintptr(index))
	return int32(r)
}

// mustCreateHIconFromPNG 从 PNG 内存创建 HICON（Vista+ 支持 PNG 资源）
func mustCreateHIconFromPNG(png []byte) uintptr {
	if len(png) == 0 {
		return 0
	}
	h, _, _ := procCreateIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(&png[0])),
		uintptr(len(png)),
		1, // fIcon = TRUE
		0x00030000,
		0, 0,
		uintptr(LR_DEFAULTCOLOR),
	)
	return h
}

// getMonitorWorkArea 取包含物理坐标点的显示器工作区（§33/§35）
func getMonitorWorkArea(pt pointW) rectW {
	ptArg := uintptr(uint64(uint32(pt.x)) | uint64(uint32(pt.y))<<32)
	mon, _, _ := procMonitorFromPoint.Call(ptArg, uintptr(MONITOR_DEFAULTTONEAREST))
	if mon == 0 {
		return rectW{}
	}
	mi := monitorInfoW{cbSize: uint32(unsafe.Sizeof(monitorInfoW{}))}
	ok, _, _ := procGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi)))
	if ok == 0 {
		return rectW{}
	}
	return mi.rcWork
}

// getWailsWindowDpiScale 取 Wails 主窗口 DPI 缩放（逻辑↔物理换算）
func getWailsWindowDpiScale() float64 {
	h := ensureWailsHwnd()
	if h == 0 {
		return 1.0
	}
	dpi, _, _ := procGetDpiForWindow.Call(h)
	if dpi == 0 {
		return 1.0
	}
	return float64(dpi) / 96.0
}

func getWindowRectPhy(h uintptr) (rectW, bool) {
	var r rectW
	ok, _, _ := procGetWindowRect.Call(h, uintptr(unsafe.Pointer(&r)))
	return r, ok != 0
}

func isMainWindowVisible() bool {
	h := ensureWailsHwnd()
	if h == 0 {
		return false
	}
	r, _, _ := procIsWindowVisible.Call(h)
	return r != 0
}

// findMainWindowByPid 兜底：按 PID 找可见顶层窗口（FindWindow 失败时）
func findWindowByTitle(title string) uintptr {
	p, _ := windows.UTF16PtrFromString(title)
	h, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(p)))
	return h
}

// Shell returns physical screen coordinates, including negative monitor origins.
type notifyIconIdentifier struct {
	cbSize   uint32
	hWnd     uintptr
	uID      uint32
	guidItem [16]byte
}

func getTrayRect() (rectW, bool) {
	id := notifyIconIdentifier{cbSize: uint32(unsafe.Sizeof(notifyIconIdentifier{})), hWnd: trayHwnd, uID: 1}
	var rect rectW
	hr, _, _ := procShellNotifyIconGetRect.Call(uintptr(unsafe.Pointer(&id)), uintptr(unsafe.Pointer(&rect)))
	return rect, int32(hr) >= 0 && rect.right > rect.left
}

func getMonitorDpiScale(pt pointW) float64 {
	mon, _, _ := procMonitorFromPoint.Call(uintptr(uint64(uint32(pt.x))|uint64(uint32(pt.y))<<32), MONITOR_DEFAULTTONEAREST)
	proc := windows.NewLazySystemDLL("shcore.dll").NewProc("GetDpiForMonitor")
	var x, y uint32
	if proc.Find() == nil {
		hr, _, _ := proc.Call(mon, 0, uintptr(unsafe.Pointer(&x)), uintptr(unsafe.Pointer(&y)))
		if int32(hr) >= 0 && x > 0 {
			return float64(x) / 96
		}
	}
	return getWailsWindowDpiScale()
}

func prepareNativeWindow(hover bool) {
	h := ensureWailsHwnd()
	if h == 0 {
		return
	}
	// Hide before changing task-switcher visibility. Restore only maximised windows.
	procShowWindow.Call(h, SW_HIDE)
	iconic, _, _ := modUser32.NewProc("IsZoomed").Call(h)
	if iconic != 0 {
		procShowWindow.Call(h, SW_RESTORE)
		procShowWindow.Call(h, SW_HIDE)
	}
	index := int32(-20) // GWL_EXSTYLE
	get := modUser32.NewProc("GetWindowLongPtrW")
	set := modUser32.NewProc("SetWindowLongPtrW")
	style, _, _ := get.Call(h, uintptr(index))
	if hover {
		style = (style &^ 0x40000) | 0x80 | 0x8000000 // TOOLWINDOW, NOACTIVATE
	} else {
		style = (style &^ (0x80 | 0x8000000)) | 0x40000
	}
	set.Call(h, uintptr(index), style)
	procSetWindowPos.Call(h, 0, 0, 0, 0, 0, 0x20|0x4|SWP_NOSIZE|SWP_NOMOVE|SWP_NOACTIVATE)
}
