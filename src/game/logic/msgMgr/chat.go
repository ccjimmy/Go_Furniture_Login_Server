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
	//	"strings"
	//	"time"

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
	//case CHAT_FRIEND_TO_ME: //朋友和我聊天
	//this.CHAT_FRIEND_TO_ME(session, msgModel)
	//	break

	default:
		fmt.Println("未知聊天消息类型")
		break
	}
}

//我向朋友聊天
func (this *Chat) CHAT_ME_TO_FRIEND(session *ace.Session, msgModel *MessageModel) {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
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
