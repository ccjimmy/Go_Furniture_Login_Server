package msgMgr

import (
	"ace"
	"database/sql"
	"encoding/json"
	"fmt"
	"game/data"
	"game/logic/protocol"
	"strconv"
	//	"tools"
	"strings"
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
	case CREATE_GROUP_CREQ: //建群
		this.CREATE_GROUP_CREQ(session, msgModel)
		break
	case ADD_GROUP_CREQ: //加群
		this.ADD_GROUP_CREQ(session, msgModel)
		break
	case AGREE_ADD_GROUP_CREQ: //群主同意入群
		this.AGREE_ADD_GROUP_CREQ(session, msgModel)
		break
	case QUIT_GROUP_CREQ: //退出一个群
		this.QUIT_GROUP_CREQ(session, msgModel)
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
	//个人信息中加入这个群
	personalInfoAddGroup(msgModel.From, groupID, 0)
	//响应
	myGroupModel := &MyGroupModel{}
	myGroupModel.GroupID = groupID
	myGroupModel.ReceiveModel = 0

	newContent, _ := json.Marshal(*myGroupModel)

	msgModel.MsgType = CREATE_GROUP_SRES
	msgModel.Content = string(newContent)
	response, _ := json.Marshal(*msgModel)
	session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, CREATE_GROUP_SRES, response})
}

//加群
func (this *Group) ADD_GROUP_CREQ(session *ace.Session, msgModel *MessageModel) {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	if err != nil {
		fmt.Println(err)
	}
	//判断是否已经在这个群了
	oldMember := ""
	stmtOut, err := db.Prepare("SELECT member FROM groups where gid =?")
	if err != nil {
		fmt.Println(err)
	}
	err = stmtOut.QueryRow(msgModel.To).Scan(&oldMember)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("之前的成员是:" + oldMember)
	memberArr := strings.Split(oldMember, ",")
	for _, v := range memberArr {
		if v == msgModel.From {
			fmt.Println("非法的入群申请，已在这个群")
			return
		}
	}

	//判断群员数量
	memberAmount := 0
	for _, v := range memberArr {
		if v != "" {
			memberAmount++
		}
	}
	fmt.Println("这个群的群员个数是：", memberAmount)
	if memberAmount > 199 {
		msgModel.MsgType = ADD_GROUP_SRES
		msgModel.Content = "too many member"
		response, _ := json.Marshal(*msgModel)
		session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, ADD_GROUP_SRES, response})
		return
	}
	//获取这个群的验证方式
	var verifymodel = 0
	stmtOut, err = db.Prepare("SELECT verifymode FROM groups where gid =?")
	if err != nil {
		fmt.Println(err)
	}
	err = stmtOut.QueryRow(msgModel.To).Scan(&verifymodel)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("这个群的验证方式", verifymodel)
	///////////////////////////////////////////////////////////需要验证
	if verifymodel == 0 {
		//找到群主
		master := ""
		stmtOut, err = db.Prepare("SELECT master FROM groups where gid =?")
		if err != nil {
			fmt.Println(err)
		}
		err = stmtOut.QueryRow(msgModel.To).Scan(&master)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("这个群的群主是:" + master)
		masterSession, ok := data.SyncAccount.AccountSession[master]
		if !ok { //群主不在线,写入离线消息
			msgModel.MsgType = ONE_WANT_ADD_GROUP_SRES
			saveOffLineMessage(&master, msgModel)

			msgModel.MsgType = ADD_GROUP_SRES
			msgModel.Content = "申请已经发出，请等待群主审核。"
			response, _ := json.Marshal(*msgModel)
			session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, ADD_GROUP_SRES, response})
			return
		} else { //群主在线
			msgModel.MsgType = ONE_WANT_ADD_GROUP_SRES
			response, _ := json.Marshal(*msgModel)
			masterSession.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, ONE_WANT_ADD_GROUP_SRES, response})
		}
		return
	}
	/////////////////////////////////////////////////////无需验证
	if verifymodel == 1 {
		gid, _ := strconv.Atoi(msgModel.To)
		//添加群成员
		addMember(gid, msgModel.From)
		//个人信息中加入这个群
		personalInfoAddGroup(msgModel.From, gid, 0)
		//响应
		myGroupModel := &MyGroupModel{}
		myGroupModel.GroupID = gid
		myGroupModel.ReceiveModel = 0

		newContent, _ := json.Marshal(*myGroupModel)
		msgModel.MsgType = ADD_GROUP_SRES
		msgModel.Content = string(newContent)
		response, _ := json.Marshal(*msgModel)
		session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, ADD_GROUP_SRES, response})
		return
	}
}

