package server

import (
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/nuominmin/biz/captcha"
	"github.com/nuominmin/biz/krs/types"
)

// 验证码
func (s *service) Captcha(captchaSvc captcha.Service) func(http.Context) error {
	return func(ctx http.Context) error {
		id, b64s := captchaSvc.Generate()
		data := map[string]interface{}{
			"id":    id,
			"image": b64s,
		}
		return ctx.JSON(200, types.NewSuccessResponse(data))
	}
}
