package controllers

import (
	"beegoHttp/models"
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
)

type GroupBaseInfoController struct {
	beego.Controller
}

func (c *GroupBaseInfoController) Get() {

	gid, err := c.GetInt("gid")
	if err != nil {
		fmt.Println(err)
		return
	}
	var groupBaseInfoModel models.GroupBaseInfoModel
	groupInfo := groupBaseInfoModel.GetGroupBaseInfoModel(gid)

	jsons, _ := json.Marshal(groupInfo)

	fmt.Println("群的信息是：", string(jsons))
	c.Ctx.WriteString(string(jsons))
}
