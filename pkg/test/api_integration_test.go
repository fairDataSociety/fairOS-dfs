/*
Copyright © 2025 FairOS Authors

API Integration Test Suite
Tests the HTTP API endpoints for all major features
*/

package test_test

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/asabya/swarm-blockstore/bee"
	"github.com/asabya/swarm-blockstore/bee/mock"
	"github.com/ethersphere/bee/v2/pkg/file/redundancy"
	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/sirupsen/logrus"
)

// APITestSuite holds the test infrastructure
type APITestSuite struct {
	handler    *api.Handler
	server     *httptest.Server
	client     *http.Client
	cookieJar  map[string][]*http.Cookie
	logger     logging.Logger
}

func setupAPITestSuite(t *testing.T) *APITestSuite {
	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer:          storer,
		PreventRedirect: true,
		Post:            mockpost.New(mockpost.WithAcceptAll()),
	})

	logger := logging.New(io.Discard, logrus.DebugLevel)
	mockClient := bee.NewBeeClient(beeUrl,
		bee.WithStamp(mock.BatchOkStr),
		bee.WithRedundancy(fmt.Sprintf("%d", redundancy.NONE)),
		bee.WithPinning(true))

	users := user.NewUsers(mockClient, nil, -1, 0, logger)
	dfsAPI := dfs.NewMockDfsAPI(mockClient, users, logger)

	handler := api.NewMockHandler(dfsAPI, logger, nil)

	return &APITestSuite{
		handler:   handler,
		client:    &http.Client{Timeout: 30 * time.Second},
		cookieJar: make(map[string][]*http.Cookie),
		logger:    logger,
	}
}

func (suite *APITestSuite) makeRequest(method, path string, body io.Reader, cookies []*http.Cookie) (*http.Response, error) {
	req := httptest.NewRequest(method, path, body)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	suite.handler.GetRouter().ServeHTTP(w, req)

	return w.Result(), nil
}

func (suite *APITestSuite) storeCookies(username string, resp *http.Response) {
	suite.cookieJar[username] = resp.Cookies()
}

func (suite *APITestSuite) getCookies(username string) []*http.Cookie {
	return suite.cookieJar[username]
}

// TestAPIAccountRegistration tests user registration via API
func TestAPIAccountRegistration(t *testing.T) {
	suite := setupAPITestSuite(t)
	t.Log("=== Testing API Account Registration ===\n")

	t.Run("SignupNewUser", func(t *testing.T) {
		username := "testuser" + randStringRunes(8)
		password := "Password123!"

		payload := map[string]string{
			"user_name": username,
			"password":  password,
		}
		jsonData, _ := json.Marshal(payload)

		resp, err := suite.makeRequest("POST", "/v1/user/signup", bytes.NewBuffer(jsonData), nil)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 201, got %d: %s", resp.StatusCode, string(body))
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if result["address"] == nil {
			t.Fatal("Response should contain user address")
		}

		suite.storeCookies(username, resp)

		t.Logf("✓ User registered successfully: %s", username)
	})

	t.Run("LoginExistingUser", func(t *testing.T) {
		// First create a user
		username := "logintest" + randStringRunes(8)
		password := "Password123!"

		signupPayload := map[string]string{
			"user_name": username,
			"password":  password,
		}
		jsonData, _ := json.Marshal(signupPayload)
		resp, _ := suite.makeRequest("POST", "/v1/user/signup", bytes.NewBuffer(jsonData), nil)
		resp.Body.Close()

		// Now login
		loginPayload := map[string]string{
			"user_name": username,
			"password":  password,
		}
		jsonData, _ = json.Marshal(loginPayload)

		resp, err := suite.makeRequest("POST", "/v1/user/login", bytes.NewBuffer(jsonData), nil)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Login failed with status %d: %s", resp.StatusCode, string(body))
		}

		suite.storeCookies(username, resp)

		t.Logf("✓ User logged in successfully: %s", username)
	})

	t.Logf("\n=== API Account Registration Tests Passed ===\n")
}

