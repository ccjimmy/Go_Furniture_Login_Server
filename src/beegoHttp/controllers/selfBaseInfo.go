package controllers

import (
	"fmt"

	"github.com/astaxie/beego"
)

type SelfBaseInfoController struct {
	beego.Controller //Go 的嵌入方式，MainController 自动拥有了所有 beego.Controller 的方法
}

func (c *MainController) Get() {
	fmt.Println("请求自己基本信息")

	c.Ctx.WriteString("你的信息是：")
}
