//go:build js

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/dustin/go-humanize"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/sirupsen/logrus"
)

var (
	ctx    context.Context
	cancel context.CancelFunc

	api *dfs.API
)

func main() {
	registerWasmFunctions()
	ctx, cancel = context.WithCancel(context.Background())
	<-ctx.Done()
}

func registerWasmFunctions() {
	js.Global().Set("connect", js.FuncOf(connect))
	js.Global().Set("stop", js.FuncOf(stop))

	js.Global().Set("connectWallet", js.FuncOf(connectWallet))
	js.Global().Set("login", js.FuncOf(login))
	js.Global().Set("walletLogin", js.FuncOf(walletLogin))
	js.Global().Set("userPresent", js.FuncOf(userPresent))
	js.Global().Set("userIsLoggedIn", js.FuncOf(userIsLoggedIn))
	js.Global().Set("userLogout", js.FuncOf(userLogout))
	js.Global().Set("userDelete", js.FuncOf(userDelete))
	js.Global().Set("userStat", js.FuncOf(userStat))
	js.Global().Set("getNameHash", js.FuncOf(getNameHash))

	js.Global().Set("podNew", js.FuncOf(podNew))
	js.Global().Set("podOpen", js.FuncOf(podOpen))
	js.Global().Set("podClose", js.FuncOf(podClose))
	js.Global().Set("podSync", js.FuncOf(podSync))
	js.Global().Set("podDelete", js.FuncOf(podDelete))
	js.Global().Set("podList", js.FuncOf(podList))
	js.Global().Set("podStat", js.FuncOf(podStat))
	js.Global().Set("podShare", js.FuncOf(podShare))
	js.Global().Set("podReceive", js.FuncOf(podReceive))
	js.Global().Set("podReceiveInfo", js.FuncOf(podReceiveInfo))

	js.Global().Set("listPodInMarketplace", js.FuncOf(listPodInMarketplace))
	js.Global().Set("changePodListStatusInMarketplace", js.FuncOf(changePodListStatusInMarketplace))
	js.Global().Set("requestSubscription", js.FuncOf(requestSubscription))
	js.Global().Set("approveSubscription", js.FuncOf(approveSubscription))
	js.Global().Set("getSubscriptions", js.FuncOf(getSubscriptions))
	js.Global().Set("openSubscribedPod", js.FuncOf(openSubscribedPod))
	js.Global().Set("getSubscribablePods", js.FuncOf(getSubscribablePods))
	js.Global().Set("getSubRequests", js.FuncOf(getSubRequests))
	js.Global().Set("getSubscribablePodInfo", js.FuncOf(getSubscribablePodInfo))
	js.Global().Set("encryptSubscription", js.FuncOf(encryptSubscription))

	js.Global().Set("dirPresent", js.FuncOf(dirPresent))
	js.Global().Set("dirMake", js.FuncOf(dirMake))
	js.Global().Set("dirRemove", js.FuncOf(dirRemove))
	js.Global().Set("dirList", js.FuncOf(dirList))
	js.Global().Set("dirStat", js.FuncOf(dirStat))

	js.Global().Set("fileShare", js.FuncOf(fileShare))
	js.Global().Set("fileReceive", js.FuncOf(fileReceive))
	js.Global().Set("fileReceiveInfo", js.FuncOf(fileReceiveInfo))
	js.Global().Set("fileDelete", js.FuncOf(fileDelete))
	js.Global().Set("fileStat", js.FuncOf(fileStat))
	js.Global().Set("fileUpload", js.FuncOf(fileUpload))
	js.Global().Set("fileDownload", js.FuncOf(fileDownload))

	js.Global().Set("kvNewStore", js.FuncOf(kvNewStore))
	js.Global().Set("kvList", js.FuncOf(kvList))
	js.Global().Set("kvOpen", js.FuncOf(kvOpen))
	js.Global().Set("kvDelete", js.FuncOf(kvDelete))
	js.Global().Set("kvCount", js.FuncOf(kvCount))
	js.Global().Set("kvEntryPut", js.FuncOf(kvEntryPut))
	js.Global().Set("kvEntryGet", js.FuncOf(kvEntryGet))
	js.Global().Set("kvEntryDelete", js.FuncOf(kvEntryDelete))
	js.Global().Set("kvLoadCSV", js.FuncOf(kvLoadCSV))
	js.Global().Set("kvSeek", js.FuncOf(kvSeek))
	js.Global().Set("kvSeekNext", js.FuncOf(kvSeekNext))

	js.Global().Set("docNewStore", js.FuncOf(docNewStore))
	js.Global().Set("docList", js.FuncOf(docList))
	js.Global().Set("docOpen", js.FuncOf(docOpen))
	js.Global().Set("docCount", js.FuncOf(docCount))
	js.Global().Set("docDelete", js.FuncOf(docDelete))
	js.Global().Set("docFind", js.FuncOf(docFind))
	js.Global().Set("docEntryPut", js.FuncOf(docEntryPut))
	js.Global().Set("docEntryGet", js.FuncOf(docEntryGet))
	js.Global().Set("docEntryDelete", js.FuncOf(docEntryDelete))
	js.Global().Set("docLoadJson", js.FuncOf(docLoadJson))
	js.Global().Set("docIndexJson", js.FuncOf(docIndexJson))
}

