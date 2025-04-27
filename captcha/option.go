package captcha

import (
	"image/color"

	cp "github.com/mojocn/base64Captcha"
)

const (
	// 默认高度
	DEFAULT_HEIGHT = 60
	// 默认宽度
	DEFAULT_WIDTH = 240
	// 默认长度
	DEFAULT_LENGTH = 4
	// 默认字体
	DEFAULT_FONT = "Flim-Flam.ttf"
	// 默认源
	DEFAULT_SOURCE = "1234567890"
)

type Options struct {
	height          int
	width           int
	length          int
	noiseCount      int
	showLineOptions int // 显示线路选项
	source          string
	fonts           []string
	bgColor         *color.RGBA
	fontsStorage    cp.FontsStorage
}

type Option func(*Options)

func newOptions(opts ...Option) Options {
	options := Options{
		height:          DEFAULT_HEIGHT,
		width:           DEFAULT_WIDTH,
		length:          DEFAULT_LENGTH,
		noiseCount:      0,
		showLineOptions: cp.OptionShowHollowLine,
		fonts:           []string{DEFAULT_FONT},
		source:          DEFAULT_SOURCE,
		bgColor:         &color.RGBA{0, 0, 0, 0},
		fontsStorage:    cp.DefaultEmbeddedFonts,
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// 设置高度
func WithHeight(height int) Option {
	return func(p *Options) {
		p.height = height
	}
}

// 设置宽度
func WithWidth(width int) Option {
	return func(p *Options) {
		p.width = width
	}
}

// 设置长度
func WithLength(length int) Option {
	return func(p *Options) {
		p.length = length
	}
}

// 设置字体
func WithSource(source string) Option {
	return func(p *Options) {
		p.source = source
	}
}

// 设置源
func WithFonts(fonts ...string) Option {
	return func(p *Options) {
		p.fonts = fonts
	}
}

// 设置噪声计数
func WithNoiseCount(noiseCount int) Option {
	return func(p *Options) {
		p.noiseCount = noiseCount
	}
}

// 设置背景颜色
func WithBgColor(bgColor *color.RGBA) Option {
	return func(p *Options) {
		p.bgColor = bgColor
	}
}

// 显示空心线
func WithShowHollowLine() Option {
	return func(p *Options) {
		p.showLineOptions = cp.OptionShowHollowLine
	}
}

// 显示黏液线
func WithShowSlimeLine() Option {
	return func(p *Options) {
		p.showLineOptions = cp.OptionShowSlimeLine
	}

}

// 显示正弦线
func WithShowSineLine() Option {
	return func(p *Options) {
		p.showLineOptions = cp.OptionShowSineLine
	}
}

// 设置文件仓储
func WithFontStorage(fontsStorage cp.FontsStorage) Option {
	return func(p *Options) {
		p.fontsStorage = fontsStorage
	}
}
