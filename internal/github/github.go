package github

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	repoOwner = "TONresistor"
	repoName  = "adnl-tunnel"
	apiBase   = "https://api.github.com/repos/" + repoOwner + "/" + repoName
)

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) GetLatestRelease() (*Release, error) {
	return c.getRelease(apiBase + "/releases/latest")
}

func (c *Client) GetReleaseByTag(tag string) (*Release, error) {
	return c.getRelease(fmt.Sprintf("%s/releases/tags/%s", apiBase, tag))
}

func (c *Client) getRelease(url string) (*Release, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	return &release, nil
}

func AssetName() string {
	arch := runtime.GOARCH
	if arch == "amd64" {
		return "tunnel-node-linux-amd64"
	}
	return "tunnel-node-linux-arm64"
}

func (r *Release) FindAsset(name string) (*Asset, error) {
	for _, a := range r.Assets {
		if a.Name == name {
			return &a, nil
		}
	}
	return nil, fmt.Errorf("asset %q not found in release %s", name, r.TagName)
}

func (c *Client) DownloadChecksums(release *Release) (map[string]string, error) {
	asset, err := release.FindAsset("checksums.txt")
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Get(asset.BrowserDownloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("checksums download returned status %d", resp.StatusCode)
	}

	checksums := make(map[string]string)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 {
			checksums[fields[1]] = fields[0]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse checksums: %w", err)
	}

	return checksums, nil
}

func (c *Client) DownloadAsset(asset *Asset, destPath string) (string, error) {
	resp, err := c.httpClient.Get(asset.BrowserDownloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	f, err := os.CreateTemp(filepath.Dir(destPath), ".tunnel-node-download-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	hasher := sha256.New()
	writer := io.MultiWriter(f, hasher)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("failed to write download: %w", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))

	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return "", err
	}

	if err := os.Chmod(f.Name(), 0755); err != nil {
		os.Remove(f.Name())
		return "", err
	}

	if err := os.Rename(f.Name(), destPath); err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("failed to move binary: %w", err)
	}

	return checksum, nil
}
