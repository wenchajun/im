package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int
	//在线用户列表
	OnlineMap map[string]*User
	maplock   sync.RWMutex
	//消息广播的channel
	Message chan string
}

//创建一个server接口
func Newserver(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server

}

//监听message广播消息channe的goroutine，一旦有消息就发给全部的在线user
func (server Server) ListenMessage() {
	for {
		msg := <-server.Message
		//发送给全部user
		server.maplock.Lock()
		for _, cli := range server.OnlineMap {
			cli.C <- msg
		}
		server.maplock.Unlock()

	}

}

//广播消息
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	server.Message <- sendMsg

}

func (server *Server) Handler(conn net.Conn) {

	fmt.Println("链接建立成功")
	user := NewUser(conn, server)

	user.Online()

	//监听用户是否活跃的channel
	isLive:=make(chan bool)

	//接收客户端发送的信息
	go func() {
		buff := make([]byte, 4096)
		for {
			n, err := conn.Read(buff)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("coon read err:", err)
				return
			}
			//提取用户的消息 （去除“\n”）
			msg := string(buff[:n-1])
			//用户针对msg进行处理
			user.Domessage(msg)
//用户的任意消息代表用户当前活跃
			isLive<-true

		}
	}()

	//当前handler阻塞
	select {
	case <-isLive:
		//当前用户活跃，应该重置定时器
		//不做任何事情，为了激活select，更新下面的定时器
	case <- time.After(time.Second*10):
		//已经超时，将当前user强制关闭
	user.SendMsg("你被踢了")
	//销毁用的资源
	close(user.C)
	//关闭连接
	conn.Close()
	//退出当前handler
		return
	//runtime.Goexit()
	}

}

func (server *Server) Start() {
	//socker listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.listen err:", err)
	}
	//close lisener
	defer listener.Close()
	go server.ListenMessage()
	for {
		//accept

		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener eaccept:", err)
			continue
		}

		//do handler
		go server.Handler(conn)
	}

}
