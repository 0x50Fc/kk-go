package main

import (
	"fmt"
	"github.com/hailongz/kk-go/kk"
	"os"
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

	var replay, getname = kk.TCPClientConnect(name, address, nil, func(message *kk.Message) {
		fmt.Println(message)
		if message.Type == "text" {
			fmt.Println(string(message.Content))
		}
	})

	go func() {

		for {

			var method string
			var to string
			var content string

			fmt.Scanf("%s %s %s", &method, &to, &content)

			func(to string, content string) {

				kk.GetDispatchMain().Async(func() {
					var m = kk.Message{method, getname(), to, "text", []byte(content)}
					replay(&m)
				})

			}(to, content)
		}
	}()

	kk.DispatchMain()

}
