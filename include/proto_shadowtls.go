//go:build with_shadowtls

package include

import (
	"github.com/sagernet/sing-box/adapter/inbound"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/protocol/shadowtls"
)

// with_shadowtls: register the ShadowTLS inbound + outbound.
func registerShadowTlsInbound(registry *inbound.Registry) {
	shadowtls.RegisterInbound(registry)
}

func registerShadowTlsOutbound(registry *outbound.Registry) {
	shadowtls.RegisterOutbound(registry)
}
