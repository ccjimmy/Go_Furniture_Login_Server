package controllers

import (
	"beegoHttp/models"
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"     //引入beego的orm
	_ "github.com/go-sql-driver/mysql" //引入beego的mysql驱动
)

type BaseInfoController struct {
	beego.Controller //Go 的嵌入方式，MainController 自动拥有了所有 beego.Controller 的方法
}

//协议
const (
	BASE_INFO        = "1" //获取一个人基本信息
	MODIFY_BASE_INFO = "2" //修改一个人的基本信息
)

func (c *BaseInfoController) Get() {
	protocal := c.GetString("protocol")
	username := c.GetString("username")

	switch protocal {
	case BASE_INFO: //获取一个人基本信息
		c.BASE_INFO(username)
		break
	case MODIFY_BASE_INFO: //修改一个人的基本信息
		nickname := c.GetString("nickname")
		desc := c.GetString("description")
		c.MODIFY_BASE_INFO(username, nickname, desc)
		break
	default:
		fmt.Println("未知个人http协议")
		break
	}
}

//获取一个人基本信息
func (c *BaseInfoController) BASE_INFO(username string) {
	var SelfBaseInfoModel models.SelfBaseInfoModel
	selfInfo := SelfBaseInfoModel.GetPersonBaseInfoModel(username)
	jsons, _ := json.Marshal(selfInfo)
	//fmt.Println("你的信息是：", string(jsons))
	c.Ctx.WriteString(string(jsons))
}

//修改一个人的基本信息
func (c *BaseInfoController) MODIFY_BASE_INFO(username string, nickname string, desc string) {
	var SelfBaseInfoModel models.SelfBaseInfoModel

	result := SelfBaseInfoModel.ModifyPersonalInfoModel(username, nickname, desc)
	c.Ctx.WriteString(result)
}

//上传头像
func (c *BaseInfoController) Post() {
	//关于数据
	username := c.GetString("username")
	fmt.Println("头像的上传者是", username)
	//uptime := c.GetString("time")

	//获取文件
	f, _, _ := c.GetFile("face")
	//fmt.Println("读到文件", h.Filename)
	f.Close()                                               //关闭上传的文件，不然的话会出现临时文件不能清除的情况
	c.SaveToFile("face", "res/face/friend"+username+".jpg") //存文件
	//修改数据库内容
	o := orm.NewOrm()
	sql := "update userinfo set face=? where username=?"
	_, err := o.Raw(sql, "friend"+username+".jpg", username).Exec()
	if err != nil {
		fmt.Println(err)
		c.Ctx.WriteString("false")
		return
	}

	c.Ctx.WriteString("true")

}
