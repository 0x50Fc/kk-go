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

const twepoch = int64(1424016000000)

var _id = twepoch

func milliseconds() int64 {
	return time.Now().UnixNano() / 1e6
}

func nextId() int64 {
	var id = milliseconds()
	for _id == id {
		id = milliseconds()
	}
	_id = id
	return _id - twepoch
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
			var r = kk.Message{message.Method, name, message.From, "text", []byte(strconv.FormatInt(nextId(), 10))}
			cli.Send(&r, nil)
		}
	}

	cli_connect()

	kk.DispatchMain()

}
