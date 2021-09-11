package main

import "net"

type User struct {
	Name   string
	Addr   string
	C      chan string
	Conn   net.Conn
	Server *Server
}

//创建一个用户的api
func NewUser(coon net.Conn, server *Server) *User {
	userAddr := coon.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		Conn:   coon,
		Server: server,
	}
	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()
	return user

}

func (user *User) Online() {
	//用户上线，将用户加入到onlinemap中
	user.Server.maplock.Lock()
	user.Server.OnlineMap[user.Name] = user
	user.Server.maplock.Unlock()
	//广播当前用户上线消息
	user.Server.BroadCast(user, "已上线")
}
func (user *User) Offline() {
	//用户下线，将用户剔除onlinemap中
	user.Server.maplock.Lock()
	delete(user.Server.OnlineMap, user.Name)
	user.Server.maplock.Unlock()
	//广播当前用户下线消息
	user.Server.BroadCast(user, "当前用户已下线")
}
func (user *User) Domessage(msg string) {
	user.Server.BroadCast(user, msg)
}

//监听当前user channe的方法，一旦有消息，就直接发送给对端客户端
func (user *User) ListenMessage() {

	for {
		msg := <-user.C
		user.Conn.Write([]byte (msg + "\n"))

	}

}
