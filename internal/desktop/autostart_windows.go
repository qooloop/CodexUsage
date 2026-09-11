// autostart_windows.go - 开机自启（需求文档 §37）
// 只写 HKEY_CURRENT_USER\...\Run，不要求管理员权限

package desktop

import (
	"os"
	"os/exec"

	"golang.org/x/sys/windows/registry"
)

const autostartRunKey = `Software\Microsoft\Windows\CurrentVersion\Run`
const autostartValueName = "CodexUsageMonitor"

func currentExePath() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	abs, err := exec.LookPath(exe)
	if err != nil {
		return exe
	}
	return abs
}

func isAutostartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, autostartRunKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	v, _, err := k.GetStringValue(autostartValueName)
	return err == nil && v != ""
}

// setAutostart 写/删 HKCU Run 键值
func setAutostart(enable bool) error {
	if enable {
		exe := currentExePath()
		if exe == "" {
			return nil
		}
		k, err := registry.OpenKey(registry.CURRENT_USER, autostartRunKey, registry.SET_VALUE)
		if err != nil {
			return err
		}
		defer k.Close()
		return k.SetStringValue(autostartValueName, `"`+exe+`"`)
	}
	k, err := registry.OpenKey(registry.CURRENT_USER, autostartRunKey, registry.SET_VALUE)
	if err != nil {
		return nil // 键不存在等情形视为成功
	}
	defer k.Close()
	_ = k.DeleteValue(autostartValueName)
	return nil
}

// applyAutostart 同步配置里的开关到注册表（并纠正与实际状态不符）
func applyAutostart(enable bool) {
	if !enable && !isAutostartEnabled() {
		return
	}
	_ = setAutostart(enable)
}
