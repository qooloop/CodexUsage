// icon.go - 动态托盘图标渲染器（需求文档 §5/§6/§8/§9）
// 形态：深色圆角底 + 白色大数字（5h 剩余%）+ 底部进度条（7d 剩余%）
// 状态色：剩余 ≥50% 绿 / 20-49% 黄 / <20% 红；无数据灰底 "--"
// 缓存：按 key 缓存 PNG 渲染结果，避免重复生成（§9）
//

package desktop

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"sync"
)

// ---- 设计 token（与前端 tokens.css 对齐） ----

var (
	colBg      = color.RGBA{0x0D, 0x16, 0x26, 0xFF} // #0D1626 深蓝底
	colBgEmpty = color.RGBA{0x3A, 0x44, 0x52, 0xFF} // 无数据灰底
	colDigit   = color.RGBA{0xF5, 0xF7, 0xFA, 0xFF} // 白色数字
	colFlash   = color.RGBA{0x46, 0xBD, 0xF4, 0xFF} // flash 高亮（品牌蓝）
	colBarBg   = color.RGBA{0x2A, 0x36, 0x48, 0xFF} // 进度条底
	colGood    = color.RGBA{0x35, 0xE6, 0xA5, 0xFF} // #35E6A5
	colWarn    = color.RGBA{0xF6, 0xC4, 0x53, 0xFF} // #F6C453
	colBad     = color.RGBA{0xFF, 0x61, 0x74, 0xFF} // #FF6174
	colIdle    = color.RGBA{0x6B, 0x7C, 0x92, 0xFF} // 灰
)

// 3x5 像素字体（每行 3bit，从高位开始）
var digitFont = map[rune][5]byte{
	'0': {0b111, 0b101, 0b101, 0b101, 0b111},
	'1': {0b010, 0b110, 0b010, 0b010, 0b111},
	'2': {0b111, 0b001, 0b111, 0b100, 0b111},
	'3': {0b111, 0b001, 0b011, 0b001, 0b111},
	'4': {0b101, 0b101, 0b111, 0b001, 0b001},
	'5': {0b111, 0b100, 0b111, 0b001, 0b111},
	'6': {0b111, 0b100, 0b111, 0b101, 0b111},
	'7': {0b111, 0b001, 0b001, 0b010, 0b010},
	'8': {0b111, 0b101, 0b111, 0b101, 0b111},
	'9': {0b111, 0b101, 0b111, 0b001, 0b111},
	'-': {0b000, 0b000, 0b111, 0b000, 0b000},
}

func quotaColorRGBA(remaining float64) color.RGBA {
	switch {
	case remaining >= 50:
		return colGood
	case remaining >= 20:
		return colWarn
	default:
		return colBad
	}
}

// TrayIconState 渲染入参（需求文档 §8）
type TrayIconState struct {
	ShortQuotaPercent int // 5h 剩余%，-1 = 无数据
	LongQuotaPercent  int // 7d 剩余%，-1 = 无数据
	Flash             bool
}

// ---- PNG 渲染缓存（§9 图标缓存） ----

var (
	iconCacheMu sync.Mutex
	iconCache   = make(map[string][]byte)
)

func trayIconSourceSize() int {
	const SM_CXSMICON = 49
	s := int(getSystemMetrics(SM_CXSMICON))
	if s < 16 {
		s = 16
	}
	if s > 64 {
		s = 64
	}
	return s
}

