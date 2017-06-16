package data

import (
	"ace"
	"database/sql"
	"encoding/json"
	"fmt"
	"game/logic/protocol"
	"time"
	"tools"
)

type HouseHandler struct {
}

var House = &HouseHandler{}

//协议
const (
	CREAT_HOUSE_BASICS_INFO_CREQ = 6 //创建户型
	CREAT_HOUSE_BASICS_INFO_SREQ = 7

	HISTORY_LIST_CREQ = 4 //获取户型列表
	HISTORY_LIST_SREQ = 5

	SAVE_VSHOW_FURNITURE_CREQ = 2 //保存家具数据
	SAVE_VSHOW_FURNITURE_SREQ = 3

	GET_VSHOW_FURNITURE_CREQ = 8 //请求房间内家具数据
	GET_VSHOW_FURNITURE_SREQ = 9

	SAVE_VSHOW_WALL_CREQ = 12 //保存墙体数据
	SAVE_VSHOW_WALL_SREQ = 13

	GET_VSHOW_WALL_CREQ = 14 //获取墙体数据
	GET_VSHOW_WALL_SREQ = 15

	SAVE_VSHOW_SETTING_CREQ = 16 //保存展厅设置信息
	SAVE_VSHOW_SETTING_SREQ = 17

	GET_VSHOW_SETTING_CREQ = 18 //获取展厅设置信息
	GET_VSHOW_SETTING_SREQ = 19
)

//创建房间的数据结构
type CreatHouseData struct {
	Uid         string //房子所属账号
	Name        string //房子的名字
	ProductCode string //此值不空，代表下载静态展厅
	IsVShow     int    //是否是一个厂家创建的展厅
	State       int    //是否对外显示
	Point       string
	Yz          string
	Furniture   string
}

//保存展厅的数据结构
type SaveHouseData struct {
	RoomName string //所要保存的展厅的名字
	RoomUid  string //房子所有者
}

//逻辑处理
func (this *HouseHandler) Process(session *ace.Session, message ace.DefaultSocketModel) {
	switch message.Command {
	case CREAT_HOUSE_BASICS_INFO_CREQ: //创建户型
		this.CreatRoom(session, message)
		break
	case HISTORY_LIST_CREQ: //获取历史列表
		this.GetRooms(session, message)
		break
	case SAVE_VSHOW_FURNITURE_CREQ: //保存房间家具数据
		this.SaveVshowFurs(session, message)
		break
	case GET_VSHOW_FURNITURE_CREQ: //请求房间里的家具
		this.GetVSHOWFurs(session, message)
		break
	case SAVE_VSHOW_SETTING_CREQ: //保存房间设置信息
		this.saveVshowSetting(session, message)
		break
	case GET_VSHOW_SETTING_CREQ: //请求房间设置信息
		this.getVshowSetting(session, message)
		break
	case SAVE_VSHOW_WALL_CREQ: //保存墙体数据
		this.saveVshowWall(session, message)
		break
	case GET_VSHOW_WALL_CREQ: //请求墙体数据
		this.getVshowWall(session, message)
		break
	default:
		fmt.Println("未知用户信息协议类型！")
		break
	}
}

//申请创建房间
func (this *HouseHandler) CreatRoom(session *ace.Session, message ace.DefaultSocketModel) {
	fmt.Println("创建房间" + string(message.Message))
	//解开json
	creatData := &CreatHouseData{}
	err := json.Unmarshal(message.Message, &creatData)
	tools.CheckErr(err)
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	tools.CheckErr(err)
	//先判断房子名字是否已存在，就不能插入数据
	stmtOut, err := db.Prepare("SELECT name FROM virtualshow WHERE uid = ? and name =?")
	var houseName string
	err = stmtOut.QueryRow(creatData.Uid, creatData.Name).Scan(&houseName)
	tools.CheckErr(err)
	//fmt.Printf("The square is: %s", username)
	if houseName == creatData.Name {
		fmt.Printf("这个房间名字已被注册")
		session.Write(&ace.DefaultSocketModel{protocol.USER, -1, CREAT_HOUSE_BASICS_INFO_SREQ, []byte("false")})
	} else {
		//插入新房子数据
		stmt, err := db.Prepare("INSERT virtualshow SET uid=?,name=?, isvshow=?,state=?,point=?,yz=?,furniture=?,createdtime=?")
		tools.CheckErr(err)
		_, err = stmt.Exec(creatData.Uid, creatData.Name, creatData.IsVShow, creatData.State, creatData.Point, creatData.Yz, creatData.Furniture, time.Now().Format("2006-01-02 15:04:05"))
		tools.CheckErr(err)
		session.Write(&ace.DefaultSocketModel{protocol.HOUSE, -1, CREAT_HOUSE_BASICS_INFO_SREQ, []byte("true")})
	}
}

