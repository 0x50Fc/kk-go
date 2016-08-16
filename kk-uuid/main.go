package main

import (
	"fmt"
	"github.com/hailongz/kk-go/kk"
	"log"
	"os"
	"strconv"
	"time"
)

func help() {
	fmt.Println("kk-uuid <name> <0.0.0.0:8080>")
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
		cli = kk.NewTCPClient(name, address, map[string]interface{}{"exclusive": true})
		cli.OnConnected = func() {
			log.Println(cli.Address())
		}
		cli.OnDisconnected = func(err error) {
			log.Println("disconnected: " + cli.Address() + " error:" + err.Error())
			kk.GetDispatchMain().AsyncDelay(cli_connect, time.Second)
		}
		cli.OnMessage = func(message *kk.Message) {
			if message.Method == "REQUEST" {
				var v = kk.Message{message.Method, name, message.From, "text", []byte(strconv.FormatInt(kk.UUID(), 10))}
				cli.Send(&v, nil)
			} else {
				var v = kk.Message{"NOIMPLEMENT", message.To, message.From, "", []byte("")}
				log.Println(v)
				cli.Send(&v, nil)
			}
		}
	}

	cli_connect()

	kk.DispatchMain()

}
