package fairos

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	dfsRoot "github.com/fairdatasociety/fairOS-dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/sirupsen/logrus"
)

var (
	api           *dfs.API
	savedPassword string
	savedUsername string
	sessionId     string
)

func Version() string {
	return dfsRoot.Version
}

// IsConnected checks if dfs.API is already initialised or password is saved or not
func IsConnected() bool {
	if api == nil {
		return false
	}
	if savedPassword == "" {
		return false
	}
	return true
}

// Connect with a bee and initialise dfs.API
func Connect(beeEndpoint, postageBlockId, network, rpc string, logLevel int) error {
	logger := logging.New(os.Stdout, logrus.Level(logLevel))
	var err error
	var ensConfig *contracts.ENSConfig
	switch network {
	case "play":
		ensConfig, _ = contracts.PlayConfig()
	case "testnet":
		ensConfig, _ = contracts.TestnetConfig(contracts.Sepolia)
	case "mainnet":
		return fmt.Errorf("not supported yet")
	default:
		return fmt.Errorf("unknown network")
	}
	ensConfig.ProviderBackend = rpc
	opts := &dfs.Options{
		Stamp:              postageBlockId,
		BeeApiEndpoint:     beeEndpoint,
		EnsConfig:          ensConfig,
		SubscriptionConfig: nil,
		Logger:             logger,
	}
	api, err = dfs.NewDfsAPI(
		context.TODO(),
		opts,
	)
	return err
}

func LoginUser(username, password string) (string, error) {
	loginResp, err := api.LoginUserV2(username, password, "")
	if err != nil {
		return "", err
	}
	ui, nameHash, publicKey := loginResp.UserInfo, loginResp.NameHash, loginResp.PublicKey
	sessionId = ui.GetSessionId()
	savedPassword = password
	savedUsername = username
	data := map[string]string{}
	data["namehash"] = nameHash
	data["publicKey"] = publicKey
	resp, _ := json.Marshal(data)
	return string(resp), nil
}

func IsUserPresent(username string) bool {
	return api.IsUserNameAvailableV2(username)
}

func IsUserLoggedIn() bool {
	return api.IsUserLoggedIn(savedUsername)
}

func LogoutUser() error {
	return api.LogoutUser(sessionId)
}

func StatUser() (string, error) {
	stat, err := api.GetUserStat(sessionId)
	if err != nil {
		return "", err
	}
	resp, _ := json.Marshal(stat)
	return string(resp), nil
}

func NewPod(podName string) (string, error) {
	_, err := api.CreatePod(podName, sessionId)
	if err != nil {
		return "", err
	}
	return "pod created successfully", nil
}

func PodOpen(podName string) (string, error) {
	_, err := api.OpenPod(podName, sessionId)
	if err != nil {
		return "", err
	}
	return "pod created successfully", nil
}

func PodClose(podName string) error {
	return api.ClosePod(podName, sessionId)
}

func PodDelete(podName string) error {
	return api.DeletePod(podName, sessionId)
}

func PodSync(podName string) error {
	return api.SyncPod(podName, sessionId)
}

func PodList() (string, error) {
	ownPods, sharedPods, err := api.ListPods(sessionId)
	if err != nil {
		return "", err
	}
	if ownPods == nil {
		ownPods = []string{}
	}
	if sharedPods == nil {
		sharedPods = []string{}
	}
	data := map[string]interface{}{}
	data["pods"] = ownPods
	data["sharedPods"] = sharedPods
	resp, _ := json.Marshal(data)
	return string(resp), nil
}

func PodStat(podName string) (string, error) {
	stat, err := api.PodStat(podName, sessionId)
	if err != nil {
		return "", err
	}
	resp, _ := json.Marshal(stat)
	return string(resp), nil
}

func IsPodPresent(podName string) bool {
	return api.IsPodExist(podName, sessionId)
}

func PodShare(podName string) (string, error) {
	reference, err := api.PodShare(podName, "", sessionId)
	if err != nil {
		return "", err
	}
	data := map[string]string{}
	data["pod_sharing_reference"] = reference
	resp, _ := json.Marshal(data)
	return string(resp), nil
}

func PodReceive(podSharingReference string) (string, error) {
	ref, err := utils.ParseHexReference(podSharingReference)
	if err != nil {
		return "", err
	}
	pi, err := api.PodReceive(sessionId, "", ref)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("public pod %q, added as shared pod", pi.GetPodName()), nil
}

func PodReceiveInfo(podSharingReference string) (string, error) {
	ref, err := utils.ParseHexReference(podSharingReference)
	if err != nil {
		return "", err
	}
	shareInfo, err := api.PodReceiveInfo(sessionId, ref)
	if err != nil {
		return "", err
	}
	resp, _ := json.Marshal(shareInfo)
	return string(resp), nil
}

