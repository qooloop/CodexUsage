// Windows desktop entry point. Application code lives in internal/.
package main

import (
	"codex-monitor/frontend"
	"codex-monitor/internal/desktop"
)

func main() { desktop.Run(frontend.Assets) }
