package models

import (
	"fmt"
	//"fmt"

	"github.com/astaxie/beego/orm"     //引入beego的orm
	_ "github.com/go-sql-driver/mysql" //引入beego的mysql驱动
)

//单品家具的模型
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

////获得一个人的所有贴图
//func (this *FurnitureModel) GetMyTextures(provider string) ([]FurnitureModel, int, error) {

//	o := orm.NewOrm()
//	var data []FurnitureModel
//	sql := "select * from commodity where provider = ? and productcode like 'TEX_%'"
//	o.Raw(sql, provider).QueryRows(&data)
//	amount := len(data)
//	data = nil
//	//获取前十条
//	sql2 := "select * from commodity where provider = ? and productcode like 'TEX_%' limit ?"
//	_, err := o.Raw(sql2, provider, MaxTexPageAmount).QueryRows(&data)

//	return data, amount, err
//}

////修改贴图数据
//func (this *FurnitureModel) MotifyTextures(oldName string, newName string, price string) bool {
//	o := orm.NewOrm()
//	if newName != oldName { //在改名字
//		//判断是否有新名字的产品已存在
//		var isExit string
//		o.Raw("SELECT name FROM commodity WHERE name =?", newName).QueryRow(&isExit)
//		if isExit == newName {
//			fmt.Printf("这个商品名字已被注册")
//			return false
//		}
//	}

//	sql := "update commodity set name=?,price=? where name =?"
//	_, err := o.Raw(sql, newName, price, oldName).Exec()
//	if err != nil {
//		fmt.Println(err)
//		return false
//	} else {
//		return true
//	}
//	return false
//}

////删除贴图数据
//func (this *FurnitureModel) DeleteTextures(Name string) bool {
//	o := orm.NewOrm()

//	sql := "delete from commodity where name =?"
//	_, err := o.Raw(sql, Name).Exec()
//	if err != nil {
//		fmt.Println(err)
//		return false
//	} else {
//		return true
//	}
//	return false
//}

////贴图列表翻页
//func (this *FurnitureModel) PageTextures(provider string, page int) ([]FurnitureModel, error) {
//	o := orm.NewOrm()
//	var data []FurnitureModel

//	sql := "select * from commodity where provider = ? and productcode like 'TEX_%' limit ? , ?"
//	_, err := o.Raw(sql, provider, page*MaxTexPageAmount, MaxTexPageAmount).QueryRows(&data)
//	return data, err
//}
