// config.go - 配置持久化（需求文档 §39）
// 路径：%APPDATA%\CodexUsageMonitor\config.json
// 敏感凭证不入此文件；在线额度复用 Codex 现有认证。

package config

import (
	"encoding/json"

	"os"
	"path/filepath"
	"sync"
)

// Config 应用配置（与需求文档 §39 字段对齐）
type Config struct {
	RefreshIntervalSec int    `json:"refreshInterval"` // 5/10/30/60/300，默认 60
	StartWithWindows   bool   `json:"startWithWindows"`
	HideOnStartup      bool   `json:"hideOnStartup"`  // 启动后隐藏主窗口（仅托盘）
	ShowHoverPopup     bool   `json:"showHoverPopup"` // 悬停托盘显示详情
	TrayMetric         string `json:"trayMetric"`     // shortQuota | longQuota（托盘主数字）
	ShowLongBar        bool   `json:"showLongBar"`    // 显示底部长期额度条
	Notify5hBelow      int    `json:"notify5hBelow"`  // 10/20/30/0=关闭
	Notify7dBelow      int    `json:"notify7dBelow"`  // 10/20/30/0=关闭
}

func Default() Config {
	return Config{
		RefreshIntervalSec: 60,
		StartWithWindows:   false,
		HideOnStartup:      true,
		ShowHoverPopup:     true,
		TrayMetric:         "shortQuota",
		ShowLongBar:        true,
		Notify5hBelow:      20,
		Notify7dBelow:      20,
	}
}

var (
	cfgMu       sync.RWMutex
	globalCfg   = Default()
	cfgLoadOnce sync.Once
)

func configDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = "."
	}
	return filepath.Join(appData, "CodexUsageMonitor")
}

func configFile() string {
	return filepath.Join(configDir(), "config.json")
}

// Load 启动时调用一次；失败静默用默认值
func Load() Config {
	cfgLoadOnce.Do(func() {
		data, err := os.ReadFile(configFile())
		if err != nil {
			return
		}
		c := Default()
		if err := json.Unmarshal(data, &c); err == nil {
			c.RefreshIntervalSec = NormalizeInterval(c.RefreshIntervalSec)
			if c.TrayMetric != "longQuota" {
				c.TrayMetric = "shortQuota"
			}
			cfgMu.Lock()
			globalCfg = c
			cfgMu.Unlock()
		}
	})
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	return globalCfg
}

// Save 全量写回（含自动应用开机启动）
func Save(c Config) error {
	c.RefreshIntervalSec = NormalizeInterval(c.RefreshIntervalSec)
	if c.TrayMetric != "longQuota" {
		c.TrayMetric = "shortQuota"
	}
	if err := os.MkdirAll(configDir(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(configFile(), data, 0o644); err != nil {
		return err
	}
	cfgMu.Lock()
	globalCfg = c
	cfgMu.Unlock()
	return nil
}

func Get() Config {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	return globalCfg
}

// NormalizeInterval migrates removed choices (including the old 2-minute option).
func NormalizeInterval(seconds int) int {
	switch seconds {
	case 5, 10, 30, 60, 300:
		return seconds
	default:
		return 60
	}
}
