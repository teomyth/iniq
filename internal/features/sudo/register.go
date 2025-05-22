package sudo

import (
	"github.com/teomyth/iniq/internal/features"
	"github.com/teomyth/iniq/pkg/osdetect"
)

// init registers the sudo feature
func init() {
	features.RegisterSudoFeature = func(registry *features.Registry, osInfo *osdetect.Info) {
		registry.Register(New(osInfo))
	}
}
