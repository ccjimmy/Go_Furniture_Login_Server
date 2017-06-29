// LogicHandler
package logic

import (
	"ace"
	"fmt"
	"game/data"
	"game/logic/login"
	"game/logic/msgMgr"
	"game/logic/protocol"
)

//它的三个方法实现了ServerSocket中的Handler接口
type GameHandler struct {
}

func (this *GameHandler) SessionOpen(session *ace.Session) {
	fmt.Println("会话 open", session)
}

func (this *GameHandler) SessionClose(session *ace.Session) {
	fmt.Println("会话 closed", session)
	data.SyncAccount.SessionClose(session)

}

func (this *GameHandler) MessageReceived(session *ace.Session, message interface{}) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("LogicHandler处理消息异常:-------------------》》》", r)
			return
		}
	}()

	m := message.(ace.DefaultSocketModel)
	//fmt.Println("收到客户端的请求：", message)
	switch m.Type {
	case protocol.HEART_PACKAGE_CREQ: //心跳
		//session.Write(&ace.DefaultSocketModel{protocol.HEART_PACKAGE_SREQ, -1, -1, []byte("im server")})
		break
	case protocol.LOGIN: //收到登录消息
		login.LoginHander.Process(session, m)
		break
	case protocol.MESSAGE: //消息相关
		msgMgr.MsgMgrHander.Process(session, m)
		break
	default:
		fmt.Println("未知协议类型！")
		session.Write(&ace.DefaultSocketModel{88, -1, -1, []byte("im server")})
		break
	}
}
