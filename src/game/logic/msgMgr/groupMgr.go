package msgMgr

import (
	"fmt"

	"database/sql"
	"encoding/json"
	"time"

	"io"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

const groupHistoryToDBAmount = 200 //写入数据库的聊天记录的条数

//消息数据结构
type GroupModel struct {
	gid        string
	Master     string
	Managers   string
	Members    string
	History    []MessageModel //聊天历史
	canDestory bool
}

type GroupManager struct {
	Groups map[string]*GroupModel
}

var GroupMgr = &GroupManager{Groups: make(map[string]*GroupModel)}

func (this *GroupManager) GetOneGroupManager(gid string) *GroupModel {
	group, ok := this.Groups[gid]
	if ok {
		return group
	} else {
		db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
		defer db.Close()
		if err != nil {
			fmt.Println(err)
		}
		//获取已存在群数据
		stmtOut, err := db.Prepare("SELECT master,manager,member,history FROM groups WHERE gid = ?")
		var master string
		var manager string
		var member string
		var history string
		err = stmtOut.QueryRow(gid).Scan(&master, &manager, &member, &history)
		if err != nil {
			fmt.Println("获取一个群失败", err)
			return nil
		}
		if history == "" {
			history = "[]"
		}
		//初始化聊天记录切片
		//解开json
		historyMsgModels := []MessageModel{}
		err = json.Unmarshal([]byte(history), &historyMsgModels)
		if err != nil {
			fmt.Println("初始化群聊天历史出错:", err)
		}
		fmt.Println("初始化一个群:", master, manager, member, len(historyMsgModels))
		var groupModel = &GroupModel{gid, master, manager, member, historyMsgModels, false}
		go groupModel.destory()
		this.Groups[gid] = groupModel
		return groupModel
	}
}

//销毁一个群模型
func (this *GroupModel) destory() {

	timer := time.NewTicker(time.Duration(60) * time.Second)
	for {
		select {
		case <-timer.C:

			if this.canDestory == false {
				this.canDestory = true
			} else { //可以销毁了
				this.saveHistory() //持久化数据
				delete(GroupMgr.Groups, this.gid)
				fmt.Println("这个群被销毁了", this.gid)
				return
			}
		}
	}
}

//一个群添加一条聊天记录
func (this *GroupModel) OnGroupActive(msg *MessageModel) {
	this.canDestory = false
	this.History = append(this.History, *msg)
}

//持久化一个群的聊天记录
func (this *GroupModel) saveHistory() {
	fmt.Println("持久化一个群的聊天记录，条数", len(this.History))
	var toDB []MessageModel
	if len(this.History) <= groupHistoryToDBAmount { //条数较少
		toDB = this.History[:len(this.History)]
	} else {
		toDB = this.History[len(this.History)-groupHistoryToDBAmount:]
	}

	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/furniture?charset=utf8")
	defer db.Close()
	if err != nil {
		fmt.Println(err)
	}
	toDBStr, _ := json.Marshal(toDB)
	//fmt.Println("写入数据库的群消息列表 ", string(toDBStr))
	stmtUp, err := db.Prepare("update groups set history=? where gid=?") //更新好友列表
	if err != nil {
		fmt.Println(err)
	}
	_, err = stmtUp.Exec(string(toDBStr), this.gid)
	if err != nil {
		fmt.Println(err)
	}
	//更多的数据保存到txt
	if len(this.History) > groupHistoryToDBAmount {
		toTXT := this.History[:len(this.History)-groupHistoryToDBAmount]
		toTXTStr, _ := json.Marshal(toTXT)

		var filename = "res/groupHistory/" + this.gid + ".txt"
		var f *os.File
		var err1 error
		if checkFileExist(filename) { //如果文件存在
			f, err1 = os.OpenFile(filename, os.O_APPEND, 0666) //打开文件
			fmt.Println("群历史文件存在")
		} else {
			f, err1 = os.Create(filename) //创建文件
			fmt.Println("群历史文件不存在")
		}
		if err1 != nil {
			fmt.Println("持久化群历史记录到txt时发生错误", err1)
		}
		_, err1 = io.WriteString(f, string(toTXTStr)+"&") //写入txt文件(留一个分隔符)
		if err1 != nil {
			fmt.Println("持久化群历史记录到txt时发生错误", err1)
		}
	}
}

/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkFileExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
