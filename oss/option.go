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

// 设置 OSS endpoint
func WithEndpoint(endpoint string) Option {
	return func(p *options) {
		p.endpoint = endpoint
	}
}

// 设置 OSS access key id
func WithAccessKeyId(accessKeyId string) Option {
	return func(p *options) {
		p.accessKeyId = accessKeyId
	}
}

// 设置 OSS access key secret
func WithAccessKeySecret(accessKeySecret string) Option {
	return func(p *options) {
		p.accessKeySecret = accessKeySecret
	}
}

// 设置 OSS bucket name
func WithBucketName(bucketName string) Option {
	return func(p *options) {
		p.bucketName = bucketName
	}
}

// 设置 OSS base url for public access
func WithBaseUrl(baseUrl string) Option {
	return func(p *options) {
		p.baseUrl = baseUrl
	}
}
