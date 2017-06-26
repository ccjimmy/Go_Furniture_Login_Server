package controllers

import (
	"beegoHttp/models"
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
)

type FindFriendController struct {
	beego.Controller //Go 的嵌入方式，MainController 自动拥有了所有 beego.Controller 的方法
}

func (c *FindFriendController) Get() {

	var findFriend models.SelfBaseInfoModel

	friends := findFriend.FindFriends(c.GetString("username"))

	jsons, _ := json.Marshal(friends)

	fmt.Println("你的信息是：", string(jsons))
	c.Ctx.WriteString(string(jsons))
}
