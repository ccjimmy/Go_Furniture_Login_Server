package models

import (
	//"fmt"
	"fmt"

	"github.com/astaxie/beego/orm"     //引入beego的orm
	_ "github.com/go-sql-driver/mysql" //引入beego的mysql驱动
)

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
