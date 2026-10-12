//go:build !with_ssh

package include

import (
	"github.com/sagernet/sing-box/adapter/outbound"
)

// !with_ssh: SSH outbound excluded from this build (Android-only slim profile).
func registerSSHOutbound(registry *outbound.Registry) {
}
