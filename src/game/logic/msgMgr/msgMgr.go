package msgMgr

import (
	"ace"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type MsgMgr struct {
}

const ( //协议类型
	FRIEND = 0 //好友相关
	GROUP  = 1 //群组相关
	CHAT   = 2 //聊天相关

	ADD_FRIEND_CREQ       = 10 //添加好友
	ADD_FRIEND_SRES       = 11 //添加好友的反馈
	ONE_ADD_YOU_SRES      = 12 //有人加你好友
	AGREE_ADD_FRIEND_CREQ = 13 //同意加好友
	AGREE_ADD_FRIEND_SRES = 14 //同意加好友的响应
	ONE_AGREED_YOU        = 15 //别人同意了你的申请
	DELETE_FRIEND_CREQ    = 16 //删除好友
	DELETE_FRIEND_SRES    = 17 //删除好友的响应
	YOU_BE_DELETED        = 18 //你被删除好友了

	CREATE_GROUP_CREQ         = 30 //创建群
	CREATE_GROUP_SRES         = 31 //创建群响应
	ADD_GROUP_CREQ            = 32 //申请入群
	ADD_GROUP_SRES            = 33 //申请响应
	ONE_WANT_ADD_GROUP_SRES   = 34 //有人想要入群
	AGREE_ADD_GROUP_CREQ      = 35 //群主同意申请入群
	AGREE_ADD_GROUP_SRES      = 36 //群主同意申请入群的响应
	YOU_BE_AGREED_ENTER_GROUP = 37 //你被同意入群
	QUIT_GROUP_CREQ           = 38 //退出一个群
	QUIT_GROUP_SRES           = 39 //退出一个群的响应

	CHAT_ME_TO_FRIEND_CREQ = 100 //和好友聊天
	CHAT_ME_TO_FRIEND_SRES = 101 //和好友聊天的响应

	CHAT_FRIEND_TO_ME_SREQ = 102 //好友和我聊天
)

//消息数据结构
type MessageModel struct {
	MsgType int
	From    string
	To      string
	Content string
	Time    string
}

var MsgMgrHander = &MsgMgr{}

func (this *MsgMgr) Process(session *ace.Session, message ace.DefaultSocketModel) {

	//解开json
	msgModel := &MessageModel{}
	err := json.Unmarshal(message.Message, &msgModel)
	if err != nil {
		fmt.Println(err)
	}

	switch message.Command {
	case FRIEND: //好友相关
		FriendHander.Process(session, msgModel)
		break
	case GROUP: //群组相关
		GroupHander.Process(session, msgModel)
		break
	case CHAT: //聊天相关
		ChatHander.Process(session, msgModel)
		break
	default:
		fmt.Println("消息管理器：未知消息协议类型")
		break
	}
}

//不在线时保存离线消息
func saveOffLineMessage(userName *string, msgModel *MessageModel) { //to不在线，存给to
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	if err != nil {
		fmt.Println(err)
	}
	//获取已存在离线消息
	stmtOut, err := db.Prepare("SELECT offlinemsg FROM userdata WHERE username = ?")
	var offlinemsg string
	err = stmtOut.QueryRow(userName).Scan(&offlinemsg)
	if err != nil {
		fmt.Println("这里有问题，他之前没有离线消息", err)
	}
	if err != nil {
		fmt.Println(err)
	}
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
	if msgModel.MsgType == ONE_WANT_ADD_GROUP_SRES {
		for _, v := range tempSlice {
			if v.MsgType == ONE_WANT_ADD_GROUP_SRES && v.From == msgModel.From {
				fmt.Println("重复的入群申请，不需要写入数据库")
				return
			}
		}
	}

	tempSlice = append(tempSlice, *msgModel)
	//更新数据库
	newofflinemsg, _ := json.Marshal(tempSlice)
	fmt.Println("最新的离线消息列表 ", string(newofflinemsg))
	stmtUp, err := db.Prepare("update userdata set offlinemsg=? where username=?") //更新好友列表
	if err != nil {
		fmt.Println(err)
	}
	_, err = stmtUp.Exec(string(newofflinemsg), userName)
	if err != nil {
		fmt.Println(err)
	}
}
