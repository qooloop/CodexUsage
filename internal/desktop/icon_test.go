// icon_test.go - 托盘图标渲染快照（本地可视化检查用）
// 运行: go test -run TestRenderTrayIcon -v
// 输出: preview/tray-icon-*.png

package desktop

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenderTrayIcon(t *testing.T) {
	out := t.TempDir()
	cases := []struct {
		name string
		st   TrayIconState
		size int
	}{
		{"tray-icon-73-60-16", TrayIconState{ShortQuotaPercent: 73, LongQuotaPercent: 60}, 16},
		{"tray-icon-73-60-24", TrayIconState{ShortQuotaPercent: 73, LongQuotaPercent: 60}, 24},
		{"tray-icon-73-60-32", TrayIconState{ShortQuotaPercent: 73, LongQuotaPercent: 60}, 32},
		{"tray-icon-15-40-16", TrayIconState{ShortQuotaPercent: 15, LongQuotaPercent: 40}, 16},
		{"tray-icon-8-12-16", TrayIconState{ShortQuotaPercent: 8, LongQuotaPercent: 12}, 16},
		{"tray-icon-empty-16", TrayIconState{ShortQuotaPercent: -1, LongQuotaPercent: -1}, 16},
	}
	for _, c := range cases {
		png := renderTrayIconPNG(c.st, c.size)
		if len(png) == 0 {
			t.Fatalf("%s: empty png", c.name)
		}
		if err := os.WriteFile(filepath.Join(out, c.name+".png"), png, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