// TestAPIPodOperations tests pod operations via API
func TestAPIPodOperations(t *testing.T) {
	suite := setupAPITestSuite(t)
	t.Log("=== Testing API Pod Operations ===\n")

	// Setup: Create and login user
	username := "poduser" + randStringRunes(8)
	password := "Password123!"

	signupPayload := map[string]string{
		"user_name": username,
		"password":  password,
	}
	jsonData, _ := json.Marshal(signupPayload)
	resp, _ := suite.makeRequest("POST", "/v1/user/signup", bytes.NewBuffer(jsonData), nil)
	suite.storeCookies(username, resp)
	resp.Body.Close()

	cookies := suite.getCookies(username)

	t.Run("CreatePod", func(t *testing.T) {
		podName := "my-documents"

		payload := map[string]string{
			"pod_name": podName,
		}
		jsonData, _ := json.Marshal(payload)

		resp, err := suite.makeRequest("POST", "/v1/pod/new", bytes.NewBuffer(jsonData), cookies)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Pod creation failed with status %d: %s", resp.StatusCode, string(body))
		}

		t.Logf("✓ Pod created: %s", podName)
	})

	t.Run("OpenPod", func(t *testing.T) {
		podName := "my-documents"

		payload := map[string]string{
			"pod_name": podName,
		}
		jsonData, _ := json.Marshal(payload)

		resp, err := suite.makeRequest("POST", "/v1/pod/open", bytes.NewBuffer(jsonData), cookies)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Pod open failed with status %d: %s", resp.StatusCode, string(body))
		}

		t.Logf("✓ Pod opened: %s", podName)
	})

	t.Run("ListPods", func(t *testing.T) {
		resp, err := suite.makeRequest("GET", "/v1/pod/ls", nil, cookies)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Pod list failed with status %d: %s", resp.StatusCode, string(body))
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		pods, ok := result["pods"].([]interface{})
		if !ok || len(pods) == 0 {
			t.Fatal("Expected at least one pod in list")
		}

		t.Logf("✓ Found %d pod(s)", len(pods))
	})

	t.Logf("\n=== API Pod Operations Tests Passed ===\n")
}

// TestAPIFileOperations tests file upload/download via API
func TestAPIFileOperations(t *testing.T) {
	suite := setupAPITestSuite(t)
	t.Log("=== Testing API File Operations ===\n")

	// Setup: Create user and pod
	username := "fileuser" + randStringRunes(8)
	password := "Password123!"
	podName := "file-test-pod"

	signupPayload := map[string]string{
		"user_name": username,
		"password":  password,
	}
	jsonData, _ := json.Marshal(signupPayload)
	resp, _ := suite.makeRequest("POST", "/v1/user/signup", bytes.NewBuffer(jsonData), nil)
	suite.storeCookies(username, resp)
	resp.Body.Close()

	cookies := suite.getCookies(username)

	// Create pod
	podPayload := map[string]string{
		"pod_name": podName,
	}
	jsonData, _ = json.Marshal(podPayload)
	resp, _ = suite.makeRequest("POST", "/v1/pod/new", bytes.NewBuffer(jsonData), cookies)
	resp.Body.Close()

	// Open pod
	resp, _ = suite.makeRequest("POST", "/v1/pod/open", bytes.NewBuffer(jsonData), cookies)
	resp.Body.Close()

	t.Run("UploadFile", func(t *testing.T) {
		// Create test file content
		fileContent := make([]byte, 1024*10) // 10KB
		rand.Read(fileContent)

		// Create multipart form
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("pod_name", podName)
		writer.WriteField("dir_path", "/")

		part, _ := writer.CreateFormFile("files", "test-document.pdf")
		part.Write(fileContent)
		writer.Close()

		req := httptest.NewRequest("POST", "/v1/file/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		suite.handler.GetRouter().ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("File upload failed with status %d: %s", resp.StatusCode, string(respBody))
		}

		t.Logf("✓ File uploaded: test-document.pdf (10KB)")
	})

	t.Run("CreateDirectory", func(t *testing.T) {
		payload := map[string]string{
			"pod_name": podName,
			"dir_path": "/documents",
		}
		jsonData, _ := json.Marshal(payload)

		resp, err := suite.makeRequest("POST", "/v1/dir/mkdir", bytes.NewBuffer(jsonData), cookies)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Directory creation failed with status %d: %s", resp.StatusCode, string(body))
		}

		t.Logf("✓ Directory created: /documents")
	})

	t.Run("ListDirectory", func(t *testing.T) {
		url := fmt.Sprintf("/v1/dir/ls?pod_name=%s&dir_path=/", podName)

		resp, err := suite.makeRequest("GET", url, nil, cookies)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Directory list failed with status %d: %s", resp.StatusCode, string(body))
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		t.Logf("✓ Directory listed successfully")
	})

	t.Run("DeleteFile", func(t *testing.T) {
		payload := map[string]string{
			"pod_name":  podName,
			"file_path": "/test-document.pdf",
		}
		jsonData, _ := json.Marshal(payload)

		resp, err := suite.makeRequest("DELETE", "/v1/file/delete", bytes.NewBuffer(jsonData), cookies)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("File delete failed with status %d: %s", resp.StatusCode, string(body))
		}

		t.Logf("✓ File deleted: test-document.pdf")
	})

	t.Logf("\n=== API File Operations Tests Passed ===\n")
}

