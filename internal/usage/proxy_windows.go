package usage

import "golang.org/x/sys/windows/registry"

// Read on every request so changing the Windows proxy does not require restart.
func systemProxySettings() (server, bypass string) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		return "", ""
	}
	defer k.Close()
	enabled, _, err := k.GetIntegerValue("ProxyEnable")
	if err != nil || enabled == 0 {
		return "", ""
	}
	server, _, _ = k.GetStringValue("ProxyServer")
	bypass, _, _ = k.GetStringValue("ProxyOverride")
	return
}