func DirPresent(podName, dirPath string) (string, error) {
	present, err := api.IsDirPresent(podName, dirPath, sessionId)
	if err != nil {
		return "", err
	}
	data := map[string]bool{}
	data["present"] = present
	resp, _ := json.Marshal(data)
	return string(resp), nil
}

func DirMake(podName, dirPath string) (string, error) {
	err := api.Mkdir(podName, dirPath, sessionId, 0)
	if err != nil {
		return "", err
	}
	return string("directory created successfully"), nil
}

func DirRemove(podName, dirPath string) (string, error) {
	err := api.RmDir(podName, dirPath, sessionId)
	if err != nil {
		return "", err
	}
	return string("directory removed successfully"), nil
}

func DirList(podName, dirPath string) (string, error) {
	dirs, files, err := api.ListDir(podName, dirPath, sessionId)
	if err != nil {
		return "", err
	}
	var fileList []string
	var dirList []string
	for _, v := range files {
		fileList = append(fileList, v.Name)
	}
	for _, v := range dirs {
		dirList = append(dirList, v.Name)
	}
	data := map[string]interface{}{}
	data["files"] = fileList
	data["dirs"] = dirList
	resp, _ := json.Marshal(data)
	return string(resp), nil
}

func DirStat(podName, dirPath string) (string, error) {
	stat, err := api.DirectoryStat(podName, dirPath, sessionId)
	if err != nil {
		return "", err
	}
	resp, _ := json.Marshal(stat)
	return string(resp), nil
}

func FileShare(podName, dirPath, destinationUser string) (string, error) {
	ref, err := api.ShareFile(podName, dirPath, destinationUser, sessionId)
	if err != nil {
		return "", err
	}
	data := map[string]string{}
	data["file_sharing_reference"] = ref
	resp, _ := json.Marshal(data)
	return string(resp), err
}

func FileReceive(podName, directory, fileSharingReference string) (string, error) {
	filePath, err := api.ReceiveFile(podName, sessionId, fileSharingReference, directory)
	if err != nil {
		return "", err
	}
	data := map[string]string{}
	data["file_name"] = filePath
	resp, _ := json.Marshal(data)
	return string(resp), err
}

func FileReceiveInfo(podName, fileSharingReference string) (string, error) {
	receiveInfo, err := api.ReceiveInfo(sessionId, fileSharingReference)
	if err != nil {
		return "", err
	}
	resp, _ := json.Marshal(receiveInfo)
	return string(resp), err
}

func FileDelete(podName, filePath string) error {
	return api.DeleteFile(podName, filePath, sessionId)
}

func FileStat(podName, filePath string) (string, error) {
	stat, err := api.FileStat(podName, filePath, sessionId)
	if err != nil {
		return "", err
	}
	resp, _ := json.Marshal(stat)
	return string(resp), err
}

func FileUpload(podName, filePath, dirPath, compression, blockSize string, overwrite bool) error {
	fileInfo, err := os.Lstat(filePath)
	if err != nil {
		return err
	}
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	// skipcq: GO-S2307
	defer f.Close()
	bs, err := humanize.ParseBytes(blockSize)
	if err != nil {
		return err
	}
	return api.UploadFile(podName, fileInfo.Name(), sessionId, fileInfo.Size(), f, dirPath, compression, uint32(bs), 0, overwrite)
}

func BlobUpload(data []byte, podName, fileName, dirPath, compression string, size, blockSize int64, overwrite bool) error {
	r := bytes.NewReader(data)
	return api.UploadFile(podName, fileName, sessionId, size, r, dirPath, compression, uint32(blockSize), 0, overwrite)
}

