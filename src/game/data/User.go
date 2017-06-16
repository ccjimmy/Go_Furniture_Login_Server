package data

import (
	"ace"
	"database/sql"
	"fmt"
	//"encoding/json"
	"encoding/json"
	"game/logic/protocol"
	"tools"

	_ "github.com/go-sql-driver/mysql"
)

//申请成为供应商的名字
type Provider struct {
	ProviderName string
}

//协议
const (
	LEVEL_CREQ = 0 //客户端请求改变账户类别
	LEVEL_SREQ = 1
)

type UserHandler struct {
}

var User = &UserHandler{}

//用户数据处理逻辑
func (this *UserHandler) Process(session *ace.Session, message ace.DefaultSocketModel) {
	switch message.Command {
	case LEVEL_CREQ: //改变用户类别
		this.UpLevel(session, message)
		break
	default:
		fmt.Println("未知用户信息协议类型！")
		break
	}
}

//提升用户类别
func (this *UserHandler) UpLevel(session *ace.Session, message ace.DefaultSocketModel) {
	//错误处理
	defer func() {
		if r := recover(); r != nil {
			//有错误的话将返回"err"
			session.Write(&ace.DefaultSocketModel{protocol.USER, -1, LEVEL_SREQ, []byte("err")})
		}
	}()
	//解开json
	providerData := &Provider{}
	err := json.Unmarshal(message.Message, &providerData)
	tools.CheckErr(err)
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	tools.CheckErr(err)
	//获得账号
	user := SyncAccount.SessionAccount[session]
	if user != "" {
		stmtUp, err := db.Prepare("update userinfo set level=? , provider=? where username=?")
		tools.CheckErr(err)
		_, err = stmtUp.Exec(2, providerData.ProviderName, user) //提升为企业用户
		tools.CheckErr(err)
		fmt.Println("这个人升级：", providerData.ProviderName)
		//响应是 本供应商名字
		session.Write(&ace.DefaultSocketModel{protocol.USER, -1, LEVEL_SREQ, []byte(providerData.ProviderName)})
	} else {
		fmt.Println("错误：这个session没有登陆，却在尝试提升用户等级" + providerData.ProviderName)
	}

}
