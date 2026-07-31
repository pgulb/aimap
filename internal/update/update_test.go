package update

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFetchLatestTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"tag_name": "release-20260731-abc1234"}`))
	}))
	defer server.Close()

	orig := httpGet
	httpGet = func(url string) (*http.Response, error) {
		return http.Get(server.URL)
	}
	defer func() { httpGet = orig }()

	tag, err := fetchLatestTag()
	if err != nil {
		t.Fatal(err)
	}
	if tag != "release-20260731-abc1234" {
		t.Errorf("got %q, want release-20260731-abc1234", tag)
	}
}

func TestFetchLatestTagEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"tag_name": ""}`))
	}))
	defer server.Close()

	orig := httpGet
	httpGet = func(url string) (*http.Response, error) {
		return http.Get(server.URL)
	}
	defer func() { httpGet = orig }()

	_, err := fetchLatestTag()
	if err == nil {
		t.Fatal("expected error for empty tag_name")
	}
}

func TestFetchLatestTagHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	orig := httpGet
	httpGet = func(url string) (*http.Response, error) {
		return http.Get(server.URL)
	}
	defer func() { httpGet = orig }()

	_, err := fetchLatestTag()
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestDownloadFile(t *testing.T) {
	content := []byte("binary-content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer server.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "archive.tar.gz")

	if err := downloadFile(server.URL, dest); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "binary-content" {
		t.Errorf("got %q, want %q", data, content)
	}
}

func TestDoFailsWithoutMarker(t *testing.T) {
	orig := markerPath
	defer func() { markerPath = orig }()

	tmpDir := t.TempDir()
	markerPath = func() string {
		return filepath.Join(tmpDir, "will_autoupdate")
	}

	err := Do("dev", false)
	if err == nil {
		t.Fatal("expected error without marker")
	}
}

func TestReplaceBinary(t *testing.T) {
	dir := t.TempDir()
	oldExe := filepath.Join(dir, "aimap")
	newExe := filepath.Join(dir, "aimap_new")

	if err := os.WriteFile(oldExe, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newExe, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := replaceBinary(newExe, oldExe); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(oldExe)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Errorf("got %q, want 'new'", data)
	}
}
