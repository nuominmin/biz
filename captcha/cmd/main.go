package main

import (
	"fmt"
	"strings"

	"github.com/nuominmin/biz/captcha"
)

func main() {
	cap := captcha.NewCaptcha()

	// 生成验证码
	id, b64s := cap.Generate()
	fmt.Println("验证码 ID：", id)

	// 打印提示
	fmt.Println("请在浏览器打开以下链接，查看验证码图片内容：")
	fmt.Printf("%s\n", b64s)

	// 让用户手动输入验证码
	var answer string
	fmt.Print("请输入验证码内容：")
	fmt.Scanln(&answer)

	// 验证输入是否正确
	if !cap.Verify(id, strings.TrimSpace(answer)) {
		fmt.Println("验证码验证失败！")
	} else {
		fmt.Println("验证码验证成功 ✅")
	}
}
