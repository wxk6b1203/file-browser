package sftp

import (
	"fmt"
	"strings"
)

// Options holds SFTP-specific connection parameters.
// Authentication supports password and/or private key (PEM-encoded).
type Options struct {
	Address        string `yaml:"address" json:"address"`
	Port           int    `yaml:"port,omitempty" json:"port,omitempty"`
	Username       string `yaml:"username" json:"username"`
	Password       string `yaml:"password,omitempty" json:"password,omitempty"`
	PrivateKey     string `yaml:"privateKey,omitempty" json:"privateKey,omitempty"`
	PrivateKeyPath string `yaml:"privateKeyPath,omitempty" json:"privateKeyPath,omitempty"`
	Passphrase     string `yaml:"passphrase,omitempty" json:"passphrase,omitempty"`
	RootPath       string `yaml:"rootPath,omitempty" json:"rootPath,omitempty"`
	DialTimeoutSec int    `yaml:"dialTimeoutSec,omitempty" json:"dialTimeoutSec,omitempty"`
}

func (o *Options) Validate() error {
	o.Address = strings.TrimSpace(o.Address)
	o.Username = strings.TrimSpace(o.Username)
	o.PrivateKey = strings.TrimSpace(o.PrivateKey)
	o.PrivateKeyPath = strings.TrimSpace(o.PrivateKeyPath)
	o.RootPath = strings.TrimSpace(o.RootPath)

	if o.Address == "" {
		return fmt.Errorf("sftp: address is required")
	}
	if o.Username == "" {
		return fmt.Errorf("sftp: username is required")
	}
	if o.Password == "" && o.PrivateKey == "" && o.PrivateKeyPath == "" {
		return fmt.Errorf("sftp: password, privateKey, or privateKeyPath is required")
	}

	if o.Port <= 0 {
		o.Port = 22
	}

	// Normalize root path: remove trailing slash.
	if o.RootPath != "" {
		o.RootPath = strings.TrimRight(o.RootPath, "/")
	}

	return nil
}
