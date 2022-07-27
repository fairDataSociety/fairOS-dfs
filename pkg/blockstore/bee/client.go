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

package bee

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/ethersphere/bee/pkg/swarm"
	bmtlegacy "github.com/ethersphere/bmt/legacy"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	lru "github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

const (
	maxIdleConnections     = 20
	maxConnectionsPerHost  = 256
	requestTimeout         = 6000
	chunkCacheSize         = 1024
	uploadBlockCacheSize   = 100
	downloadBlockCacheSize = 100
	healthUrl              = "/health"
	chunkUploadDownloadUrl = "/chunks"
	bytesUploadDownloadUrl = "/bytes"
	pinsUrl                = "/pins/"
	swarmPinHeader         = "Swarm-Pin"
	swarmEncryptHeader     = "Swarm-Encrypt"
	swarmPostageBatchId    = "Swarm-Postage-Batch-Id"
)

// Client is a bee http client that satisfies blockstore.Client
type Client struct {
	url                string
	client             *http.Client
	hasher             *bmtlegacy.Hasher
	chunkCache         *lru.Cache
	uploadBlockCache   *lru.Cache
	downloadBlockCache *lru.Cache
	postageBlockId     string
	logger             logging.Logger
}

func hashFunc() hash.Hash {
	return sha3.NewLegacyKeccak256()
}

type bytesPostResponse struct {
	Reference swarm.Address `json:"reference"`
}

// NewBeeClient creates a new client which connects to the Swarm bee node to access the Swarm network.
func NewBeeClient(apiUrl, postageBlockId string, logger logging.Logger) *Client {
	p := bmtlegacy.NewTreePool(hashFunc, swarm.Branches, bmtlegacy.PoolSize)
	cache, err := lru.New(chunkCacheSize)
	if err != nil {
		logger.Warningf("could not initialise chunkCache. system will be slow")
	}
	uploadBlockCache, err := lru.New(uploadBlockCacheSize)
	if err != nil {
		logger.Warningf("could not initialise blockCache. system will be slow")
	}
	downloadBlockCache, err := lru.New(downloadBlockCacheSize)
	if err != nil {
		logger.Warningf("could not initialise blockCache. system will be slow")
	}

	return &Client{
		url:                apiUrl,
		client:             createHTTPClient(),
		hasher:             bmtlegacy.New(p),
		chunkCache:         cache,
		uploadBlockCache:   uploadBlockCache,
		downloadBlockCache: downloadBlockCache,
		postageBlockId:     postageBlockId,
		logger:             logger,
	}
}

type chunkAddressResponse struct {
	Reference swarm.Address `json:"reference"`
}

func socResource(owner, id, sig string) string {
	return fmt.Sprintf("/soc/%s/%s?sig=%s", owner, id, sig)
}

// CheckConnection is used to check if the nbe client is up and running.
func (s *Client) CheckConnection(isProxy bool) bool {
	url := s.url
	matchString := "Ethereum Swarm Bee\n"
	if isProxy {
		url += healthUrl
		matchString = "OK"
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}

	response, err := s.client.Do(req)
	if err != nil {
		return false
	}
	defer response.Body.Close()

	req.Close = true

	if response.StatusCode != http.StatusOK {
		return false
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false
	}

	if string(data) != matchString {
		return false
	}
	return true
}

// UploadSOC is used construct and send a Single Owner Chunk to the Swarm bee client.
func (s *Client) UploadSOC(owner, id, signature string, data []byte) (address []byte, err error) {
	to := time.Now()
	socResStr := socResource(owner, id, signature)
	fullUrl := fmt.Sprintf(s.url + socResStr)

	req, err := http.NewRequest(http.MethodPost, fullUrl, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	// the postage block id to store the SOC chunk
	req.Header.Set(swarmPostageBatchId, s.postageBlockId)

	// TODO change this in the future when we have some alternative to pin SOC
	// This is a temporary fix to force soc pinning
	req.Header.Set(swarmPinHeader, "true")

	response, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	req.Close = true

	if response.StatusCode != http.StatusCreated {
		return nil, errors.New("error uploading data")
	}

	addrData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("error downloading data")
	}

	var addrResp *chunkAddressResponse
	err = json.Unmarshal(addrData, &addrResp)
	if err != nil {
		return nil, err
	}

	if s.inChunkCache(addrResp.Reference.String()) {
		s.addToChunkCache(addrResp.Reference.String(), data)
	}
	fields := logrus.Fields{
		"reference": addrResp.Reference.String(),
		"duration":  time.Since(to).String(),
	}
	s.logger.WithFields(fields).Log(logrus.DebugLevel, "upload chunk: ")
	return addrResp.Reference.Bytes(), nil
}

