package constant

const ACMETLS1Protocol = "acme-tls/1"

const (
	// RealityKeyShare* — KeyShare picks the key exchange offered in the REALITY ClientHello
	RealityKeyShareDefault   = ""
	RealityKeyShareHybrid    = "hybrid"
	RealityKeyShareClassical = "classical"
)

const (
	TLSEngineDefault = ""
	TLSEngineGo      = "go"
	TLSEngineApple   = "apple"
	TLSEngineWindows = "windows"
)
