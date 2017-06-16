// LogicHandler
package logic

import (
	"ace"
	"fmt"
	"game/data"
	//	"game/logic/User"
	"game/logic/login"
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
	m := message.(ace.DefaultSocketModel)
	//fmt.Println("收到客户端的请求：", message)
	switch m.Type {
	case protocol.LOGIN: //登录
		if m.Command == 0 { //注册
			login.LoginHander.RegistProcess(session, m)
		}
		if m.Command == 2 { //登陆
			login.LoginHander.LoginProcess(session, m)
		}
		if m.Command == 4 { //心跳
			login.LoginHander.HeartPackage(session)
		}
		if m.Command == 10 { //退出
			login.LoginHander.ExitProcess(session)
		}
		break
	case protocol.USER: //客户数据
		data.User.Process(session, m)
		break
	case protocol.HOUSE: //房子
		//data.House.Process(session, m)
		break
	default:
		fmt.Println("未知协议类型！")
		break
	}
}
