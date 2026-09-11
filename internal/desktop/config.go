package desktop

import (
	"codex-monitor/internal/config"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func getConfig() config.Config  { return config.Get() }
func loadConfig() config.Config { return config.Load() }
func saveConfig(c config.Config) error {
	if err := config.Save(c); err != nil {
		return err
	}
	applyAutostart(c.StartWithWindows)
	if appCtx != nil {
		wailsruntime.EventsEmit(appCtx, "config:update", c)
	}
	return nil
}
