package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/astaxie/beego/orm"     //引入beego的orm
	_ "github.com/go-sql-driver/mysql" //引入beego的mysql驱动
)

//vshowModel
type VShowModel struct {
	Username      string
	Name          string
	Productcode   string
	Provider      string
	Style         int
	Createdtime   string
	Lastsavedtime string
	Version       string
}

//创建展厅的数据结构
type CreateHouseData struct {
	Username   string //房子所有者
	Name       string //所要保存的展厅的名字
	Provider   string //供应商名字
	Lib        int    //库别
	Style      int    //风格
	Furnitures string
	Settings   string
	Walls      string
}

const (
	//我的展厅每页个数
	MyVshowListPageAmount = 8
	//每页展厅的个数
	AllVshowListPageAmount = 10
)

//创建房间
func (this *VShowModel) CreatRoom(info string) string {
	//解析json 获取用户名及 展厅名
	createData := &CreateHouseData{}
	err := json.Unmarshal([]byte(info), &createData)
	if err != nil {
		fmt.Println("json解析失败")
		return "fail"
	}
	//	fmt.Println(createData.RoomUid, createData.RoomName, createData.Provider)
	//操作数据库
	o := orm.NewOrm()
	var isExit string
	o.Raw("SELECT name FROM virtualshow WHERE username = ? and name =?", createData.Username, createData.Name).QueryRow(&isExit)
	if isExit == createData.Name {
		fmt.Printf("这个房间名字已被注册")
		return "exist"
	} else {
		//插入新房子数据
		_, sqlerr := o.Raw("INSERT virtualshow SET username=?,name=?,provider=?,productcode=?,lib=?,style=?,point=?,yz=?,wall=?,furniture=?,setting=?,createdtime=?,lastsavedtime=?,version=?",
			createData.Username, createData.Name, createData.Provider, "", 1, createData.Style, info, "", "", "", "", time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"), 0).Exec()
		if sqlerr != nil {
			fmt.Println(sqlerr) //可以打印出错误
			return "fail"
		}
		return "succeed"
	}
	return "fail"
}

//我的展厅列表
func (this *VShowModel) MyVshowList(userName string, page string) ([]VShowModel, int) {

	//操作数据库
	o := orm.NewOrm()
	var vshows []VShowModel
	//计算数量
	sql := "select username from virtualshow where username = ?"
	o.Raw(sql, userName).QueryRows(&vshows)
	amount := len(vshows)
	vshows = nil
	//前8条数据
	sql2 := "SELECT username,name,productcode,provider,style,createdtime,lastsavedtime,version FROM virtualshow WHERE username =? Limit ?,?"
	intPage, _ := strconv.Atoi(page) //获得页数
	_, err := o.Raw(sql2, userName, intPage*MyVshowListPageAmount, MyVshowListPageAmount).QueryRows(&vshows)
	if err != nil {
		fmt.Println(err)
	}
	return vshows, amount
}

//删除一个房间
func (this *VShowModel) DeleteVshow(uid string, name string) bool {
	o := orm.NewOrm()
	sql := "delete from virtualshow where username =? and name =?"
	_, err := o.Raw(sql, uid, name).Exec()
	if err != nil {
		fmt.Println(err)
		return false
	} else {
		return true
	}
	return false
}

//修改一个房间
func (this *VShowModel) MotifyVshow(uid string, oldName string, newName string, newStyle string) bool {
	o := orm.NewOrm()
	if newName != oldName { //在改名字
		//判断自己是否已经有这个房间名字了
		var isExit string
		o.Raw("SELECT name FROM virtualshow WHERE username = ? and name =?", uid, newName).QueryRow(&isExit)
		if isExit == newName {
			fmt.Println("这个房间名字已被注册")
			return false
		}
	}
	//不是改名字 而是改属性
	sql := "update virtualshow set name=?, style=? where username=? and name =?"
	res, err := o.Raw(sql, newName, newStyle, uid, oldName).Exec()
	if err != nil {
		fmt.Println(err)
	}
	count, _ := res.RowsAffected()
	if count == 1 {
		return true
	} else {
		return false
	}
}

//获取4场景展厅列表
func (this *VShowModel) VShowList(provider string, lib string, style string, page string) ([]VShowModel, int) {
	o := orm.NewOrm()
	//1、计算满足需求的房间个数
	var vshows []VShowModel
	numsql := "SELECT id FROM virtualshow WHERE "
	this.prepareSql(&numsql, lib, style)
	fmt.Println("查询个数的语句", numsql)

	o.Raw(numsql, lib, provider, style, 0, 100000).QueryRows(&vshows)
	fmt.Println("查询参数", numsql, lib, provider, style, 0, 100000)
	amount := len(vshows)
	vshows = nil
	//2、获取10条具体信息
	sql := "SELECT username,name,productcode,provider,createdtime,lastsavedtime,version FROM virtualshow WHERE "
	this.prepareSql(&sql, lib, style)
	//fmt.Println("展厅列表sql语句", sql)
	intPage, _ := strconv.Atoi(page) //获得页数
	o.Raw(sql, lib, provider, style, intPage*AllVshowListPageAmount, AllVshowListPageAmount).QueryRows(&vshows)

	return vshows, amount
}

func (this *VShowModel) prepareSql(ori *string, lib string, style string) {
	//对库做出区别
	if lib == "0" { //共享库
		*ori += " lib =?"
		*ori += " and provider !=?"
	} else { //公司库
		*ori += " lib <= ?"
		*ori += " and provider = ?"
	}
	//对风格做出区别
	if style == "0" { //不做区别
		*ori += " and style >= ?"
	} else {
		*ori += " and style = ?"
	}
	//页数处理
	*ori += " limit ?,?"
}
