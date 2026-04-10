package webdav

import "testing"

func TestOptionsValidate(t *testing.T) {
	t.Run("defaults and normalize root", func(t *testing.T) {
		cfg := &Options{
			Endpoint: "https://dav.example.com/remote.php/dav/files/user",
			RootPath: "/Documents/Projects/",
		}
		if err := cfg.Validate(); err != nil {
			t.Fatalf("Validate() error = %v", err)
		}
		if cfg.RootPath != "Documents/Projects" {
			t.Fatalf("RootPath = %q", cfg.RootPath)
		}
		if cfg.TimeoutSec != 30 {
			t.Fatalf("TimeoutSec = %d", cfg.TimeoutSec)
		}
	})

	t.Run("requires endpoint", func(t *testing.T) {
		cfg := &Options{}
		if err := cfg.Validate(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects non-http schemes", func(t *testing.T) {
		cfg := &Options{Endpoint: "ftp://dav.example.com/files"}
		if err := cfg.Validate(); err == nil {
			t.Fatal("expected error")
		}
	})
}
