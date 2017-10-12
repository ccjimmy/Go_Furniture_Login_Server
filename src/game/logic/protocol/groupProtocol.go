package protocol

//关于群的套接字协议
const (
	MODIFY_GROUP_INFO_SREQ = 1 //群数据发生了改变
	MODIFY_GROUP_FACE_SREQ = 2 //群头像发生了改变
)

//关于群的http协议
const (
	GROUP_BASE_INFO           = "1" //获取一个群的基本信息
	MODIFY_GROUP_BASE_INFO    = "2" //修改一个群的基本信息
	MODIFY_ENTER_GROUP_METHOD = "3" //修改一个群的加群方式
)
