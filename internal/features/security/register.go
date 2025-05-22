package security

import (
	"github.com/teomyth/iniq/internal/features"
	"github.com/teomyth/iniq/pkg/osdetect"
)

// init registers the security feature
func init() {
	features.RegisterSecurityFeature = func(registry *features.Registry, osInfo *osdetect.Info) {
		registry.Register(New(osInfo))
	}
}
