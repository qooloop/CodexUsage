package usage

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/http/httpproxy"
)

func quotaProxy(req *http.Request) (*url.URL, error) {
	cfg := httpproxy.FromEnvironment()
	// Explicit environment settings (including NO_PROXY) take precedence.
	if cfg.HTTPProxy != "" || cfg.HTTPSProxy != "" {
		return cfg.ProxyFunc()(req.URL)
	}
	server, bypass := systemProxySettings()
	if cfg.NoProxy != "" {
		bypass += ";" + strings.ReplaceAll(cfg.NoProxy, ",", ";")
	}
	return proxyForURL(req.URL, server, bypass)
}

// Windows supports one proxy for all schemes or a semicolon-separated map.
func proxyForURL(target *url.URL, server, bypass string) (*url.URL, error) {
	var httpServer, httpsServer string
	if !strings.Contains(server, "=") {
		httpServer, httpsServer = strings.TrimSpace(server), strings.TrimSpace(server)
	} else {
		for _, entry := range strings.Split(server, ";") {
			scheme, value, ok := strings.Cut(strings.TrimSpace(entry), "=")
			if !ok {
				continue
			}
			switch strings.ToLower(strings.TrimSpace(scheme)) {
			case "http":
				httpServer = strings.TrimSpace(value)
			case "https":
				httpsServer = strings.TrimSpace(value)
			}
		}
	}
	var exclusions []string
	for _, item := range strings.FieldsFunc(bypass, func(r rune) bool { return r == ';' || r == ',' }) {
		item = strings.TrimSpace(item)
		if strings.EqualFold(item, "<local>") {
			host := target.Hostname()
			if !strings.Contains(host, ".") && net.ParseIP(host) == nil {
				return nil, nil
			}
			continue
		}
		exclusions = append(exclusions, item)
	}
	cfg := httpproxy.Config{HTTPProxy: httpServer, HTTPSProxy: httpsServer, NoProxy: strings.Join(exclusions, ",")}
	return cfg.ProxyFunc()(target)
}
