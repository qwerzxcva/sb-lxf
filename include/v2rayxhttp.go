package include

// Blank-import the XHTTP transport so its init() registers the "xhttp" client
// transport constructor with the v2ray registry. In sb-lxf (tsbAI) XHTTP is
// always compiled, so no build tag gate is needed.
import _ "github.com/sagernet/sing-box/transport/v2rayxhttp"
