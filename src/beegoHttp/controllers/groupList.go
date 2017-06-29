package controllers

import (
	//	"beegoHttp/models"

	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm" //引入beego的orm
)

type GroupListController struct {
	beego.Controller
}

func (c *GroupListController) Get() {
	result := c.pullGroupList(c.GetString("username"))
	c.Ctx.WriteString(result)
}

//获取离线信息
func (c *GroupListController) pullGroupList(username string) string {
	o := orm.NewOrm()
	var groups string
	sql := "select groups from userdata where username =?"

	err := o.Raw(sql, username).QueryRow(&groups)
	if err != nil {
		fmt.Println("查询群组列表出错：", err)
	}
	return groups
}
