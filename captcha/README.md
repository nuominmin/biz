# captcha

## 例子
```go
// 初始化
func NewCaptcha() captcha.Service {
    return captcha.NewCaptcha()
}

// 验证码接口
srv.Route("/api").GET("/captcha", func(ctx http.Context) error {
    id, b64s := captchaSvc.Generate()
        return ctx.JSON(200, map[string]interface{}{
            "code":    0,
            "message": "success",
            "data": map[string]interface{}{
            "id":    id,
            "image": b64s,
        },
    })
})

// 验证码检查
if !s.captchaSvc.Verify(req.CaptchaId, req.CaptchaCode) {
    return &pb.LoginResponse{}, errors.New("验证码错误")
}

```