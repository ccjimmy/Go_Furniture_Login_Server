package models

import (
	"fmt"
	//"fmt"

	"github.com/astaxie/beego/orm"     //引入beego的orm
	_ "github.com/go-sql-driver/mysql" //引入beego的mysql驱动
)

//个人的模型
type SelfBaseInfoModel struct {
	Username    string
	Nickname    string
	Face        string
	Description string
}

//获取自己基本信息
func (this *SelfBaseInfoModel) GetPersonBaseInfoModel(username string) SelfBaseInfoModel {
	o := orm.NewOrm()
	var data SelfBaseInfoModel
	sql := "select username,nickname,face,description from userinfo where username =?"

	o.Raw(sql, username).QueryRow(&data)
	return data
}

//修改自己的个人资料
func (this *SelfBaseInfoModel) ModifyPersonalInfoModel(username string, nickname string, description string) string {
	o := orm.NewOrm()
	sql := "update userinfo set nickname=?, description=? where username=?"
	_, err := o.Raw(sql, nickname, description, username).Exec()
	if err != nil {
		fmt.Println(err)
		return "false"
	}
	return "true"
}

//查找好友 根据账号或昵称
func (this *SelfBaseInfoModel) FindFriends(username string) []SelfBaseInfoModel {
	o := orm.NewOrm()
	var data []SelfBaseInfoModel
	sql := "select username,nickname,face,description from userinfo where username = ? or nickname like ?"

	_, err := o.Raw(sql, username, "%"+username+"%").QueryRows(&data)
	if err != nil {
		fmt.Println(err)
	}
	return data
}