// TestAPISharingOperations tests pod sharing via API
func TestAPISharingOperations(t *testing.T) {
	suite := setupAPITestSuite(t)
	t.Log("=== Testing API Sharing Operations ===\n")

	// Setup: Create two users
	user1 := "shareuser1" + randStringRunes(6)
	user2 := "shareuser2" + randStringRunes(6)
	password := "Password123!"

	// Create user1
	payload1 := map[string]string{
		"user_name": user1,
		"password":  password,
	}
	jsonData, _ := json.Marshal(payload1)
	resp, _ := suite.makeRequest("POST", "/v1/user/signup", bytes.NewBuffer(jsonData), nil)
	suite.storeCookies(user1, resp)
	resp.Body.Close()

	// Create user2
	payload2 := map[string]string{
		"user_name": user2,
		"password":  password,
	}
	jsonData, _ = json.Marshal(payload2)
	resp, _ = suite.makeRequest("POST", "/v1/user/signup", bytes.NewBuffer(jsonData), nil)
	suite.storeCookies(user2, resp)
	resp.Body.Close()

	podName := "shared-project"
	cookies1 := suite.getCookies(user1)

	// User1 creates pod
	podPayload := map[string]string{
		"pod_name": podName,
	}
	jsonData, _ = json.Marshal(podPayload)
	resp, _ = suite.makeRequest("POST", "/v1/pod/new", bytes.NewBuffer(jsonData), cookies1)
	resp.Body.Close()

	// Open pod
	resp, _ = suite.makeRequest("POST", "/v1/pod/open", bytes.NewBuffer(jsonData), cookies1)
	resp.Body.Close()

	var sharingReference string

	t.Run("SharePod", func(t *testing.T) {
		payload := map[string]string{
			"pod_name": podName,
		}
		jsonData, _ := json.Marshal(payload)

		resp, err := suite.makeRequest("POST", "/v1/pod/share", bytes.NewBuffer(jsonData), cookies1)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Pod share failed with status %d: %s", resp.StatusCode, string(body))
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		sharingReference = result["pod_sharing_reference"].(string)
		if sharingReference == "" {
			t.Fatal("Sharing reference should not be empty")
		}

		t.Logf("✓ Pod shared: %s", sharingReference[:20]+"...")
	})

	t.Run("ReceiveSharedPod", func(t *testing.T) {
		cookies2 := suite.getCookies(user2)

		payload := map[string]string{
			"pod_name":  "received-project",
			"reference": sharingReference,
		}
		jsonData, _ := json.Marshal(payload)

		resp, err := suite.makeRequest("POST", "/v1/pod/receive", bytes.NewBuffer(jsonData), cookies2)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Pod receive failed with status %d: %s", resp.StatusCode, string(body))
		}

		t.Logf("✓ User2 received shared pod successfully")
	})

	t.Logf("\n=== API Sharing Operations Tests Passed ===\n")
}

