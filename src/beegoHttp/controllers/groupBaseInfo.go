package controllers

import (
	"ace"
	"beegoHttp/models"
	"encoding/json"
	"fmt"
	"game/data"
	"game/logic/msgMgr"
	"game/logic/protocol"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"     //引入beego的orm
	_ "github.com/go-sql-driver/mysql" //引入beego的mysql驱动
)

//群基本信息与功能
type GroupBaseInfoController struct {
	beego.Controller
}

func (c *GroupBaseInfoController) Get() {
	protocal := c.GetString("protocol")
	gid := c.GetString("gid")

	switch protocal {
	case protocol.GROUP_BASE_INFO: //获取一个群的基本信息
		c.GROUP_BASE_INFO(gid)
		break
	case protocol.MODIFY_GROUP_BASE_INFO: //修改一个群的基本信息
		c.MODIFY_GROUP_BASE_INFO(gid, c.GetString("name"), c.GetString("description"))
		break
	case protocol.MODIFY_ENTER_GROUP_METHOD: //修改一个群的加群方式
		c.MODIFY_ENTER_GROUP_METHOD(gid, c.GetString("method"))
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
	if result == "false" {
		c.Ctx.WriteString(result)
		return
	}
	c.Ctx.WriteString(result)
	//广播给所有群员
	group := msgMgr.GroupMgr.GetOneGroupManager(gid)
	//得到所有成员
	allMembers := group.Master + "," + group.Managers + "," + group.Members
	allMembersArr := strings.Split(allMembers, ",")
	for _, v := range allMembersArr {
		if v != "" { //得到每一个人
			memSe, ok := data.SyncAccount.AccountSession[v]
			if ok { //如果这个人在线
				memSe.Write(&ace.DefaultSocketModel{protocol.SETTING, -1, protocol.MODIFY_GROUP_INFO_SREQ, []byte(gid)})
			}
		}
	}
}

//修改入群方式
func (c *GroupBaseInfoController) MODIFY_ENTER_GROUP_METHOD(gid string, method string) {
	var groupBaseInfoModel models.GroupBaseInfoModel
	result := groupBaseInfoModel.ModifyEnterMethod(gid, method)
	if result == "false" {
		c.Ctx.WriteString(result)
		return
	}
	c.Ctx.WriteString(result)
	//广播给所有群员
	group := msgMgr.GroupMgr.GetOneGroupManager(gid)
	//得到所有成员
	allMembers := group.Master + "," + group.Managers + "," + group.Members
	allMembersArr := strings.Split(allMembers, ",")
	for _, v := range allMembersArr {
		if v != "" { //得到每一个人
			memSe, ok := data.SyncAccount.AccountSession[v]
			if ok { //如果这个人在线
				memSe.Write(&ace.DefaultSocketModel{protocol.SETTING, -1, protocol.MODIFY_GROUP_INFO_SREQ, []byte(gid)})
			}
		}
	}
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
	//广播给所有群员
	group := msgMgr.GroupMgr.GetOneGroupManager(gid)
	//得到所有成员
	allMembers := group.Master + "," + group.Managers + "," + group.Members
	allMembersArr := strings.Split(allMembers, ",")
	for _, v := range allMembersArr {
		if v != "" { //得到每一个人
			memSe, ok := data.SyncAccount.AccountSession[v]
			if ok { //如果这个人在线
				memSe.Write(&ace.DefaultSocketModel{protocol.SETTING, -1, protocol.MODIFY_GROUP_FACE_SREQ, []byte(gid)})
			}
		}
	}
}
