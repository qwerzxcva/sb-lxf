package libbox

const (
	CommandLog int32 = iota
	CommandStatus
	CommandGroup
	CommandClashMode
	CommandConnections
	CommandOutbounds
	// lx: SPEC 018 — multiplexed structured DNS-resolution events to the libbox
	// command channel (mirrors CommandConnections' event stream).
	CommandDNS
)
