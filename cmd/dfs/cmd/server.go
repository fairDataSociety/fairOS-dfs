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
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	dfs "github.com/fairdatasociety/fairOS-dfs"

	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	httpPort     string
	cookieDomain string
	corsOrigins  []string
	handler      *api.Handler
)

// startCmd represents the start command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "starts a HTTP server for the dfs",
	Long: `Serves all the dfs commands through an HTTP server so that the upper layers
can consume it.`,
	Run: func(cmd *cobra.Command, args []string) {
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
		logger.Info("version      : ", dfs.Version)
		logger.Info("dataDir      : ", dataDir)
		logger.Info("beeHost      : ", beeHost)
		logger.Info("beePort      : ", beePort)
		logger.Info("verbosity    : ", verbosity)
		logger.Info("httpPort     : ", httpPort)
		logger.Info("cookieDomain : ", cookieDomain)
		logger.Info("corsOrigins  : ", corsOrigins)
		hdlr, err := api.NewHandler(dataDir, beeHost, beePort, cookieDomain, logger)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		handler = hdlr
		startHttpService(logger)
	},
}

func init() {
	serverCmd.Flags().StringVar(&httpPort, "httpPort", "9090", "http port")
	serverCmd.Flags().StringVar(&cookieDomain, "cookieDomain", "api.fairos.io", "the domain to use in the cookie")
	serverCmd.Flags().StringSliceVar(&corsOrigins, "cors-origins", []string{}, "allow CORS headers for the given origins")
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
	})
	router.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, dfs.Version)
		if err != nil {
			logger.Errorf("error in API /version: ", err)
			return
		}
	})

	apiVersion := "v0"
	baseRouter := router.PathPrefix("/" + apiVersion).Subrouter()
	baseRouter.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "User-agent: *\nDisallow: /")
		if err != nil {
			logger.Errorf("error in API /robots.txt: ", err)
			return
		}
	})

	// User account related handlers which does not login need middleware
	baseRouter.Use(handler.LogMiddleware)
	baseRouter.HandleFunc("/user/signup", handler.UserSignupHandler).Methods("POST")
	baseRouter.HandleFunc("/user/login", handler.UserLoginHandler).Methods("POST")
	baseRouter.HandleFunc("/user/import", handler.ImportUserHandler).Methods("POST")
	baseRouter.HandleFunc("/user/present", handler.UserPresentHandler).Methods("GET")
	baseRouter.HandleFunc("/user/isloggedin", handler.IsUserLoggedInHandler).Methods("GET")

	// user account related handlers which require login middleware
	userRouter := baseRouter.PathPrefix("/user/").Subrouter()
	userRouter.Use(handler.LoginMiddleware)
	userRouter.Use(handler.LogMiddleware)
	userRouter.HandleFunc("/logout", handler.UserLogoutHandler).Methods("POST")
	userRouter.HandleFunc("/avatar", handler.SaveUserAvatarHandler).Methods("POST")
	userRouter.HandleFunc("/name", handler.SaveUserNameHandler).Methods("POST")
	userRouter.HandleFunc("/contact", handler.SaveUserContactHandler).Methods("POST")
	userRouter.HandleFunc("/export", handler.ExportUserHandler).Methods("POST")

	userRouter.HandleFunc("/delete", handler.UserDeleteHandler).Methods("DELETE")
	userRouter.HandleFunc("/stat", handler.GetUserStatHandler).Methods("GET")
	userRouter.HandleFunc("/avatar", handler.GetUserAvatarHandler).Methods("GET")
	userRouter.HandleFunc("/name", handler.GetUserNameHandler).Methods("GET")
	userRouter.HandleFunc("/contact", handler.GetUserContactHandler).Methods("GET")
	userRouter.HandleFunc("/share/inbox", handler.GetUserSharingInboxHandler).Methods("GET")
	userRouter.HandleFunc("/share/outbox", handler.GetUserSharingOutboxHandler).Methods("GET")

	// pod related handlers
	podRouter := baseRouter.PathPrefix("/pod/").Subrouter()
	podRouter.Use(handler.LoginMiddleware)
	podRouter.Use(handler.LogMiddleware)
	podRouter.HandleFunc("/new", handler.PodCreateHandler).Methods("POST")
	podRouter.HandleFunc("/open", handler.PodOpenHandler).Methods("POST")
	podRouter.HandleFunc("/close", handler.PodCloseHandler).Methods("POST")
	podRouter.HandleFunc("/sync", handler.PodSyncHandler).Methods("POST")
	podRouter.HandleFunc("/delete", handler.PodDeleteHandler).Methods("DELETE")
	podRouter.HandleFunc("/ls", handler.PodListHandler).Methods("GET")
	podRouter.HandleFunc("/stat", handler.PodStatHandler).Methods("GET")

	// directory related handlers
	dirRouter := baseRouter.PathPrefix("/dir/").Subrouter()
	dirRouter.Use(handler.LoginMiddleware)
	dirRouter.Use(handler.LogMiddleware)
	dirRouter.HandleFunc("/mkdir", handler.DirectoryMkdirHandler).Methods("POST")
	dirRouter.HandleFunc("/rmdir", handler.DirectoryRmdirHandler).Methods("DELETE")
	dirRouter.HandleFunc("/ls", handler.DirectoryLsHandler).Methods("GET")
	dirRouter.HandleFunc("/stat", handler.DirectoryStatHandler).Methods("GET")
	dirRouter.HandleFunc("/present", handler.DirectoryPresentHandler).Methods("GET")

	// file related handlers
	fileRouter := baseRouter.PathPrefix("/file/").Subrouter()
	fileRouter.Use(handler.LoginMiddleware)
	fileRouter.Use(handler.LogMiddleware)
	fileRouter.HandleFunc("/download", handler.FileDownloadHandler).Methods("POST")
	fileRouter.HandleFunc("/upload", handler.FileUploadHandler).Methods("POST")
	fileRouter.HandleFunc("/share", handler.FileShareHandler).Methods("POST")
	fileRouter.HandleFunc("/receive", handler.FileReceiveHandler).Methods("POST")
	fileRouter.HandleFunc("/receiveinfo", handler.FileReceiveInfoHandler).Methods("POST")
	fileRouter.HandleFunc("/delete", handler.FileDeleteHandler).Methods("DELETE")
	fileRouter.HandleFunc("/stat", handler.FileStatHandler).Methods("GET")

	kvRouter := baseRouter.PathPrefix("/kv/").Subrouter()
	kvRouter.Use(handler.LoginMiddleware)
	kvRouter.Use(handler.LogMiddleware)
	kvRouter.HandleFunc("/new", handler.DocCreateHandler).Methods("POST")
	kvRouter.HandleFunc("/ls", handler.KVListHandler).Methods("POST")
	kvRouter.HandleFunc("/open", handler.DocOpenHandler).Methods("POST")
	kvRouter.HandleFunc("/count", handler.KVCountHandler).Methods("POST")
	kvRouter.HandleFunc("/delete", handler.KVDeleteHandler).Methods("DELETE")
	kvRouter.HandleFunc("/entry/put", handler.KVPutHandler).Methods("POST")
	kvRouter.HandleFunc("/entry/get", handler.KVGetHandler).Methods("GET")
	kvRouter.HandleFunc("/entry/del", handler.KVDelHandler).Methods("DELETE")
	kvRouter.HandleFunc("/loadcsv", handler.KVLoadCSVHandler).Methods("POST")
	kvRouter.HandleFunc("/seek", handler.KVSeekHandler).Methods("POST")
	kvRouter.HandleFunc("/seek/next", handler.KVGetNextHandler).Methods("GET")

	docRouter := baseRouter.PathPrefix("/doc/").Subrouter()
	docRouter.Use(handler.LoginMiddleware)
	docRouter.Use(handler.LogMiddleware)
	docRouter.HandleFunc("/new", handler.DocCreateHandler).Methods("POST")
	docRouter.HandleFunc("/ls", handler.DocListHandler).Methods("POST")
	docRouter.HandleFunc("/open", handler.DocOpenHandler).Methods("POST")
	docRouter.HandleFunc("/loadjson", handler.DocLoadJsonHandler).Methods("POST")
	//docRouter.HandleFunc("/entry/put", handler.KVPutHandler).Methods("POST")
	docRouter.HandleFunc("/entry/get", handler.DocGetHandler).Methods("GET")

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
	err := http.ListenAndServe(":"+httpPort, handler)
	if err != nil {
		logger.Errorf("listenAndServe: %w", err)
		return
	}
}