//获取房间历史列表
func (this *HouseHandler) GetRooms(session *ace.Session, message ace.DefaultSocketModel) {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	tools.CheckErr(err)
	user := SyncAccount.SessionAccount[session]
	if user != "" {
		stmtOut, err := db.Prepare("SELECT rooms FROM userdata WHERE uid = ?")
		tools.CheckErr(err)
		var rooms string
		err = stmtOut.QueryRow(user).Scan(&rooms)
		tools.CheckErr(err)
		//响应
		session.Write(&ace.DefaultSocketModel{protocol.HOUSE, -1, HISTORY_LIST_SREQ, []byte(rooms)})
		fmt.Println("获取户型:", rooms)
	} else {
		fmt.Println("没有这个用户")
	}
}

//保存家具数据
func (this *HouseHandler) SaveVshowFurs(session *ace.Session, message ace.DefaultSocketModel) {
	//错误处理
	defer func() {
		if r := recover(); r != nil {
			//有错误的话将返回"err"
			session.Write(&ace.DefaultSocketModel{protocol.HOUSE, -1, SAVE_VSHOW_FURNITURE_SREQ, []byte("err")})
		}
	}()
	//解析json 获取展厅名字
	saveData := &SaveHouseData{}
	err := json.Unmarshal(message.Message, &saveData)
	tools.CheckErr(err)
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	tools.CheckErr(err)
	user := SyncAccount.SessionAccount[session]
	if user != "" {
		stmtUp, err := db.Prepare("update virtualshow set furniture=? where uid=? and name=?")
		tools.CheckErr(err)
		_, err = stmtUp.Exec(string(message.Message), user, saveData.RoomName) //保存户型数据
		tools.CheckErr(err)
		//响应
		session.Write(&ace.DefaultSocketModel{protocol.HOUSE, -1, SAVE_VSHOW_FURNITURE_SREQ, []byte("true")})
		fmt.Println("保存家具数据:", len(string(message.Message)))
	} else {
		fmt.Println("一个空用户正在保存户型，这不该发生！")
	}
}

//请求家具数据
func (this *HouseHandler) GetVSHOWFurs(session *ace.Session, message ace.DefaultSocketModel) {
	//错误处理
	defer func() {
		if r := recover(); r != nil {
			//有错误的话将返回"err"
			session.Write(&ace.DefaultSocketModel{protocol.HOUSE, -1, GET_VSHOW_FURNITURE_SREQ, []byte("err")})
		}
	}()
	//解析json 获取展厅名字
	furData := &SaveHouseData{}
	err := json.Unmarshal(message.Message, &furData)
	tools.CheckErr(err)
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	tools.CheckErr(err)
	stmtOut, err := db.Prepare("select furniture from virtualshow where uid=? and name=?")
	tools.CheckErr(err)
	var furs string
	err = stmtOut.QueryRow(furData.RoomUid, furData.RoomName).Scan(&furs) //参数是：所有者及房子名字
	tools.CheckErr(err)
	//响应
	session.Write(&ace.DefaultSocketModel{protocol.HOUSE, -1, GET_VSHOW_FURNITURE_SREQ, []byte(furs)})
	fmt.Println("家具数据：", len(furs))
}

