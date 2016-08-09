package main

import (
	"fmt"
	"github.com/hailongz/kk-go/kk"
	"log"
	"os"
	"time"
)

func help() {
	fmt.Println("kk-cli <name> <0.0.0.0:8080>")
}

func main() {

	var args = os.Args
	var name string = ""
	var address string = ""

	if len(args) > 2 {
		name = args[1]
		address = args[2]
	} else {
		help()
		return
	}

	var cli *kk.TCPClient = nil
	var cli_connect func() = nil

	cli_connect = func() {
		log.Println("connect " + address + " ...")
		cli = kk.NewTCPClient(name, address)
		cli.OnConnected = func() {
			log.Println(cli.Address())
		}
		cli.OnDisconnected = func(err error) {
			log.Println("disconnected: " + cli.Address() + " error:" + err.Error())
			kk.GetDispatchMain().AsyncDelay(cli_connect, time.Second)
		}
		cli.OnMessage = func(message *kk.Message) {
			fmt.Println(message)
			if message.Type == "text" {
				fmt.Println(string(message.Content))
			}
		}
	}

	cli_connect()

	go func() {

		for {

			var method string
			var to string
			var content string

			fmt.Scanf("%s %s %s", &method, &to, &content)

			func(to string, content string) {

				kk.GetDispatchMain().Async(func() {
					var m = kk.Message{method, cli.Name(), to, "text", []byte(content)}
					cli.Send(&m, nil)
				})

			}(to, content)
		}
	}()

	kk.DispatchMain()

}
