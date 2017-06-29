package msgMgr

import (
	"ace"

	"encoding/json"
	"fmt"
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

	CREATE_GROUP_CREQ = 30 //创建群
	CREATE_GROUP_SRES = 31 //创建群响应
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

		break
	default:
		fmt.Println("消息管理器：未知消息协议类型")
		break
	}
}
