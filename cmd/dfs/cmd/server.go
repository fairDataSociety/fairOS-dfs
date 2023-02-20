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
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	dfs "github.com/fairdatasociety/fairOS-dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	_ "github.com/fairdatasociety/fairOS-dfs/swagger"
	docs "github.com/fairdatasociety/fairOS-dfs/swagger"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	httpSwagger "github.com/swaggo/http-swagger"
)

var (
	pprof          bool
	swag           bool
	httpPort       string
	pprofPort      string
	cookieDomain   string
	postageBlockId string
	corsOrigins    []string
	handler        *api.Handler
)

const (
	zeroBatchId = "0000000000000000000000000000000000000000000000000000000000000000"
)

// @title           FairOS-dfs server
// @version         v0.0.0
// @description     A list of the currently provided Interfaces to interact with FairOS decentralised file system(dfs), implementing user, pod, file system, key value store and document store
// @contact.name	Sabyasachi Patra
// @contact.email	sabyasachi@datafund.io
// @license.name  	Apache 2.0
// @license.url   	http://www.apache.org/licenses/LICENSE-2.0.html
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
	RunE: func(cmd *cobra.Command, args []string) error {

		docs.SwaggerInfo.Host = "fairos.dev.fairdatasociety.org"
		docs.SwaggerInfo.Schemes = []string{"https"}
		docs.SwaggerInfo.Version = dfs.Version

		httpPort = config.GetString(optionDFSHttpPort)
		pprofPort = config.GetString(optionDFSPprofPort)
		cookieDomain = config.GetString(optionCookieDomain)
		postageBlockId = config.GetString(optionBeePostageBatchId)
		corsOrigins = config.GetStringSlice(optionCORSAllowedOrigins)
		verbosity = config.GetString(optionVerbosity)

		if postageBlockId == "" {
			_ = cmd.Help()
			fmt.Println("\npostageBlockId is required to run server")
			return fmt.Errorf("postageBlockId is required to run server")
		} else if postageBlockId != zeroBatchId && postageBlockId != "0" {
			if len(postageBlockId) != 64 {
				fmt.Println("\npostageBlockId is invalid")
				return fmt.Errorf("postageBlockId is invalid")
			}
			_, err := hex.DecodeString(postageBlockId)
			if err != nil {
				fmt.Println("\npostageBlockId is invalid")
				return fmt.Errorf("postageBlockId is invalid")
			}
		}
		ensConfig := &contracts.Config{}
		network := config.GetString("network")
		rpc := config.GetString(optionRPC)
		if rpc == "" {
			fmt.Println("\nrpc endpoint is missing")
			return fmt.Errorf("rpc endpoint is missing")
		}
		if network != "testnet" && network != "mainnet" && network != "play" {
			if network != "" {
				fmt.Println("\nunknown network")
				return fmt.Errorf("unknown network")
			}
			network = "custom"
			providerDomain := config.GetString(optionProviderDomain)
			publicResolverAddress := config.GetString(optionPublicResolverAddress)
			fdsRegistrarAddress := config.GetString(optionFDSRegistrarAddress)
			ensRegistryAddress := config.GetString(optionENSRegistryAddress)

			if providerDomain == "" {
				fmt.Println("\nens provider domain is missing")
				return fmt.Errorf("ens provider domain is missing")
			}
			if publicResolverAddress == "" {
				fmt.Println("\npublicResolver contract address is missing")
				return fmt.Errorf("publicResolver contract address is missing")
			}
			if fdsRegistrarAddress == "" {
				fmt.Println("\nfdsRegistrar contract address is missing")
				return fmt.Errorf("fdsRegistrar contract address is missing")
			}
			if ensRegistryAddress == "" {
				fmt.Println("\nensRegistry contract address is missing")
				return fmt.Errorf("ensRegistry contract address is missing")
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
				return fmt.Errorf("ens is not available for mainnet yet")
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
			logger = logging.New(io.Discard, 0)
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
			fmt.Println("unknown verbosity level", v)
			return fmt.Errorf("unknown verbosity level")
		}

		logger.Info("configuration values")
		logger.Info("version        : ", dfs.Version)
		logger.Info("network        : ", network)
		logger.Info("beeApi         : ", beeApi)
		logger.Info("verbosity      : ", verbosity)
		logger.Info("httpPort       : ", httpPort)
		logger.Info("pprofPort      : ", pprofPort)
		logger.Info("cookieDomain   : ", cookieDomain)
		logger.Info("postageBlockId : ", postageBlockId)
		logger.Info("corsOrigins    : ", corsOrigins)

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()
		// datadir will be removed in some future version. it is kept for migration purpose only
		hdlr, err := api.New(ctx, beeApi, cookieDomain, postageBlockId, corsOrigins, ensConfig, logger)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
		defer hdlr.Close()
		handler = hdlr
		if pprof {
			go startPprofService(logger)
		}

		srv := startHttpService(logger)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Minute)
		defer func() {
			err = srv.Shutdown(shutdownCtx)
			if err != nil {
				logger.Error("failed to shutdown server", err.Error())
			}
			shutdownCancel()
		}()

		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
		case <-done:
		}
		return nil
	},
}

