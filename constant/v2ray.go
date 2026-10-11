package constant

const (
	V2RayTransportTypeHTTP        = "http"
	V2RayTransportTypeWebsocket   = "ws"
	V2RayTransportTypeQUIC        = "quic"
	V2RayTransportTypeGRPC        = "grpc"
	V2RayTransportTypeHTTPUpgrade = "httpupgrade"
	// lx: XHTTP transport (Xray-compatible, xmux + range), always compiled in sb-lxf
	V2RayTransportTypeXHTTP = "xhttp"
)
