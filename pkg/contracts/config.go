package contracts

const (
	Sepolia = "11155111"
	Goerli  = "5"
)

// ENSConfig handles the ENS configuration
type ENSConfig struct {
	ChainID               string
	ENSRegistryAddress    string
	FDSRegistrarAddress   string
	PublicResolverAddress string
	ProviderDomain        string
	ProviderBackend       string
}

// SubscriptionConfig handles the Subscription Management
type SubscriptionConfig struct {
	RPC            string
	DataHubAddress string
}

// TestnetConfig defines the configuration for goerli testnet
func TestnetConfig(chainId string) (*ENSConfig, *SubscriptionConfig) {
	e := &ENSConfig{
		ChainID:        chainId,
		ProviderDomain: "fds",
	}
	s := &SubscriptionConfig{}
	switch chainId {
	case Sepolia:
		e.ENSRegistryAddress = "0x42a96D45d787685ac4b36292d218B106Fb39be7F"
		e.FDSRegistrarAddress = "0xFBF00389140C00384d88d458239833E3231a7414"
		e.PublicResolverAddress = "0xC904989B579c2B216A75723688C784038AA99B56"

		s.DataHubAddress = "0xbF38b92a9baE1e23e150A66c7A44412828210371"
	case Goerli:
		e.ENSRegistryAddress = "0x42B22483e3c8dF794f351939620572d1a3193c12"
		e.FDSRegistrarAddress = "0xF4C9Cd25031E3BB8c5618299bf35b349c1aAb6A9"
		e.PublicResolverAddress = "0xbfeCC6c32B224F7D0026ac86506Fe40A9607BD14"

		s.DataHubAddress = "0x982d3A3516E08763DEf73485e5762bdBbD932Ce9"
	}

	return e, s
}

// PlayConfig defines the configuration for fdp-play
func PlayConfig() (*ENSConfig, *SubscriptionConfig) {
	return &ENSConfig{
		ChainID:               "4020",
		ENSRegistryAddress:    "0xDb56f2e9369E0D7bD191099125a3f6C370F8ed15",
		FDSRegistrarAddress:   "0xA94B7f0465E98609391C623d0560C5720a3f2D33",
		PublicResolverAddress: "0xFC628dd79137395F3C9744e33b1c5DE554D94882",
		ProviderDomain:        "fds",
	}, nil
}
