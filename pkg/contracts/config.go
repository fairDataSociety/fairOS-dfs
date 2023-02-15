package contracts

// Config handles the ENS configuration
type Config struct {
	ChainID               string
	ENSRegistryAddress    string
	FDSRegistrarAddress   string
	PublicResolverAddress string
	ProviderDomain        string
	ProviderBackend       string
}

// TestnetConfig defines the configuration for goerli testnet
func TestnetConfig() *Config {
	return &Config{
		ChainID:               "5",
		ENSRegistryAddress:    "0x42B22483e3c8dF794f351939620572d1a3193c12",
		FDSRegistrarAddress:   "0xF4C9Cd25031E3BB8c5618299bf35b349c1aAb6A9",
		PublicResolverAddress: "0xbfeCC6c32B224F7D0026ac86506Fe40A9607BD14",
		ProviderDomain:        "fds",
	}
}

// PlayConfig defines the configuration for fdp-play
func PlayConfig() *Config {
	return &Config{
		ChainID:               "4020",
		ENSRegistryAddress:    "0xDb56f2e9369E0D7bD191099125a3f6C370F8ed15",
		FDSRegistrarAddress:   "0xA94B7f0465E98609391C623d0560C5720a3f2D33",
		PublicResolverAddress: "0xFC628dd79137395F3C9744e33b1c5DE554D94882",
		ProviderDomain:        "fds",
	}
}
