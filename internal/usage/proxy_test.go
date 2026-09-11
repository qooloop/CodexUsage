package usage

import (
	"net/http"
	"net/url"
	"testing"
)

func TestWindowsProxyRules(t *testing.T) {
	for _, tc := range []struct{ name, target, server, bypass, want string }{
		{"shared", "https://chatgpt.com/", "127.0.0.1:7890", "", "http://127.0.0.1:7890"},
		{"by scheme", "https://chatgpt.com/", "http=localhost:8080;https=localhost:7890", "", "http://localhost:7890"},
		{"http only", "https://chatgpt.com/", "http=localhost:8080", "", ""},
		{"disabled", "https://chatgpt.com/", "", "", ""},
		{"bypass", "https://chatgpt.com/", "localhost:7890", "*.example.com;chatgpt.com", ""},
		{"local", "http://intranet/", "localhost:7890", "<local>", ""},
		{"loopback", "http://127.0.0.1/", "localhost:7890", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, _ := url.Parse(tc.target)
			got, err := proxyForURL(target, tc.server, tc.bypass)
			if err != nil {
				t.Fatal(err)
			}
			actual := ""
			if got != nil {
				actual = got.String()
			}
			if actual != tc.want {
				t.Fatalf("got %q want %q", actual, tc.want)
			}
		})
	}
}

func TestExplicitProxyAndNoProxyTakePrecedence(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "http://explicit.invalid:7890")
	t.Setenv("NO_PROXY", "chatgpt.com")
	req, _ := http.NewRequest("GET", "https://chatgpt.com/", nil)
	if got, err := quotaProxy(req); err != nil || got != nil {
		t.Fatalf("NO_PROXY ignored: %v %v", got, err)
	}
	t.Setenv("NO_PROXY", "")
	got, err := quotaProxy(req)
	if err != nil || got == nil || got.Host != "explicit.invalid:7890" {
		t.Fatalf("explicit proxy ignored: %v %v", got, err)
	}
}
