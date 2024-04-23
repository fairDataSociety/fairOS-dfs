package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/file"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"github.com/fairdatasociety/fairOS-dfs/pkg/auth/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/auth/jwt"
	"github.com/gorilla/mux"
	"resenje.org/jsonhttp"
)

var (
	refFile = "refFile"
)

// GitAuthMiddleware checks the Authorization header for git auth credentials
func (h *Handler) GitAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}

		_, err := jwt.GetSessionIdFromGitRequest(r)
		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// validateCredentials checks the provided username and password
func (h *Handler) validateCredentials(w http.ResponseWriter, username, password string) bool {
	loginResp, err := h.dfsAPI.LoginUserV2(username, password, "")
	if err != nil {
		return false
	}
	err = cookie.SetSession(loginResp.UserInfo.GetSessionId(), w, h.cookieDomain)
	if err != nil {
		return false
	}
	return true
}

func (h *Handler) GitInfoRef(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetSessionIdFromGitRequest(r)
	if err != nil {
		h.logger.Errorf("sessionId parse failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrUnauthorized.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Error("sessionId not set: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrUnauthorized.Error()})
		return
	}

	vars := mux.Vars(r)
	pod := vars["repo"]
	serviceType := r.FormValue("service")

	// check if pod exists
	if _, err = h.dfsAPI.OpenPod(pod, sessionId); err != nil {
		h.logger.Errorf("IsPodExist failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: "Repo does not exist"})
		return
	}
	refLine := ""
	// check if ref file exists
	reader, _, err := h.dfsAPI.DownloadFile(pod, fmt.Sprintf("/%s", refFile), sessionId, false)
	if err == nil {
		defer reader.Close()
		refData, _ := ioutil.ReadAll(reader)
		if len(refData) != 0 {
			refLine = fmt.Sprintf("%s\n", refData)
		}
	}

	w.Header().Set("Expires", "Fri, 01 Jan 2080 00:00:00 GMT")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", serviceType))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(packetWrite("# service=" + serviceType + "\n")))
	_, _ = w.Write([]byte("0000"))
	if refLine != "" {
		_, _ = w.Write([]byte(packetWrite(refLine)))
	} else {
		capabilities := "report-status object-format=sha1"
		emptyList := fmt.Sprintf("0000000000000000000000000000000000000000 capabilities^{}\000%s\n", capabilities)
		_, _ = w.Write([]byte(packetWrite(emptyList)))
	}
	_, _ = w.Write([]byte("0000"))
}

func (h *Handler) GitUploadPack(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetSessionIdFromGitRequest(r)
	if err != nil {
		h.logger.Errorf("sessionId parse failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrUnauthorized.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Error("sessionId not set: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrUnauthorized.Error()})
		return
	}
	vars := mux.Vars(r)
	pod := vars["repo"]
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-upload-pack-result"))

	reader, _, err := h.dfsAPI.DownloadFile(pod, fmt.Sprintf("/%s", refFile), sessionId, false)
	if err != nil {
		h.logger.Error("ref not found: ", err)
		jsonhttp.BadRequest(w, &response{Message: "ref not found"})
		return
	}
	commitDetailsBytes, err := io.ReadAll(reader)
	if err != nil {
		h.logger.Error("ref not found: ", err)
		jsonhttp.BadRequest(w, &response{Message: "ref not found"})
		return
	}
	commitDetailsArr := strings.Split(string(commitDetailsBytes), " ")

	packReader, _, err := h.dfsAPI.DownloadFile(pod, fmt.Sprintf("/%s", commitDetailsArr[0]), sessionId, false)
	if err != nil {
		h.logger.Error("ref not found: ", err)
		jsonhttp.BadRequest(w, &response{Message: "ref not found"})
		return
	}
	_, _ = w.Write([]byte(packetWrite(fmt.Sprintf("ACK %s\n", commitDetailsArr[0]))))
	_, err = io.Copy(w, packReader)
	if err != nil {
		h.logger.Errorf("download: %v", err)
		w.Header().Set("Content-Type", " application/json")
		jsonhttp.InternalServerError(w, "download: "+err.Error())
	}
}

func (h *Handler) GitReceivePack(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetSessionIdFromGitRequest(r)
	if err != nil {
		h.logger.Errorf("sessionId parse failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrUnauthorized.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Error("sessionId not set: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrUnauthorized.Error()})
		return
	}

	defer r.Body.Close()
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		h.logger.Errorf("Error reading request body: %v", err)
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusInternalServerError)
		return
	}

	packIndex := bytes.Index(buf.Bytes(), []byte("PACK"))
	if packIndex == -1 {
		h.logger.Errorf("PACK signature not found in request body")
		http.Error(w, "PACK signature not found in request body", http.StatusBadRequest)
		return
	}
	commitDetails := strings.TrimSpace(buf.String()[:packIndex])
	commitDetailsArr := strings.Split(commitDetails, " ")

	vars := mux.Vars(r)
	pod := vars["repo"]
	newHash, ref := commitDetailsArr[1], commitDetailsArr[2]

	_, _, _, err = h.dfsAPI.StatusFile(pod, fmt.Sprintf("/%s", refFile), sessionId, false)
	if err != nil && !errors.Is(err, file.ErrFileNotFound) {
		h.logger.Errorf("Error checking commit status: %v", err)
		http.Error(w, fmt.Sprintf("Error checking file status: %v", err), http.StatusInternalServerError)
		return
	}
	if err == nil {
		h.logger.Errorf("Cannot push. ref file already exists")
		http.Error(w, fmt.Sprintf("Cannot push"), http.StatusInternalServerError)
		return
	}
	err = h.dfsAPI.UploadFile(pod, refFile, sessionId, int64(len(newHash+" "+ref)), strings.NewReader(newHash+" "+ref), "/", "", file.MinBlockSize, 0, false, false)
	if err != nil {
		h.logger.Errorf("Error uploading commit: %v", err)
		http.Error(w, fmt.Sprintf("Error uploading file: %v", err), http.StatusInternalServerError)
		return
	}

	packData := bytes.NewReader(buf.Bytes()[packIndex:])
	err = h.dfsAPI.UploadFile(pod, newHash, sessionId, int64(packData.Len()), packData, "/", "", file.MinBlockSize, 0, false, false)
	if err != nil {
		h.logger.Errorf("Error uploading packfile: %v", err)
		http.Error(w, fmt.Sprintf("Error uploading file: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-receive-pack-result"))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(packetWrite("unpack ok\n")))
	w.Write([]byte(packetWrite("ok " + ref + "\n")))
	w.Write([]byte("0000"))
}

func packetWrite(data string) string {
	length := len(data) + 4 // 4 bytes for the length itself
	return fmt.Sprintf("%04x%s", length, data)
}
