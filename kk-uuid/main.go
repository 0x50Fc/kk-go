package main

import (
	"fmt"
	"github.com/hailongz/kk-go/kk"
	"log"
	"os"
	"strconv"
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

	var replay func(message *kk.Message) bool = nil

	replay, _ = kk.TCPClientConnect(name, address, map[string]interface{}{"exclusive": true}, func(message *kk.Message) {
		if message.Method == "REQUEST" {
			var v = kk.Message{message.Method, name, message.From, "text", []byte(strconv.FormatInt(kk.UUID(), 10))}
			replay(&v)
		} else {
			var v = kk.Message{"NOIMPLEMENT", message.To, message.From, "", []byte("")}
			log.Println(v)
			replay(&v)
		}
	})

	kk.DispatchMain()

}