// UploadChunk uploads a chunk to Swarm network.
func (s *Client) UploadChunk(ch swarm.Chunk, pin bool) (address []byte, err error) {
	to := time.Now()
	fullUrl := fmt.Sprintf(s.url + chunkUploadDownloadUrl)
	req, err := http.NewRequest(http.MethodPost, fullUrl, bytes.NewBuffer(ch.Data()))
	if err != nil {
		return nil, err
	}

	if pin {
		req.Header.Set(swarmPinHeader, "true")
	}

	// the postage block id to store the chunk
	req.Header.Set(swarmPostageBatchId, s.postageBlockId)

	response, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	req.Close = true

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("error uploading data")
	}

	addrData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("error downloading data")
	}

	var addrResp *chunkAddressResponse
	err = json.Unmarshal(addrData, &addrResp)
	if err != nil {
		return nil, err
	}

	if s.inChunkCache(ch.Address().String()) {
		s.addToChunkCache(ch.Address().String(), ch.Data())
	}
	fields := logrus.Fields{
		"reference": ch.Address().String(),
		"duration":  time.Since(to).String(),
	}
	s.logger.WithFields(fields).Log(logrus.DebugLevel, "upload chunk: ")

	return addrResp.Reference.Bytes(), nil
}

// DownloadChunk downloads a chunk with given address from the Swarm network
func (s *Client) DownloadChunk(ctx context.Context, address []byte) (data []byte, err error) {
	to := time.Now()
	addrString := swarm.NewAddress(address).String()
	if s.inChunkCache(addrString) {
		return s.getFromChunkCache(swarm.NewAddress(address).String()), nil
	}

	path := filepath.Join(chunkUploadDownloadUrl, addrString)
	fullUrl := fmt.Sprintf(s.url + path)
	req, err := http.NewRequest(http.MethodGet, fullUrl, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	response, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	req.Close = true

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("error downloading data")
	}

	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("error downloading data")
	}

	s.addToChunkCache(addrString, data)
	fields := logrus.Fields{
		"reference": addrString,
		"duration":  time.Since(to).String(),
	}
	s.logger.WithFields(fields).Log(logrus.DebugLevel, "download chunk: ")
	return data, nil
}

// UploadBlob uploads a binary blob of data to Swarm network. It also optionally pins and encrypts the data.
func (s *Client) UploadBlob(data []byte, pin, encrypt bool) (address []byte, err error) {
	to := time.Now()

	// return the ref if this data is already in swarm
	if s.inBlockCache(s.uploadBlockCache, string(data)) {
		return s.getFromBlockCache(s.uploadBlockCache, string(data)), nil
	}

	fullUrl := s.url + bytesUploadDownloadUrl
	req, err := http.NewRequest(http.MethodPost, fullUrl, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	if pin {
		req.Header.Set(swarmPinHeader, "true")
	}

	if encrypt {
		req.Header.Set(swarmEncryptHeader, "true")
	}

	// the postage block id to store the blob
	req.Header.Set(swarmPostageBatchId, s.postageBlockId)

	response, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	req.Close = true

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		return nil, errors.New("error uploading blob")
	}

	respData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("error uploading blob")
	}

	var resp bytesPostResponse
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response")
	}
	fields := logrus.Fields{
		"reference": resp.Reference.String(),
		"size":      len(data),
		"duration":  time.Since(to).String(),
	}
	s.logger.WithFields(fields).Log(logrus.DebugLevel, "upload blob: ")

	// add the data and ref if itis not in cache
	if !s.inBlockCache(s.uploadBlockCache, string(data)) {
		s.addToBlockCache(s.uploadBlockCache, string(data), resp.Reference.Bytes())
	}

	return resp.Reference.Bytes(), nil
}