// renderTrayIconPNG 渲染一帧托盘图标 PNG
func renderTrayIconPNG(st TrayIconState, size int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	bg := colBg
	if st.ShortQuotaPercent < 0 && st.LongQuotaPercent < 0 {
		bg = colBgEmpty
	}
	drawRoundedRect(img, 0, 0, size, size, size/5, bg)

	// 数字部分
	text := "--"
	if st.ShortQuotaPercent >= 0 {
		text = fmt.Sprintf("%d", st.ShortQuotaPercent)
	} else if st.LongQuotaPercent >= 0 {
		text = fmt.Sprintf("%d", st.LongQuotaPercent)
	}
	digitCol := colDigit
	if st.Flash {
		digitCol = colFlash
	}

	scale := 1
	switch {
	case size >= 48:
		scale = 4
	case size >= 32:
		scale = 3
	case size >= 24:
		scale = 2
	}

	// 数字宽 = 每字 3*scale + 字间 1*scale
	textW := len(text)*3*scale + (len(text)-1)*scale
	textH := 5 * scale
	x0 := (size - textW) / 2
	y0 := (size-textH)/2 - scale/2 // 略偏上，给底部进度条让位
	if y0 < 1 {
		y0 = 1
	}
	cx := x0
	for _, ch := range text {
		glyph, ok := digitFont[ch]
		if !ok {
			cx += 4 * scale
			continue
		}
		for row := 0; row < 5; row++ {
			for col := 0; col < 3; col++ {
				if glyph[row]&(0b100>>col) != 0 {
					drawRect(img, cx+col*scale, y0+row*scale, cx+(col+1)*scale, y0+(row+1)*scale, digitCol)
				}
			}
		}
		cx += 4 * scale
	}

	if !getConfig().ShowLongBar {
		var buf bytes.Buffer
		_ = png.Encode(&buf, img)
		return buf.Bytes()
	}
	// 底部进度条（7d 剩余）
	barH := scale + 1
	if barH < 2 {
		barH = 2
	}
	barY1 := size - 1
	barY0 := barY1 - barH
	barX0 := scale
	barX1 := size - scale
	if barX1 <= barX0 {
		barX1 = size
	}
	// 槽底
	drawRect(img, barX0, barY0, barX1, barY1, colBarBg)
	fillPct := 0
	if st.LongQuotaPercent >= 0 {
		fillPct = st.LongQuotaPercent
	}
	barCol := colIdle
	if st.LongQuotaPercent >= 0 {
		barCol = quotaColorRGBA(float64(st.LongQuotaPercent))
	}
	fillW := (barX1 - barX0) * fillPct / 100
	if fillW > 0 {
		drawRect(img, barX0, barY0, barX0+fillW, barY1, barCol)
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// renderTrayIconPNGCached 带缓存的渲染
func renderTrayIconPNGCached(st TrayIconState, size int) []byte {
	key := fmt.Sprintf("%t_", getConfig().ShowLongBar) + fmt.Sprintf("%d_%d_%v_%d", st.ShortQuotaPercent, st.LongQuotaPercent, st.Flash, size)
	iconCacheMu.Lock()
	defer iconCacheMu.Unlock()
	if v, ok := iconCache[key]; ok {
		return v
	}
	v := renderTrayIconPNG(st, size)
	if len(iconCache) > 512 { // 粗暴防膨胀：清空重来（缓存 key 有界且有限）
		iconCache = make(map[string][]byte)
	}
	iconCache[key] = v
	return v
}

// ---- 简单绘图原语 ----

func drawRect(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if x >= 0 && y >= 0 && x < img.Rect.Dx() && y < img.Rect.Dy() {
				img.Set(x, y, c)
			}
		}
	}
}

// drawRoundedRect 圆角矩形（切角近似，小尺寸下视觉足够）
func drawRoundedRect(img *image.RGBA, x0, y0, x1, y1 int, r int, c color.Color) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			dx, dy := 0, 0
			corner := false
			if x-x0 < r && y-y0 < r {
				corner, dx, dy = true, x-x0, y-y0
			} else if x1-1-x < r && y-y0 < r {
				corner, dx, dy = true, x1-1-x, y-y0
			} else if x-x0 < r && y1-1-y < r {
				corner, dx, dy = true, x-x0, y1-1-y
			} else if x1-1-x < r && y1-1-y < r {
				corner, dx, dy = true, x1-1-x, y1-1-y
			}
			if corner && dx*dx+dy*dy > r*r {
				continue
			}
			img.Set(x, y, c)
		}
	}
}
