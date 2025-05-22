package ssh

import (
	"github.com/teomyth/iniq/internal/features"
	"github.com/teomyth/iniq/pkg/osdetect"
)

// init registers the SSH feature
func init() {
	features.RegisterSSHFeature = func(registry *features.Registry, osInfo *osdetect.Info) {
		registry.Register(New(osInfo))
	}
}
