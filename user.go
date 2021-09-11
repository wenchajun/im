package main

import (
	"net"
	"strings"
)

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
//给当前user对应的客户端发送消息
func (user *User) SendMsg(msg string)  {
	user.Conn.Write([]byte(msg))
}

//用户处理消息的业务
func (user *User) Domessage(msg string) {
	if msg=="who" {
		//查询当前用户有哪些
		user.Server.maplock.Lock()
         for _, useronline :=range user.Server.OnlineMap{
         	onlineMsg:= "["+useronline.Addr+"]"+useronline.Name+":"+"在线"
         	user.SendMsg(onlineMsg)
		 }
		user.Server.maplock.Unlock()
	}else if len(msg)>7 && msg[:7]=="rename|" {
  //消息格式：rename|张三
		newName:=strings.Split(msg,"|")[1]
		//判断name是否已经存在
		_,ok:=user.Server.OnlineMap[newName]
		if ok{
			user.SendMsg("当前用户已经存在")
		}else{
			user.Server.maplock.Lock()
			delete(user.Server.OnlineMap,user.Name)
			user.Server.OnlineMap[newName]=user
			user.Server.maplock.Unlock()

			user.Name=newName
			user.SendMsg("你已经更新了用户名"+user.Name+"\n")
		}
	}else if len(msg)>4&& msg[:3]=="to|"{
		//消息格式 to|张三|消息内容

		//获取对方的用户名
		remotename:=strings.Split(msg,"|")[1]
		if remotename ==""{
			user.SendMsg("消息格式不正确")
			return
		}


		//根据用户名得到对方user对象

		remoteUser,ok := user.Server.OnlineMap[remotename]
		if !ok {
			user.SendMsg("用户名不存在")
			return
		}

		//获取消息内容，，通过对方的user对象将消息发送过去
		content:=strings.Split(msg,"|")[2]
		if content==""{
			user.SendMsg("无消息，请重发")
			return
		}
		remoteUser.SendMsg(user.Name+"对你说"+content)

	}else {
		user.Server.BroadCast(user, msg)
	}

}

//监听当前user channe的方法，一旦有消息，就直接发送给对端客户端
func (user *User) ListenMessage() {

	for {
		msg := <-user.C
		user.Conn.Write([]byte (msg + "\n"))

	}

}
