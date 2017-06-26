package controllers

import (
	"beegoHttp/models"
	"encoding/json"
	//"fmt"

	"github.com/astaxie/beego"
)

type BaseInfoController struct {
	beego.Controller //Go 的嵌入方式，MainController 自动拥有了所有 beego.Controller 的方法
}

func (c *BaseInfoController) Get() {
	var SelfBaseInfoModel models.SelfBaseInfoModel
	selfInfo := SelfBaseInfoModel.GetPersonBaseInfoModel(c.GetString("username"))

	jsons, _ := json.Marshal(selfInfo)

	//fmt.Println("你的信息是：", string(jsons))
	c.Ctx.WriteString(string(jsons))
}
