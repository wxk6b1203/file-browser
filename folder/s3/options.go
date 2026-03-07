package s3

type Options struct {
	Region          string `yaml:"region,omitempty" json:"region,omitempty"`
	Bucket          string `yaml:"bucket,omitempty" json:"bucket,omitempty"`
	AccessKeyID     string `yaml:"accessKeyID,omitempty" json:"accessKeyID,omitempty"`
	AccessKeySecret string `yaml:"accessKeySecret,omitempty" json:"accessKeySecret,omitempty"`
	SessionToken    string `yaml:"sessionToken,omitempty" json:"sessionToken,omitempty"`
	Endpoint        string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
}
