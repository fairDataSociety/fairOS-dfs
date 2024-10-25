package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
)

func actNew(actName, publicKey string) {
	url := fmt.Sprintf("%s/%s?grantee=%s", actGrantee, actName, publicKey)
	data, err := fdfsAPI.postReq(http.MethodPost, url, nil)
	if err != nil {
		fmt.Println("could not create act: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func actGrantRevoke(actName, key, action string) {
	url := fmt.Sprintf("%s/%s", actGrantee, actName)
	switch action {
	case "grant":
		url = fmt.Sprintf("%s?grant=%s", url, key)
	case "revoke":
		url = fmt.Sprintf("%s?revoke=%s", url, key)
	default:
		fmt.Println("invalid action")
	}
	data, err := fdfsAPI.postReq(http.MethodPatch, url, nil)
	if err != nil {
		fmt.Println("could not create act: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func actListGrantees(actName string) {
	url := fmt.Sprintf("%s/%s", actGrantee, actName)
	data, err := fdfsAPI.postReq(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("could not create act: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func actPodShare(actName, podName string) {
	url := fmt.Sprintf("%s/%s/%s", actSharePod, actName, podName)
	data, err := fdfsAPI.postReq(http.MethodPost, url, nil)
	if err != nil {
		fmt.Println("could not create act: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func actListAll() {
	data, err := fdfsAPI.postReq(http.MethodGet, actList, nil)
	if err != nil {
		fmt.Println("could not create act: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func actPodsShared(actName string) {
	url := fmt.Sprintf("%s/%s", actSharedPods, actName)
	data, err := fdfsAPI.postReq(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("could not create act: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func actSaveSharedPod(actName, reference, topic, owner, ownerPublicKey string) {
	url := fmt.Sprintf("%s/%s", actSavePod, actName)
	topicBytes, err := base64.StdEncoding.DecodeString(topic)
	if err != nil {
		fmt.Println("could not save act: ", err)
		return
	}
	content := &api.Content{
		Reference:      reference,
		Topic:          topicBytes,
		Owner:          owner,
		OwnerPublicKey: ownerPublicKey,
	}
	data, err := json.Marshal(content)
	if err != nil {
		fmt.Println("could not save act: ", err)
		return
	}
	resp, err := fdfsAPI.postReq(http.MethodPost, url, data)
	if err != nil {
		fmt.Println("could not create act: ", err)
		return
	}
	message := strings.ReplaceAll(string(resp), "\n", "")
	fmt.Println(message)
}

func actOpenSharedPod(actName string) {
	url := fmt.Sprintf("%s/%s", actOpenPod, actName)
	data, err := fdfsAPI.postReq(http.MethodPost, url, nil)
	if err != nil {
		fmt.Println("could not open act: ", err)
		return
	}
	currentPod = "podone"
	currentDirectory = utils.PathSeparator
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}
