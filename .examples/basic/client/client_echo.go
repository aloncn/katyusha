package main

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/katyusha/.examples/basic/protocol"
	"github.com/gogf/katyusha/krpc"
	"golang.org/x/net/context"
	"time"
)

// go run client_echo.go
func main() {
	conn, err := krpc.Client.NewGrpcClientConn("echo")
	if err != nil {
		g.Log().Fatal(err)
	}
	defer conn.Close()

	echoClient := protocol.NewEchoClient(conn)
	for i := 0; i < 500; i++ {
		res, err := echoClient.Say(context.Background(), &protocol.SayReq{Content: "Hello"})
		if err != nil {
			g.Log().Error(err)
			time.Sleep(time.Second)
			continue
		}
		time.Sleep(time.Second)
		g.Log().Print("Response:", res.Content)
	}
}
