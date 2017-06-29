package msgMgr

import (
	"ace"
	"database/sql"
	"encoding/json"
	"fmt"
	"game/data"
	"game/logic/protocol"
	"strings"
	"tools"

	_ "github.com/go-sql-driver/mysql"
)

type Friend struct {
}

var FriendHander = &Friend{}

func (this *Friend) Process(session *ace.Session, msgModel *MessageModel) {
	switch msgModel.MsgType {
	case ADD_FRIEND_CREQ: //申请添加好友
		this.ADD_FRIEND_CREQ(session, msgModel)
		break
	case AGREE_ADD_FRIEND_CREQ: //同意加好友
		this.AGREE_ADD_FRIEND(session, msgModel)
		break
	case DELETE_FRIEND_CREQ: //删除好友
		this.DELETE_FRIEND_CREQ(session, msgModel)
		break
	default:
		fmt.Println("未知好友消息类型")
		break
	}
}

//申请添加好友
func (this *Friend) ADD_FRIEND_CREQ(session *ace.Session, message *MessageModel) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("添加好友时异常:-------------》》》", r)
			return
		}
	}()

	var isHeOnLine bool = false
	//遍历在线人员
	for otherSe, acc := range data.SyncAccount.SessionAccount {
		if message.To == acc {
			message.MsgType = ONE_ADD_YOU_SRES
			response, _ := json.Marshal(*message)
			otherSe.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, ONE_ADD_YOU_SRES, response})
			isHeOnLine = true
			break
		}
	}
	//不在线则存入数据库
	if isHeOnLine == false {
		fmt.Println("要申请的人不在线")
		//申请加好友消息--->有人加你消息
		var offlineMsg = message
		offlineMsg.MsgType = ONE_ADD_YOU_SRES
		saveOffLineMessage(message)
	}
	//给自己的响应
	message.MsgType = ADD_FRIEND_SRES
	response, _ := json.Marshal(*message)
	session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, ADD_FRIEND_SRES, response})
}

//同意添加好友
func (this *Friend) AGREE_ADD_FRIEND(session *ace.Session, message *MessageModel) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	//更新好友列表
	updateFriendList(message.From, message.To, 0)
	updateFriendList(message.To, message.From, 0)
	//回复同意的人
	message.MsgType = AGREE_ADD_FRIEND_SRES
	response, _ := json.Marshal(*message)
	session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, AGREE_ADD_FRIEND_SRES, response})

	//回复申请的人
	//遍历在线人员
	var isHeOnLine bool = false
	for otherSe, acc := range data.SyncAccount.SessionAccount {
		if message.To == acc {
			message.MsgType = ONE_AGREED_YOU //别人同意了你的申请
			response, _ := json.Marshal(*message)
			otherSe.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, ONE_AGREED_YOU, response})
			isHeOnLine = true
			break
		}
	}
	//不在线则存入数据库
	if isHeOnLine == false {
		fmt.Println("申请人不在线")
		//同意加好友--->别人同意了你的申请
		var offlineMsg = message
		offlineMsg.MsgType = ONE_AGREED_YOU
		saveOffLineMessage(message)
	}
}

//删除好友
func (this *Friend) DELETE_FRIEND_CREQ(session *ace.Session, message *MessageModel) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	//互相删除好友列表
	updateFriendList(message.From, message.To, 1)
	updateFriendList(message.To, message.From, 1)

	var isHeOnLine bool = false
	//遍历在线人员
	for otherSe, acc := range data.SyncAccount.SessionAccount {
		if message.To == acc {
			message.MsgType = YOU_BE_DELETED
			response, _ := json.Marshal(*message)
			otherSe.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, YOU_BE_DELETED, response})
			isHeOnLine = true
			break
		}
	}
	//不在线则存入数据库
	if isHeOnLine == false {
		fmt.Println("要申请的人不在线")
		var offlineMsg = message
		offlineMsg.MsgType = YOU_BE_DELETED
		saveOffLineMessage(message)
	}
	//给自己的响应
	message.MsgType = DELETE_FRIEND_SRES
	response, _ := json.Marshal(*message)
	session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, message.MsgType, response})
}

