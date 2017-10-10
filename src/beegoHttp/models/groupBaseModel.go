package models

import (
	"fmt"
	//"game/logic/msgMgr"

	"github.com/astaxie/beego/orm"     //引入beego的orm
	_ "github.com/go-sql-driver/mysql" //引入beego的mysql驱动
)

//群基本模型
type GroupBaseInfoModel struct {
	Gid         int
	Name        string
	Description string
	Face        string
	Level       int
	Master      string
	Manager     string
	Member      string
	Verifymode  int
	Createdtime string
}

//获取群基本信息
func (this *GroupBaseInfoModel) GetGroupBaseInfoModel(gid string) GroupBaseInfoModel {
	o := orm.NewOrm()
	var data GroupBaseInfoModel
	sql := "select gid,name,description,face,level,master,manager,member,verifymode,createdtime from groups where gid =?"

	o.Raw(sql, gid).QueryRow(&data)
	return data
}

//修改群资料
func (this *GroupBaseInfoModel) ModifyGroupInfoModel(gid string, nickname string, description string) string {
	o := orm.NewOrm()
	sql := "update groups set name=?, description=? where gid=?"
	_, err := o.Raw(sql, nickname, description, gid).Exec()
	if err != nil {
		fmt.Println(err)
		return "false"
	}
	return "true"
}

//修改入群方式
func (this *GroupBaseInfoModel) ModifyEnterMethod(gid string, method string) string {
	fmt.Print("你好好", method)
	o := orm.NewOrm()
	sql := "update groups set verifymode=? where gid=?"
	_, err := o.Raw(sql, method, gid).Exec()
	if err != nil {
		fmt.Println(err)
		return "false"
	}
	return "true"
}
