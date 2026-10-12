//go:build !with_anytls

package include

import (
	"github.com/sagernet/sing-box/adapter/inbound"
	"github.com/sagernet/sing-box/adapter/outbound"
)

// !with_anytls: anytls is excluded from this build (Android-only slim profile).
func registerAnytlsInbound(registry *inbound.Registry) {
}

func registerAnytlsOutbound(registry *outbound.Registry) {
}