func updateFriendList(self string, other string, op int) { //op=0是增加好友 op=1是删除好友
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	tools.CheckErr(err)
	stmtOut, err := db.Prepare("SELECT friends FROM userdata WHERE username = ?")
	var friends string
	err = stmtOut.QueryRow(self).Scan(&friends)
	tools.CheckErr(err)
	fmt.Print("我有这么多的好友", friends)
	if op == 0 { //增加好友
		friendsArr := strings.Split(friends, ",")
		for _, v := range friendsArr {
			if v == other { //已经有这个好友了
				fmt.Println("已经有这个好友了，这！不该发生")
				return
			}
		}
		//增加好友
		friends = friends + "," + other
		stmtUp, err := db.Prepare("update userdata set friends=? where username=?") //更新好友列表
		tools.CheckErr(err)
		_, err = stmtUp.Exec(friends, self)
		tools.CheckErr(err)
	} else { //删除好友
		newFriendList := ""
		friendsArr := strings.Split(friends, ",")
		for _, v := range friendsArr {
			if v != other && v != "" {
				newFriendList += v + ","
			}
		}
		stmtUp, err := db.Prepare("update userdata set friends=? where username=?") //更新好友列表
		tools.CheckErr(err)
		_, err = stmtUp.Exec(newFriendList, self)
		tools.CheckErr(err)
	}
}

//不在线时保存离线消息
func saveOffLineMessage(msgModel *MessageModel) { //to不在线，存给to
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	tools.CheckErr(err)
	//获取已存在离线消息
	stmtOut, err := db.Prepare("SELECT offlinemsg FROM userdata WHERE username = ?")
	var offlinemsg string
	err = stmtOut.QueryRow(msgModel.To).Scan(&offlinemsg)
	if err != nil {
		fmt.Println("这里有问题，他之前没有离线消息", err)
	}
	tools.CheckErr(err)
	fmt.Println("之前的离线消息是：", offlinemsg)
	tempSlice := []MessageModel{}

	//解开json,变成切片
	err = json.Unmarshal([]byte(offlinemsg), &tempSlice)
	if err != nil {
		fmt.Println(err)
	}
	//追加
	if msgModel.MsgType == ONE_ADD_YOU_SRES { //重复的加好友，不需要写入数据库
		for _, v := range tempSlice {
			if v.MsgType == ONE_ADD_YOU_SRES && v.From == msgModel.From {
				fmt.Println("重复的加好友，不需要写入数据库")
				return
			}
		}
	}
	if msgModel.MsgType == ONE_AGREED_YOU { //重复的同意加好友，不需要写入数据库
		for _, v := range tempSlice {
			if v.MsgType == ONE_AGREED_YOU && v.From == msgModel.From {
				fmt.Println("重复的同意加好友，不需要写入数据库")
				return
			}
		}
	}
	if msgModel.MsgType == YOU_BE_DELETED { //重复的你被删除好友，不需要写入数据库
		for _, v := range tempSlice {
			if v.MsgType == YOU_BE_DELETED && v.From == msgModel.From {
				fmt.Println("重复的被删除好友，不需要写入数据库")
				return
			}
		}
	}

	tempSlice = append(tempSlice, *msgModel)
	//更新数据库
	newofflinemsg, _ := json.Marshal(tempSlice)
	fmt.Println("最新的离线消息列表 ", string(newofflinemsg))
	stmtUp, err := db.Prepare("update userdata set offlinemsg=? where username=?") //更新好友列表
	tools.CheckErr(err)
	_, err = stmtUp.Exec(string(newofflinemsg), msgModel.To)
	tools.CheckErr(err)
}
