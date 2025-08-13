package captcha

import (
	cp "github.com/mojocn/base64Captcha"
)

type Service interface {
	Verify(id, answer string) bool
	Generate() (string, string)
}

type captcha struct {
	c *cp.Captcha
}

// 创建字符串验证码实例
func NewCaptcha(optFns ...Option) Service {
	opts := newOptions(optFns...)
	driver := cp.NewDriverString(
		opts.height,
		opts.width,
		opts.noiseCount,
		opts.showLineOptions,
		opts.length,
		opts.source,
		opts.bgColor,
		opts.fontsStorage,
		opts.fonts,
	)
	return &captcha{cp.NewCaptcha(driver, cp.DefaultMemStore)}
}

// 验证是否有效
func (r *captcha) Verify(id, answer string) bool {
	get := cp.DefaultMemStore.Get(id, false)
	if get == "" {
		return false
	}
	return r.c.Verify(id, answer, true)
}

// 生成base64
func (r *captcha) Generate() (string, string) {
	id, b64s, _, _ := r.c.Generate()
	return id, b64s
}
