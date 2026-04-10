package webdav

import (
	"fmt"
	"net/url"
	pathpkg "path"
	"strings"
)

type Options struct {
	Endpoint           string `yaml:"endpoint" json:"endpoint"`
	Username           string `yaml:"username,omitempty" json:"username,omitempty"`
	Password           string `yaml:"password,omitempty" json:"password,omitempty"`
	BearerToken        string `yaml:"bearerToken,omitempty" json:"bearerToken,omitempty"`
	RootPath           string `yaml:"rootPath,omitempty" json:"rootPath,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify,omitempty" json:"insecureSkipVerify,omitempty"`
	TimeoutSec         int    `yaml:"timeoutSec,omitempty" json:"timeoutSec,omitempty"`
}

func (o *Options) Validate() error {
	o.Endpoint = strings.TrimSpace(o.Endpoint)
	o.Username = strings.TrimSpace(o.Username)
	o.Password = strings.TrimSpace(o.Password)
	o.BearerToken = strings.TrimSpace(o.BearerToken)
	o.RootPath = normalizeRootPath(o.RootPath)

	if o.Endpoint == "" {
		return fmt.Errorf("webdav: endpoint is required")
	}

	parsed, err := url.Parse(o.Endpoint)
	if err != nil {
		return fmt.Errorf("webdav: invalid endpoint %q: %w", o.Endpoint, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("webdav: endpoint %q must use http or https", o.Endpoint)
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return fmt.Errorf("webdav: endpoint %q is missing host", o.Endpoint)
	}

	if o.TimeoutSec <= 0 {
		o.TimeoutSec = 30
	}

	return nil
}

func normalizeRootPath(value string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	if trimmed == "" || trimmed == "." || trimmed == "/" {
		return ""
	}

	cleaned := pathpkg.Clean("/" + strings.TrimPrefix(trimmed, "/"))
	if cleaned == "/" || cleaned == "." {
		return ""
	}
	return strings.TrimPrefix(cleaned, "/")
}
