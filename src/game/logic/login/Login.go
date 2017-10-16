// RegistHandler
package login

import (
	"ace"
	"database/sql"
	"encoding/json"
	"fmt"
	"game/data"
	"game/logic/protocol"
	"time"
	"tools"

	_ "github.com/go-sql-driver/mysql"
)

type AccountDTO struct {
	Username string
	Password string
	Phone    string
}

//协议
const (
	REGIST_CREQ = 0
	REGIST_SREQ = 1 //command=1代表这是注册结果
	LOGIN_CREQ  = 2
	LOGIN_SREQ  = 3 //3代表登陆成功

	RELOGIN_CREQ = 4 //重新登陆
	RELOGIN_SRES = 5 //重新登陆

	EXIT_CREQ = 10 //退出登录
)

//登陆结果变量
const (
	USER_NO      = "10" //用户名不存在
	USER_RELOGIN = "11" //重复登录
	USER_PSDERR  = "12" //密码错误
	//USER_WAIT    = "13" //需要进行是否重复登陆的判断，请等待
	//LOGIN_OK = "14" //登陆成功 返回用户昵称
)

type Handler struct {
}

var LoginHander = &Handler{}

func (this *Handler) Process(session *ace.Session, message ace.DefaultSocketModel) {
	switch message.Command {
	case REGIST_CREQ: //注册
		this.RegistProcess(session, message)
		break
	case LOGIN_CREQ: //登陆
		this.LoginProcess(session, message)
		break
	case RELOGIN_CREQ: //重新登录
		this.ReLoginProcess(session, message)
		break
	case EXIT_CREQ: //推出登陆
		this.ExitProcess(session)
		break
	}
}

//注册账号
func (this *Handler) RegistProcess(session *ace.Session, message ace.DefaultSocketModel) {
	//错误处理
	defer func() {
		if r := recover(); r != nil {
			//有错误的话将返回"err"
			session.Write(&ace.DefaultSocketModel{protocol.LOGIN, -1, REGIST_SREQ, []byte("err")})
		}
	}()

	//解开json
	registData := &AccountDTO{}
	err := json.Unmarshal(message.Message, &registData)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println("申请注册:", registData.Username, registData.Password, registData.Phone)
	//注册具体逻辑
	regidtResult := this.reg(&registData.Username, &registData.Password, &registData.Phone)
	if regidtResult == false {
		fmt.Println("注册失败")
		session.Write(&ace.DefaultSocketModel{protocol.LOGIN, -1, REGIST_SREQ, []byte("false")})
	} else {
		fmt.Println("注册成功")
		session.Write(&ace.DefaultSocketModel{protocol.LOGIN, -1, REGIST_SREQ, []byte("true")})
	}
}

//登陆
func (this *Handler) LoginProcess(session *ace.Session, message ace.DefaultSocketModel) {
	//错误处理
	defer func() {
		if r := recover(); r != nil {
			//有错误的话将返回"err"
			session.Write(&ace.DefaultSocketModel{protocol.LOGIN, -1, LOGIN_SREQ, []byte("err")})
		}
	}()
	//解开json
	loginData := &AccountDTO{}
	err := json.Unmarshal(message.Message, &loginData)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println("申请登录", loginData.Username, loginData.Password)
	//登陆具体逻辑
	loginResult := this.login(session, &loginData.Username, &loginData.Password)
	//登录结果 响应
	session.Write(&ace.DefaultSocketModel{protocol.LOGIN, 88, LOGIN_SREQ, []byte(loginResult)})
}

//断线重新登陆
func (this *Handler) ReLoginProcess(session *ace.Session, message ace.DefaultSocketModel) {
	//错误处理
	defer func() {
		if r := recover(); r != nil {
			//有错误的话将返回"err"
			session.Write(&ace.DefaultSocketModel{protocol.LOGIN, -1, RELOGIN_SRES, []byte("err")})
		}
	}()
	//解开json
	loginData := &AccountDTO{}
	err := json.Unmarshal(message.Message, &loginData)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println("申请登录", loginData.Username, loginData.Password)
	//登陆具体逻辑
	loginResult := this.login(session, &loginData.Username, &loginData.Password)
	//登录结果 响应
	session.Write(&ace.DefaultSocketModel{protocol.LOGIN, -1, RELOGIN_SRES, []byte(loginResult)})
}

//退出登录
func (this *Handler) ExitProcess(session *ace.Session) {
	data.SyncAccount.SessionClose(session)
}

//发送心跳包后，客户端的回信. 一旦运行此方法，则说明旧的session是活跃的
func (this *Handler) HeartPackage(oldSession *ace.Session) {
	oldSession.IsColse = false
}

