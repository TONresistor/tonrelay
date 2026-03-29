package updater

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadVersion(t *testing.T) {
	dir := t.TempDir()

	// No version file
	v := readVersion(dir)
	if v != "unknown" {
		t.Errorf("readVersion with no file = %q, want %q", v, "unknown")
	}

	// With version file
	versionPath := filepath.Join(dir, "version")
	os.WriteFile(versionPath, []byte("v0.1.8\n"), 0644)

	v = readVersion(dir)
	if v != "v0.1.8" {
		t.Errorf("readVersion = %q, want %q", v, "v0.1.8")
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")

	content := []byte("test binary content")
	os.WriteFile(src, content, 0755)

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("copy content = %q, want %q", string(data), string(content))
	}
}

func TestVerifyBinary(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "binary")
	os.WriteFile(binPath, []byte("fake binary"), 0755)

	// No checksum stored
	_, err := VerifyBinary(binPath, dir)
	if err == nil {
		t.Error("VerifyBinary should fail with no stored checksum")
	}

	// Wrong checksum
	checksumPath := filepath.Join(dir, "tunnel-node.sha256")
	os.WriteFile(checksumPath, []byte("0000000000000000000000000000000000000000000000000000000000000000"), 0600)

	valid, err := VerifyBinary(binPath, dir)
	if err != nil {
		t.Fatalf("VerifyBinary: %v", err)
	}
	if valid {
		t.Error("VerifyBinary should be false with wrong checksum")
	}
}
