package controllers

import (
	"beegoHttp/models"
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"     //引入beego的orm
	_ "github.com/go-sql-driver/mysql" //引入beego的mysql驱动
)

//协议
const (
	GROUP_BASE_INFO        = "1" //获取一个群的基本信息
	MODIFY_GROUP_BASE_INFO = "2" //修改一个群的基本信息
)

//群基本信息与功能
type GroupBaseInfoController struct {
	beego.Controller
}

func (c *GroupBaseInfoController) Get() {
	protocal := c.GetString("protocol")
	gid := c.GetString("gid")

	switch protocal {
	case GROUP_BASE_INFO: //获取一个群的基本信息
		c.GROUP_BASE_INFO(gid)
		break
	case MODIFY_GROUP_BASE_INFO: //修改一个群的基本信息
		c.MODIFY_GROUP_BASE_INFO(gid, c.GetString("name"), c.GetString("description"))
		break
	default:
		fmt.Println("未知群http协议")
		break
	}

}

//请求一个群的基本数据
func (c *GroupBaseInfoController) GROUP_BASE_INFO(gid string) {
	var groupBaseInfoModel models.GroupBaseInfoModel
	groupInfo := groupBaseInfoModel.GetGroupBaseInfoModel(gid)
	jsons, _ := json.Marshal(groupInfo)
	fmt.Println("群的信息是：", string(jsons))
	c.Ctx.WriteString(string(jsons))
}

//修改一个群的信息
func (c *GroupBaseInfoController) MODIFY_GROUP_BASE_INFO(gid string, name string, description string) {
	var groupBaseInfoModel models.GroupBaseInfoModel
	result := groupBaseInfoModel.ModifyGroupInfoModel(gid, name, description)
	c.Ctx.WriteString(result)
}

//上传群头像
func (c *GroupBaseInfoController) Post() {
	//关于数据
	gid := c.GetString("gid")
	fmt.Println("头像的上传者是", gid)
	//获取文件
	f, _, _ := c.GetFile("face")
	//fmt.Println("读到文件", h.Filename)
	f.Close()                                         //关闭上传的文件，不然的话会出现临时文件不能清除的情况
	c.SaveToFile("face", "res/face/group"+gid+".jpg") //存文件
	//修改数据库内容
	o := orm.NewOrm()
	sql := "update groups set face=? where gid=?"
	_, err := o.Raw(sql, "group"+gid+".jpg", gid).Exec()
	if err != nil {
		fmt.Println(err)
		c.Ctx.WriteString("false")
		return
	}
	c.Ctx.WriteString("true")
}
