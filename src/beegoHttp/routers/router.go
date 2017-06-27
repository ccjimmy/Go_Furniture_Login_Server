package routers

import (
	"beegoHttp/controllers"

	"github.com/astaxie/beego"
)

func init() {

	//默认路由
	beego.Router("/", &controllers.MainController{})
	//一个人的基本信息
	beego.Router("/baseInfo", &controllers.BaseInfoController{})
	//下载文件
	beego.Router("/res/*", &controllers.DownLoadController{})
	//查找好友
	beego.Router("/findFriend", &controllers.FindFriendController{})
	//拉取离线消息
	beego.Router("/offlinemsg", &controllers.OfflineMsgController{})
	//拉取好友列表
	beego.Router("/friendList", &controllers.FriendListController{})
}
