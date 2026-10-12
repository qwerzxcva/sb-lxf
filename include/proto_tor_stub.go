//go:build !with_tor

package include

import (
	"github.com/sagernet/sing-box/adapter/outbound"
)

// !with_tor: Tor outbound excluded from this build (Android-only slim profile).
func registerTorOutbound(registry *outbound.Registry) {
}
