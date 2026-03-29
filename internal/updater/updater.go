package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/TONresistor/tonrelay/internal/github"
	"github.com/TONresistor/tonrelay/internal/service"
)

type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	UpdateAvailable bool
}

func CheckUpdate(dataDir string) (*UpdateInfo, error) {
	current := readVersion(dataDir)

	gh := github.NewClient()
	release, err := gh.GetLatestRelease()
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	return &UpdateInfo{
		CurrentVersion:  current,
		LatestVersion:   release.TagName,
		UpdateAvailable: current != release.TagName,
	}, nil
}

func Update(version, binaryPath, dataDir string) error {
	gh := github.NewClient()

	var release *github.Release
	var err error
	if version == "" || version == "latest" {
		release, err = gh.GetLatestRelease()
	} else {
		release, err = gh.GetReleaseByTag(version)
	}
	if err != nil {
		return err
	}

	assetName := github.AssetName()
	asset, err := release.FindAsset(assetName)
	if err != nil {
		return err
	}

	checksums, err := gh.DownloadChecksums(release)
	if err != nil {
		fmt.Printf("Warning: could not download checksums.txt: %v\n", err)
	}

	wasRunning := service.IsActive()
	if wasRunning {
		fmt.Println("Stopping service...")
		if err := service.Stop(); err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}
	}

	backupPath := filepath.Join(dataDir, "tunnel-node.bak")
	if _, err := os.Stat(binaryPath); err == nil {
		fmt.Println("Backing up current binary...")
		if err := copyFile(binaryPath, backupPath); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}

	checksumPath := filepath.Join(dataDir, "tunnel-node.sha256")
	versionPath := filepath.Join(dataDir, "version")
	oldChecksum, _ := os.ReadFile(checksumPath)
	oldVersion, _ := os.ReadFile(versionPath)

	rollback := func() {
		if _, err := os.Stat(backupPath); err == nil {
			os.Rename(backupPath, binaryPath)
		}
		if oldChecksum != nil {
			os.WriteFile(checksumPath, oldChecksum, 0600)
		}
		if oldVersion != nil {
			os.WriteFile(versionPath, oldVersion, 0644)
		}
		if wasRunning {
			service.Start()
		}
	}

	fmt.Printf("Downloading %s (%s)...\n", assetName, release.TagName)
	checksum, err := gh.DownloadAsset(asset, binaryPath)
	if err != nil {
		rollback()
		return fmt.Errorf("download failed: %w", err)
	}

	if expected, ok := checksums[assetName]; ok {
		if checksum != expected {
			rollback()
			return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, checksum)
		}
		fmt.Println("Checksum verified against upstream checksums.txt")
	}

	if err := os.WriteFile(checksumPath, []byte(checksum), 0600); err != nil {
		rollback()
		return fmt.Errorf("failed to write checksum: %w", err)
	}

	if err := os.WriteFile(versionPath, []byte(release.TagName+"\n"), 0644); err != nil {
		rollback()
		return fmt.Errorf("failed to write version: %w", err)
	}

	fmt.Printf("Updated to %s (SHA256: %s)\n", release.TagName, checksum[:16]+"...")

	ok, verr := VerifyBinary(binaryPath, dataDir)
	if verr != nil {
		rollback()
		return fmt.Errorf("binary verification failed: %w", verr)
	}
	if !ok {
		rollback()
		return fmt.Errorf("binary checksum mismatch after download")
	}

	if wasRunning {
		fmt.Println("Starting service...")
		if err := service.Start(); err != nil {
			rollback()
			return fmt.Errorf("failed to start service: %w", err)
		}
		fmt.Println("Service started.")
	}

	return nil
}

func VerifyBinary(binaryPath, dataDir string) (bool, error) {
	checksumPath := filepath.Join(dataDir, "tunnel-node.sha256")
	stored, err := os.ReadFile(checksumPath)
	if err != nil {
		return false, fmt.Errorf("no stored checksum found: %w", err)
	}

	f, err := os.Open(binaryPath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return false, err
	}

	current := hex.EncodeToString(hasher.Sum(nil))
	return current == strings.TrimSpace(string(stored)), nil
}

func readVersion(dataDir string) string {
	data, err := os.ReadFile(filepath.Join(dataDir, "version"))
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