//保存房间设置
func (this *HouseHandler) saveVshowSetting(session *ace.Session, message ace.DefaultSocketModel) {
	//错误处理
	defer func() {
		if r := recover(); r != nil {
			//有错误的话将返回"err"
			session.Write(&ace.DefaultSocketModel{protocol.HOUSE, -1, SAVE_VSHOW_SETTING_SREQ, []byte("err")})
		}
	}()
	//解析json 获取展厅名字
	saveData := &SaveHouseData{}
	err := json.Unmarshal(message.Message, &saveData)
	tools.CheckErr(err)
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	tools.CheckErr(err)
	user := SyncAccount.SessionAccount[session]
	if user != "" {
		stmtUp, err := db.Prepare("update virtualshow set setting=? where uid=? and name=?")
		tools.CheckErr(err)
		_, err = stmtUp.Exec(string(message.Message), user, saveData.RoomName) //保存设置数据
		tools.CheckErr(err)
		//响应
		session.Write(&ace.DefaultSocketModel{protocol.HOUSE, -1, SAVE_VSHOW_SETTING_SREQ, []byte("true")})
		fmt.Println("保存设置数据:", len(string(message.Message)))
	} else {
		fmt.Println("一个空用户正在保存户型设置，这不该发生！")
	}
}

//请求房间设置信息
func (this *HouseHandler) getVshowSetting(session *ace.Session, message ace.DefaultSocketModel) {
	//错误处理
	defer func() {
		if r := recover(); r != nil {
			//有错误的话将返回"err"
			session.Write(&ace.DefaultSocketModel{protocol.HOUSE, -1, GET_VSHOW_SETTING_SREQ, []byte("err")})
		}
	}()
	//解析json 获取展厅名字
	furData := &SaveHouseData{}
	err := json.Unmarshal(message.Message, &furData)
	tools.CheckErr(err)
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	tools.CheckErr(err)
	stmtOut, err := db.Prepare("select setting from virtualshow where uid=? and name=?")
	tools.CheckErr(err)
	var furs string
	err = stmtOut.QueryRow(furData.RoomUid, furData.RoomName).Scan(&furs) //参数是：所有者及房子名字
	tools.CheckErr(err)
	//响应
	session.Write(&ace.DefaultSocketModel{protocol.HOUSE, -1, GET_VSHOW_SETTING_SREQ, []byte(furs)})
	fmt.Println("设置数据：", len(furs))
}

//保存房间墙体
func (this *HouseHandler) saveVshowWall(session *ace.Session, message ace.DefaultSocketModel) {
	//错误处理
	defer func() {
		if r := recover(); r != nil {
			//有错误的话将返回"err"
			session.Write(&ace.DefaultSocketModel{protocol.HOUSE, -1, SAVE_VSHOW_WALL_SREQ, []byte("err")})
		}
	}()
	//解析json 获取展厅名字
	saveData := &SaveHouseData{}
	err := json.Unmarshal(message.Message, &saveData)
	tools.CheckErr(err)
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	tools.CheckErr(err)
	user := SyncAccount.SessionAccount[session]
	if user != "" {
		stmtUp, err := db.Prepare("update virtualshow set wall=? where uid=? and name=?")
		tools.CheckErr(err)
		_, err = stmtUp.Exec(string(message.Message), user, saveData.RoomName)
		tools.CheckErr(err)
		//响应
		session.Write(&ace.DefaultSocketModel{protocol.HOUSE, -1, SAVE_VSHOW_WALL_SREQ, []byte("true")})
		fmt.Println("保存墙体数据:", len(string(message.Message)))
	} else {
		fmt.Println("一个空用户正在保存户型墙体，这不该发生！")
	}
}

//请求房间墙体
func (this *HouseHandler) getVshowWall(session *ace.Session, message ace.DefaultSocketModel) {
	//错误处理
	defer func() {
		if r := recover(); r != nil {
			//有错误的话将返回"err"
			session.Write(&ace.DefaultSocketModel{protocol.HOUSE, -1, GET_VSHOW_WALL_SREQ, []byte("err")})
		}
	}()
	//解析json 获取展厅名字
	furData := &SaveHouseData{}
	err := json.Unmarshal(message.Message, &furData)
	tools.CheckErr(err)
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	tools.CheckErr(err)

	stmtOut, err := db.Prepare("select wall from virtualshow where uid=? and name=?")
	tools.CheckErr(err)
	var furs string
	err = stmtOut.QueryRow(furData.RoomUid, furData.RoomName).Scan(&furs) //参数是：所有者及房子名字
	tools.CheckErr(err)
	//响应
	session.Write(&ace.DefaultSocketModel{protocol.HOUSE, -1, GET_VSHOW_WALL_SREQ, []byte(furs)})
	fmt.Println("墙体数据：", len(furs))
}
