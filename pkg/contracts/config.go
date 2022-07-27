package contracts

// Config handles the ENS configuration
type Config struct {
	ENSRegistryAddress    string
	FDSRegistrarAddress   string
	PublicResolverAddress string
	ProviderDomain        string
	ProviderBackend       string
}

// TestnetConfig defines the configuration for goerli testnet
func TestnetConfig() *Config {
	return &Config{
		ENSRegistryAddress:    "0x42B22483e3c8dF794f351939620572d1a3193c12",
		FDSRegistrarAddress:   "0xF4C9Cd25031E3BB8c5618299bf35b349c1aAb6A9",
		PublicResolverAddress: "0xbfeCC6c32B224F7D0026ac86506Fe40A9607BD14",
		ProviderDomain:        "fds",
	}
}

// PlayConfig defines the configuration for fdp-play
func PlayConfig() *Config {
	return &Config{
		ENSRegistryAddress:    "0x26b4AFb60d6C903165150C6F0AA14F8016bE4aec",
		FDSRegistrarAddress:   "0x630589690929E9cdEFDeF0734717a9eF3Ec7Fcfe",
		PublicResolverAddress: "0xA94B7f0465E98609391C623d0560C5720a3f2D33",
		ProviderDomain:        "fds",
	}
}