// DownloadBlob downloads a blob of binary data from the Swarm network.
func (s *Client) DownloadBlob(address []byte) ([]byte, int, error) {
	to := time.Now()

	// return the data if this address is already in cache
	addrString := swarm.NewAddress(address).String()
	if s.inBlockCache(s.downloadBlockCache, addrString) {
		return s.getFromBlockCache(s.downloadBlockCache, addrString), 200, nil
	}

	fullUrl := s.url + bytesUploadDownloadUrl + "/" + addrString
	req, err := http.NewRequest(http.MethodGet, fullUrl, nil)
	if err != nil {
		return nil, http.StatusNotFound, err
	}

	response, err := s.client.Do(req)
	if err != nil {
		return nil, http.StatusNotFound, err
	}
	defer response.Body.Close()

	req.Close = true

	if response.StatusCode != http.StatusOK {
		return nil, response.StatusCode, errors.New("error downloading blob ")
	}

	respData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, response.StatusCode, errors.New("error downloading blob")
	}

	fields := logrus.Fields{
		"reference": addrString,
		"size":      len(respData),
		"duration":  time.Since(to).String(),
	}
	s.logger.WithFields(fields).Log(logrus.DebugLevel, "download blob: ")

	// add the data and ref if it is not in cache
	if !s.inBlockCache(s.downloadBlockCache, addrString) {
		s.addToBlockCache(s.downloadBlockCache, addrString, respData)
	}
	return respData, response.StatusCode, nil
}

// DeleteReference unpins a reference so that it will be garbage collected by the Swarm network.
func (s *Client) DeleteReference(address []byte) error {
	to := time.Now()
	addrString := swarm.NewAddress(address).String()

	fullUrl := s.url + pinsUrl + addrString
	req, err := http.NewRequest(http.MethodDelete, fullUrl, nil)
	if err != nil {
		return err
	}

	response, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	req.Close = true
	if response.StatusCode != http.StatusOK {
		respData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to unpin reference : %s", respData)
	}

	fields := logrus.Fields{
		"reference": addrString,
		"duration":  time.Since(to).String(),
	}
	s.logger.WithFields(fields).Log(logrus.DebugLevel, "delete chunk: ")
	return nil
}

// createHTTPClient for connection re-use
func createHTTPClient() *http.Client {
	client := &http.Client{
		Timeout: time.Second * requestTimeout,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: maxIdleConnections,
			MaxConnsPerHost:     maxConnectionsPerHost,
		},
	}
	return client
}

func (s *Client) addToChunkCache(key string, value []byte) {
	if s.chunkCache != nil {
		s.chunkCache.Add(key, hex.EncodeToString(value))
	}
}

func (s *Client) inChunkCache(key string) bool {
	if s.chunkCache != nil {
		return s.chunkCache.Contains(key)
	}
	return false
}

func (s *Client) getFromChunkCache(key string) []byte {
	if s.chunkCache != nil {
		value, ok := s.chunkCache.Get(key)
		if ok {
			data, err := hex.DecodeString(fmt.Sprintf("%v", value))
			if err != nil {
				return nil
			}
			return data
		}
		return nil
	}
	return nil
}

func (*Client) addToBlockCache(cache *lru.Cache, key string, value []byte) {
	if cache != nil {
		cache.Add(key, value)
	}
}

func (*Client) inBlockCache(cache *lru.Cache, key string) bool {
	if cache != nil {
		return cache.Contains(key)
	}
	return false
}

func (*Client) getFromBlockCache(cache *lru.Cache, key string) []byte {
	if cache != nil {
		value, ok := cache.Get(key)
		if ok {
			return value.([]byte)
		}
		return nil
	}
	return nil
}
