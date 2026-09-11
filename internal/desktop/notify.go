// notify.go - 阈值通知（需求文档 §50）
// 只在剩余额度跌破阈值"跨越"时提醒一次，恢复后重新武装；不做轮询骚扰

package desktop

import (
	"codex-monitor/internal/usage"
	"fmt"
	"log"

	toast "git.sr.ht/~jackmordaunt/go-toast/v2"
)

var (
	notify5hArmed = true // true = 高于阈值（允许提醒）
	notify7dArmed = true
)

// checkThresholds 每次快照更新时调用
func checkThresholds(cur *usage.Snapshot) {
	if cur == nil {
		return
	}
	cfg := getConfig()

	if cur.Primary != nil {
		short := usage.DisplayPercent(cur.Primary.RemainingPct)
		if thr := cfg.Notify5hBelow; thr > 0 {
			if notify5hArmed && short <= thr {
				notify5hArmed = false
				pushQuotaToast("5小时额度", short, thr)
			} else if short > thr {
				notify5hArmed = true
			}
		}
	}
	if cur.Secondary != nil {
		long := usage.DisplayPercent(cur.Secondary.RemainingPct)
		if thr := cfg.Notify7dBelow; thr > 0 {
			if notify7dArmed && long <= thr {
				notify7dArmed = false
				pushQuotaToast("7天额度", long, thr)
			} else if long > thr {
				notify7dArmed = true
			}
		}
	}
}

func pushQuotaToast(label string, pct, thr int) {
	n := toast.Notification{
		AppID: "Codex Usage Monitor",
		Title: fmt.Sprintf("Codex %s 剩余 %d%%", label, pct),
		Body:  fmt.Sprintf("%s 额度已低于 %d%%，注意使用节奏。", label, thr),
	}
	if err := n.Push(); err != nil {
		log.Println("toast notify:", err)
	}
}
