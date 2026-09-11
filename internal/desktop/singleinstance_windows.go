// singleinstance_windows.go - 单实例（需求文档 §38）
// CreateMutex 检测；二次启动时唤起已有实例的面板后退出

package desktop

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

const mutexName = `Local\CodexUsageMonitor.SingleInstance`

// acquireSingleInstance 返回 true 表示本进程是唯一实例
func acquireSingleInstance() bool {
	name, err := windows.UTF16PtrFromString(mutexName)
	if err != nil {
		return true
	}
	h, err := windows.CreateMutex(nil, false, name)
	if err != nil {
		// ERROR_ALREADY_EXISTS 等 → 已有实例
		activateExistingWindow()
		return false
	}
	_ = h // 持有句柄直到进程退出
	return true
}

func activateExistingWindow() {
	class, _ := windows.UTF16PtrFromString("CodexMonitorTrayWnd")
	h, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(class)), 0)
	if h == 0 {
		fmt.Fprintln(os.Stderr, "Codex Usage Monitor 已在运行")
		return
	}
	procPostMessageW.Call(h, appWM_OPENPANEL, 0, 0)
}
