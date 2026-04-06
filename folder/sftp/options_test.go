package sftp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateAcceptsPrivateKeyPath(t *testing.T) {
	cfg := &Options{
		Address:        " 127.0.0.1 ",
		Username:       " user ",
		PrivateKeyPath: " ~/.ssh/id_rsa ",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if cfg.Address != "127.0.0.1" {
		t.Fatalf("Address = %q, want trimmed address", cfg.Address)
	}
	if cfg.Username != "user" {
		t.Fatalf("Username = %q, want trimmed username", cfg.Username)
	}
	if cfg.PrivateKeyPath != "~/.ssh/id_rsa" {
		t.Fatalf("PrivateKeyPath = %q, want trimmed path", cfg.PrivateKeyPath)
	}
	if cfg.Port != 22 {
		t.Fatalf("Port = %d, want default 22", cfg.Port)
	}
}

func TestLoadPrivateKeyMaterialReadsPrivateKeyPath(t *testing.T) {
	privateKey := "-----BEGIN OPENSSH PRIVATE KEY-----\nbody\n-----END OPENSSH PRIVATE KEY-----\n"
	keyPath := filepath.Join(t.TempDir(), "id_rsa")
	if err := os.WriteFile(keyPath, []byte(privateKey), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := loadPrivateKeyMaterial(&Options{PrivateKeyPath: keyPath})
	if err != nil {
		t.Fatalf("loadPrivateKeyMaterial() error = %v", err)
	}
	if string(got) != privateKey {
		t.Fatalf("loadPrivateKeyMaterial() = %q, want %q", string(got), privateKey)
	}
}

func TestLoadPrivateKeyMaterialTreatsLegacyPrivateKeyPathAsPath(t *testing.T) {
	privateKey := "-----BEGIN OPENSSH PRIVATE KEY-----\nlegacy\n-----END OPENSSH PRIVATE KEY-----\n"
	keyPath := filepath.Join(t.TempDir(), "id_rsa")
	if err := os.WriteFile(keyPath, []byte(privateKey), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := loadPrivateKeyMaterial(&Options{PrivateKey: keyPath})
	if err != nil {
		t.Fatalf("loadPrivateKeyMaterial() error = %v", err)
	}
	if string(got) != privateKey {
		t.Fatalf("loadPrivateKeyMaterial() = %q, want %q", string(got), privateKey)
	}
}

func TestLoadPrivateKeyMaterialPrefersInlineTextWhenExplicitPathAlsoExists(t *testing.T) {
	inlineKey := "-----BEGIN OPENSSH PRIVATE KEY-----\ninline\n-----END OPENSSH PRIVATE KEY-----"
	keyPath := filepath.Join(t.TempDir(), "id_rsa")
	if err := os.WriteFile(keyPath, []byte("path-key"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := loadPrivateKeyMaterial(&Options{
		PrivateKey:     inlineKey,
		PrivateKeyPath: keyPath,
	})
	if err != nil {
		t.Fatalf("loadPrivateKeyMaterial() error = %v", err)
	}
	if string(got) != inlineKey {
		t.Fatalf("loadPrivateKeyMaterial() = %q, want inline key", string(got))
	}
}

func TestLoadPrivateKeyMaterialUsesExplicitPathWhenPrivateKeyIsNotInlineText(t *testing.T) {
	privateKey := "-----BEGIN OPENSSH PRIVATE KEY-----\npath\n-----END OPENSSH PRIVATE KEY-----\n"
	keyPath := filepath.Join(t.TempDir(), "id_rsa")
	if err := os.WriteFile(keyPath, []byte(privateKey), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := loadPrivateKeyMaterial(&Options{
		PrivateKey:     "not-a-private-key-block",
		PrivateKeyPath: keyPath,
	})
	if err != nil {
		t.Fatalf("loadPrivateKeyMaterial() error = %v", err)
	}
	if string(got) != privateKey {
		t.Fatalf("loadPrivateKeyMaterial() = %q, want explicit path key", string(got))
	}
}

func TestFullPathWithAbsoluteRootPath(t *testing.T) {
	drv := &Driver{cfg: &Options{RootPath: "/home/ping"}}

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "root", path: "", want: "/home/ping"},
		{name: "child", path: "docs/readme.txt", want: "/home/ping/docs/readme.txt"},
		{name: "absolute caller path stays scoped", path: "/docs/readme.txt", want: "/home/ping/docs/readme.txt"},
		{name: "sibling traversal rejected", path: "../other", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := drv.fullPath(tt.path); got != tt.want {
				t.Fatalf("fullPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestFullPathWithRootSlash(t *testing.T) {
	drv := &Driver{cfg: &Options{RootPath: "/"}}

	if got := drv.fullPath(""); got != "/" {
		t.Fatalf("fullPath root = %q, want /", got)
	}
	if got := drv.fullPath("home/ping"); got != "/home/ping" {
		t.Fatalf("fullPath child = %q, want /home/ping", got)
	}
}
