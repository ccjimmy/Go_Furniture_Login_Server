package controllers

import (
	//	"beegoHttp/models"

	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm" //引入beego的orm
)

type FriendListController struct {
	beego.Controller //Go 的嵌入方式，MainController 自动拥有了所有 beego.Controller 的方法
}

func (c *FriendListController) Get() {
	result := c.pullFriendList(c.GetString("username"))
	c.Ctx.WriteString(result)
}

//获取离线信息
func (c *FriendListController) pullFriendList(username string) string {
	o := orm.NewOrm()
	var friends string
	sql := "select friends from userdata where username =?"

	err := o.Raw(sql, username).QueryRow(&friends)
	if err != nil {
		fmt.Println("查询好友列表出错：", err)
	}
	return friends
}
