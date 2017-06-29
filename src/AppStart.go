// AppStart
package main

import (
	"ace"
	_ "beegoHttp/routers"
	_ "bufio"

	"game/logic"
	_ "os"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

func main() {

	server := ace.CreateServer()
	//此Handler即LogicHandler文件
	server.SetHandler(&logic.GameHandler{})
	go server.Start()
	orm.RegisterDataBase("default", "mysql", "root:@/furniture?charset=utf8", 30)
	orm.SetMaxIdleConns("default", 30) //设置数据库最大空闲连接
	orm.SetMaxOpenConns("default", 30) //设置数据库最大连接数
	beego.Run()
}
