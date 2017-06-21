package models

import (
	"fmt"

	"github.com/astaxie/beego/orm"     //引入beego的orm
	_ "github.com/go-sql-driver/mysql" //引入beego的mysql驱动
)

//表go_archives的结构
type UploadTypes struct {
	//	Id       int
	Maintype string
	Name     string
}

//单品家具的模型
type FurnitureModel struct {
	Id          int
	Lib         int
	Style       int
	Color       int
	Mat         int
	Brand       string
	Maintype    string
	Secondtype  string
	Thirdtype   string
	Name        string
	Provider    string
	Productcode string
	Price       string
	Length      string
	Width       string
	Height      string
	Version     int
}

const (
	MaxTexPageAmount = 10 //我的贴图列表
)

func init() {
	//orm.RegisterDriver("mysql", orm.DR_MySQL)                                //注册数据库驱动
	/*	orm.RegisterDataBase("default", "mysql", "root:@/furniture?charset=utf8") //注册一个别名为default的数据库
		orm.SetMaxIdleConns("default", 30)                                        //设置数据库最大空闲连接
		orm.SetMaxOpenConns("default", 30)        */ //设置数据库最大连接数
	//orm.RegisterModelWithPrefix("go_", new(UploadTypes)) //注册模型并使用表前缀
	//orm.RegisterModelWithPrefix("go_", new(Archives))                         //注册模型并使用表前缀
}

func (this *UploadTypes) GetUploadTypes() []UploadTypes {

	//旧的写法
	//	o := orm.NewOrm()
	//	var data []UploadTypes
	//	//_, err := o.QueryTable("maintype").All(&data)
	//	//return data, err
	//	sql := "select maintype,secondtype,thirdtype,name from thirdtype where maintype=00 and (secondtype=02 or secondtype=03)"
	//	_, err := o.Raw(sql).QueryRows(&data)
	var types []UploadTypes
	for _, v := range ShopManager.shopTypes {
		if v.Parent == "12" || v.Parent == "13" { //12、13分别是地面和墙面
			data := &UploadTypes{v.Index, v.TypeName}
			types = append(types, *data)
		}
	}
	return types
}

//获得一个人的所有贴图
func (this *FurnitureModel) GetMyTextures(provider string) ([]FurnitureModel, int, error) {

	o := orm.NewOrm()
	var data []FurnitureModel
	sql := "select * from commodity where provider = ? and productcode like 'TEX_%'"
	o.Raw(sql, provider).QueryRows(&data)
	amount := len(data)
	data = nil
	//获取前十条
	sql2 := "select * from commodity where provider = ? and productcode like 'TEX_%' limit ?"
	_, err := o.Raw(sql2, provider, MaxTexPageAmount).QueryRows(&data)

	return data, amount, err
}

//修改贴图数据
func (this *FurnitureModel) MotifyTextures(oldName string, newName string, price string) bool {
	o := orm.NewOrm()
	if newName != oldName { //在改名字
		//判断是否有新名字的产品已存在
		var isExit string
		o.Raw("SELECT name FROM commodity WHERE name =?", newName).QueryRow(&isExit)
		if isExit == newName {
			fmt.Printf("这个商品名字已被注册")
			return false
		}
	}

	sql := "update commodity set name=?,price=? where name =?"
	_, err := o.Raw(sql, newName, price, oldName).Exec()
	if err != nil {
		fmt.Println(err)
		return false
	} else {
		return true
	}
	return false
}

//删除贴图数据
func (this *FurnitureModel) DeleteTextures(Name string) bool {
	o := orm.NewOrm()

	sql := "delete from commodity where name =?"
	_, err := o.Raw(sql, Name).Exec()
	if err != nil {
		fmt.Println(err)
		return false
	} else {
		return true
	}
	return false
}

//贴图列表翻页
func (this *FurnitureModel) PageTextures(provider string, page int) ([]FurnitureModel, error) {
	o := orm.NewOrm()
	var data []FurnitureModel

	sql := "select * from commodity where provider = ? and productcode like 'TEX_%' limit ? , ?"
	_, err := o.Raw(sql, provider, page*MaxTexPageAmount, MaxTexPageAmount).QueryRows(&data)
	return data, err
}

////表go_archives的增加
//func (this *Archives) Add(title, body string, typeid int) (int64, error) {
//	o := orm.NewOrm()
//	arc := Archives{Title: title, Body: body, Typeid: typeid}
//	id, err := o.Insert(&arc)
//	return id, err
//}

//func (this *Archives) Edit(title, body string, typeid, id int) error {
//	o := orm.NewOrm()
//	arc := Archives{Title: title, Body: body, Typeid: typeid, Id: id}
//	_, err := o.Update(&arc)
//	return err
//}

//func (this *Archives) Delete(id int) error {
//	o := orm.NewOrm()
//	arc := Archives{Id: id}
//	_, err := o.Delete(&arc)
//	return err
//}
