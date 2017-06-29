// Session
package ace

import (
	"net"
	"time"
)

type Session struct {
	Conn        net.Conn
	Attribute   map[string]string //特性
	Encode      Encode
	IsHandshake bool
	IsColse     bool
}

func CreateSession(conn net.Conn, encode Encode) *Session {
	var session = &Session{conn, make(map[string]string), encode, true, false}
	go session.IsConnecting()
	return session
}

func (session *Session) Close() { //会话的内存并没有删除？？？
	session.Conn.Close()
	session.IsColse = true
	session.IsHandshake = false
	session = nil
}

//判断当前会话是否通畅
func (session *Session) IsConnecting() {
	//		tempSlice := make([]byte, 1)
	//		session.Conn.Write(tempSlice)
	timer := time.NewTicker(time.Duration(3000) * time.Millisecond)
	for {
		select {
		case <-timer.C:

			if session.IsHandshake == true {
				//fmt.Println("3秒到了,握手过")
				session.IsHandshake = false
			} else { //没有握手,视为掉线
				//fmt.Println("3秒到了,没有握手")
				session.Close()
				return
			}
		}
	}
}

//暂时用不到这两个方法
//func (session *Session) SetAttribute(key string, value string) {
//	fmt.Println("这里不会运行 Session中 设置特性")
//	session.Attribute[key] = value
//}

//func (session *Session) RemoveAttribute(key string) {
//	fmt.Println("这里不会运行 Session中 删除特性")
//	delete(session.Attribute, key)
//}

func (session *Session) Write(msg *DefaultSocketModel) {
	//	if session.IsColse {
	//		return
	//	}
	bytes := session.Encode.Encode(msg)
	session.Conn.Write(bytes)
	//fmt.Println("Session中写入数据，在这之前应该编码过了")
}

/**消息接收**/
func (session *Session) Read(buffer []byte) (int, bool) {
	//读取消息长度 错误则关闭客户端链接 并从客户端列表中移除客户端信息 否则返回长度
	session.IsHandshake = true //标记为已握手
	readLength, err := session.Conn.Read(buffer)
	if err != nil {
		session.Close()
		return 0, false
	}
	//fmt.Println("Session中读取到数据，这里应该最先运行，随后应该是解码")
	return readLength, true
}
