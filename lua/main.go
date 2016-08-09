package main

import (
	"github.com/aarzilli/golua/lua"
	"github.com/hailongz/kk-go/kk"
	"log"
	"os"
	"time"
)

func help() {
	log.Println("kk-go-lua <name> <0.0.0.0:8080>")
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
	var L = lua.NewState()

	L.OpenLibs()

	defer L.Close()

	var onmessage = func(message *kk.Message) {

		if 0 != L.LoadFile("./"+message.From+"lua") {
			var err = L.ToString(-1)
			log.Fatal(err)
		} else {
			var err = L.Call(0, 1)
			if err != nil {
				log.Fatal(err)
			} else if L.IsFunction(-1) {
				L.PushGoStruct(message)
				L.PushGoStruct(cli)
				err = L.Call(2, 0)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		L.Pop(L.GetTop())
	}

	cli_connect = func() {
		log.Println("connect " + localAddr + " ...")
		cli = kk.NewTCPClient(name, localAddr)
		cli.OnConnected = func() {
			log.Println(cli.Address())
		}
		cli.OnDisconnected = func(err error) {
			log.Println("disconnected: " + cli.Address() + " error:" + err.Error())
			kk.GetDispatchMain().AsyncDelay(cli_connect, time.Second)
		}
		cli.OnMessage = onmessage
	}

	cli_connect()

	kk.DispatchMain()
}
