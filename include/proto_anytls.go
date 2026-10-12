//go:build with_anytls

package include

import (
	"github.com/sagernet/sing-box/adapter/inbound"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/protocol/anytls"
)

// with_anytls: register the anytls (AnyTLS) inbound + outbound.
func registerAnytlsInbound(registry *inbound.Registry) {
	anytls.RegisterInbound(registry)
}

func registerAnytlsOutbound(registry *outbound.Registry) {
	anytls.RegisterOutbound(registry)
}
