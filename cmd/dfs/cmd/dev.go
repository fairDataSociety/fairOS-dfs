package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/asabya/swarm-blockstore/bee"
	"github.com/asabya/swarm-blockstore/bee/mock"
	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
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
	logger := logging.New(os.Stdout, logrus.DebugLevel)
	mockClient := bee.NewBeeClient(beeUrl, bee.WithStamp(mock.BatchOkStr), bee.WithRedundancy("0"))
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
	batchOk := make([]byte, 32)
	_, _ = rand.Read(batchOk)
	fmt.Println(hex.EncodeToString(batchOk))
	<-done
}