func FileDownload(podName, filePath string) ([]byte, error) {
	r, _, err := api.DownloadFile(podName, filePath, sessionId)
	if err != nil {
		return nil, err
	}
	// skipcq: GO-S2307
	defer r.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func KVNewStore(podName, tableName, indexType string) (string, error) {
	if indexType == "" {
		indexType = "string"
	}

	var idxType collection.IndexType
	switch indexType {
	case "string":
		idxType = collection.StringIndex
	case "number":
		idxType = collection.NumberIndex
	case "bytes":
	default:
		return "", fmt.Errorf("invalid indexType. only string and number are allowed")
	}
	err := api.KVCreate(sessionId, podName, tableName, idxType)
	if err != nil {
		return "", err
	}
	return "kv store created", nil
}

func KVList(podName string) (string, error) {
	collections, err := api.KVList(sessionId, podName)
	if err != nil {
		return "", err
	}
	resp, _ := json.Marshal(collections)
	return string(resp), err
}

func KVOpen(podName, tableName string) error {
	return api.KVOpen(sessionId, podName, tableName)
}

func KVDelete(podName, tableName string) error {
	return api.KVDelete(sessionId, podName, tableName)
}

func KVCount(podName, tableName string) (string, error) {
	count, err := api.KVCount(sessionId, podName, tableName)
	if err != nil {
		return "", err
	}
	resp, _ := json.Marshal(count)
	return string(resp), err
}

func KVEntryPut(podName, tableName, key string, value []byte) error {
	return api.KVPut(sessionId, podName, tableName, key, value)
}

func KVEntryGet(podName, tableName, key string) ([]byte, error) {
	_, data, err := api.KVGet(sessionId, podName, tableName, key)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func KVEntryDelete(podName, tableName, key string) error {
	_, err := api.KVDel(sessionId, podName, tableName, key)
	return err
}

func KVLoadCSV(podName, tableName, filePath, memory string) (string, error) {
	_, err := os.Lstat(filePath)
	if err != nil {
		return "", err
	}
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	// skipcq: GO-S2307
	defer f.Close()
	mem := true
	if memory == "" {
		mem = false
	}
	reader := bufio.NewReader(f)
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
				return "", err
			}

			err = batch.Put(collection.CSVHeaderKey, []byte(record), false, mem)
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
		err = batch.Put(key, []byte(record), false, mem)
		if err != nil {
			failureCount++
			continue
		}
		successCount++
	}
	_, err = batch.Write("")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("csv file loaded in to kv table (%s) with total:%d, success: %d, failure: %d rows", tableName, rowCount, successCount, failureCount), nil
}

func KVSeek(podName, tableName, start, end string, limit int64) error {
	_, err := api.KVSeek(sessionId, podName, tableName, start, end, limit)
	return err
}

func KVSeekNext(podName, tableName string) (string, error) {
	_, key, data, err := api.KVGetNext(sessionId, podName, tableName)
	if err != nil {
		return "", err
	}
	d := map[string]interface{}{}
	d["key"] = key
	d["value"] = data
	resp, _ := json.Marshal(data)
	return string(resp), nil
}

func DocNewStore(podName, tableName, simpleIndexes string, mutable bool) error {
	indexes := make(map[string]collection.IndexType)
	if simpleIndexes != "" {
		idxs := strings.Split(simpleIndexes, ",")
		for _, idx := range idxs {
			nt := strings.Split(idx, "=")
			if len(nt) != 2 {
				return fmt.Errorf("invalid argument")
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
				return fmt.Errorf("invalid indexType")
			}
		}
	}
	return api.DocCreate(sessionId, podName, tableName, indexes, mutable)
}

func DocList(podName string) (string, error) {
	collections, err := api.DocList(sessionId, podName)
	if err != nil {
		return "", err
	}
	resp, _ := json.Marshal(collections)
	return string(resp), err
}

func DocOpen(podName, tableName string) error {
	return api.DocOpen(sessionId, podName, tableName)
}

func DocCount(podName, tableName, expression string) (string, error) {
	count, err := api.DocCount(sessionId, podName, tableName, expression)
	if err != nil {
		return "", err
	}
	resp, _ := json.Marshal(count)
	return string(resp), err
}

func DocDelete(podName, tableName string) error {
	return api.DocDelete(sessionId, podName, tableName)
}

func DocFind(podName, tableName, expression string, limit int) (string, error) {
	count, err := api.DocFind(sessionId, podName, tableName, expression, limit)
	if err != nil {
		return "", err
	}
	resp, _ := json.Marshal(count)
	return string(resp), err
}

func DocEntryPut(podName, tableName, value string) error {
	return api.DocPut(sessionId, podName, tableName, []byte(value))
}

type DocGetResponse struct {
	Doc []byte `json:"doc"`
}

func DocEntryGet(podName, tableName, id string) (string, error) {
	data, err := api.DocGet(sessionId, podName, tableName, id)
	if err != nil {
		return "", err
	}
	var getResponse DocGetResponse
	getResponse.Doc = data

	resp, _ := json.Marshal(getResponse)
	return string(resp), err
}

func DocEntryDelete(podName, tableName, id string) error {
	return api.DocDel(sessionId, podName, tableName, id)
}

func DocLoadJson(podName, tableName, filePath string) (string, error) {
	_, err := os.Lstat(filePath)
	if err != nil {
		return "", err
	}
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	// skipcq: GO-S2307
	defer f.Close()
	reader := bufio.NewReader(f)

	rowCount := 0
	successCount := 0
	failureCount := 0
	docBatch, err := api.DocBatch(sessionId, podName, tableName)
	if err != nil {
		return "", err
	}
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
		return "", err
	}
	return fmt.Sprintf("json file loaded in to document db (%s) with total:%d, success: %d, failure: %d rows", tableName, rowCount, successCount, failureCount), nil
}

func DocIndexJson(podName, tableName, filePath string) error {
	return api.DocIndexJson(sessionId, podName, tableName, filePath)
}
