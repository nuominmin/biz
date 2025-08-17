package server

type options struct {
	allowedTypes map[string]struct{}
}

type Option func(*options)

func newOptions(optFns ...Option) options {
	opts := options{
		allowedTypes: make(map[string]struct{}),
	}
	for _, opt := range optFns {
		opt(&opts)
	}
	return opts
}

// 设置上传允许类型
func WithAllowedTypes(allowedTypes ...string) Option {
	return func(o *options) {
		for i := 0; i < len(allowedTypes); i++ {
			o.allowedTypes[allowedTypes[i]] = struct{}{}
		}
	}
}
