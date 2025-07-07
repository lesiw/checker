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
)

// Lint runs golangci-lint on the current package.
//
// Version should be, e.g. "1.2.3" and not contain a "v".
func Lint(t *testing.T, version string) {
	bin := lintBin(t, version)
	out, err := exec.Command(bin, "run").CombinedOutput()
	if ee := new(exec.ExitError); err != nil && errors.As(err, &ee) {
		t.Errorf("[lesiw.io/checker] golangci-lint failed\n%s", string(out))
	} else if err != nil {
		t.Fatalf("[lesiw.io/checker] golangci-lint failed to run: %v", err)
	}
}

func lintBin(t *testing.T, version string) string {
	cache, err := os.UserCacheDir()
	if err != nil {
		fatal(t, "failed to get user cache directory: %v", err)
	}
	dir := filepath.Join(cache, "gochecker")
	if err = os.MkdirAll(dir, 0755); err != nil {
		fatal(t, "failed to create cache directory: %v", err)
	}
	bin := fmt.Sprintf("golangci-lint-%s-%s-%s",
		version, runtime.GOOS, runtime.GOARCH)
	if _, err = os.Stat(filepath.Join(dir, bin)); os.IsNotExist(err) {
		lintFetch(t, filepath.Join(dir, bin), version)
	} else if err != nil {
		fatal(t, "failed to stat %v: %v", filepath.Join(dir, bin), err)
	}
	return filepath.Join(dir, bin)
}

const lintReleaseURL = "https://github.com/golangci/golangci-lint/releases" +
	"/download/v%[1]s"
const lintSumURL = lintReleaseURL + "/golangci-lint-%[1]s-checksums.txt"
const lintTarFile = "golangci-lint-%[1]s-%[2]s-%[3]s.tar.gz"
const lintTarURL = lintReleaseURL + "/" + lintTarFile

func lintSum(t *testing.T, version string) []byte {
	url := fmt.Sprintf(lintSumURL, version)
	resp, err := http.Get(url)
	if err != nil {
		fatal(t, "failed to fetch %v: %v", url, err)
	}
	defer resp.Body.Close()
	hashtext := lintHash(t, resp.Body, version)
	hash, err := hex.DecodeString(hashtext)
	if err != nil {
		fatal(t, "failed to decode golangci-lint sum %v: %v", hashtext, err)
	}
	return hash
}

func lintHash(t *testing.T, r io.Reader, version string) string {
	wantfile := fmt.Sprintf("golangci-lint-%s-%s-%s.tar.gz",
		version, runtime.GOOS, runtime.GOARCH)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		hash, file, ok := strings.Cut(line, "  ")
		if !ok {
			fatal(t, "failed to parse golangci-lint checksum file: "+
				"failed to parse line: %v", line)
		} else if file == wantfile {
			return hash
		}
	}
	fatal(t, "failed to find %q in golangci-lint checksum file", wantfile)
	return ""
}

func lintFetch(t *testing.T, path, version string) {
	url := fmt.Sprintf(lintTarURL, version, runtime.GOOS, runtime.GOARCH)
	resp, err := http.Get(url)
	if err != nil {
		fatal(t, "failed to fetch %v: %v", url, err)
	}
	defer resp.Body.Close()

	sum := sha256.New()
	var buf bytes.Buffer
	r := io.TeeReader(resp.Body, &buf)
	if _, err = io.Copy(sum, r); err != nil {
		fatal(t, "failed to read body from %v: %v", url, err)
	}
	gr, err := gzip.NewReader(&buf)
	if err != nil {
		fatal(t, "failed to create gzip reader for %v: %v", url, err)
	}
	tr := tar.NewReader(gr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			fatal(t, "failed to find 'golangci-lint' binary in %v", url)
		}
		if strings.HasSuffix(hdr.Name, "golangci-lint") {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				fatal(t, "failed to create directory %v: %v", dir, err)
			}
			f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				_ = os.Remove(path)
				fatal(t, "failed to create file %v: %v", path, err)
			}
			defer f.Close()
			if _, err = io.Copy(f, tr); err != nil {
				_ = os.Remove(path)
				fatal(t, "failed to write file %v: %v", path, err)
			}
			break
		}
	}
	if !bytes.Equal(sum.Sum(nil), lintSum(t, version)) {
		_ = os.Remove(path) // Attempt to remove untrusted binary.
		fatal(t, "checksum mismatch for golangci-lint")
	}
}
