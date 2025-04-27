package pagination

const (
	defaultPage = 1 // 默认页码
	minPage     = 1 // 最小页码

	defaultPageSize = 20  // 默认每页数量
	maxPageSize     = 500 // 最大每页数量
	minPageSize     = 1   // 最小每页数量
)

// PageReq 代表分页请求参数
type PageReq struct {
	Page     uint `form:"page" json:"page"`
	PageSize uint `form:"page_size" json:"page_size"`
}

// Sanitize 处理 PageSize 的默认值、最小值和最大值
func (p *PageReq) Sanitize(opts ...Option) {
	options := newOptions(opts...)

	if p.Page < minPage {
		p.Page = defaultPage
	}

	// 小于最小值则用默认值，没有填也是默认值
	if p.PageSize < minPageSize {
		p.PageSize = options.defaultPageSize
		return
	}

	// 最大值处理
	if p.PageSize > options.maxPageSize {
		p.PageSize = options.maxPageSize
		return
	}
}

// Offset 计算偏移量
func (p *PageReq) Offset() uint {
	if p.Page <= 0 || p.PageSize <= 0 {
		p.Sanitize()
	}

	return (p.Page - 1) * p.PageSize
}

func (p *PageReq) Default() {
	p.Page = defaultPage
	p.PageSize = defaultPageSize
}
