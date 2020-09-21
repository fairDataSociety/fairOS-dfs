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

	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var handler *api.Handler

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
		hdlr, err := api.NewHandler(dataDir, beeHost, beePort, logger)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		handler = hdlr
		startHttpService(logger)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func startHttpService(logger logging.Logger) {
	fs := http.FileServer(http.Dir("build/static"))
	http.Handle("/static/", fs)
	router := mux.NewRouter()

	// Web page handlers
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "FairOS-dfs")
	})

	apiVersion := "v0"
	baseRouter := router.PathPrefix("/" + apiVersion).Subrouter()
	baseRouter.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User-agent: *\nDisallow: /")
	})

	// User account related handlers which does not login need middleware
	baseRouter.Use(handler.LogMiddleware)
	baseRouter.HandleFunc("/user/signup", handler.UserSignupHandler).Methods("POST")
	baseRouter.HandleFunc("/user/login", handler.UserLoginHandler).Methods("POST")
	baseRouter.HandleFunc("/import", handler.ImportUserHandler).Methods("POST")
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

	// Web page handlers
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./build/")))
	http.Handle("/", router)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8000", "http://localhost:9090", "http://fairdrive.org"},
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
