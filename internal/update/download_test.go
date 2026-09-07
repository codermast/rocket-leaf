package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func digestOf(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// assetServer serves one payload with a Content-Length, the way the release
// assets are actually served. Go only sets one itself for small bodies, and a
// big enough payload would otherwise go out chunked.
func assetServer(t *testing.T, payload []byte) (*httptest.Server, *http.Client) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if agent := r.Header.Get("User-Agent"); !strings.HasPrefix(agent, "MQ-Studio") {
			t.Errorf("User-Agent = %q", agent)
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
		_, _ = w.Write(payload)
	}))
	t.Cleanup(server.Close)
	return server, server.Client()
}

func TestDownloadVerifiesAndRenamesIntoPlace(t *testing.T) {
	payload := []byte(strings.Repeat("mq", 400_000))
	server, client := assetServer(t, payload)
	destination := filepath.Join(t.TempDir(), "nested", "mq-studio-1.0.0-mac-arm64.dmg")

	var mu sync.Mutex
	var reports [][2]int64
	err := Download(context.Background(), client, server.URL, destination, digestOf(payload),
		func(done, total int64) {
			mu.Lock()
			defer mu.Unlock()
			reports = append(reports, [2]int64{done, total})
		})
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}

	written, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("reading the download: %v", err)
	}
	if len(written) != len(payload) {
		t.Fatalf("wrote %d bytes, want %d", len(written), len(payload))
	}
	if _, err := os.Stat(destination + ".part"); !os.IsNotExist(err) {
		t.Error("the partial file should be gone once the download lands")
	}
	if len(reports) < 2 {
		t.Fatalf("progress reported %d times, want several", len(reports))
	}
	last := reports[len(reports)-1]
	if last[0] != int64(len(payload)) {
		t.Errorf("final progress = %d bytes, want %d", last[0], len(payload))
	}
	if last[1] != int64(len(payload)) {
		t.Errorf("reported total = %d, want the Content-Length %d", last[1], len(payload))
	}
}

func TestDownloadRefusesContentThatDoesNotMatchItsChecksum(t *testing.T) {
	server, client := assetServer(t, []byte("the real package"))
	destination := filepath.Join(t.TempDir(), "package.dmg")

	err := Download(context.Background(), client, server.URL, destination, digestOf([]byte("something else")), nil)
	if !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("Download() error = %v, want ErrChecksumMismatch", err)
	}
	for _, path := range []string{destination, destination + ".part"} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s should not survive a failed verification", filepath.Base(path))
		}
	}
}

func TestDownloadRefusesToRunWithoutAChecksum(t *testing.T) {
	server, client := assetServer(t, []byte("payload"))
	destination := filepath.Join(t.TempDir(), "package.dmg")

	if err := Download(context.Background(), client, server.URL, destination, "", nil); err == nil {
		t.Fatal("Download() should refuse an asset with no published checksum")
	}
}

func TestDownloadStopsWhenTheContextIsCancelled(t *testing.T) {
	// Big enough that the copy loop is still running when the cancel lands.
	payload := []byte(strings.Repeat("x", 8<<20))
	server, client := assetServer(t, payload)
	destination := filepath.Join(t.TempDir(), "package.dmg")

	ctx, cancel := context.WithCancel(context.Background())
	err := Download(ctx, client, server.URL, destination, digestOf(payload),
		func(int64, int64) { cancel() })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Download() error = %v, want context.Canceled", err)
	}
	if _, err := os.Stat(destination + ".part"); !os.IsNotExist(err) {
		t.Error("a cancelled download should leave no partial file behind")
	}
}

func TestDownloadReportsAnHTTPFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)
	destination := filepath.Join(t.TempDir(), "package.dmg")

	err := Download(context.Background(), server.Client(), server.URL, destination, digestOf(nil), nil)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("Download() error = %v, want the status code", err)
	}
}

// A server that sends no length must still download; the bar just cannot say
// how far along it is.
func TestDownloadHandlesAnUnknownContentLength(t *testing.T) {
	payload := []byte("chunked package")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Transfer-Encoding", "chunked")
		_, _ = w.Write(payload)
	}))
	t.Cleanup(server.Close)
	destination := filepath.Join(t.TempDir(), "package.dmg")

	var total int64
	if err := Download(context.Background(), server.Client(), server.URL, destination, digestOf(payload),
		func(_, reported int64) { total = reported }); err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	if total != -1 {
		t.Errorf("reported total = %d, want -1 for an unknown length", total)
	}
}
