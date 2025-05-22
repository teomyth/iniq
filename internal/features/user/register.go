package user

import (
	"github.com/teomyth/iniq/internal/features"
	"github.com/teomyth/iniq/pkg/osdetect"
)

// init registers the user feature
func init() {
	features.RegisterUserFeature = func(registry *features.Registry, osInfo *osdetect.Info) {
		registry.Register(New(osInfo))
	}
}
