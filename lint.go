package checker

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/danjacques/gofslock/fslock"
)

// Lint runs golangci-lint on the current package.
//
// Version should be, e.g. "1.2.3" and not contain a "v".
func Lint(t *testing.T, version string) { lint(testingT{t}, version) }

func lint(t testingT, version string) {
	var bin string
	_ = fslock.With(filepath.Join(cacheDir(t), "lint.lock"), func() error {
		bin = lintBin(t, version)
		return nil
	})
	out, err := exec.Command(
		bin, "run", "--allow-parallel-runners",
	).CombinedOutput()
	if ee := new(exec.ExitError); err != nil && errors.As(err, &ee) {
		t.Errorf("golangci-lint failed\n%s", string(out))
	} else if err != nil {
		t.Fatalf("golangci-lint failed to run: %v", err)
	}
}

func lintBin(t testingT, version string) (path string) {
	path = filepath.Join(
		cacheDir(t),
		fmt.Sprintf("golangci-lint-%s-%s-%s",
			version, runtime.GOOS, runtime.GOARCH),
	)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		lintFetch(t, path, version)
	} else if err != nil {
		t.Fatalf("failed to stat %v: %v", path, err)
	}
	return
}

const lintReleaseURL = "https://github.com/golangci/golangci-lint/releases" +
	"/download/v%[1]s"
const lintSumURL = lintReleaseURL + "/golangci-lint-%[1]s-checksums.txt"
const lintTarFile = "golangci-lint-%[1]s-%[2]s-%[3]s.tar.gz"
const lintTarURL = lintReleaseURL + "/" + lintTarFile

func lintSum(t testingT, version string) []byte {
	url := fmt.Sprintf(lintSumURL, version)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to fetch %v: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	hashtext := lintHash(t, resp.Body, version)
	hash, err := hex.DecodeString(hashtext)
	if err != nil {
		t.Fatalf("failed to decode golangci-lint sum %v: %v", hashtext, err)
	}
	return hash
}

func lintHash(t testingT, r io.Reader, version string) string {
	wantfile := fmt.Sprintf("golangci-lint-%s-%s-%s.tar.gz",
		version, runtime.GOOS, runtime.GOARCH)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		hash, file, ok := strings.Cut(scanner.Text(), "  ")
		if !ok {
			t.Fatalf("failed to parse golangci-lint checksum file: "+
				"failed to parse line: %v", scanner.Text())
		}
		if file == wantfile {
			return hash
		}
	}
	t.Fatalf("failed to find %q in golangci-lint checksum file", wantfile)
	return ""
}

func lintFetch(t testingT, path, version string) {
	url := fmt.Sprintf(lintTarURL, version, runtime.GOOS, runtime.GOARCH)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to fetch %v: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	sum := sha256.New()
	var buf bytes.Buffer
	if _, err = io.Copy(io.MultiWriter(sum, &buf), resp.Body); err != nil {
		t.Fatalf("failed to read body from %v: %v", url, err)
	}

	gr, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatalf("failed to create gzip reader for %v: %v", url, err)
	}
	defer func() { _ = gr.Close() }()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			t.Fatalf("failed to find 'golangci-lint' binary in %v", url)
		}
		if err != nil {
			t.Fatalf("failed to read tar entry: %v", err)
		}
		if strings.HasSuffix(hdr.Name, "golangci-lint") {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("failed to create directory %v: %v", dir, err)
			}
			f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				_ = os.Remove(path)
				t.Fatalf("failed to create file %v: %v", path, err)
			}
			defer func() { _ = f.Close() }()
			if _, err = io.Copy(f, tr); err != nil {
				_ = os.Remove(path)
				t.Fatalf("failed to write file %v: %v", path, err)
			}
			break
		}
	}
	if !bytes.Equal(sum.Sum(nil), lintSum(t, version)) {
		_ = os.Remove(path) // Attempt to remove untrusted binary.
		t.Fatalf("checksum mismatch for golangci-lint")
	}
}
