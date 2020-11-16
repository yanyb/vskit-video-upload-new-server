package service

import (
	"fmt"
	"go-app/app"
	"net"
)

// 启动服务
func Start() {
	l, err := net.Listen("tcp", app.GetAddr())
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}
		go func() {
			if err := recover(); err != nil {
				fmt.Println("handle conn:", err)
			}
			HandleConn(c)
		}()
	}
}
