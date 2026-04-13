package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestPresignAndServe(t *testing.T) {
	skipIfNoIntLive(t)

	custToken := intLogin(t, intCustUser, intCustPass)
	bID := intCreateBooking(t, custToken)

	// Upload a minimal PNG.
	attID := intUploadAttachment(t, custToken, bID, "", intMiniPNG)

	// Request a presigned URL.
	presignResp := intAPI(t, http.MethodPost, "/api/v1/attachments/"+attID+"/presign",
		map[string]int{"ttlSeconds": 60}, custToken)
	presignBody := intMustStatus(t, presignResp, http.StatusOK)
	var presignOut struct{ Url string `json:"url"` }
	if err := json.Unmarshal(presignBody, &presignOut); err != nil || presignOut.Url == "" {
		t.Fatalf("presign: bad response %s", presignBody)
	}

	// Fetch the file via the presigned URL.
	u, err := url.Parse(presignOut.Url)
	if err != nil {
		t.Fatalf("invalid presign url: %v", err)
	}
	// Rewrite host to the test server URL (presign URL might contain the
	// internal container hostname; we connect via TEST_SERVER_URL).
	serverBase, _ := url.Parse(intLiveServerURL)
	u.Host = serverBase.Host
	u.Scheme = serverBase.Scheme

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		t.Fatalf("build serve request: %v", err)
	}
	serveResp, err := intLiveClient.Do(req)
	if err != nil {
		t.Fatalf("serve request: %v", err)
	}
	defer serveResp.Body.Close()

	if serveResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(serveResp.Body)
		t.Fatalf("serve: expected 200, got %d — %s", serveResp.StatusCode, body)
	}
	if serveResp.Header.Get("Content-Disposition") == "" {
		t.Fatal("serve: missing Content-Disposition header")
	}
	body, _ := io.ReadAll(serveResp.Body)
	if len(body) == 0 {
		t.Fatal("serve: empty body")
	}
}
