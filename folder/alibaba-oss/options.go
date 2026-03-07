package alibaba_oss

type Options struct {
	Region          string `yaml:"region,omitempty" json:"region,omitempty"`
	Bucket          string `yaml:"bucket,omitempty" json:"bucket,omitempty"`
	AccessKeyID     string `yaml:"accessKeyID,omitempty" json:"accessKeyID,omitempty"`
	AccessKeySecret string `yaml:"accessKeySecret,omitempty" json:"accessKeySecret,omitempty"`
}
