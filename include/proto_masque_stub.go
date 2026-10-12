//go:build !with_masque

package include

import (
	"github.com/sagernet/sing-box/adapter/endpoint"
	"github.com/sagernet/sing-box/adapter/outbound"
)

// !with_masque: MASQUE endpoint + outbound excluded from this build (Android-only slim profile).
func registerMasqueEndpoint(registry *endpoint.Registry) {
}

func registerMasqueOutbound(registry *outbound.Registry) {
}
