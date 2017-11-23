package data

import (
	"ace"
	"database/sql"
	"fmt"
	"tools"

	_ "github.com/go-sql-driver/mysql"
)

type Vector3 struct {
	X float64
	Y float64
	Z float64
}
type Vector4 struct {
	X float64
	Y float64
	Z float64
	W float64
}

type Sync struct {
	AccountSession map[string]*ace.Session //根据Account得到Session ,踢下线时，根据账号把一个session踢掉
	SessionAccount map[*ace.Session]string //根据Session得到Account ,离线时 ，根据session 清理在线列表
}

var SyncAccount = &Sync{AccountSession: make(map[string]*ace.Session), SessionAccount: make(map[*ace.Session]string)}

//处理离线
func (this *Sync) SessionClose(session *ace.Session) {
	tempacc, ok := this.SessionAccount[session]
	//说明此session没有登陆 ，那就没有什么可以需要操作的
	if !ok {
		return
	}
	//更新用户表的最后登录时间
	db, err := sql.Open("mysql", tools.GetSQLStr())
	defer db.Close()
	tools.CheckErr(err)
	//更新是否在线的状态
	stmtUp, err := db.Prepare("update userinfo set online=? where username=?")
	tools.CheckErr(err)
	_, err = stmtUp.Exec(0, tempacc)
	tools.CheckErr(err)
	fmt.Println(this.SessionAccount[session], "----------->>>>>离开了", "持久化数据")
	//清除session与账号相关联的 map数据
	delete(this.AccountSession, tempacc)
	delete(this.SessionAccount, session)
}
