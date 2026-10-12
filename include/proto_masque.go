//go:build with_masque

package include

import (
	"github.com/sagernet/sing-box/adapter/endpoint"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/protocol/masque"
)

// with_masque: register the MASQUE (RFC 9484 / Cloudflare WARP) endpoint + outbound.
func registerMasqueEndpoint(registry *endpoint.Registry) {
	masque.RegisterEndpoint(registry)
}

func registerMasqueOutbound(registry *outbound.Registry) {
	masque.RegisterOutbound(registry)
}
