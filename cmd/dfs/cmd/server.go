/*
Copyright © 2020 FairOS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	dfs "github.com/fairdatasociety/fairOS-dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	httpPort       string
	pprofPort      string
	cookieDomain   string
	postageBlockId string
	corsOrigins    []string
	handler        *api.Handler
)

// startCmd represents the start command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "starts a HTTP server for the dfs",
	Long: `Serves all the dfs commands through an HTTP server so that the upper layers
can consume it.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := config.BindPFlag(optionDFSHttpPort, cmd.Flags().Lookup("httpPort")); err != nil {
			return err
		}
		if err := config.BindPFlag(optionDFSPprofPort, cmd.Flags().Lookup("pprofPort")); err != nil {
			return err
		}
		if err := config.BindPFlag(optionCookieDomain, cmd.Flags().Lookup("cookieDomain")); err != nil {
			return err
		}
		if err := config.BindPFlag(optionCORSAllowedOrigins, cmd.Flags().Lookup("cors-origins")); err != nil {
			return err
		}
		if err := config.BindPFlag(optionNetwork, cmd.Flags().Lookup("network")); err != nil {
			return err
		}
		if err := config.BindPFlag(optionRPC, cmd.Flags().Lookup("rpc")); err != nil {
			return err
		}
		return config.BindPFlag(optionBeePostageBatchId, cmd.Flags().Lookup("postageBlockId"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		httpPort = config.GetString(optionDFSHttpPort)
		pprofPort = config.GetString(optionDFSPprofPort)
		cookieDomain = config.GetString(optionCookieDomain)
		postageBlockId = config.GetString(optionBeePostageBatchId)
		corsOrigins = config.GetStringSlice(optionCORSAllowedOrigins)
		verbosity = config.GetString(optionVerbosity)
		isGatewayProxy := config.GetBool(optionIsGatewayProxy)
		if !isGatewayProxy {
			if postageBlockId == "" {
				_ = cmd.Help()
				fmt.Println("\npostageBlockId is required to run server")
				return
			} else if len(postageBlockId) != 64 {
				fmt.Println("\npostageBlockId is invalid")
				return
			}
			_, err := hex.DecodeString(postageBlockId)
			if err != nil {
				fmt.Println("\npostageBlockId is invalid")
				return
			}
		}
		ensConfig := &contracts.Config{}
		network := config.GetString("network")
		rpc := config.GetString(optionRPC)
		if rpc == "" {
			fmt.Println("\nrpc endpoint is missing")
			return
		}
		if network != "testnet" && network != "mainnet" && network != "play" {
			if network != "" {
				fmt.Println("\nunknown network")
				return
			}
			network = "custom"
			providerDomain := config.GetString(optionProviderDomain)
			publicResolverAddress := config.GetString(optionPublicResolverAddress)
			fdsRegistrarAddress := config.GetString(optionFDSRegistrarAddress)
			ensRegistryAddress := config.GetString(optionENSRegistryAddress)

			if providerDomain == "" {
				fmt.Println("\nens provider domain is missing")
				return
			}
			if publicResolverAddress == "" {
				fmt.Println("\npublicResolver contract address is missing")
				return
			}
			if fdsRegistrarAddress == "" {
				fmt.Println("\nfdsRegistrar contract address is missing")
				return
			}
			if ensRegistryAddress == "" {
				fmt.Println("\nensRegistry contract address is missing")
				return
			}

			ensConfig = &contracts.Config{
				ENSRegistryAddress:    ensRegistryAddress,
				FDSRegistrarAddress:   fdsRegistrarAddress,
				PublicResolverAddress: publicResolverAddress,
				ProviderDomain:        providerDomain,
			}
		} else {
			switch v := strings.ToLower(network); v {
			case "mainnet":
				fmt.Println("\nens is not available for mainnet yet")
				return
			case "testnet":
				ensConfig = contracts.TestnetConfig()
			case "play":
				ensConfig = contracts.PlayConfig()
			}
		}
		ensConfig.ProviderBackend = rpc
		var logger logging.Logger
		switch v := strings.ToLower(verbosity); v {
		case "0", "silent":
			logger = logging.New(ioutil.Discard, 0)
		case "1", "error":
			logger = logging.New(cmd.OutOrStdout(), logrus.ErrorLevel)
		case "2", "warn":
			logger = logging.New(cmd.OutOrStdout(), logrus.WarnLevel)
		case "3", "info":
			logger = logging.New(cmd.OutOrStdout(), logrus.InfoLevel)
		case "4", "debug":
			logger = logging.New(cmd.OutOrStdout(), logrus.DebugLevel)
		case "5", "trace":
			logger = logging.New(cmd.OutOrStdout(), logrus.TraceLevel)
		default:
			fmt.Println("unknown verbosity level ", v)
			return
		}

		logger.Info("configuration values")
		logger.Info("version        : ", dfs.Version)
		logger.Info("network        : ", network)
		logger.Info("beeApi         : ", beeApi)
		logger.Info("isGatewayProxy : ", isGatewayProxy)
		logger.Info("verbosity      : ", verbosity)
		logger.Info("httpPort       : ", httpPort)
		logger.Info("pprofPort      : ", pprofPort)
		logger.Info("cookieDomain   : ", cookieDomain)
		logger.Info("postageBlockId : ", postageBlockId)
		logger.Info("corsOrigins    : ", corsOrigins)

		// datadir will be removed in some future version. it is kept for migration purpose only
		hdlr, err := api.NewHandler(dataDir, beeApi, cookieDomain, postageBlockId, corsOrigins, isGatewayProxy, ensConfig, logger)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		handler = hdlr
		startHttpService(logger)
	},
}

func init() {
	serverCmd.Flags().String("httpPort", defaultDFSHttpPort, "http port")
	serverCmd.Flags().String("pprofPort", defaultDFSPprofPort, "pprof port")
	serverCmd.Flags().String("cookieDomain", defaultCookieDomain, "the domain to use in the cookie")
	serverCmd.Flags().String("postageBlockId", "", "the postage block used to store the data in bee")
	serverCmd.Flags().StringSlice("cors-origins", defaultCORSAllowedOrigins, "allow CORS headers for the given origins")
	serverCmd.Flags().String("network", "", "network to use for authentication (mainnet/testnet/play)")
	serverCmd.Flags().String("rpc", "", "rpc endpoint for ens network. xDai for mainnet | Goerli for testnet | local fdp-play rpc endpoint for play")
	rootCmd.AddCommand(serverCmd)
}

func startHttpService(logger logging.Logger) {
	router := mux.NewRouter()

	// Web page handlers
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "FairOS-dfs")
		if err != nil {
			logger.Errorf("error in API /: ", err)
			return
		}
		_, err = fmt.Fprintln(w, dfs.Version)
		if err != nil {
			logger.Errorf("error in API /: ", err)
			return
		}
		_, err = fmt.Fprintln(w, beeApi)
		if err != nil {
			logger.Errorf("error in API /: ", err)
			return
		}
		_, err = fmt.Fprintln(w, verbosity)
		if err != nil {
			logger.Errorf("error in API /: ", err)
			return
		}
		_, err = fmt.Fprintln(w, httpPort)
		if err != nil {
			logger.Errorf("error in API /: ", err)
			return
		}
		_, err = fmt.Fprintln(w, pprofPort)
		if err != nil {
			logger.Errorf("error in API /: ", err)
			return
		}
		_, err = fmt.Fprintln(w, cookieDomain)
		if err != nil {
			logger.Errorf("error in API /: ", err)
			return
		}
		_, err = fmt.Fprintln(w, corsOrigins)
		if err != nil {
			logger.Errorf("error in API /: ", err)
			return
		}
	})
	router.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, dfs.Version)
		if err != nil {
			logger.Errorf("error in API /version: ", err)
			return
		}
	})
	apiVersion := "v1"

	// v2 introduces user credentials storage on secondary location and identity storage on ens registry
	apiVersionV2 := "v2"

	wsRouter := router.PathPrefix("/ws/" + apiVersion).Subrouter()
	wsRouter.HandleFunc("/", handler.WebsocketHandler)

	baseRouter := router.PathPrefix("/" + apiVersion).Subrouter()
	baseRouter.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "User-agent: *\nDisallow: /")
		if err != nil {
			logger.Errorf("error in API /robots.txt: ", err)
			return
		}
	})
	baseRouterV2 := router.PathPrefix("/" + apiVersionV2).Subrouter()
	baseRouterV2.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "User-agent: *\nDisallow: /")
		if err != nil {
			logger.Errorf("error in API /robots.txt: ", err)
			return
		}
	})

	// User account related handlers which does not need login middleware
	baseRouterV2.Use(handler.LogMiddleware)
	baseRouterV2.HandleFunc("/user/signup", handler.UserSignupV2Handler).Methods("POST")
	baseRouterV2.HandleFunc("/user/login", handler.UserLoginV2Handler).Methods("POST")
	baseRouterV2.HandleFunc("/user/present", handler.UserPresentV2Handler).Methods("GET")
	userRouterV2 := baseRouterV2.PathPrefix("/user/").Subrouter()
	userRouterV2.Use(handler.LoginMiddleware)
	userRouterV2.HandleFunc("/delete", handler.UserDeleteV2Handler).Methods("DELETE")
	userRouterV2.HandleFunc("/migrate", handler.UserMigrateHandler).Methods("POST")

	baseRouter.Use(handler.LogMiddleware)
	// TODO remove signup before merging into master. this is kept for testing purpose only
	baseRouter.HandleFunc("/user/signup", handler.UserSignupHandler).Methods("POST")
	baseRouter.HandleFunc("/user/login", handler.UserLoginHandler).Methods("POST")
	baseRouter.HandleFunc("/user/present", handler.UserPresentHandler).Methods("GET")
	baseRouter.HandleFunc("/user/isloggedin", handler.IsUserLoggedInHandler).Methods("GET")

	// user account related handlers which require login middleware
	userRouter := baseRouter.PathPrefix("/user/").Subrouter()
	userRouter.Use(handler.LoginMiddleware)
	userRouter.HandleFunc("/logout", handler.UserLogoutHandler).Methods("POST")
	userRouter.HandleFunc("/export", handler.ExportUserHandler).Methods("POST")
	userRouter.HandleFunc("/delete", handler.UserDeleteHandler).Methods("DELETE")
	userRouter.HandleFunc("/stat", handler.UserStatHandler).Methods("GET")

	// pod related handlers
	baseRouter.HandleFunc("/pod/receive", handler.PodReceiveHandler).Methods("GET")
	baseRouter.HandleFunc("/pod/receiveinfo", handler.PodReceiveInfoHandler).Methods("GET")

	podRouter := baseRouter.PathPrefix("/pod/").Subrouter()
	podRouter.Use(handler.LoginMiddleware)
	podRouter.HandleFunc("/present", handler.PodPresentHandler).Methods("GET")
	podRouter.HandleFunc("/new", handler.PodCreateHandler).Methods("POST")
	podRouter.HandleFunc("/open", handler.PodOpenHandler).Methods("POST")
	podRouter.HandleFunc("/close", handler.PodCloseHandler).Methods("POST")
	podRouter.HandleFunc("/sync", handler.PodSyncHandler).Methods("POST")
	podRouter.HandleFunc("/share", handler.PodShareHandler).Methods("POST")
	podRouter.HandleFunc("/delete", handler.PodDeleteHandler).Methods("DELETE")
	podRouter.HandleFunc("/ls", handler.PodListHandler).Methods("GET")
	podRouter.HandleFunc("/stat", handler.PodStatHandler).Methods("GET")

	// directory related handlers
	dirRouter := baseRouter.PathPrefix("/dir/").Subrouter()
	dirRouter.Use(handler.LoginMiddleware)
	dirRouter.HandleFunc("/mkdir", handler.DirectoryMkdirHandler).Methods("POST")
	dirRouter.HandleFunc("/rmdir", handler.DirectoryRmdirHandler).Methods("DELETE")
	dirRouter.HandleFunc("/ls", handler.DirectoryLsHandler).Methods("GET")
	dirRouter.HandleFunc("/stat", handler.DirectoryStatHandler).Methods("GET")
	dirRouter.HandleFunc("/present", handler.DirectoryPresentHandler).Methods("GET")

	// file related handlers
	fileRouter := baseRouter.PathPrefix("/file/").Subrouter()
	fileRouter.Use(handler.LoginMiddleware)
	fileRouter.HandleFunc("/download", handler.FileDownloadHandler).Methods("GET")
	fileRouter.HandleFunc("/download", handler.FileDownloadHandler).Methods("POST")
	fileRouter.HandleFunc("/upload", handler.FileUploadHandler).Methods("POST")
	fileRouter.HandleFunc("/share", handler.FileShareHandler).Methods("POST")
	fileRouter.HandleFunc("/receive", handler.FileReceiveHandler).Methods("GET")
	fileRouter.HandleFunc("/receiveinfo", handler.FileReceiveInfoHandler).Methods("GET")
	fileRouter.HandleFunc("/delete", handler.FileDeleteHandler).Methods("DELETE")
	fileRouter.HandleFunc("/stat", handler.FileStatHandler).Methods("GET")

	kvRouter := baseRouter.PathPrefix("/kv/").Subrouter()
	kvRouter.Use(handler.LoginMiddleware)

	kvRouter.HandleFunc("/new", handler.KVCreateHandler).Methods("POST")
	kvRouter.HandleFunc("/ls", handler.KVListHandler).Methods("GET")
	kvRouter.HandleFunc("/open", handler.KVOpenHandler).Methods("POST")
	kvRouter.HandleFunc("/count", handler.KVCountHandler).Methods("POST")
	kvRouter.HandleFunc("/delete", handler.KVDeleteHandler).Methods("DELETE")
	kvRouter.HandleFunc("/present", handler.KVPresentHandler).Methods("GET")
	kvRouter.HandleFunc("/entry/put", handler.KVPutHandler).Methods("POST")
	kvRouter.HandleFunc("/entry/get", handler.KVGetHandler).Methods("GET")
	kvRouter.HandleFunc("/entry/get-data", handler.KVGetDataHandler).Methods("GET")
	kvRouter.HandleFunc("/entry/del", handler.KVDelHandler).Methods("DELETE")
	kvRouter.HandleFunc("/loadcsv", handler.KVLoadCSVHandler).Methods("POST")
	kvRouter.HandleFunc("/export", handler.KVExportHandler).Methods("POST")
	kvRouter.HandleFunc("/seek", handler.KVSeekHandler).Methods("POST")
	kvRouter.HandleFunc("/seek/next", handler.KVGetNextHandler).Methods("GET")

	docRouter := baseRouter.PathPrefix("/doc/").Subrouter()
	docRouter.Use(handler.LoginMiddleware)
	docRouter.HandleFunc("/new", handler.DocCreateHandler).Methods("POST")
	docRouter.HandleFunc("/ls", handler.DocListHandler).Methods("GET")
	docRouter.HandleFunc("/open", handler.DocOpenHandler).Methods("POST")
	docRouter.HandleFunc("/count", handler.DocCountHandler).Methods("POST")
	docRouter.HandleFunc("/delete", handler.DocDeleteHandler).Methods("DELETE")
	docRouter.HandleFunc("/find", handler.DocFindHandler).Methods("GET")
	docRouter.HandleFunc("/loadjson", handler.DocLoadJsonHandler).Methods("POST")
	docRouter.HandleFunc("/indexjson", handler.DocIndexJsonHandler).Methods("POST")
	docRouter.HandleFunc("/entry/put", handler.DocPutHandler).Methods("POST")
	docRouter.HandleFunc("/entry/get", handler.DocGetHandler).Methods("GET")
	docRouter.HandleFunc("/entry/del", handler.DocDelHandler).Methods("DELETE")

	var origins []string
	for _, c := range corsOrigins {
		c = strings.TrimSpace(c)
		origins = append(origins, c)
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"Origin", "Accept", "Authorization", "Content-Type", "X-Requested-With", "Access-Control-Request-Headers", "Access-Control-Request-Method"},
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
		MaxAge:           3600,
	})

	// Insert the middleware
	handler := c.Handler(router)

	// starting the pprof server
	go func() {
		logger.Infof("fairOS-dfs pprof listening on port: %v", pprofPort)
		err := http.ListenAndServe("localhost"+pprofPort, nil)
		if err != nil {
			logger.Errorf("pprof listenAndServe: %v ", err.Error())
			return
		}
	}()

	logger.Infof("fairOS-dfs API server listening on port: %v", httpPort)
	err := http.ListenAndServe(httpPort, handler)
	if err != nil {
		logger.Errorf("http listenAndServe: %v ", err.Error())
		return
	}
}
