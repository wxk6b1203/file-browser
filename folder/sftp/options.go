package sftp

type Options struct {
	Address    string `yaml:"address,omitempty" json:"address,omitempty"`
	Port       int    `yaml:"port,omitempty" json:"port,omitempty"`
	Username   string `yaml:"username,omitempty" json:"username,omitempty"`
	Password   string `yaml:"password,omitempty" json:"password,omitempty"`
	PrivateKey string `yaml:"privateKey,omitempty" json:"privateKey,omitempty"`
	RootPath   string `yaml:"rootPath,omitempty" json:"rootPath,omitempty"`
}
