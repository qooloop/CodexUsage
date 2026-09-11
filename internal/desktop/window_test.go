package desktop

import (
	"codex-monitor/internal/usage"
	"testing"
	"unsafe"
)

func TestWindowsABILayout(t *testing.T) {
	if unsafe.Sizeof(monitorInfoW{}) != 40 || unsafe.Offsetof(monitorInfoW{}.rcWork) != 20 {
		t.Fatal("MONITORINFO must be 40 bytes with rcWork at byte 20; padding breaks GetMonitorInfo")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (unsafe.Sizeof(notifyIconDataW{}) != 976 || unsafe.Offsetof(notifyIconDataW{}.uVersion) != 816) {
		t.Fatal("NOTIFYICONDATAW layout does not match the Windows x64 ABI")
	}
	if HWND_TOPMOST != ^uintptr(0) || HWND_NOTOPMOST != ^uintptr(1) {
		t.Fatal("incorrect z-order pseudo handles")
	}
}

func TestMonitorWorkAreaNative(t *testing.T) {
	var pt pointW
	if !getCursorPos(&pt) {
		t.Skip("no interactive desktop")
	}
	work := getMonitorWorkArea(pt)
	if work.right <= work.left || work.bottom <= work.top {
		t.Fatalf("GetMonitorInfo failed: %+v", work)
	}
}

func TestPopupPlacement(t *testing.T) {
	cases := []struct {
		name         string
		work, anchor rectW
		w, h         int
	}{
		{"bottom", rectW{0, 0, 1920, 1040}, rectW{1780, 1040, 1804, 1080}, 344, 288},
		{"top", rectW{0, 40, 1920, 1080}, rectW{800, 0, 824, 40}, 344, 288},
		{"left", rectW{40, 0, 1920, 1080}, rectW{0, 600, 40, 624}, 344, 288},
		{"right", rectW{0, 0, 1880, 1080}, rectW{1880, 600, 1920, 624}, 344, 288},
		{"negative origin 150%", rectW{-2560, -400, 0, 1000}, rectW{-150, 1000, -114, 1040}, 516, 432},
		{"200%", rectW{1920, 0, 5760, 2080}, rectW{5200, 2080, 5248, 2160}, 688, 576},
		{"small display", rectW{0, 0, 320, 240}, rectW{280, 240, 304, 264}, 344, 288},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := popupBounds(c.work, c.anchor, c.w, c.h, 8)
			if got.left < c.work.left || got.top < c.work.top || got.right > c.work.right || got.bottom > c.work.bottom {
				t.Fatalf("offscreen: %+v", got)
			}
			if got.right-got.left != int32(min(c.w, int(c.work.right-c.work.left))) {
				t.Fatalf("bad width: %+v", got)
			}
			if c.name == "bottom" && got.bottom > c.anchor.top {
				t.Fatal("bottom popup overlaps tray")
			}
			if c.name == "top" && got.top < c.anchor.bottom {
				t.Fatal("top popup overlaps tray")
			}
		})
	}
}

func TestVersion4TrayActivation(t *testing.T) {
	oldEvents, oldVersion := trayEvents, trayVersion4
	defer func() { trayEvents, trayVersion4 = oldEvents, oldVersion }()
	trayEvents = make(chan trayEvent, 8)
	trayVersion4 = true
	trayWndProc(0, appWM_TRAYCALLBACK, 0, uintptr(WM_LBUTTONUP))
	trayWndProc(0, appWM_TRAYCALLBACK, 0, uintptr(NIN_SELECT))
	if len(trayEvents) != 1 || (<-trayEvents).kind != "click" {
		t.Fatal("one click must toggle exactly once")
	}
	trayWndProc(0, appWM_TRAYCALLBACK, 0, uintptr(NIN_POPUPOPEN))
	if len(trayEvents) != 1 || (<-trayEvents).kind != "hover" {
		t.Fatal("rich popup event ignored")
	}
	trayVersion4 = false
	trayWndProc(0, appWM_TRAYCALLBACK, 0, uintptr(WM_LBUTTONUP))
	if len(trayEvents) != 1 || (<-trayEvents).kind != "click" {
		t.Fatal("legacy activation ignored")
	}
}

func TestTrayFractionalQuota(t *testing.T) {
	state := trayStateFromSnapshot(&usage.Snapshot{Primary: &usage.QuotaWindow{RemainingPct: 57.9}, Secondary: &usage.QuotaWindow{RemainingPct: 51.9}})
	if state.ShortQuotaPercent != 57 || state.LongQuotaPercent != 51 {
		t.Fatalf("rounded quota: %+v", state)
	}
}
