package controllers

import (
	"fmt"

	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller //Go 的嵌入方式，MainController 自动拥有了所有 beego.Controller 的方法
}

func (c *MainController) Get() {
	fmt.Println("来了")
	c.Data["Website"] = "beego.me"
	c.Data["Email"] = "astaxie@gmail.com"
	//c.TplName = "index.tpl" //(文件、文件夹必须小写)

	c.Ctx.WriteString("哈哈哈哈哈哈")
}
