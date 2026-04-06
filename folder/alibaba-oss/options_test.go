package alibaba_oss

import (
	"context"
	"testing"

	"github.com/wxk6b1203/file-util-manager/folder"
)

func TestNormalizePrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{name: "empty", prefix: "", want: ""},
		{name: "slash root", prefix: "/", want: ""},
		{name: "absolute prefix", prefix: "/home/ping", want: "home/ping/"},
		{name: "relative prefix", prefix: "home/ping", want: "home/ping/"},
		{name: "extra slashes", prefix: " /home/ping/ ", want: "home/ping/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizePrefix(tt.prefix); got != tt.want {
				t.Fatalf("normalizePrefix(%q) = %q, want %q", tt.prefix, got, tt.want)
			}
		})
	}
}

func TestNewMergesRootAndPrefixWithoutLeadingSlash(t *testing.T) {
	drv, err := New(context.Background(), &folder.DriverOptions{Root: "/root"}, &Options{
		Region:          "cn-hangzhou",
		Bucket:          "bucket",
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		Prefix:          "/child/",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ossDrv := drv.(*Driver)
	if ossDrv.cfg.Prefix != "root/child/" {
		t.Fatalf("Prefix = %q, want root/child/", ossDrv.cfg.Prefix)
	}
	if got := ossDrv.fullKey(""); got != "root/child/" {
		t.Fatalf("fullKey root = %q, want root/child/", got)
	}
	if got := ossDrv.fullKey("/file.txt"); got != "root/child/file.txt" {
		t.Fatalf("fullKey child = %q, want root/child/file.txt", got)
	}
}
