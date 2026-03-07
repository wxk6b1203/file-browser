package alibaba_oss

import "github.com/wxk6b1203/file-util-manager/folder"

type ossDriver struct{}

func NewDriver(opt *Options) (folder.Driver, error) {
	return &ossDriver{}, nil
}
