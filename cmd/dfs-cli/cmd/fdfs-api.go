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
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
)

const (
	MaxIdleConnections int = 20
)

type FdfsClient struct {
	url    string
	client *http.Client
	cookie *http.Cookie
}

func NewFdfsClient(host, port string) (*FdfsClient, error) {
	client, err := createHTTPClient()
	if err != nil {
		return nil, err
	}
	return &FdfsClient{
		url:    fmt.Sprintf("http://" + host + ":" + port),
		client: client,
	}, nil
}

func createHTTPClient() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil { // error handling
		return nil, err
	}
	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: MaxIdleConnections,
		},
	}
	return client, nil
}

func (s *FdfsClient) CheckConnection() bool {
	req, err := http.NewRequest(http.MethodGet, s.url, nil)
	if err != nil {
		return false
	}

	response, err := s.client.Do(req)
	if err != nil {
		return false
	}

	if response.StatusCode != http.StatusOK {
		return false
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false
	}
	if string(data) != "FairOS-dfs\n" {
		return false
	}
	return true
}

func (s *FdfsClient) callFdfsApi(method, urlPath string, arguments map[string]string) ([]byte, error) {
	// prepare the  request
	fullUrl := fmt.Sprintf(s.url + urlPath)
	var req *http.Request
	var err error
	if arguments != nil {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		for k, v := range arguments {
			err := writer.WriteField(k, v)
			if err != nil {
				return nil, err
			}
		}
		err = writer.Close()
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, fullUrl, body)
		if err != nil {
			return nil, err
		}
		// add the headers

		contentType := fmt.Sprintf("multipart/form-data;boundary=%v", writer.Boundary())
		req.Header.Add("Content-Type", contentType)
		req.Header.Add("Content-Length", strconv.Itoa(len(body.Bytes())))
	} else {
		req, err = http.NewRequest(method, fullUrl, nil)
		if err != nil {
			return nil, err
		}
	}

	if s.cookie != nil {
		req.AddCookie(s.cookie)
	}

	// execute the request
	response, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		errStr := fmt.Sprintf("received invalid status: %s", response.Status)
		return nil, errors.New(errStr)
	}

	if len(response.Cookies()) > 0 {
		s.cookie = response.Cookies()[0]
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("error downloading data")
	}

	return data, nil
}

func (s *FdfsClient) uploadMultipartFile(urlPath, fileName string, fileSize int64, fd *os.File, arguments map[string]string, formFileArgument, compression string) ([]byte, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add parameters
	for k, v := range arguments {
		err := writer.WriteField(k, v)
		if err != nil {
			return nil, err
		}
	}

	part, err := writer.CreateFormFile(formFileArgument, fileName)
	if err != nil {
		return nil, err
	}
	n, err := io.Copy(part, fd)
	if err != nil {
		return nil, err
	}

	if n != fileSize {
		return nil, fmt.Errorf("partial write")
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	fullUrl := fmt.Sprintf(s.url + urlPath)
	req, err := http.NewRequest("POST", fullUrl, body)
	if err != nil {
		return nil, err
	}

	contentType := fmt.Sprintf("multipart/form-data;boundary=%v", writer.Boundary())
	req.Header.Set("Content-Type", contentType)
	if strings.ToLower(compression) == "true" {
		req.Header.Set(api.CompressionHeader, "true")
	}

	// execute the request
	response, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		errStr := fmt.Sprintf("received invalid status: %v", response.StatusCode)
		return nil, errors.New(errStr)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("error downloading data")
	}

	return data, nil

}

func (s *FdfsClient) downloadMultipartFile(method, urlPath string, arguments map[string]string, out *os.File) (int64, error) {
	// prepare the  request
	fullUrl := fmt.Sprintf(s.url + urlPath)
	var req *http.Request
	var err error

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	for k, v := range arguments {
		err := writer.WriteField(k, v)
		if err != nil {
			return 0, err
		}
	}
	err = writer.Close()
	if err != nil {
		return 0, err
	}
	req, err = http.NewRequest(method, fullUrl, body)
	if err != nil {
		return 0, err
	}
	// add the headers

	contentType := fmt.Sprintf("multipart/form-data;boundary=%v", writer.Boundary())
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Content-Length", strconv.Itoa(len(body.Bytes())))

	// execute the request
	response, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		errStr := fmt.Sprintf("received invalid status: %s", response.Status)
		return 0, errors.New(errStr)
	}

	// Write the body to file
	return io.Copy(out, response.Body)

}
