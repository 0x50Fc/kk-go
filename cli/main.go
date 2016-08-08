package main

import (
	"fmt"
	"github.com/hailongz/kk-go/kk"
	"os"
	"time"
)

func help() {
	fmt.Println("kk-go-cli <name> <0.0.0.0:8080>")
}

func main() {

	var args = os.Args
	var name string = ""
	var localAddr string = ""

	if len(args) > 2 {
		name = args[1]
		localAddr = args[2]
	} else {
		help()
		return
	}

	var cli *kk.TCPClient = nil
	var cli_connect func() = nil

	cli_connect = func() {
		fmt.Println("connect " + localAddr + " ...")
		cli = kk.NewTCPClient(name, localAddr)
		cli.OnConnected = func() {
			fmt.Println(cli.Address())
		}
		cli.OnDisconnected = func(err error) {
			fmt.Println("disconnected: " + cli.Address() + " error:" + err.Error())
			kk.GetDispatchMain().AsyncDelay(cli_connect, time.Second)
		}
		cli.OnMessage = func(message *kk.Message) {
			fmt.Println(message)
		}
	}

	cli_connect()

	go func() {

		for {

			var to string
			var content string

			fmt.Scanf("%s %s", &to, &content)

			fmt.Printf("%s %s\n", to, content)

			func(to string, content string) {
				kk.GetDispatchMain().Async(func() {
					var m = kk.Message{"MESSAGE", cli.Name(), to, "text", []byte(content)}
					cli.Send(&m, nil)
				})

			}(to, content)
		}
	}()

	kk.DispatchMain()

}
