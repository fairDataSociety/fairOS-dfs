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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

func podNew(podName string) {
	password := getPassword()
	newPod := common.PodRequest{
		PodName:  podName,
		Password: password,
	}
	jsonData, err := json.Marshal(newPod)
	if err != nil {
		fmt.Println("create pod: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiPodNew, jsonData)
	if err != nil {
		fmt.Println("could not create pod: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func deletePod(podName string) {
	delPod := common.PodRequest{
		PodName: podName,
	}
	jsonData, err := json.Marshal(delPod)
	if err != nil {
		fmt.Println("delete pod: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodDelete, apiPodDelete, jsonData)
	if err != nil {
		fmt.Println("could not delete pod: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func openPod(podName string) {
	data, err := fdfsAPI.getReq(apiPodLs, "")
	if err != nil {
		fmt.Println("error while listing pods: %w", err)
		return
	}
	var resp api.PodListResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("open pod: ", err)
		return
	}
	invalidPodName := true
	password := ""
	for _, pod := range resp.Pods {
		if pod == podName {
			password = getPassword()
			invalidPodName = false
		}
	}
	for _, pod := range resp.SharedPods {
		if pod == podName {
			invalidPodName = false
		}
	}
	if invalidPodName {
		fmt.Println("invalid pod name")
		return
	}

	openPodReq := common.PodRequest{
		PodName:  podName,
		Password: password,
	}
	jsonData, err := json.Marshal(openPodReq)
	if err != nil {
		fmt.Println("open pod: error marshalling request")
		return
	}
	data, err = fdfsAPI.postReq(http.MethodPost, apiPodOpen, jsonData)
	if err != nil {
		fmt.Println("pod open failed: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func closePod(podName string) {
	newPod := common.PodRequest{
		PodName: podName,
	}
	jsonData, err := json.Marshal(newPod)
	if err != nil {
		fmt.Println("create pod: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiPodClose, jsonData)
	if err != nil {
		fmt.Println("error closing pod: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func syncPod(podName string) {
	newPod := common.PodRequest{
		PodName: podName,
	}
	jsonData, err := json.Marshal(newPod)
	if err != nil {
		fmt.Println("create pod: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiPodSync, jsonData)
	if err != nil {
		fmt.Println("could not sync pod: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func sharePod(podName string) {
	password := getPassword()
	sharePodReq := common.PodRequest{
		PodName:  podName,
		Password: password,
	}
	jsonData, err := json.Marshal(sharePodReq)
	if err != nil {
		fmt.Println("share pod: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiPodShare, jsonData)
	if err != nil {
		fmt.Println("pod share failed: ", err)
		return
	}
	var sharingRef api.PodSharingReference
	err = json.Unmarshal(data, &sharingRef)
	if err != nil {
		fmt.Println("pod share failed: ", err)
		return
	}
	fmt.Println("Pod Sharing Reference : ", sharingRef.Reference)
}

func listPod() {
	data, err := fdfsAPI.getReq(apiPodLs, "")
	if err != nil {
		fmt.Println("error while listing pods: %w", err)
		return
	}
	var resp api.PodListResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("pod list: ", err)
		return
	}
	for _, pod := range resp.Pods {
		fmt.Println("<Pod>: ", pod)
	}
	for _, pod := range resp.SharedPods {
		fmt.Println("<Shared Pod>: ", pod)
	}
}

func podStat(podName string) {
	data, err := fdfsAPI.getReq(apiPodStat, "pod_name="+podName)
	if err != nil {
		fmt.Println("error getting stat: ", err)
		return
	}
	var resp api.PodStatResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("pod stat: ", err)
		return
	}

	fmt.Println("pod Name         : ", resp.PodName)
	fmt.Println("pod Address      : ", resp.PodAddress)
}

func receive(sharingRef string) {
	data, err := fdfsAPI.getReq(apiPodReceive, "sharing_ref="+sharingRef)
	if err != nil {
		fmt.Println("pod receive failed: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func receiveInfo(sharingRef string) {
	data, err := fdfsAPI.getReq(apiPodReceiveInfo, "sharing_ref="+sharingRef)
	if err != nil {
		fmt.Println("pod receive info failed: ", err)
		return
	}
	var podSharingInfo pod.ShareInfo
	err = json.Unmarshal(data, &podSharingInfo)
	if err != nil {
		fmt.Println("pod receive info failed: ", err)
		return
	}
	fmt.Println("Pod Name  : ", podSharingInfo.PodName)
	fmt.Println("Pod Ref.  : ", podSharingInfo.Address)
	fmt.Println("User Name : ", podSharingInfo.UserName)
	fmt.Println("User Ref. : ", podSharingInfo.UserAddress)
	fmt.Println("Shared Time : ", podSharingInfo.SharedTime)
}
