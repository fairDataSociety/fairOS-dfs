package cmd

import (
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	optionCORSAllowedOrigins    = "cors-allowed-origins"
	optionDFSHttpPort           = "dfs.ports.http-port"
	optionDFSPprofPort          = "dfs.ports.pprof-port"
	optionVerbosity             = "verbosity"
	optionBeeApi                = "bee.bee-api-endpoint"
	optionBeePostageBatchId     = "bee.postage-batch-id"
	optionCookieDomain          = "cookie-domain"
	optionNetwork               = "ens-network"
	optionRPC                   = "rpc"
	optionProviderDomain        = "ens.provider-domain"
	optionPublicResolverAddress = "ens.public-resolver-address"
	optionFDSRegistrarAddress   = "ens.fds-registrar-address"
	optionENSRegistryAddress    = "ens.ens-registry-address"

	defaultCORSAllowedOrigins = []string{}
	defaultDFSHttpPort        = ":9090"
	defaultDFSPprofPort       = ":9091"
	defaultVerbosity          = "trace"
	defaultBeeApi             = "http://localhost:1633"
	defaultCookieDomain       = "api.fairos.io"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print default or provided configuration in yaml format",
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		if len(args) > 0 {
			return cmd.Help()
		}

		d := config.AllSettings()
		ym, err := yaml.Marshal(d)
		if err != nil { // skipcq: TCV-001
			return err
		}
		cmd.Println(string(ym))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