//群主同意入群
func (this *Group) AGREE_ADD_GROUP_CREQ(session *ace.Session, msgModel *MessageModel) {
	fmt.Println("群主同意", msgModel.From, "入群", msgModel.To)
	//入群手续
	gid, _ := strconv.Atoi(msgModel.To)
	//个人信息中加入这个群
	personalInfoAddGroup(msgModel.From, gid, 0)
	//群加入这个成员
	addMember(gid, msgModel.From)
	//告诉申请人你已被通过
	proposerSession, ok := data.SyncAccount.AccountSession[msgModel.From]
	if !ok { //不在线,写入离线消息
		msgModel.MsgType = YOU_BE_AGREED_ENTER_GROUP
		saveOffLineMessage(&msgModel.From, msgModel)

	} else { //在线
		msgModel.MsgType = YOU_BE_AGREED_ENTER_GROUP
		response, _ := json.Marshal(*msgModel)
		proposerSession.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, YOU_BE_AGREED_ENTER_GROUP, response})
	}
	//响应
	msgModel.MsgType = AGREE_ADD_GROUP_SRES
	response, _ := json.Marshal(*msgModel)
	session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, AGREE_ADD_GROUP_SRES, response})
}

//退群
func (this *Group) QUIT_GROUP_CREQ(session *ace.Session, msgModel *MessageModel) {
	gid, _ := strconv.Atoi(msgModel.To)
	personalInfoRemoveGroup(msgModel.From, gid)
	removeMember(gid, msgModel.From)
	//响应
	msgModel.MsgType = QUIT_GROUP_SRES
	response, _ := json.Marshal(*msgModel)
	session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, QUIT_GROUP_SRES, response})
}

//个人信息中加入这个群
func personalInfoAddGroup(user string, gid int, receiveMode int) {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	if err != nil {
		fmt.Println(err)
	}

	var myGroupModel = &MyGroupModel{}
	myGroupModel.GroupID = gid
	myGroupModel.ReceiveModel = receiveMode

	//获取已存在群数据
	stmtOut, err := db.Prepare("SELECT groups FROM userdata WHERE username = ?")
	var groups string
	err = stmtOut.QueryRow(user).Scan(&groups)
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
	for _, v := range tempSlice {
		if v.GroupID == gid {
			fmt.Println("我之前已经拥有这个群了，不需要重复拥有")
			return
		}
	}
	//追加
	tempSlice = append(tempSlice, *myGroupModel)
	//更新数据库
	newGroups, _ := json.Marshal(tempSlice)
	fmt.Println("最新的群列表 ", string(newGroups))
	stmtUp, err := db.Prepare("update userdata set groups=? where username=?") //更新好友列表
	_, err = stmtUp.Exec(string(newGroups), user)
	if err != nil {
		fmt.Println(err)
	}
}

//个人信息中移除这个群
func personalInfoRemoveGroup(user string, gid int) {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	if err != nil {
		fmt.Println(err)
	}
	//获取已存在群数据
	stmtOut, err := db.Prepare("SELECT groups FROM userdata WHERE username = ?")
	var groups string
	err = stmtOut.QueryRow(user).Scan(&groups)
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
	newGroupSlice := []MyGroupModel{}
	for _, v := range tempSlice {
		if v.GroupID != gid {
			newGroupSlice = append(newGroupSlice, v)
		}
	}
	//更新数据库
	newGroups, _ := json.Marshal(newGroupSlice)
	fmt.Println("最新的群列表 ", string(newGroups))
	stmtUp, err := db.Prepare("update userdata set groups=? where username=?") //更新好友列表
	_, err = stmtUp.Exec(string(newGroups), user)
	if err != nil {
		fmt.Println(err)
	}
}

//群增加新成员
func addMember(gid int, newMember string) {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	if err != nil {
		fmt.Println(err)
	}
	//获取之前的群成员
	member := ""
	stmtOut, err := db.Prepare("SELECT member FROM groups where gid =?")
	if err != nil {
		fmt.Println(err)
	}
	err = stmtOut.QueryRow(gid).Scan(&member)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("之前的成员是:" + member)
	memberArr := strings.Split(member, ",")
	for _, v := range memberArr {
		if v == newMember {
			fmt.Println("本群之前已经有这个成员了，不需要重复增加这个成员")
			return
		}
	}
	//更新成员列表
	member += "," + newMember
	fmt.Println("之后的成员是:" + member)
	stmtUp, err := db.Prepare("update groups set member=? where gid=?")
	_, err = stmtUp.Exec(member, gid)
	if err != nil {
		fmt.Println(err)
	}
}

//移除成员
func removeMember(gid int, removeMember string) {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	if err != nil {
		fmt.Println(err)
	}
	//获取之前的群成员
	member := ""
	stmtOut, err := db.Prepare("SELECT member FROM groups where gid =?")
	if err != nil {
		fmt.Println(err)
	}
	err = stmtOut.QueryRow(gid).Scan(&member)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("之前的成员是:" + member)
	newMemberList := ""
	memberArr := strings.Split(member, ",")
	for _, v := range memberArr {
		if v != "" && v != removeMember {
			newMemberList += v + ","
			return
		}
	}
	//更新成员列表
	fmt.Println("之后的成员是:" + newMemberList)
	stmtUp, err := db.Prepare("update groups set member=? where gid=?")
	_, err = stmtUp.Exec(newMemberList, gid)
	if err != nil {
		fmt.Println(err)
	}
}
