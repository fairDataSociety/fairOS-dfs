package contracts

type Config struct {
	ENSRegistryAddress    string
	FDSRegistrarAddress   string
	PublicResolverAddress string
	ProviderDomain        string
	ProviderBackend       string
}

func TestnetConfig() *Config {
	return &Config{
		ENSRegistryAddress:    "0xE687f17858382C6FCbAe02b31B0aAB607D396059",
		FDSRegistrarAddress:   "0x3adfB0D6B9662c9F711c2Ab18Cf5D7B0cc369C6B",
		PublicResolverAddress: "0x200C9d891F5b480D6210a252539c473e3Ae4771a",
		ProviderDomain:        "fds",
	}
}

func PlayConfig() *Config {
	return &Config{
		ENSRegistryAddress:    "0x26b4AFb60d6C903165150C6F0AA14F8016bE4aec",
		FDSRegistrarAddress:   "0x630589690929E9cdEFDeF0734717a9eF3Ec7Fcfe",
		PublicResolverAddress: "0xA94B7f0465E98609391C623d0560C5720a3f2D33",
		ProviderDomain:        "fds",
	}
}
