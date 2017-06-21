// AppStart
package main

import (
	"ace"
	_ "beegoHttp/routers"
	_ "bufio"
	"fmt"
	"game/logic"
	_ "os"

	"github.com/astaxie/beego"
)

func main() {
	server := ace.CreateServer()
	//此Handler即LogicHandler文件
	server.SetHandler(&logic.GameHandler{})
	fmt.Println("聊天服务器开启")
	go server.Start(10101)
	fmt.Println("beego开启")
	beego.Run()
}
