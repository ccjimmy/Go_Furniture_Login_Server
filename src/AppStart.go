// AppStart
package main

import (
	"ace"
	_ "bufio"
	"fmt"
	"game/logic"
	_ "os"
)

func main() {

	server := ace.CreateServer()
	//此Handler即LogicHandler文件
	server.SetHandler(&logic.GameHandler{})
	fmt.Println("家具服务器开启")
	server.Start(10101)

}
