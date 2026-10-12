//go:build with_ssh

package include

import (
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/protocol/ssh"
)

// with_ssh: register the SSH outbound.
func registerSSHOutbound(registry *outbound.Registry) {
	ssh.RegisterOutbound(registry)
}
