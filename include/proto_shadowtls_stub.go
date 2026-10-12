//go:build !with_shadowtls

package include

import (
	"github.com/sagernet/sing-box/adapter/inbound"
	"github.com/sagernet/sing-box/adapter/outbound"
)

// !with_shadowtls: ShadowTLS excluded from this build (Android-only slim profile).
func registerShadowTlsInbound(registry *inbound.Registry) {
}

func registerShadowTlsOutbound(registry *outbound.Registry) {
}
