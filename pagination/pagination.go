package pagination

const (
	defaultPage = 1 // 默认页码
	minPage     = 1 // 最小页码

	defaultPageSize = 20  // 默认每页数量
	maxPageSize     = 500 // 最大每页数量
	minPageSize     = 1   // 最小每页数量
)

// 代表分页请求参数
type PageReq struct {
	Page     uint `form:"page" json:"page"`
	PageSize uint `form:"page_size" json:"page_size"`
}

// Sanitize 对分页请求参数进行合法性校验与修正
// 根据传入的可选参数（Option）来设置默认页大小和最大页大小，
// 并对 Page 和 PageSize 参数进行以下处理：
// 1. 如果 Page 小于允许的最小页码（minPage），则设置为默认页码（defaultPage）
// 2. 如果 PageSize 小于允许的最小页大小（minPageSize），则设置为默认页大小（defaultPageSize）
// 3. 如果 PageSize 大于允许的最大页大小（maxPageSize），则设置为最大页大小（maxPageSize）
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

// 计算偏移量
func (p *PageReq) Offset() uint {
	if p.Page <= 0 || p.PageSize <= 0 {
		p.Sanitize()
	}

	return (p.Page - 1) * p.PageSize
}
