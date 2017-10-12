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
	case FORCE_REMOVE_GROUP_CREQ: //申请把一个人移除出群
		this.FORCE_REMOVE_GROUP_CREQ(session, msgModel)
		break
	case INVITE_TO_GROUP_CREQ: //申请邀请一个人入群
		this.INVITE_TO_GROUP_CREQ(session, msgModel)
		break
	case INVITE_PROCESS_CREQ: //被邀请人的操作（他可以同意或拒绝）
		this.INVITE_PROCESS_CREQ(session, msgModel)
		break
	default:
		fmt.Println("未知群消息类型")
		break
	}
}

//被邀请人的操作（他可以同意或拒绝）
func (this *Group) INVITE_PROCESS_CREQ(session *ace.Session, msgModel *MessageModel) {
	//根据session获得这个人的账号
	beInvite := data.SyncAccount.SessionAccount[session]

	if msgModel.Content == "yes" { //同意了邀请
		//入群逻辑
		//群中加入这个人
		intgid, _ := strconv.Atoi(msgModel.To)
		addMember(intgid, beInvite)
		//这个人中加入群
		personalInfoAddGroup(beInvite, intgid, 0)
		//内存中修改群成员
		GroupMgr.ChangeMember(msgModel.To, beInvite, 1)
	}
	if msgModel.Content == "no" { //拒绝了邀请
	}
	//告诉邀请人 对方是否在线 先判断邀请人是否在线
	yqrSession, ok := data.SyncAccount.AccountSession[msgModel.From]
	if !ok { //邀请人不在线
		msgModel.MsgType = OTHER_PROCESS_OF_INVITE_SRES
		msgModel.From = data.SyncAccount.SessionAccount[session] //改变from 为被邀请人！
		saveOffLineMessage(&msgModel.From, msgModel)
	} else { //邀请人在线
		msgModel.MsgType = OTHER_PROCESS_OF_INVITE_SRES
		response, _ := json.Marshal(*msgModel)
		yqrSession.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, -1, response})
	}
	//给被邀请人自己的响应
	msgModel.MsgType = INVITE_PROCESS_SRES
	response, _ := json.Marshal(*msgModel)
	session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, -1, response})
	//广播群成员们刷新成员列表
	GroupMgr.Broadcast(msgModel.To, ace.DefaultSocketModel{protocol.SETTING, -1, protocol.MODIFY_GROUP_INFO_SREQ, []byte(msgModel.To)})
}

//申请邀请一个人入群
func (this *Group) INVITE_TO_GROUP_CREQ(session *ace.Session, msgModel *MessageModel) {
	fmt.Println("申请邀请一个人入群", msgModel.From, msgModel.To, msgModel.Content)
	//给邀请人的响应
	msgModel.MsgType = INVITE_TO_GROUP_SRES
	response, _ := json.Marshal(*msgModel)
	session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, -1, response})
	//给被邀请人的响应
	msgModel.MsgType = BE_INVITE_TO_GROUP_SRES
	memberArr := strings.Split(msgModel.Content, ",")
	for _, v := range memberArr {
		if v != "" {
			fmt.Println("通知这个人被邀请了", v)
			beInviteSession, ok := data.SyncAccount.AccountSession[v]
			if !ok { //被邀请人不在线
				saveOffLineMessage(&v, msgModel)
			} else {
				response, _ := json.Marshal(*msgModel)
				beInviteSession.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, -1, response})
			}
		}
	}
}

