package pagination

type Options struct {
	maxPageSize     uint
	defaultPageSize uint
}

type Option func(*Options)

func newOptions(opts ...Option) Options {
	options := Options{
		defaultPageSize: defaultPageSize,
		maxPageSize:     maxPageSize,
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// 设置每页数量默认值
func WithDefaultPageSize(pageSize uint) Option {
	return func(o *Options) {
		o.defaultPageSize = pageSize
	}
}

// 设置每页数量最大值
func WithMaxPageSize(pageSize uint) Option {
	return func(o *Options) {
		o.maxPageSize = pageSize
	}
}
