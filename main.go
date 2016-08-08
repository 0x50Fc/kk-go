package main

import (
	"./kk"
	"log"
)

func main() {

	var server = kk.NewTCPServer("kk", "0.0.0.0:8080")

	server.OnStart = func() {
		log.Println(server.Address())
	}

	kk.DispatchMain()

}
