package controllers

import (
	//	"beegoHttp/models"

	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm" //引入beego的orm
)

type OfflineMsgController struct {
	beego.Controller //Go 的嵌入方式，MainController 自动拥有了所有 beego.Controller 的方法
}

const (
	PullOfflineMsg  = "0"
	ClearOfflineMsg = "1"
)

func (c *OfflineMsgController) Get() {
	protocol := c.GetString("protocol")
	if protocol == PullOfflineMsg {
		offlineMsg := c.getOfflineMsg(c.GetString("username"))
		fmt.Println("你的离线信息是：", offlineMsg)
		c.Ctx.WriteString(offlineMsg)
	} else {
		result := c.clearOfflineMsg(c.GetString("username"))
		if result == true {
			c.Ctx.WriteString("true")
		} else {
			c.Ctx.WriteString("false")
		}

	}

}

//获取离线信息
func (c *OfflineMsgController) getOfflineMsg(username string) string {
	o := orm.NewOrm()
	var offlineMsg string
	sql := "select offlinemsg from userdata where username =?"

	o.Raw(sql, username).QueryRow(&offlineMsg)
	return offlineMsg
}

//清除离线消息
func (c *OfflineMsgController) clearOfflineMsg(username string) bool {
	o := orm.NewOrm()
	sql := "update userdata set offlinemsg =? where username =?"
	_, err := o.Raw(sql, "[]", username).Exec()
	if err == nil {
		fmt.Println("离线数据已经清除")
		return true
	}
	return false
}
