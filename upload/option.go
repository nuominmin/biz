package upload

// oss 配置
type ossConfig struct {
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

type options struct {
	ossConfig
}

type Option func(*options)

func newOptions(optFns ...Option) options {
	opts := options{}
	for _, opt := range optFns {
		opt(&opts)
	}
	return opts
}

// 设置 OSS 配置
func WithOssConfig(endpoint, accessKeyId, accessKeySecret, bucketName, baseUrl string) Option {
	return func(p *options) {
		p.endpoint = endpoint
		p.accessKeyId = accessKeyId
		p.accessKeySecret = accessKeySecret
		p.bucketName = bucketName
		p.baseUrl = baseUrl
	}
}
