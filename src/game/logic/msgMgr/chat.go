package msgMgr

import (
	"fmt"

	"ace"
	"database/sql"
	"encoding/json"

	"game/data"
	"game/logic/protocol"
	//	"strconv"
	//	//	"tools"
	"strings"
	//	"time"
	"tools"

	_ "github.com/go-sql-driver/mysql"
)

type Chat struct {
}

var ChatHander = &Chat{}

func (this *Chat) Process(session *ace.Session, msgModel *MessageModel) {
	switch msgModel.MsgType {
	case CHAT_ME_TO_FRIEND_CREQ: //我和朋友聊天
		this.CHAT_ME_TO_FRIEND(session, msgModel)
		break
	case CHAT_ME_TO_GROUP_CREQ: //我向群聊天
		this.CHAT_ME_TO_GROUP(session, msgModel)
		break

	default:
		fmt.Println("未知聊天消息类型")
		break
	}
}

//我向朋友聊天
func (this *Chat) CHAT_ME_TO_FRIEND(session *ace.Session, msgModel *MessageModel) {
	db, err := sql.Open("mysql", tools.GetSQLStr())
	defer db.Close()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("聊天消息:", msgModel.From, " -> ", msgModel.To, " : ", msgModel.Content)
	//找到那个朋友
	friendsession, ok := data.SyncAccount.AccountSession[msgModel.To]
	if ok {
		msgModel.MsgType = CHAT_FRIEND_TO_ME_SREQ
		response, _ := json.Marshal(*msgModel)
		friendsession.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, CHAT, response})
	} else { //朋友不在线
		msgModel.MsgType = CHAT_FRIEND_TO_ME_SREQ
		saveOffLineMessage(&msgModel.To, msgModel)
	}
	//响应
	msgModel.MsgType = CHAT_ME_TO_FRIEND_SRES
	response, _ := json.Marshal(*msgModel)
	session.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, CHAT, response})
}

//向群聊天
func (this *Chat) CHAT_ME_TO_GROUP(session *ace.Session, msgModel *MessageModel) {

	group := GroupMgr.GetOneGroupManager(msgModel.To)
	group.OnGroupActive(msgModel)
	//把消息分发给所有成员
	msgModel.MsgType = CHAT_GROUP_TO_ME_SRES //转换类型
	response, _ := json.Marshal(*msgModel)   //转发给所有人的消息

	allMembers := group.Master + "," + group.Managers + "," + group.Members
	allMembersArr := strings.Split(allMembers, ",")
	count := 0
	for _, v := range allMembersArr {
		if v != "" { //得到每一个人

			//	fmt.Println("群成员", v)
			memSe, ok := data.SyncAccount.AccountSession[v]
			if ok { //如果这个人在线
				count++
				memSe.Write(&ace.DefaultSocketModel{protocol.MESSAGE, -1, CHAT, response})
			}
		}
	}
	fmt.Println("群发", count, "条消息")
}
