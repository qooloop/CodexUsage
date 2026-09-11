// Codex 在线额度补充：GET https://chatgpt.com/backend-api/wham/usage。
// 本地 JSONL 与在线快照按记录时间选取最新数据；网络失败不阻挡本地同步。
// Windows 请求跟随系统手动代理，显式 HTTP(S)_PROXY 环境变量优先。
// 认证：~/.codex/auth.json 的 ChatGPT OAuth；401 时 refresh 后重试一次。

package usage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	whamUsageURL   = "https://chatgpt.com/backend-api/wham/usage"
	oauthTokenURL  = "https://auth.openai.com/oauth/token"
	codexClientID  = "app_EMoamEEZ73f0CkXaXp7hrann"
	whamHTTPTimout = 10 * time.Second
)

var whamHTTP = newQuotaHTTPClient()

func newQuotaHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = quotaProxy
	return &http.Client{Timeout: whamHTTPTimout, Transport: transport}
}

// ---- auth.json ----

type codexAuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	AccountID    string `json:"account_id"`
}

func codexAuthPath() string {
	home := os.Getenv("CODEX_HOME")
	if home == "" {
		if h, err := os.UserHomeDir(); err == nil {
			home = filepath.Join(h, ".codex")
		}
	}
	return filepath.Join(home, "auth.json")
}

// ---- wham/usage 响应 ----

type whamWindow struct {
	UsedPercent        float64 `json:"used_percent"`
	ResetAt            int64   `json:"reset_at"`
	LimitWindowSeconds int64   `json:"limit_window_seconds"`
}

type whamUsageResponse struct {
	PlanType  string `json:"plan_type"`
	RateLimit *struct {
		PrimaryWindow   *whamWindow `json:"primary_window"`
		SecondaryWindow *whamWindow `json:"secondary_window"`
	} `json:"rate_limit"`
	Credits *struct {
		HasCredits bool            `json:"has_credits"`
		Unlimited  bool            `json:"unlimited"`
		Balance    json.RawMessage `json:"balance"`
	} `json:"credits"`
}

// fetchWhamUsage 拉实时额度并转成 quota Event（窗口按 limit_window_seconds 归类）
func fetchWhamUsage() (*Event, error) {
	authPath := codexAuthPath()
	tokens, err := readAuthTokens(authPath)
	if err != nil {
		return nil, err
	}
	if tokens.AccessToken == "" {
		return nil, fmt.Errorf("auth.json 无 access_token（可能未登录 ChatGPT）")
	}

	status, body, err := whamGet(tokens)
	if err != nil {
		return nil, err
	}
	if status == 401 {
		// token 过期：refresh 后重试一次（与 codex-rs client.rs 行为一致）
		if err := refreshCodexToken(authPath, tokens.RefreshToken); err != nil {
			return nil, fmt.Errorf("token 刷新失败: %w", err)
		}
		tokens, err = readAuthTokens(authPath)
		if err != nil {
			return nil, err
		}
		status, body, err = whamGet(tokens)
		if err != nil {
			return nil, err
		}
	}
	if status != 200 {
		return nil, fmt.Errorf("额度服务 HTTP %d", status)
	}

	var resp whamUsageResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("wham/usage 响应解析失败: %w", err)
	}
	return whamToEvent(&resp), nil
}

func whamGet(tokens *codexAuthTokens) (int, []byte, error) {
	req, err := http.NewRequest("GET", whamUsageURL, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "codex-cli")
	if tokens.AccountID != "" {
		req.Header.Set("ChatGPT-Account-Id", tokens.AccountID)
	}
	resp, err := whamHTTP.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, body, err
}

// whamToEvent 按窗口秒数归类（18000=5h, 604800=7d），与 JSONL 口径一致
func whamToEvent(r *whamUsageResponse) *Event {
	ev := &Event{Kind: "quota", Ts: time.Now().UnixMilli()}
	if r.PlanType != "" {
		s := r.PlanType
		ev.PlanType = &s
	}
	if r.RateLimit != nil {
		toWindow := func(w *whamWindow) *Window {
			if w == nil {
				return nil
			}
			out := &Window{
				UsedPercent:   w.UsedPercent,
				WindowMinutes: int(w.LimitWindowSeconds / 60),
			}
			if w.ResetAt > 0 {
				ra := w.ResetAt
				out.ResetsAt = &ra
			}
			return out
		}
		primary := toWindow(r.RateLimit.PrimaryWindow)
		secondary := toWindow(r.RateLimit.SecondaryWindow)
		ev.Primary = pickWindow(primary, secondary, Window5H)
		ev.Secondary = pickWindow(primary, secondary, Window7D)
	}
	if c := r.Credits; c != nil {
		ev.Credits = &Credits{
			HasCredits: c.HasCredits,
			Unlimited:  c.Unlimited,
			Balance:    strings.Trim(string(c.Balance), `"`),
		}
	}
	return ev
}

// ---- token 刷新 ----

func readAuthTokens(path string) (*codexAuthTokens, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Tokens *codexAuthTokens `json:"tokens"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, err
	}
	if wrapper.Tokens == nil {
		return nil, fmt.Errorf("auth.json 缺少 tokens 字段")
	}
	return wrapper.Tokens, nil
}

func refreshCodexToken(authPath, refreshToken string) error {
	if refreshToken == "" {
		return fmt.Errorf("无 refresh_token")
	}
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {codexClientID},
		"refresh_token": {refreshToken},
	}
	req, err := http.NewRequest("POST", oauthTokenURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := whamHTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != 200 {
		return fmt.Errorf("oauth/token HTTP %d", resp.StatusCode)
	}

	var tok struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &tok); err != nil || tok.AccessToken == "" {
		return fmt.Errorf("oauth/token 响应无 access_token")
	}

	// 回写 auth.json：保留其它字段，只更新 tokens + last_refresh
	raw, err := os.ReadFile(authPath)
	if err != nil {
		return err
	}
	var full map[string]interface{}
	if err := json.Unmarshal(raw, &full); err != nil {
		return err
	}
	tokens, _ := full["tokens"].(map[string]interface{})
	if tokens == nil {
		tokens = map[string]interface{}{}
	}
	tokens["access_token"] = tok.AccessToken
	if tok.RefreshToken != "" {
		tokens["refresh_token"] = tok.RefreshToken
	}
	if tok.IDToken != "" {
		tokens["id_token"] = tok.IDToken
	}
	full["tokens"] = tokens
	full["last_refresh"] = time.Now().UTC().Format(time.RFC3339Nano)

	out, err := json.MarshalIndent(full, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(authPath, out, 0600)
}
