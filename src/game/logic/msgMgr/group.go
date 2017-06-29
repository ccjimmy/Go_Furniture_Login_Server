package msgMgr

import (
	"ace"
	"database/sql"
	"encoding/json"
	"fmt"
	//"game/data"
	"game/logic/protocol"
	//"strconv"
	//	"tools"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Group struct {
}

//创建群的数据模型
type CreateGroupModel struct {
	Groupname   string
	VerifyModel int
}

//我的群数据结构
type MyGroupModel struct {
	GroupID      int
	ReceiveModel int
}

var GroupHander = &Group{}

func (this *Group) Process(session *ace.Session, msgModel *MessageModel) {

	switch msgModel.MsgType {
	case CREATE_GROUP_CREQ:

		this.CREATE_GROUP_CREQ(session, msgModel)
		break

	default:
		fmt.Println("未知群消息类型")
		break
	}
}

//创建群
func (this *Group) CREATE_GROUP_CREQ(session *ace.Session, msgModel *MessageModel) {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	if err != nil {
		fmt.Println(err)
	}
	//解析创建群基本信息
	var createGroupModel = &CreateGroupModel{}
	//解开json
	err = json.Unmarshal([]byte(msgModel.Content), &createGroupModel)
	if err != nil {
		fmt.Println(err)
	}

	//设置群号
	var groupID = 0
	stmtOut, err := db.Prepare("SELECT max(gid) FROM groups")
	if err != nil {
		fmt.Println(err)
	}
	err = stmtOut.QueryRow().Scan(&groupID)
	if err != nil {
		fmt.Println(err)
	}
	groupID = groupID + 1
	fmt.Println("群号是", groupID)
	//数据库中加入这个群
	stmt, err := db.Prepare("INSERT groups SET gid=?,name=?,face=?,level=?,master=?,manager=?,member=?,verifymode=?,createdtime=?")
	if err != nil {
		fmt.Println(err)
	}
	_, err = stmt.Exec(0, createGroupModel.Groupname, "default.jpg", 0, msgModel.From, "", "", createGroupModel.VerifyModel, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		fmt.Println(err)
	}
	//群主个人信息中加入这个群
	var myGroupModel = &MyGroupModel{}
	myGroupModel.GroupID = groupID
	myGroupModel.ReceiveModel = 0

	//获取已存在群数据
	stmtOut, err = db.Prepare("SELECT groups FROM userdata WHERE username = ?")
	var groups string
	err = stmtOut.QueryRow(msgModel.From).Scan(&groups)
	if err != nil {
		fmt.Println("他之前没有群", err)
	}
	fmt.Println("之前的群是：", groups)
	tempSlice := []MyGroupModel{}

	//解开json,变成切片
	err = json.Unmarshal([]byte(groups), &tempSlice)
	if err != nil {
		fmt.Println(err)
	}
	//追加
	tempSlice = append(tempSlice, *myGroupModel)
	//更新数据库
	newGroups, _ := json.Marshal(tempSlice)
	fmt.Println("最新的群列表 ", string(newGroups))
	stmtUp, err := db.Prepare("update userdata set groups=? where username=?") //更新好友列表
	_, err = stmtUp.Exec(string(newGroups), msgModel.From)
	if err != nil {
		fmt.Println(err)
	}
	//响应
	createGroupModel.VerifyModel = groupID
	newContent, _ := json.Marshal(*createGroupModel)

	msgModel.MsgType = CREATE_GROUP_SRES
	msgModel.Content = string(newContent)
	response, _ := json.Marshal(*msgModel)
	session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, CREATE_GROUP_SRES, response})
}
