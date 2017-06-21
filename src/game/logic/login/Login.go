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
	REGIST_SREQ        = 1  //command=1代表这是注册结果
	LOGIN_SREQ         = 3  //3代表登陆成功
	HEART_PACKAGE_CREQ = 4  //心跳检查
	HEART_PACKAGE_SREQ = 5  //心跳检查
	RETRY_LOGIN_SREQ   = 6  //建议客户端尝试重新登陆
	EXIT_CREQ          = 10 //退出登录
)

//登陆结果变量
//如果登陆成功直接返回用户类别
const (
	USER_NO      = "10" //用户名不存在
	USER_RELOGIN = "11" //重复登录
	USER_PSDERR  = "12" //密码错误
	USER_WAIT    = "13" //需要进行是否重复登陆的判断，请等待
)

type Handler struct {
	OldNewSession map[*ace.Session]*ace.Session //判断重复登陆时，保存已存在的和新来的Session
}

var LoginHander = &Handler{OldNewSession: make(map[*ace.Session]*ace.Session)}

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
			session.Write(&ace.DefaultSocketModel{protocol.LOGIN, -1, REGIST_SREQ, []byte("err")})
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
	_, err = stmt.Exec(*un, *psw, "default", "default", "default", *phone, 0, 0, "不是厂家", time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))
	tools.CheckErr(err)
	//添加数据
	stmtIns, err := db.Prepare("INSERT userdata SET username=?,commodity=?,rooms=?,likes=?")
	tools.CheckErr(err)
	_, err = stmtIns.Exec(*un, "[]", "[]", "[]")
	tools.CheckErr(err)
	return true
}

//******************************************************************
//                       登陆具体逻辑
//如果登录成功返回用户级别，如果是企业用户返回企业名字，如果需要判断是否重登录则返回空字符串
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
		tempSession, ok := data.SyncAccount.AccountSession[*un] //****************
		if ok {                                                 //如果能在此切片中取出值，说明已登录
			go this.heartPackage(tempSession, session)
			return USER_WAIT
		} else { //可以登录
			fmt.Println(*un, "<<<<<-------------可以登录")
			stmtUp, err := db.Prepare("update userinfo set online=?,lasttime=? where username=?") //更新最后登录时间
			tools.CheckErr(err)
			_, err = stmtUp.Exec(1, time.Now().Format("2006-01-02 15:04:05"), *un) //更改登录状态为1
			tools.CheckErr(err)
			//此账号与session相关联
			data.SyncAccount.AccountSession[*un] = session
			data.SyncAccount.SessionAccount[session] = *un
			//登陆成功:普通用户只返回等级
			if level == "0" {
				//userLevel, _ := strconv.Atoi(level)
				return level
			}
			if level == "1" {
				//userLevel, _ := strconv.Atoi(level)
				return level + "&" + provider //经销商登陆成功
			}
			//登陆成功:企业用户返回等级+供应商名字
			if level == "2" {
				return level + "&" + provider //厂家登陆成功
			}
		}
	} else {
		fmt.Printf("账号密码不对")
		return USER_PSDERR
	}
	return USER_NO
}

//心跳包
func (this *Handler) heartPackage(oldSession *ace.Session, newSession *ace.Session) {
	//保存老的、新的session
	this.OldNewSession[oldSession] = newSession
	oldSession.Write(&ace.DefaultSocketModel{protocol.LOGIN, -1, HEART_PACKAGE_SREQ, []byte("are you there?")})
	oldSession.IsColse = true
	//3秒后判断心跳是否活跃
	timer := time.NewTicker(time.Duration(2000) * time.Millisecond)
	for {
		select {
		case <-timer.C:
			//			fmt.Println("3秒到了")
			//老的很活跃，通知新的不可以登陆
			if oldSession.IsColse == false {
				newSession.Write(&ace.DefaultSocketModel{protocol.LOGIN, -1, LOGIN_SREQ, []byte(USER_RELOGIN)})
			} else { //老的不活跃，让客户端重新登陆
				oldSession.Close()
				newSession.Write(&ace.DefaultSocketModel{protocol.LOGIN, -1, RETRY_LOGIN_SREQ, []byte("")})
			}
			delete(this.OldNewSession, oldSession)
		}
		return
	}
}
