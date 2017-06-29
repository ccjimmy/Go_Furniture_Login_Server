package models

import (
	//"fmt"
	//"fmt"

	"github.com/astaxie/beego/orm"     //引入beego的orm
	_ "github.com/go-sql-driver/mysql" //引入beego的mysql驱动
)

type GroupBaseInfoModel struct {
	Gid         int
	Name        string
	Face        string
	Level       int
	Master      string
	Manager     string
	Member      string
	Verifymode  int
	Createdtime string
}

//获取群基本信息
func (this *GroupBaseInfoModel) GetGroupBaseInfoModel(gid int) GroupBaseInfoModel {
	o := orm.NewOrm()
	var data GroupBaseInfoModel
	sql := "select gid,name,face,level,master,manager,member,verifymode,createdtime from groups where gid =?"

	o.Raw(sql, gid).QueryRow(&data)
	return data
}
