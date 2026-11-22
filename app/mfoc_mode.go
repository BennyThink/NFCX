package app

import (
	"fmt"
	"strings"

	"github.com/BennyThink/NFCX/internal/attack"
)

// mfoCModeFromDTO maps the optional GUI mode string onto the engine mode.
// An empty value keeps the upstream (balanced) behaviour so older frontends
// and scripts keep working unchanged.
func mfoCModeFromDTO(raw string) (attack.MFoCMode, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(attack.MFoCModeBalanced):
		return attack.MFoCModeBalanced, nil
	case string(attack.MFoCModeHighPerformance), "high-performance", "performance", "highperf":
		return attack.MFoCModeHighPerformance, nil
	case string(attack.MFoCModeX64), "amd64", "high_performance_x64":
		return attack.MFoCModeX64, nil
	default:
		return "", fmt.Errorf("未知的 mfoc 模式：%q（可用：balanced、high_performance、x64）", raw)
	}
}
