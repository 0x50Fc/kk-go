package main

import (
	"container/list"
	"fmt"
	"github.com/hailongz/kk-go/kk"
	"os"
	"strconv"
	"time"
)

func help() {
	fmt.Println("kk-go-test <name> <0.0.0.0:8080> <100>")
}

func main() {

	var args = os.Args
	var name string = ""
	var localAddr string = ""
	var count int = 100

	if len(args) > 3 {
		name = args[1]
		localAddr = args[2]
		count, _ = strconv.Atoi(args[3])
	} else {
		help()
		return
	}

	var connects = list.New()

	var cli_connect = func() {
		fmt.Println("connect " + localAddr + " ...")
		var cli = kk.NewTCPClient(name, localAddr)
		cli.OnConnected = func() {
			fmt.Println(cli.Address())
		}
		cli.OnDisconnected = func(err error) {
			fmt.Println("disconnected: " + cli.Address() + " error:" + err.Error())
			var e = connects.Front()
			for e != nil {
				var c = e.Value.(*kk.TCPClient)
				if c == cli {
					var n = e.Next()
					connects.Remove(e)
					e = n
				} else {
					e = e.Next()
				}
			}
		}
		cli.OnMessage = func(message *kk.Message) {
			fmt.Println(message)
		}
		connects.PushBack(cli)
	}

	var i int

	for i = 0; i < count; i++ {
		time.Sleep(time.Millisecond * 20)
		kk.GetDispatchMain().AsyncDelay(cli_connect, time.Second)
	}

	kk.DispatchMain()

}