func init() {
	serverCmd.Flags().BoolVar(&pprof, "pprof", false, "should run pprof")
	serverCmd.Flags().BoolVar(&swag, "swag", false, "should run swagger-ui")
	serverCmd.Flags().String("httpPort", defaultDFSHttpPort, "http port")
	serverCmd.Flags().String("pprofPort", defaultDFSPprofPort, "pprof port")
	serverCmd.Flags().String("cookieDomain", defaultCookieDomain, "the domain to use in the cookie")
	serverCmd.Flags().String("postageBlockId", "", "the postage block used to store the data in bee")
	serverCmd.Flags().StringSlice("cors-origins", defaultCORSAllowedOrigins, "allow CORS headers for the given origins")
	serverCmd.Flags().String("network", "", "network to use for authentication (mainnet/testnet/play)")
	serverCmd.Flags().String("rpc", "", "rpc endpoint for ens network. xDai for mainnet | Goerli for testnet | local fdp-play rpc endpoint for play")
	rootCmd.AddCommand(serverCmd)
}

func startHttpService(logger logging.Logger) *http.Server {
	router := mux.NewRouter()

	// Web page handlers
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "OK")
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
	if swag {
		router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
			httpSwagger.URL("./swagger/doc.json"),
		)).Methods(http.MethodGet)
	}

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
	podRouter := baseRouter.PathPrefix("/pod/").Subrouter()
	podRouter.Use(handler.LoginMiddleware)
	podRouter.HandleFunc("/present", handler.PodPresentHandler).Methods("GET")
	podRouter.HandleFunc("/new", handler.PodCreateHandler).Methods("POST")
	podRouter.HandleFunc("/open", handler.PodOpenHandler).Methods("POST")
	podRouter.HandleFunc("/open-async", handler.PodOpenAsyncHandler).Methods("POST")
	podRouter.HandleFunc("/close", handler.PodCloseHandler).Methods("POST")
	podRouter.HandleFunc("/sync", handler.PodSyncHandler).Methods("POST")
	podRouter.HandleFunc("/sync-async", handler.PodSyncAsyncHandler).Methods("POST")
	podRouter.HandleFunc("/share", handler.PodShareHandler).Methods("POST")
	podRouter.HandleFunc("/delete", handler.PodDeleteHandler).Methods("DELETE")
	podRouter.HandleFunc("/ls", handler.PodListHandler).Methods("GET")
	podRouter.HandleFunc("/stat", handler.PodStatHandler).Methods("GET")
	podRouter.HandleFunc("/receive", handler.PodReceiveHandler).Methods("GET")
	podRouter.HandleFunc("/receiveinfo", handler.PodReceiveInfoHandler).Methods("GET")
	podRouter.HandleFunc("/fork", handler.PodForkHandler).Methods("POST")
	podRouter.HandleFunc("/fork-from-reference", handler.PodForkFromReferenceHandler).Methods("POST")

	// directory related handlers
	dirRouter := baseRouter.PathPrefix("/dir/").Subrouter()
	dirRouter.Use(handler.LoginMiddleware)
	dirRouter.HandleFunc("/mkdir", handler.DirectoryMkdirHandler).Methods("POST")
	dirRouter.HandleFunc("/rmdir", handler.DirectoryRmdirHandler).Methods("DELETE")
	dirRouter.HandleFunc("/ls", handler.DirectoryLsHandler).Methods("GET")
	dirRouter.HandleFunc("/stat", handler.DirectoryStatHandler).Methods("GET")
	dirRouter.HandleFunc("/chmod", handler.DirectoryModeHandler).Methods("POST")
	dirRouter.HandleFunc("/present", handler.DirectoryPresentHandler).Methods("GET")
	dirRouter.HandleFunc("/rename", handler.DirectoryRenameHandler).Methods("POST")

	// file related handlers
	fileRouter := baseRouter.PathPrefix("/file/").Subrouter()
	fileRouter.Use(handler.LoginMiddleware)
	fileRouter.HandleFunc("/status", handler.FileStatusHandler).Methods("GET")
	fileRouter.HandleFunc("/download", handler.FileDownloadHandlerGet).Methods("GET")
	fileRouter.HandleFunc("/download", handler.FileDownloadHandlerPost).Methods("POST")
	fileRouter.HandleFunc("/update", handler.FileUpdateHandler).Methods("POST")
	fileRouter.HandleFunc("/upload", handler.FileUploadHandler).Methods("POST")
	fileRouter.HandleFunc("/share", handler.FileShareHandler).Methods("POST")
	fileRouter.HandleFunc("/receive", handler.FileReceiveHandler).Methods("GET")
	fileRouter.HandleFunc("/receiveinfo", handler.FileReceiveInfoHandler).Methods("GET")
	fileRouter.HandleFunc("/delete", handler.FileDeleteHandler).Methods("DELETE")
	fileRouter.HandleFunc("/stat", handler.FileStatHandler).Methods("GET")
	fileRouter.HandleFunc("/chmod", handler.FileModeHandler).Methods("POST")
	fileRouter.HandleFunc("/rename", handler.FileRenameHandler).Methods("POST")

	kvRouter := baseRouter.PathPrefix("/kv/").Subrouter()
	kvRouter.Use(handler.LoginMiddleware)

	kvRouter.HandleFunc("/new", handler.KVCreateHandler).Methods("POST")
	kvRouter.HandleFunc("/ls", handler.KVListHandler).Methods("GET")
	kvRouter.HandleFunc("/open", handler.KVOpenHandler).Methods("POST")
	kvRouter.HandleFunc("/count", handler.KVCountHandler).Methods("POST")
	kvRouter.HandleFunc("/delete", handler.KVDeleteHandler).Methods("DELETE")
	kvRouter.HandleFunc("/entry/present", handler.KVPresentHandler).Methods("GET")
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
	docRouter.HandleFunc("/entry/put", handler.DocEntryPutHandler).Methods("POST")
	docRouter.HandleFunc("/entry/get", handler.DocEntryGetHandler).Methods("GET")
	docRouter.HandleFunc("/entry/del", handler.DocEntryDelHandler).Methods("DELETE")

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

	logger.Infof("fairOS-dfs API server listening on port: %v", httpPort)
	srv := &http.Server{
		Addr:    httpPort,
		Handler: handler,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			logger.Errorf("http listenAndServe: %v ", err.Error())
			return
		}
	}()

	return srv
}

func startPprofService(logger logging.Logger) {
	logger.Infof("fairOS-dfs pprof listening on port: %v", pprofPort)
	err := http.ListenAndServe("localhost"+pprofPort, nil)
	if err != nil {
		logger.Errorf("pprof listenAndServe: %v ", err.Error())
		return
	}
}