func connect(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]
		if len(funcArgs) != 6 {
			reject.Invoke("not enough arguments. \"connect(beeEndpoint, stampId, false, rpc, play, subRpc, subContractAddress)\"")
			return nil
		}
		beeEndpoint := funcArgs[0].String()
		stampId := funcArgs[1].String()
		rpc := funcArgs[2].String()
		network := funcArgs[3].String()
		subRpc := funcArgs[4].String()
		subContractAddress := funcArgs[5].String()
		if network != "testnet" && network != "play" {
			reject.Invoke("unknown network. \"use play or testnet\"")
			return nil
		}
		var config *contracts.ENSConfig
		if network == "play" {
			config = contracts.PlayConfig()
		} else {
			config = contracts.TestnetConfig()
		}
		config.ProviderBackend = rpc
		logger := logging.New(os.Stdout, logrus.DebugLevel)

		subConfig := &contracts.SubscriptionConfig{
			RPC:              subRpc,
			SwarmMailAddress: subContractAddress,
		}

		go func() {
			var err error
			api, err = dfs.NewDfsAPI(
				beeEndpoint,
				stampId,
				config,
				subConfig,
				logger,
			)
			if err != nil {
				reject.Invoke(fmt.Sprintf("failed to connect to fairOS: %s", err.Error()))
			}
			resolve.Invoke("connected")
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func stop(js.Value, []js.Value) interface{} {
	cancel()
	return nil
}

func login(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"login(username, password)\"")
			return nil
		}
		username := funcArgs[0].String()
		password := funcArgs[1].String()

		go func() {
			ui, _, _, err := api.LoginUserV2(username, password, "")
			if err != nil {
				reject.Invoke(fmt.Sprintf("Failed to create user : %s", err.Error()))
				return
			}
			object := js.Global().Get("Object").New()
			object.Set("user", ui.GetUserName())
			object.Set("sessionId", ui.GetSessionId())

			resolve.Invoke(object)
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func walletLogin(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"walletLogin(addressHex, signature)\"")
			return nil
		}
		address := funcArgs[0].String()
		signature := funcArgs[1].String()

		go func() {
			ui, err := api.LoginWithWallet(address, signature, "")
			if err != nil {
				reject.Invoke(fmt.Sprintf("Failed to create user : %s", err.Error()))
				return
			}

			object := js.Global().Get("Object").New()
			object.Set("user", ui.GetUserName())
			object.Set("sessionId", ui.GetSessionId())
			resolve.Invoke(object)
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func connectWallet(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 4 {
			reject.Invoke("not enough arguments. \"connectWallet(username, password, walletAddress, signature)\"")
			return nil
		}
		username := funcArgs[0].String()
		password := funcArgs[1].String()
		walletAddress := funcArgs[2].String()
		signature := funcArgs[3].String()

		go func() {
			err := api.ConnectPortableAccountWithWallet(username, password, walletAddress, signature)
			if err != nil {
				reject.Invoke(fmt.Sprintf("Failed to create user : %s", err.Error()))
				return
			}
			resolve.Invoke("wallet connected")
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func userPresent(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 1 {
			reject.Invoke("not enough arguments. \"userPresent(username)\"")
			return nil
		}
		username := funcArgs[0].String()

		go func() {
			present := api.IsUserNameAvailableV2(username)
			object := js.Global().Get("Object").New()
			object.Set("present", present)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func userIsLoggedIn(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 1 {
			reject.Invoke("not enough arguments. \"userIsLoggedIn(username)\"")
			return nil
		}
		username := funcArgs[0].String()

		go func() {
			loggedin := api.IsUserLoggedIn(username)

			object := js.Global().Get("Object").New()
			object.Set("loggedin", loggedin)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func userLogout(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 1 {
			reject.Invoke("not enough arguments. \"userLogout(sessionId)\"")
			return nil
		}
		sessionId := funcArgs[0].String()

		go func() {
			err := api.LogoutUser(sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("userLogout failed : %s", err.Error()))
				return
			}
			resolve.Invoke("user logged out")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func userDelete(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"userDelete(sessionId, password)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		password := funcArgs[1].String()

		go func() {
			err := api.DeleteUserV2(password, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("userDelete failed : %s", err.Error()))
				return
			}
			resolve.Invoke("user deleted")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func userStat(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 1 {
			reject.Invoke("not enough arguments. \"userStat(sessionId)\"")
			return nil
		}
		sessionId := funcArgs[0].String()

		go func() {
			stat, err := api.GetUserStat(sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("userStat failed : %s", err.Error()))
				return
			}

			object := js.Global().Get("Object").New()
			object.Set("userName", stat.Name)
			object.Set("address", stat.Reference)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func podNew(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"podNew(sessionId, podName)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()

		go func() {
			_, err := api.CreatePod(podName, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("podNew failed : %s", err.Error()))
				return
			}
			resolve.Invoke("pod created successfully")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func podOpen(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"podOpen(sessionId, podName)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()

		go func() {
			_, err := api.OpenPod(podName, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("podOpen failed : %s", err.Error()))
				return
			}
			resolve.Invoke("pod opened successfully")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func podClose(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"podOpen(sessionId, podName)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()

		go func() {
			err := api.ClosePod(podName, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("podClose failed : %s", err.Error()))
				return
			}
			resolve.Invoke("pod closed")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func podSync(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"podSync(sessionId, podName)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()

		go func() {
			err := api.SyncPod(podName, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("podSync failed : %s", err.Error()))
				return
			}
			resolve.Invoke("pod sync in progress")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func podDelete(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"podDelete(sessionId, podName)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()

		go func() {
			err := api.DeletePod(podName, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("podDelete failed : %s", err.Error()))
				return
			}
			resolve.Invoke("pod deleted")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func podList(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 1 {
			reject.Invoke("not enough arguments. \"podList(sessionId)\"")
			return nil
		}
		sessionId := funcArgs[0].String()

		go func() {
			ownPods, sharedPods, err := api.ListPods(sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("podList failed : %s", err.Error()))
				return
			}

			object := js.Global().Get("Object").New()
			pods := js.Global().Get("Array").New(len(ownPods))
			for i, v := range ownPods {
				pods.SetIndex(i, js.ValueOf(v))
			}

			sPods := js.Global().Get("Array").New(len(sharedPods))
			for i, v := range sharedPods {
				sPods.SetIndex(i, js.ValueOf(v))
			}

			object.Set("pods", pods)
			object.Set("sharedPods", sPods)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func podStat(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"podStat(sessionId, podName)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()

		go func() {
			stat, err := api.PodStat(podName, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("podStat failed : %s", err.Error()))
				return
			}
			object := js.Global().Get("Object").New()
			object.Set("podName", stat.PodName)
			object.Set("address", stat.PodAddress)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func podShare(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"podShare(sessionId, podName, shareAs)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		shareAs := funcArgs[2].String()

		go func() {
			reference, err := api.PodShare(podName, shareAs, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("podShare failed : %s", err.Error()))
				return
			}
			object := js.Global().Get("Object").New()
			object.Set("podSharingReference", reference)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func podReceive(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"podReceive(sessionId, newPodName, podSharingReference)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		newPodName := funcArgs[1].String()
		podSharingReference := funcArgs[2].String()

		go func() {
			ref, err := utils.ParseHexReference(podSharingReference)
			if err != nil {
				reject.Invoke(fmt.Sprintf("podReceive failed : %s", err.Error()))
				return
			}
			pi, err := api.PodReceive(sessionId, newPodName, ref)
			if err != nil {
				reject.Invoke(fmt.Sprintf("podReceive failed : %s", err.Error()))
				return
			}
			resolve.Invoke(fmt.Sprintf("public pod \"%s\", added as shared pod", pi.GetPodName()))
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func podReceiveInfo(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"podReceiveInfo(sessionId, pod_sharing_reference)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podSharingReference := funcArgs[1].String()

		go func() {
			ref, err := utils.ParseHexReference(podSharingReference)
			if err != nil {
				reject.Invoke(fmt.Sprintf("podReceiveInfo failed : %s", err.Error()))
				return
			}
			shareInfo, err := api.PodReceiveInfo(sessionId, ref)
			if err != nil {
				reject.Invoke(fmt.Sprintf("podReceiveInfo failed : %s", err.Error()))
				return
			}

			object := js.Global().Get("Object").New()
			object.Set("podName", shareInfo.PodName)
			object.Set("podAddress", shareInfo.Address)
			object.Set("password", shareInfo.Password)
			object.Set("userAddress", shareInfo.UserAddress)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func dirPresent(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"dirPresent(sessionId, podName, dirPath)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		dirPath := funcArgs[2].String()

		go func() {
			present, err := api.IsDirPresent(podName, dirPath, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("dirPresent failed : %s", err.Error()))
				return
			}

			object := js.Global().Get("Object").New()
			object.Set("present", present)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func dirMake(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"dirMake(sessionId, podName, dirPath)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		dirPath := funcArgs[2].String()

		go func() {
			err := api.Mkdir(podName, dirPath, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("dirMake failed : %s", err.Error()))
				return
			}
			resolve.Invoke("directory created successfully")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func dirRemove(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"dirRemove(sessionId, podName, dirPath)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		dirPath := funcArgs[2].String()

		go func() {
			err := api.RmDir(podName, dirPath, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("dirRemove failed : %s", err.Error()))
				return
			}
			resolve.Invoke("directory removed successfully")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func dirList(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"dirList(sessionId, podName, dirPath)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		dirPath := funcArgs[2].String()

		go func() {
			dirs, files, err := api.ListDir(podName, dirPath, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("dirList failed : %s", err.Error()))
				return
			}
			filesList := js.Global().Get("Array").New(len(files))
			for i, v := range files {
				file := js.Global().Get("Object").New()
				file.Set("name", v.Name)
				file.Set("contentType", v.ContentType)
				file.Set("size", v.Size)
				file.Set("blockSize", v.BlockSize)
				file.Set("creationTime", v.CreationTime)
				file.Set("modificationTime", v.ModificationTime)
				file.Set("accessTime", v.AccessTime)
				file.Set("mode", v.Mode)
				filesList.SetIndex(i, file)
			}
			dirsList := js.Global().Get("Array").New(len(dirs))
			for i, v := range dirs {
				dir := js.Global().Get("Object").New()
				dir.Set("name", v.Name)
				dir.Set("contentType", v.ContentType)
				dir.Set("size", v.Size)
				dir.Set("mode", v.Mode)
				dir.Set("blockSize", v.BlockSize)
				dir.Set("creationTime", v.CreationTime)
				dir.Set("modificationTime", v.ModificationTime)
				dir.Set("accessTime", v.AccessTime)
				dirsList.SetIndex(i, dir)
			}
			object := js.Global().Get("Object").New()
			object.Set("files", filesList)
			object.Set("dirs", dirsList)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func dirStat(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"dirStat(sessionId, podName, dirPath)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		dirPath := funcArgs[2].String()

		go func() {
			stat, err := api.DirectoryStat(podName, dirPath, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("dirStat failed : %s", err.Error()))
				return
			}
			object := js.Global().Get("Object").New()
			object.Set("podName", stat.PodName)
			object.Set("dirPath", stat.DirPath)
			object.Set("dirName", stat.DirName)
			object.Set("mode", stat.Mode)
			object.Set("creationTime", stat.CreationTime)
			object.Set("modificationTime", stat.ModificationTime)
			object.Set("accessTime", stat.AccessTime)
			object.Set("noOfDirectories", stat.NoOfDirectories)
			object.Set("noOfFiles", stat.NoOfFiles)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func fileDownload(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]
		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"fileDownload(sessionId, podName, filePath)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		filePath := funcArgs[2].String()

		go func() {
			r, _, err := api.DownloadFile(podName, filePath, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("fileDownload failed : %s", err.Error()))
				return
			}
			defer r.Close()

			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(r)
			if err != nil {
				reject.Invoke(fmt.Sprintf("fileDownload failed : %s", err.Error()))
				return
			}
			a := js.Global().Get("Uint8Array").New(buf.Len())
			js.CopyBytesToJS(a, buf.Bytes())
			resolve.Invoke(a)
		}()
		return nil
	})
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func fileUpload(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]
		if len(funcArgs) != 8 {
			reject.Invoke("not enough arguments. \"fileUpload(sessionId, podName, dirPath, file, name, size, blockSize, compression)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		dirPath := funcArgs[2].String()
		array := funcArgs[3]
		fileName := funcArgs[4].String()
		size := funcArgs[5].Int()
		blockSize := funcArgs[6].String()
		compression := funcArgs[7].String()
		if compression != "" {
			if compression != "snappy" && compression != "gzip" {
				reject.Invoke("invalid compression value")
				return nil
			}
		}
		bs, err := humanize.ParseBytes(blockSize)
		if err != nil {
			reject.Invoke("invalid blockSize value")
			return nil
		}

		go func() {
			inBuf := make([]uint8, array.Get("byteLength").Int())
			js.CopyBytesToGo(inBuf, array)
			reader := bytes.NewReader(inBuf)

			err := api.UploadFile(podName, fileName, sessionId, int64(size), reader, dirPath, compression, uint32(bs), true)
			if err != nil {
				reject.Invoke(fmt.Sprintf("fileUpload failed : %s", err.Error()))
				return
			}
			resolve.Invoke("file uploaded")
		}()
		return nil
	})
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func fileShare(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 4 {
			reject.Invoke("not enough arguments. \"fileShare(sessionId, podName, dirPath, destinationUser)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		dirPath := funcArgs[2].String()
		destinationUser := funcArgs[3].String()

		go func() {
			ref, err := api.ShareFile(podName, dirPath, destinationUser, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("fileShare failed : %s", err.Error()))
				return
			}

			object := js.Global().Get("Object").New()
			object.Set("fileSharingReference", ref)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func fileReceive(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 4 {
			reject.Invoke("not enough arguments. \"fileReceive(sessionId, podName, directory, file_sharing_reference)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		directory := funcArgs[2].String()
		fileSharingReference := funcArgs[3].String()

		go func() {
			ref, err := utils.ParseSharingReference(fileSharingReference)
			if err != nil {
				reject.Invoke(fmt.Sprintf("fileReceive failed : %s", err.Error()))
				return
			}
			filePath, err := api.ReceiveFile(podName, sessionId, ref, directory)
			if err != nil {
				reject.Invoke(fmt.Sprintf("fileReceive failed : %s", err.Error()))
				return
			}
			object := js.Global().Get("Object").New()
			object.Set("fileName", filePath)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func fileReceiveInfo(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"fileReceiveInfo(sessionId, fileSharingReference)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		fileSharingReference := funcArgs[2].String()

		go func() {
			ref, err := utils.ParseSharingReference(fileSharingReference)
			if err != nil {
				reject.Invoke(fmt.Sprintf("fileReceiveInfo failed : %s", err.Error()))
				return
			}
			receiveInfo, err := api.ReceiveInfo(sessionId, ref)
			if err != nil {
				reject.Invoke(fmt.Sprintf("fileReceiveInfo failed : %s", err.Error()))
				return
			}
			object := js.Global().Get("Object").New()
			object.Set("name", receiveInfo.FileName)
			object.Set("size", receiveInfo.Size)
			object.Set("blockSize", receiveInfo.BlockSize)
			object.Set("numberOfBlocks", receiveInfo.NumberOfBlocks)
			object.Set("contentType", receiveInfo.ContentType)
			object.Set("compression", receiveInfo.Compression)
			object.Set("sourceAddress", receiveInfo.Sender)
			object.Set("destAddress", receiveInfo.Receiver)
			object.Set("sharedTime", receiveInfo.SharedTime)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func fileDelete(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"fileDelete(sessionId, podName, podFileWithPath)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		filePath := funcArgs[2].String()

		go func() {
			err := api.DeleteFile(podName, filePath, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("fileDelete failed : %s", err.Error()))
				return
			}
			resolve.Invoke("file deleted successfully")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func fileStat(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"fileStat(sessionId, podName, podFileWithPath)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		filePath := funcArgs[2].String()

		go func() {
			stat, err := api.FileStat(podName, filePath, sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("fileStat failed : %s", err.Error()))
				return
			}
			object := js.Global().Get("Object").New()
			object.Set("podName", stat.PodName)
			object.Set("mode", stat.Mode)
			object.Set("filePath", stat.FilePath)
			object.Set("fileName", stat.FileName)
			object.Set("fileSize", stat.FileSize)
			object.Set("blockSize", stat.BlockSize)
			object.Set("compression", stat.Compression)
			object.Set("contentType", stat.ContentType)
			object.Set("creationTime", stat.CreationTime)
			object.Set("modificationTime", stat.ModificationTime)
			object.Set("accessTime", stat.AccessTime)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func kvNewStore(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 4 {
			reject.Invoke("not enough arguments. \"kvNewStore(sessionId, podName, tableName, indexType)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()
		idxType := funcArgs[3].String()
		if idxType == "" {
			idxType = "string"
		}

		var indexType collection.IndexType
		switch idxType {
		case "string":
			indexType = collection.StringIndex
		case "number":
			indexType = collection.NumberIndex
		case "bytes":
		default:
			reject.Invoke("invalid indexType. only string and number are allowed")
			return nil
		}

		go func() {
			err := api.KVCreate(sessionId, podName, tableName, indexType)
			if err != nil {
				reject.Invoke(fmt.Sprintf("kvNewStore failed : %s", err.Error()))
				return
			}
			resolve.Invoke("kv store created")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func kvList(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"kvList(sessionId, podName)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()

		go func() {
			collections, err := api.KVList(sessionId, podName)
			if err != nil {
				reject.Invoke(fmt.Sprintf("kvList failed : %s", err.Error()))
				return
			}
			resp, _ := json.Marshal(collections)
			resolve.Invoke(string(resp))
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func kvOpen(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"kvOpen(sessionId, podName, tableName)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()

		go func() {
			err := api.KVOpen(sessionId, podName, tableName)
			if err != nil {
				reject.Invoke(fmt.Sprintf("kvOpen failed : %s", err.Error()))
				return
			}
			resolve.Invoke("kv store opened")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func kvDelete(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"kvDelete(sessionId, podName, tableName)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()

		go func() {
			err := api.KVDelete(sessionId, podName, tableName)
			if err != nil {
				reject.Invoke(fmt.Sprintf("kvDelete failed : %s", err.Error()))
				return
			}
			resolve.Invoke("kv store deleted")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func kvCount(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"kvCount(sessionId, podName, tableName)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()

		go func() {
			count, err := api.KVCount(sessionId, podName, tableName)
			if err != nil {
				reject.Invoke(fmt.Sprintf("kvCount failed : %s", err.Error()))
				return
			}
			resp, _ := json.Marshal(count)
			resolve.Invoke(resp)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func kvEntryPut(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 5 {
			reject.Invoke("not enough arguments. \"kvEntryPut(sessionId, podName, tableName, key, value)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()
		key := funcArgs[3].String()
		value := funcArgs[4].String()

		go func() {
			err := api.KVPut(sessionId, podName, tableName, key, []byte(value))
			if err != nil {
				reject.Invoke(fmt.Sprintf("kvEntryPut failed : %s", err.Error()))
				return
			}
			resolve.Invoke("key added")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

type KVResponse struct {
	Keys   []string `json:"keys,omitempty"`
	Values []byte   `json:"values"`
}

func kvEntryGet(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 4 {
			reject.Invoke("not enough arguments. \"kvEntryGet(sessionId, podName, tableName, key)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()
		key := funcArgs[3].String()

		go func() {
			columns, data, err := api.KVGet(sessionId, podName, tableName, key)
			if err != nil {
				reject.Invoke(fmt.Sprintf("kvEntryGet failed : %s", err.Error()))
				return
			}
			var res KVResponse
			if columns != nil {
				res.Keys = columns
			} else {
				res.Keys = []string{key}
			}
			res.Values = data
			resp, _ := json.Marshal(res)
			resolve.Invoke(resp)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func kvEntryDelete(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 4 {
			reject.Invoke("not enough arguments. \"kvEntryDelete(sessionId, podName, tableName, key)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()
		key := funcArgs[3].String()

		go func() {
			_, err := api.KVDel(sessionId, podName, tableName, key)
			if err != nil {
				reject.Invoke(fmt.Sprintf("kvEntryDelete failed : %s", err.Error()))
				return
			}
			resolve.Invoke("key deleted")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func kvLoadCSV(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]
		if len(funcArgs) != 5 {
			reject.Invoke("not enough arguments. \"kvLoadCSV(sessionId, podName, tableName, memory, file)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()
		memory := funcArgs[3].Bool()
		array := funcArgs[4]

		go func() {
			inBuf := make([]uint8, array.Get("byteLength").Int())
			js.CopyBytesToGo(inBuf, array)
			r := bytes.NewReader(inBuf)
			reader := bufio.NewReader(r)
			readHeader := false
			rowCount := 0
			successCount := 0
			failureCount := 0
			var batch *collection.Batch
			for {
				// read one row from csv (assuming
				record, err := reader.ReadString('\n')
				if err == io.EOF {
					break
				}
				rowCount++
				if err != nil {
					failureCount++
					continue
				}

				record = strings.TrimSuffix(record, "\n")
				record = strings.TrimSuffix(record, "\r")
				if !readHeader {
					columns := strings.Split(record, ",")
					batch, err = api.KVBatch(sessionId, podName, tableName, columns)
					if err != nil {
						reject.Invoke(fmt.Sprintf("kv loadcsv: %s", err.Error()))
						return
					}

					err = batch.Put(collection.CSVHeaderKey, []byte(record), false, memory)
					if err != nil {
						failureCount++
						readHeader = true
						continue
					}
					readHeader = true
					successCount++
					continue
				}

				key := strings.Split(record, ",")[0]
				err = batch.Put(key, []byte(record), false, memory)
				if err != nil {
					failureCount++
					continue
				}
				successCount++
			}
			_, err := batch.Write("")
			if err != nil {
				reject.Invoke(fmt.Sprintf("kv loadcsv: %s", err.Error()))
				return
			}
			resolve.Invoke(fmt.Sprintf("csv file loaded in to kv table (%s) with total:%d, success: %d, failure: %d rows", tableName, rowCount, successCount, failureCount))
		}()
		return nil
	})
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func kvSeek(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 6 {
			reject.Invoke("not enough arguments. \"kvSeek(sessionId, podName, tableName, start, end, limit)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()
		start := funcArgs[3].String()
		end := funcArgs[4].String()
		limit := funcArgs[5].Int()
		if limit == 0 {
			limit = 10
		}

		go func() {
			_, err := api.KVSeek(sessionId, podName, tableName, start, end, int64(limit))
			if err != nil {
				reject.Invoke(fmt.Sprintf("kvSeek failed : %s", err.Error()))
				return
			}
			resolve.Invoke("seeked closest to the start key")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func kvSeekNext(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"kvSeekNext(sessionId, podName, tableName)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()

		go func() {
			columns, key, data, err := api.KVGetNext(sessionId, podName, tableName)
			if err != nil {
				reject.Invoke(fmt.Sprintf("kvSeekNext failed : %s", err.Error()))
				return
			}
			var res KVResponse
			if columns != nil {
				res.Keys = columns
			} else {
				res.Keys = []string{key}
			}
			res.Values = data
			resp, _ := json.Marshal(res)
			resolve.Invoke(resp)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func docNewStore(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 5 {
			reject.Invoke("not enough arguments. \"docNewStore(sessionId, podName, tableName, simpleIndexes, mutable)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()
		si := funcArgs[3].String()
		mutable := funcArgs[4].Bool()
		indexes := make(map[string]collection.IndexType)
		if si != "" {
			idxs := strings.Split(si, ",")
			for _, idx := range idxs {
				nt := strings.Split(idx, "=")
				if len(nt) != 2 {
					reject.Invoke("invalid argument")
					return nil
				}
				switch nt[1] {
				case "string":
					indexes[nt[0]] = collection.StringIndex
				case "number":
					indexes[nt[0]] = collection.NumberIndex
				case "map":
					indexes[nt[0]] = collection.MapIndex
				case "list":
					indexes[nt[0]] = collection.ListIndex
				case "bytes":
				default:
					reject.Invoke("invalid indexType")
					return nil
				}
			}
		}

		go func() {
			err := api.DocCreate(sessionId, podName, tableName, indexes, mutable)
			if err != nil {
				reject.Invoke(fmt.Sprintf("docNewStore failed : %s", err.Error()))
				return
			}
			resolve.Invoke("doc store created")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func docList(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"docList(sessionId, podName)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()

		go func() {
			collections, err := api.DocList(sessionId, podName)
			if err != nil {
				reject.Invoke(fmt.Sprintf("docList failed : %s", err.Error()))
				return
			}
			resp, _ := json.Marshal(collections)
			resolve.Invoke(string(resp))
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func docOpen(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"docOpen(sessionId, podName, tableName)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()

		go func() {
			err := api.DocOpen(sessionId, podName, tableName)
			if err != nil {
				reject.Invoke(fmt.Sprintf("docOpen failed : %s", err.Error()))
				return
			}
			resolve.Invoke("doc store opened")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func docCount(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 4 {
			reject.Invoke("not enough arguments. \"docCount(sessionId, podName, tableName, expression)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()
		expr := funcArgs[3].String()

		go func() {
			count, err := api.DocCount(sessionId, podName, tableName, expr)
			if err != nil {
				reject.Invoke(fmt.Sprintf("docCount failed : %s", err.Error()))
				return
			}
			resp, _ := json.Marshal(count)
			resolve.Invoke(resp)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func docDelete(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"docDelete(sessionId, podName, tableName)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()

		go func() {
			err := api.DocDelete(sessionId, podName, tableName)
			if err != nil {
				reject.Invoke(fmt.Sprintf("docDelete failed : %s", err.Error()))
				return
			}
			resolve.Invoke("doc store deleted")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func docFind(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 5 {
			reject.Invoke("not enough arguments. \"docFind(sessionId, podName, tableName, expression, limit)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()
		expr := funcArgs[3].String()
		limit := funcArgs[4].Int()

		go func() {
			count, err := api.DocFind(sessionId, podName, tableName, expr, limit)
			if err != nil {
				reject.Invoke(fmt.Sprintf("docCount failed : %s", err.Error()))
				return
			}
			resp, _ := json.Marshal(count)
			resolve.Invoke(resp)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func docEntryPut(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 4 {
			reject.Invoke("not enough arguments. \"docEntryPut(sessionId, podName, tableName, value)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()
		value := funcArgs[3].String()

		go func() {
			err := api.DocPut(sessionId, podName, tableName, []byte(value))
			if err != nil {
				reject.Invoke(fmt.Sprintf("docEntryPut failed : %s", err.Error()))
				return
			}
			resolve.Invoke("added document to db")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

type DocGetResponse struct {
	Doc []byte `json:"doc"`
}

func docEntryGet(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 4 {
			reject.Invoke("not enough arguments. \"docEntryGet(sessionId, podName, tableName, id)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()
		id := funcArgs[3].String()

		go func() {
			data, err := api.DocGet(sessionId, podName, tableName, id)
			if err != nil {
				reject.Invoke(fmt.Sprintf("docEntryGet failed : %s", err.Error()))
				return
			}
			var getResponse DocGetResponse
			getResponse.Doc = data

			resp, _ := json.Marshal(getResponse)
			resolve.Invoke(resp)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func docEntryDelete(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 4 {
			reject.Invoke("not enough arguments. \"docEntryDelete(sessionId, podName, tableName, id)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()
		id := funcArgs[3].String()

		go func() {
			err := api.DocDel(sessionId, podName, tableName, id)
			if err != nil {
				reject.Invoke(fmt.Sprintf("docEntryDelete failed : %s", err.Error()))
				return
			}
			resolve.Invoke("deleted document from db")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func docLoadJson(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]
		if len(funcArgs) != 4 {
			reject.Invoke("not enough arguments. \"docLoadJson(sessionId, podName, tableName, file)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()
		array := funcArgs[3]

		go func() {
			inBuf := make([]uint8, array.Get("byteLength").Int())
			js.CopyBytesToGo(inBuf, array)
			r := bytes.NewReader(inBuf)
			reader := bufio.NewReader(r)

			rowCount := 0
			successCount := 0
			failureCount := 0
			docBatch, err := api.DocBatch(sessionId, podName, tableName)
			for {
				// read one row from csv (assuming
				record, err := reader.ReadString('\n')
				if err == io.EOF {
					break
				}
				rowCount++
				if err != nil {
					failureCount++
					continue
				}

				record = strings.TrimSuffix(record, "\n")
				record = strings.TrimSuffix(record, "\r")

				err = api.DocBatchPut(sessionId, podName, []byte(record), docBatch)
				if err != nil {
					failureCount++
					continue
				}
				successCount++
			}
			err = api.DocBatchWrite(sessionId, podName, docBatch)
			if err != nil {
				reject.Invoke(fmt.Sprintf("doc loadjson: %s", err.Error()))
				return
			}
			resolve.Invoke(fmt.Sprintf("json file loaded in to document db (%s) with total:%d, success: %d, failure: %d rows", tableName, rowCount, successCount, failureCount))
		}()
		return nil
	})
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func docIndexJson(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 4 {
			reject.Invoke("not enough arguments. \"docIndexJson(sessionId, podName, tableName, filePath)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		tableName := funcArgs[2].String()
		filePath := funcArgs[3].String()

		go func() {
			err := api.DocIndexJson(sessionId, podName, tableName, filePath)
			if err != nil {
				reject.Invoke(fmt.Sprintf("docIndexJson failed : %s", err.Error()))
				return
			}
			resolve.Invoke("indexing started")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func listPodInMarketplace(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 7 {
			reject.Invoke("not enough arguments. \"listPodInMarketplace(sessionId, podName, title, desc, thumbnail, price, category)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		title := funcArgs[2].String()
		desc := funcArgs[3].String()
		thumbnail := funcArgs[4].String()
		priceStr := funcArgs[5].String()
		categoryStr := funcArgs[6].String()

		// convert priceStr to uint64
		price, err := strconv.ParseUint(priceStr, 10, 64)
		if err != nil {
			reject.Invoke(fmt.Sprintf("listPodInMarketplace failed : %s", err.Error()))
			return nil
		}

		category, err := utils.Decode(categoryStr)
		if err != nil {
			reject.Invoke(fmt.Sprintf("listPodInMarketplace failed : %s", err.Error()))
			return nil
		}

		var c [32]byte
		copy(c[:], category)
		go func() {
			err := api.ListPodInMarketplace(sessionId, podName, title, desc, thumbnail, price, c)
			if err != nil {
				reject.Invoke(fmt.Sprintf("listPodInMarketplace failed : %s", err.Error()))
				return
			}
			resolve.Invoke("pod listed")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func changePodListStatusInMarketplace(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"changePodListStatusInMarketplace(sessionId, subHash, show)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		subHashStr := funcArgs[1].String()
		show := funcArgs[2].Bool()

		subHash, err := utils.Decode(subHashStr)
		if err != nil {
			reject.Invoke(fmt.Sprintf("changePodListStatusInMarketplace failed : %s", err.Error()))
			return nil
		}

		var s [32]byte
		copy(s[:], subHash)
		go func() {
			err := api.ChangePodListStatusInMarketplace(sessionId, s, show)
			if err != nil {
				reject.Invoke(fmt.Sprintf("listPodInMarketplace failed : %s", err.Error()))
				return
			}
			resolve.Invoke("pod list status changed successfully")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func requestSubscription(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"requestSubscription(sessionId, subHash)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		subHashStr := funcArgs[1].String()

		subHash, err := utils.Decode(subHashStr)
		if err != nil {
			reject.Invoke(fmt.Sprintf("requestSubscription failed : %s", err.Error()))
			return nil
		}

		var s [32]byte
		copy(s[:], subHash)
		go func() {
			err := api.RequestSubscription(sessionId, s)
			if err != nil {
				reject.Invoke(fmt.Sprintf("requestSubscription failed : %s", err.Error()))
				return
			}
			resolve.Invoke("request submitted successfully")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func approveSubscription(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 4 {
			reject.Invoke("not enough arguments. \"approveSubscription(sessionId, podName, reqHash, subscriberNameHash)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		reqHashStr := funcArgs[2].String()
		subscriberNameHashStr := funcArgs[3].String()

		reqHash, err := utils.Decode(reqHashStr)
		if err != nil {
			reject.Invoke(fmt.Sprintf("approveSubscription failed : %s", err.Error()))
			return nil
		}

		var r [32]byte
		copy(r[:], reqHash)

		nameHash, err := utils.Decode(subscriberNameHashStr)
		if err != nil {
			reject.Invoke(fmt.Sprintf("approveSubscription failed : %s", err.Error()))
			return nil
		}

		var nh [32]byte
		copy(nh[:], nameHash)
		go func() {
			err := api.ApproveSubscription(sessionId, podName, r, nh)
			if err != nil {
				reject.Invoke(fmt.Sprintf("approveSubscription failed : %s", err.Error()))
				return
			}
			resolve.Invoke("request approved successfully")
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func encryptSubscription(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"encryptSubscription(sessionId, podName, subscriberNameHash)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		podName := funcArgs[1].String()
		subscriberNameHashStr := funcArgs[2].String()

		nameHash, err := utils.Decode(subscriberNameHashStr)
		if err != nil {
			reject.Invoke(fmt.Sprintf("approveSubscription failed : %s", err.Error()))
			return nil
		}

		var nh [32]byte
		copy(nh[:], nameHash)
		go func() {
			ref, err := api.EncryptSubscription(sessionId, podName, nh)
			if err != nil {
				reject.Invoke(fmt.Sprintf("encryptSubscription failed : %s", err.Error()))
				return
			}
			object := js.Global().Get("Object").New()
			object.Set("reference", ref)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func getSubscriptions(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 3 {
			reject.Invoke("not enough arguments. \"getSubscriptions(sessionId, start, limit)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		start := funcArgs[1].Int()
		limit := funcArgs[2].Int()

		go func() {
			subs, err := api.GetSubscriptions(sessionId, uint64(start), uint64(limit))
			if err != nil {
				reject.Invoke(fmt.Sprintf("getSubscriptions failed : %s", err.Error()))
				return
			}
			object := js.Global().Get("Object").New()
			subscriptions := js.Global().Get("Array").New(len(subs))
			for i, v := range subs {
				subscription := js.Global().Get("Object").New()
				subscription.Set("podName", v.PodName)
				subscription.Set("subHash", utils.Encode(v.SubHash[:]))
				subscription.Set("podAddress", v.PodAddress)
				subscription.Set("validTill", v.ValidTill)
				subscription.Set("infoLocation", utils.Encode(v.InfoLocation))
				subscriptions.SetIndex(i, js.ValueOf(v))
			}
			object.Set("subscriptions", subscriptions)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func openSubscribedPod(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"openSubscribedPod(sessionId, subHash)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		subHashStr := funcArgs[1].String()

		subHash, err := utils.Decode(subHashStr)
		if err != nil {
			reject.Invoke(fmt.Sprintf("openSubscribedPod failed : %s", err.Error()))
			return nil
		}

		var s [32]byte
		copy(s[:], subHash)

		go func() {
			pi, err := api.OpenSubscribedPod(sessionId, s)
			if err != nil {
				reject.Invoke(fmt.Sprintf("openSubscribedPod failed : %s", err.Error()))
				return
			}

			resolve.Invoke(fmt.Sprintf("%s opened successfully", pi.GetPodName()))
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func getSubscribablePods(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 1 {
			reject.Invoke("not enough arguments. \"getSubscribablePods(sessionId)\"")
			return nil
		}
		sessionId := funcArgs[0].String()

		go func() {
			subs, err := api.GetSubscribablePods(sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("getSubscribablePods failed : %s", err.Error()))
				return
			}
			object := js.Global().Get("Object").New()
			subscriptions := js.Global().Get("Array").New(len(subs))
			for i, v := range subs {
				subscription := js.Global().Get("Object").New()
				subscription.Set("subHash", utils.Encode(v.SubHash[:]))
				subscription.Set("sellerNameHash", utils.Encode(v.FdpSellerNameHash[:]))
				subscription.Set("seller", v.Seller.Hex())
				subscription.Set("swarmLocation", utils.Encode(v.SwarmLocation[:]))
				subscription.Set("price", v.Price.Int64())
				subscription.Set("active", v.Active)
				subscription.Set("earned", v.Earned.Int64())
				subscription.Set("bid", v.Bids)
				subscription.Set("sells", v.Sells)
				subscription.Set("reports", v.Reports)
				subscriptions.SetIndex(i, js.ValueOf(v))
			}
			object.Set("subscribablePods", subscriptions)
			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func getSubRequests(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 1 {
			reject.Invoke("not enough arguments. \"getSubRequests(sessionId)\"")
			return nil
		}
		sessionId := funcArgs[0].String()

		go func() {
			requests, err := api.GetSubsRequests(sessionId)
			if err != nil {
				reject.Invoke(fmt.Sprintf("getSubRequests failed : %s", err.Error()))
				return
			}
			object := js.Global().Get("Object").New()
			subRequests := js.Global().Get("Array").New(len(requests))
			for i, v := range requests {
				request := js.Global().Get("Object").New()
				request.Set("subHash", utils.Encode(v.SubHash[:]))
				request.Set("buyerNameHash", utils.Encode(v.FdpBuyerNameHash[:]))
				request.Set("requestHash", utils.Encode(v.RequestHash[:]))
				request.Set("buyer", v.Buyer.Hex())
				subRequests.SetIndex(i, js.ValueOf(v))
			}
			object.Set("requests", subRequests)
			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func getSubscribablePodInfo(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"getSubscribablePodInfo(sessionId, subHash)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		subHashStr := funcArgs[1].String()

		subHash, err := utils.Decode(subHashStr)
		if err != nil {
			reject.Invoke(fmt.Sprintf("getSubscribablePodInfo failed : %s", err.Error()))
			return nil
		}

		var s [32]byte
		copy(s[:], subHash)

		go func() {
			info, err := api.GetSubscribablePodInfo(sessionId, s)
			if err != nil {
				reject.Invoke(fmt.Sprintf("getSubscribablePodInfo failed : %s", err.Error()))
				return
			}
			object := js.Global().Get("Object").New()
			object.Set("category", info.Category)
			object.Set("description", info.Description)
			object.Set("fdpSellerNameHash", info.FdpSellerNameHash)
			object.Set("imageUrl", info.ImageURL)
			object.Set("podAddress", info.PodAddress)
			object.Set("podName", info.PodName)
			object.Set("price", info.Price)
			object.Set("title", info.Title)

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func getNameHash(_ js.Value, funcArgs []js.Value) interface{} {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		if len(funcArgs) != 2 {
			reject.Invoke("not enough arguments. \"getNameHash(sessionId, username)\"")
			return nil
		}
		sessionId := funcArgs[0].String()
		username := funcArgs[1].String()

		go func() {
			nameHash, err := api.GetNameHash(sessionId, username)
			if err != nil {
				reject.Invoke(fmt.Sprintf("getNameHash failed : %s", err.Error()))
				return
			}
			object := js.Global().Get("Object").New()
			object.Set("namehash", utils.Encode(nameHash[:]))

			resolve.Invoke(object)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}
