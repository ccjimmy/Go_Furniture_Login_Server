//获取一个群的群成员。
package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"     //引入beego的orm
	_ "github.com/go-sql-driver/mysql" //引入beego的mysql驱动
)

type GroupMemberController struct {
	beego.Controller //Go 的嵌入方式，MainController 自动拥有了所有 beego.Controller 的方法
}

type groupMembers struct {
	Master  string
	Manager string
	Member  string
}

func (c *GroupMemberController) Get() {

	mems := c.GetGroupMember(c.GetString("gid"))
	fmt.Println("你好", (*mems).Member)
	jsons, _ := json.Marshal((*mems))

	fmt.Println("群成员的信息是：", string(jsons))
	c.Ctx.WriteString(string(jsons))
}

//获取自己基本信息
func (c *GroupMemberController) GetGroupMember(gid string) *groupMembers {
	o := orm.NewOrm()
	var mas string
	var man string
	var mes string
	sql := "select master,manager,member from groups where gid =?"

	o.Raw(sql, gid).QueryRow(&mas, &man, &mes)

	allMembers := &groupMembers{}
	allMembers.Master = mas
	allMembers.Manager = man
	allMembers.Member = mes

	return allMembers
}
