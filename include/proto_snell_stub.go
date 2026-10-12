//go:build !with_snell

package include

import (
	"github.com/sagernet/sing-box/adapter/inbound"
	"github.com/sagernet/sing-box/adapter/outbound"
)

// !with_snell: snell excluded from this build (Android-only slim profile).
func registerSnellInbound(registry *inbound.Registry) {
}

func registerSnellOutbound(registry *outbound.Registry) {
}
