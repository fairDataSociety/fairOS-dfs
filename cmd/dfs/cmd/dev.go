package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"testing"

	mockpost "github.com/ethersphere/bee/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "run fairOS-dfs in development mode",
	Run: func(cmd *cobra.Command, args []string) {
		startDevServer()
	},
}

func init() {
	rootCmd.AddCommand(devCmd)
}

func startDevServer() {
	storer := mockstorer.New()
	t := &testing.T{}
	fmt.Println(`▓█████▄ ▓█████ ██▒   █▓
▒██▀ ██▌▓█   ▀▓██░   █▒
░██   █▌▒███   ▓██  █▒░
░▓█▄   ▌▒▓█  ▄  ▒██ █░░
░▒████▓ ░▒████▒  ▒▀█░  
 ▒▒▓  ▒ ░░ ▒░ ░  ░ ▐░  
 ░ ▒  ▒  ░ ░  ░  ░ ░░  
 ░ ░  ░    ░       ░░  
   ░       ░  ░     ░  
 ░                 ░   `)
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer:          storer,
		PreventRedirect: true,
		Post:            mockpost.New(mockpost.WithAcceptAll()),
	})
	fmt.Println("Bee running at: ", beeUrl)
	logger := logging.New(io.Discard, logrus.DebugLevel)
	mockClient := bee.NewBeeClient(beeUrl, mock.BatchOkStr, true, logger)
	ens := mock2.NewMockNamespaceManager()

	users := user.NewUsers(mockClient, ens, -1, 0, logger)
	dfsApi := dfs.NewMockDfsAPI(mockClient, users, logger)
	handler = api.NewMockHandler(dfsApi, logger, []string{"http://localhost:3000"})
	defer handler.Close()
	httpPort = ":9090"
	pprofPort = ":9091"
	srv := startHttpService(logger)
	fmt.Printf("Server running at:http://127.0.0.1%s\n", httpPort)
	defer func() {
		err := srv.Shutdown(context.TODO())
		if err != nil {
			logger.Error("failed to shutdown server", err.Error())
		}
	}()
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-done:
	}
}