//群管理者申请移除一位群成员
func (this *Group) FORCE_REMOVE_GROUP_CREQ(session *ace.Session, msgModel *MessageModel) {
	fmt.Println("要移除的群是", msgModel.From, "移除的人是", msgModel.To)
	gid, err := strconv.Atoi(msgModel.From)
	if err != nil {
		fmt.Println("err--->FORCE_REMOVE_GROUP_CREQ:", err)
	}
	//1、群中移除这个人
	removeMember(gid, msgModel.To)
	//2、人中移除这个群
	personalInfoRemoveGroup(msgModel.To, gid)
	//3、群管理器中移除这个成员
	GroupMgr.ChangeMember(msgModel.From, msgModel.To, 1)
	//给操作者响应
	msgModel.MsgType = FORCE_REMOVE_GROUP_SRES
	response, _ := json.Marshal(*msgModel)
	session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, -1, response})
	//给被删除人的响应
	beRemoveSession, ok := data.SyncAccount.AccountSession[msgModel.To]
	if !ok { //他不在线,写入离线消息
		msgModel.MsgType = BE_REMOVE_GROUP_SRES
		saveOffLineMessage(&msgModel.To, msgModel)
	} else { //他在线
		msgModel.MsgType = BE_REMOVE_GROUP_SRES
		response, _ := json.Marshal(*msgModel)
		beRemoveSession.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, -1, response})
	}
	//广播群成员们刷新成员列表
	GroupMgr.Broadcast(msgModel.From, ace.DefaultSocketModel{protocol.SETTING, -1, protocol.MODIFY_GROUP_INFO_SREQ, []byte(msgModel.From)})
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
	stmt, err := db.Prepare("INSERT groups SET gid=?,name=?,description=?,face=?,level=?,master=?,manager=?,member=?,verifymode=?,history=?,createdtime=?")
	if err != nil {
		fmt.Println(err)
	}
	_, err = stmt.Exec(0, createGroupModel.Groupname, "还没有说明", "default.jpg", 0, msgModel.From, "", "", createGroupModel.VerifyModel, "[]", time.Now().Format("2006-01-02 15:04:05"))
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

			//	return
		} else { //群主在线
			msgModel.MsgType = ONE_WANT_ADD_GROUP_SRES
			response, _ := json.Marshal(*msgModel)
			masterSession.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, ONE_WANT_ADD_GROUP_SRES, response})
		}
		//给申请人的回应
		msgModel.MsgType = ADD_GROUP_SRES
		msgModel.Content = "申请已经发出，请等待群主审核。"
		response, _ := json.Marshal(*msgModel)
		session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, ADD_GROUP_SRES, response})
		fmt.Println("给申请人响应")

	} else /////////////////////////////////////////////////////无需验证
	{
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
		//广播群成员们刷新成员列表
		GroupMgr.Broadcast(msgModel.To, ace.DefaultSocketModel{protocol.SETTING, -1, protocol.MODIFY_GROUP_INFO_SREQ, []byte(msgModel.To)})
		//改变内存群成员
		GroupMgr.ChangeMember(msgModel.To, msgModel.From, 1)
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
	//广播群成员们刷新成员列表
	GroupMgr.Broadcast(msgModel.To, ace.DefaultSocketModel{protocol.SETTING, -1, protocol.MODIFY_GROUP_INFO_SREQ, []byte(msgModel.To)})
	//改变内存群成员
	GroupMgr.ChangeMember(msgModel.To, msgModel.From, 1)
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
	//广播群成员们刷新成员列表
	GroupMgr.Broadcast(msgModel.To, ace.DefaultSocketModel{protocol.SETTING, -1, protocol.MODIFY_GROUP_INFO_SREQ, []byte(msgModel.To)})
	//改变内存群成员
	GroupMgr.ChangeMember(msgModel.To, msgModel.From, 0)
}

//数据库个人信息中加入这个群
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

//数据库个人信息中移除这个群
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

//数据库中群增加新成员
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

//数据库中群移除一名成员
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
	//fmt.Println("之前的成员是:" + member)
	newMemberList := ""
	memberArr := strings.Split(member, ",")
	for _, v := range memberArr {
		if v != "" && v != removeMember {
			newMemberList += v + ","
		}
	}
	//更新成员列表
	//fmt.Println("之后的成员是:" + newMemberList)
	stmtUp, err := db.Prepare("update groups set member=? where gid=?")
	_, err = stmtUp.Exec(newMemberList, gid)
	if err != nil {
		fmt.Println(err)
	}
}
