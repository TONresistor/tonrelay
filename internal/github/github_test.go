package github

import (
	"runtime"
	"testing"
)

func TestAssetName(t *testing.T) {
	name := AssetName()

	switch runtime.GOARCH {
	case "amd64":
		if name != "tunnel-node-linux-amd64" {
			t.Errorf("AssetName() = %q, want tunnel-node-linux-amd64", name)
		}
	case "arm64":
		if name != "tunnel-node-linux-arm64" {
			t.Errorf("AssetName() = %q, want tunnel-node-linux-arm64", name)
		}
	}
}

func TestReleaseFindAsset(t *testing.T) {
	release := &Release{
		TagName: "v0.1.8",
		Assets: []Asset{
			{Name: "tunnel-node-linux-amd64", BrowserDownloadURL: "https://example.com/amd64", Size: 1000},
			{Name: "tunnel-node-linux-arm64", BrowserDownloadURL: "https://example.com/arm64", Size: 1000},
		},
	}

	asset, err := release.FindAsset("tunnel-node-linux-amd64")
	if err != nil {
		t.Fatalf("FindAsset: %v", err)
	}
	if asset.BrowserDownloadURL != "https://example.com/amd64" {
		t.Errorf("URL = %q, want https://example.com/amd64", asset.BrowserDownloadURL)
	}

	_, err = release.FindAsset("nonexistent")
	if err == nil {
		t.Error("FindAsset should error for missing asset")
	}
}

func TestReleaseEmpty(t *testing.T) {
	release := &Release{
		TagName: "v0.0.0",
		Assets:  []Asset{},
	}

	_, err := release.FindAsset("tunnel-node-linux-amd64")
	if err == nil {
		t.Error("FindAsset should error for empty release")
	}
}