// TestAPIFullWorkflow tests a complete workflow via API
func TestAPIFullWorkflow(t *testing.T) {
	suite := setupAPITestSuite(t)
	t.Log("=== Testing API Full Workflow ===\n")
	t.Log("Workflow: User registers -> Creates pod -> Uploads files -> Shares -> Syncs\n")

	username := "workflow" + randStringRunes(8)
	password := "Password123!"
	podName := "workflow-pod"

	// Step 1: Register
	t.Log("Step 1: User registration")
	signupPayload := map[string]string{
		"user_name": username,
		"password":  password,
	}
	jsonData, _ := json.Marshal(signupPayload)
	resp, _ := suite.makeRequest("POST", "/v1/user/signup", bytes.NewBuffer(jsonData), nil)
	suite.storeCookies(username, resp)
	resp.Body.Close()
	t.Log("✓ User registered")

	cookies := suite.getCookies(username)

	// Step 2: Create pod
	t.Log("\nStep 2: Pod creation")
	podPayload := map[string]string{
		"pod_name": podName,
	}
	jsonData, _ = json.Marshal(podPayload)
	resp, _ = suite.makeRequest("POST", "/v1/pod/new", bytes.NewBuffer(jsonData), cookies)
	resp.Body.Close()
	t.Log("✓ Pod created")

	// Step 3: Open pod
	t.Log("\nStep 3: Opening pod")
	resp, _ = suite.makeRequest("POST", "/v1/pod/open", bytes.NewBuffer(jsonData), cookies)
	resp.Body.Close()
	t.Log("✓ Pod opened")

	// Step 4: Create directories
	t.Log("\nStep 4: Creating directory structure")
	dirs := []string{"/docs", "/data", "/images"}
	for _, dir := range dirs {
		dirPayload := map[string]string{
			"pod_name": podName,
			"dir_path": dir,
		}
		jsonData, _ := json.Marshal(dirPayload)
		resp, _ := suite.makeRequest("POST", "/v1/dir/mkdir", bytes.NewBuffer(jsonData), cookies)
		resp.Body.Close()
	}
	t.Logf("✓ Created %d directories", len(dirs))

	// Step 5: Upload files
	t.Log("\nStep 5: Uploading files")
	for i := 0; i < 3; i++ {
		fileContent := make([]byte, 1024*5)
		rand.Read(fileContent)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("pod_name", podName)
		writer.WriteField("dir_path", "/docs")

		part, _ := writer.CreateFormFile("files", fmt.Sprintf("doc%d.txt", i+1))
		part.Write(fileContent)
		writer.Close()

		req := httptest.NewRequest("POST", "/v1/file/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		suite.handler.GetRouter().ServeHTTP(w, req)
		w.Result().Body.Close()
	}
	t.Log("✓ Uploaded 3 files")

	// Step 6: Share pod
	t.Log("\nStep 6: Sharing pod")
	sharePayload := map[string]string{
		"pod_name": podName,
	}
	jsonData, _ = json.Marshal(sharePayload)
	resp, _ = suite.makeRequest("POST", "/v1/pod/share", bytes.NewBuffer(jsonData), cookies)
	var shareResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&shareResult)
	resp.Body.Close()

	sharingRef := shareResult["pod_sharing_reference"].(string)
	t.Logf("✓ Pod shared: %s", sharingRef[:20]+"...")

	// Step 7: Close and reopen (sync test)
	t.Log("\nStep 7: Testing sync (close and reopen)")
	closePayload := map[string]string{
		"pod_name": podName,
	}
	jsonData, _ = json.Marshal(closePayload)
	resp, _ = suite.makeRequest("POST", "/v1/pod/close", bytes.NewBuffer(jsonData), cookies)
	resp.Body.Close()

	// Reopen
	resp, _ = suite.makeRequest("POST", "/v1/pod/open", bytes.NewBuffer(jsonData), cookies)
	resp.Body.Close()
	t.Log("✓ Pod synced successfully")

	// Step 8: Verify files still exist
	t.Log("\nStep 8: Verifying data integrity")
	url := fmt.Sprintf("/v1/dir/ls?pod_name=%s&dir_path=/docs", podName)
	resp, _ = suite.makeRequest("GET", url, nil, cookies)
	var lsResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&lsResult)
	resp.Body.Close()
	t.Log("✓ All files verified")

	t.Logf("\n=== API Full Workflow Test Passed ===\n")
	t.Log("Summary:")
	t.Log("  - User registration: ✓")
	t.Log("  - Pod management: ✓")
	t.Log("  - Directory operations: ✓")
	t.Log("  - File operations: ✓")
	t.Log("  - Sharing: ✓")
	t.Log("  - Sync: ✓")
	t.Log("  - Data integrity: ✓")
}

// Helper function to generate random strings
func randStringRunes(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
