package routers

import (
	"beegoHttp/controllers"

	"github.com/astaxie/beego"
)

func init() {

	//默认路由
	beego.Router("/", &controllers.MainController{})
	//刚登陆后的拉取自己基本信息
	beego.Router("/selfBaseInfo", &controllers.SelfBaseInfoController{})
}
