package main

import "net"

type User struct {
	Name string
	Addr string
	C  chan string
	Conn net.Conn
}
//创建一个用户的api
func NewUser(coon net.Conn) *User{
	userAddr :=coon.RemoteAddr().String()
	user:=&User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		Conn: coon,
	}
		//启动监听当前user channel消息的goroutine
	go user.ListenMessage()
	return user

}

//监听当前user channe的方法，一旦有消息，就直接发送给对端客户端
func (user *User) ListenMessage()  {

	for{
		msg:=<-user.C
		user.Conn.Write([]byte (msg+"\n"))

	}




}

