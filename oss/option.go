package oss

type options struct {
	// OSS endpoint
	endpoint string
	// OSS access key id
	accessKeyId string
	// OSS access key secret
	accessKeySecret string
	// OSS bucket name
	bucketName string
	// OSS base url for public access
	baseUrl string
}

type Option func(*options)

func newOptions(optFns ...Option) options {
	opts := options{}
	for _, opt := range optFns {
		opt(&opts)
	}
	return opts
}
