package s3

import (
	"fmt"
	"strings"
)

// Options holds S3-specific connection parameters.
// All credential fields are passed explicitly — no environment variable fallback.
type Options struct {
	Region          string `yaml:"region" json:"region"`
	Bucket          string `yaml:"bucket" json:"bucket"`
	AccessKeyID     string `yaml:"accessKeyID" json:"accessKeyID"`
	AccessKeySecret string `yaml:"accessKeySecret" json:"accessKeySecret"`
	SessionToken    string `yaml:"sessionToken,omitempty" json:"sessionToken,omitempty"`
	Endpoint        string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`

	// ForcePathStyle forces path-style addressing (e.g. for MinIO / self-hosted S3).
	ForcePathStyle bool `yaml:"forcePathStyle,omitempty" json:"forcePathStyle,omitempty"`
	// DisableSSL uses HTTP instead of HTTPS when true.
	DisableSSL bool `yaml:"disableSSL,omitempty" json:"disableSSL,omitempty"`
	// Prefix is prepended to every key, acting as a virtual sub-directory.
	Prefix string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
}

func (o *Options) Validate() error {
	if o.Region == "" {
		return fmt.Errorf("s3: region is required")
	}
	if o.Bucket == "" {
		return fmt.Errorf("s3: bucket is required")
	}
	if o.AccessKeyID == "" || o.AccessKeySecret == "" {
		return fmt.Errorf("s3: accessKeyID and accessKeySecret are required")
	}
	o.Region = strings.TrimSpace(o.Region)
	o.Bucket = strings.TrimSpace(o.Bucket)
	o.AccessKeyID = strings.TrimSpace(o.AccessKeyID)
	o.AccessKeySecret = strings.TrimSpace(o.AccessKeySecret)
	o.SessionToken = strings.TrimSpace(o.SessionToken)
	o.Endpoint = strings.TrimSpace(o.Endpoint)
	o.Prefix = strings.TrimSpace(o.Prefix)

	// Normalize prefix: ensure it ends with "/" if non-empty.
	if o.Prefix != "" {
		o.Prefix = strings.TrimRight(o.Prefix, "/") + "/"
	}
	return nil
}
