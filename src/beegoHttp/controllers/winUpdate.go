package controllers

import (
	//"fmt"
	"io/ioutil"

	"github.com/astaxie/beego"
)

type WinUpdateController struct {
	beego.Controller
}

func (c *WinUpdateController) Get() {
	info := LoadFile("conf/winUpdate.conf")
	c.Ctx.WriteString(info)

}

func LoadFile(path string) string {
	dat, err := ioutil.ReadFile(path)
	check(err)
	return string(dat)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