//******************************************************************
//                       注册具体逻辑
//******************************************************************
func (this *Handler) reg(un *string, psw *string, phone *string) bool {
	//错误处理
	defer func() bool {
		if r := recover(); r != nil {
			return false
		}
		return false
	}()

	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	tools.CheckErr(err)
	//先对比数据库 看是否已被注册
	stmtOut, err := db.Prepare("SELECT username FROM userinfo WHERE username = ?")
	var username string
	err = stmtOut.QueryRow(*un).Scan(&username)
	//fmt.Printf("The square is: %s", username)
	if *un == username {
		fmt.Println("这账号已被注册")
		return false
	}
	//添加账户
	stmt, err := db.Prepare("INSERT userinfo SET username=?,password=?,face=?,nickname=?,description=?,phone=?,online=?,level=?,provider=?,lasttime=?,createdtime=?")
	tools.CheckErr(err)
	_, err = stmt.Exec(*un, *psw, "default.jpg", "叮叮小鸟", "我还没有个性签名", *phone, 0, 0, "不是厂家", time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))
	tools.CheckErr(err)
	//添加数据
	stmtIns, err := db.Prepare("INSERT userdata SET username=?,offlinemsg=?,friends=?,groups=?,commodity=?,rooms=?,likes=?")
	tools.CheckErr(err)
	_, err = stmtIns.Exec(*un, "[]", "111111", "[]", "[]", "[]", "[]")
	tools.CheckErr(err)
	return true
}

//******************************************************************
//                       登陆具体逻辑
//如果登录成功//返回用户级别，如果是企业用户返回企业名字，如果需要判断是否重登录则返回空字符串
//******************************************************************
func (this *Handler) login(session *ace.Session, un *string, psw *string) string {
	//错误处理
	defer func() bool {
		if r := recover(); r != nil {
			return false
		}
		return false
	}()
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	tools.CheckErr(err)
	//验证账号与密码
	stmtOut, err := db.Prepare("SELECT password ,level,provider FROM userinfo WHERE username = ?")
	var password string
	var level string
	var provider string
	stmtOut.QueryRow(*un).Scan(&password, &level, &provider)
	//fmt.Printf("The square is: %s", password)
	if password == "" {
		return USER_NO //用户名不存在
	}
	if *psw == password {
		//fmt.Printf("账号%s与密码%s匹配  ", *un, *psw)
		//检验此账号是否已经登录
		_, ok := data.SyncAccount.AccountSession[*un] //****************
		if ok {                                       //如果能在此切片中取出值，说明已登录
			//go this.heartPackage(tempSession, session)

			return USER_RELOGIN
		} else { //可以登录

			fmt.Println(*un, "<<<<<-------------可以登录")
			stmtUp, err := db.Prepare("update userinfo set online=?,lasttime=? where username=?") //更新最后登录时间
			tools.CheckErr(err)
			_, err = stmtUp.Exec(1, time.Now().Format("2006-01-02 15:04:05"), *un) //更改登录状态为1
			tools.CheckErr(err)
			//此账号与session相关联
			data.SyncAccount.AccountSession[*un] = session
			data.SyncAccount.SessionAccount[session] = *un
			//登陆成功
			return *un
		}
	} else {
		fmt.Printf("账号密码不对")
		return USER_PSDERR
	}
	return USER_NO
}

//旧的心跳包
//func (this *Handler) heartPackage(oldSession *ace.Session, newSession *ace.Session) {
//	//保存老的、新的session
//	this.OldNewSession[oldSession] = newSession
//	oldSession.Write(&ace.DefaultSocketModel{protocol.LOGIN, -1, HEART_PACKAGE_SREQ, []byte("are you there?")})
//	oldSession.IsColse = true
//	//3秒后判断心跳是否活跃
//	timer := time.NewTicker(time.Duration(2000) * time.Millisecond)
//	for {
//		select {
//		case <-timer.C:
//			//			fmt.Println("3秒到了")
//			//老的很活跃，通知新的不可以登陆
//			if oldSession.IsColse == false {
//				newSession.Write(&ace.DefaultSocketModel{protocol.LOGIN, -1, LOGIN_SREQ, []byte(USER_RELOGIN)})
//			} else { //老的不活跃，让客户端重新登陆
//				oldSession.Close()
//				newSession.Write(&ace.DefaultSocketModel{protocol.LOGIN, -1, RETRY_LOGIN_SREQ, []byte("")})
//			}
//			delete(this.OldNewSession, oldSession)
//		}
//		return
//	}
//}
