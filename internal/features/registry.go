// Package features provides the core feature interface and registry
package features

import (
	"github.com/teomyth/iniq/pkg/osdetect"
)

// RegisterFeatures registers all features with the registry
func RegisterFeatures(registry *Registry, osInfo *osdetect.Info) {
	// Call registration functions from each feature package
	// These are set by the init() functions in each package
	if RegisterUserFeature != nil {
		RegisterUserFeature(registry, osInfo)
	}

	if RegisterSSHFeature != nil {
		RegisterSSHFeature(registry, osInfo)
	}

	if RegisterSudoFeature != nil {
		RegisterSudoFeature(registry, osInfo)
	}

	if RegisterSecurityFeature != nil {
		RegisterSecurityFeature(registry, osInfo)
	}
}
