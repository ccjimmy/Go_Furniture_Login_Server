package msgMgr

import (
	"database/sql"
	"fmt"
	"time"
	"tools"

	_ "github.com/go-sql-driver/mysql"
)

type AllGroupManager struct {
	Groups map[string]*GroupManager
}

//单个群管理器
type GroupManager struct {
	gid      string
	Master   string
	Managers string
	Members  string
	canDes   bool //是否可以销毁
	//	CloseTimer *time.Ticker
}

var GroupMgr = &AllGroupManager{make(map[string]*GroupManager)}

//获取一个群管理器
func (this *AllGroupManager) GetOneGroupManager(gid string) *GroupManager {
	groupManager, ok := this.Groups[gid]
	if ok {
		return groupManager
	} else {
		this.initGroupManager(&gid)
		return this.Groups[gid]
	}
}

//初始化一个群
func (this *AllGroupManager) initGroupManager(gid *string) {
	//读取这个群的数据库
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	tools.CheckErr(err)
	stmtOut, err := db.Prepare("SELECT master,manager,member FROM groups WHERE gid = ?")
	var master string
	var managers string
	var members string
	err = stmtOut.QueryRow(*gid).Scan(&master, &managers, &members)
	tools.CheckErr(err)
	//实例化
	fmt.Print("初始化一个群", *gid, "此群的成员信息是", master, managers, members, "\n")
	GroupMgr.Groups[*gid] = &GroupManager{*gid, master, managers, members, true}
	go GroupMgr.Groups[*gid].destory()
}

//销毁一个群管理器，每5分钟判断一次是否应该销毁
func (this *GroupManager) destory() {
	timer := time.NewTicker(time.Duration(10) * time.Second)
	for {
		select {
		case <-timer.C:
			if this.canDes {
				fmt.Println("这个群而已销毁了" + this.gid)
				timer.Stop()
				delete(GroupMgr.Groups, this.gid)
				return
			} else {
				this.canDes = true
			}
		}
	}
}

func (this *GroupManager) OnGroupActive() {
	this.canDes = false
}
