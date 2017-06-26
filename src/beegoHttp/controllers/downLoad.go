package controllers

import (
	//"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/astaxie/beego"
)

type DownLoadController struct {
	beego.Controller
}

func (c *DownLoadController) Get() {
	//	fmt.Println("文件下载", c.Ctx.Request.URL.Path)
	path := Substr(c.Ctx.Request.URL.Path, 1, len(c.Ctx.Request.URL.Path))
	//fmt.Print("下载路径", path)
	//下载文件
	c.Ctx.Output.Download(path)
}

//获取程序路径
func getCurrentPath() string {
	s, err := exec.LookPath(os.Args[0])
	checkErr(err)
	i := strings.LastIndex(s, "\\")
	path := string(s[0 : i+1])
	return path
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0
	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}
	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}
