package kk

import (
	"container/list"
	"log"
	"net"
	"strings"
)

type TCPServer struct {
	Neuron
	chan_break chan bool
	clients    *list.List

	OnMessage      func(message *Message, from INeuron)
	OnStart        func()
	OnFail         func(err error)
	OnAccept       func(client *TCPClient)
	OnDisconnected func(client *TCPClient)
}

func (c *TCPServer) Break() {
	c.chan_break <- true
}

func (c *TCPServer) Send(message *Message, from INeuron) {
	var e = c.clients.Front()
	for e != nil {
		var f = e.Value.(*TCPClient)
		if (from == nil || (from != f)) && f.name != "" && strings.HasPrefix(message.To, f.name) {
			f.Send(message, from)
		}
		e = e.Next()
	}
}

func NewTCPServer(name string, address string, maxconnections int) *TCPServer {

	var v = TCPServer{}

	v.name = name
	v.address = address
	v.chan_break = make(chan bool)
	v.clients = list.New()

	go func() {

		var listen, err = net.Listen("tcp", address)

		if err != nil {
			func(err error) {
				GetDispatchMain().Async(func() {
					if v.OnFail != nil {
						v.OnFail(err)
					}
				})
			}(err)
			return
		} else {
			GetDispatchMain().Async(func() {
				if v.OnStart != nil {
					v.OnStart()
				}
			})
		}

		defer close(v.chan_break)
		defer listen.Close()

		var num_connections = 0
		var chan_num_connections = make(chan bool)

		defer close(chan_num_connections)

		go func() {

			for {

				for num_connections >= maxconnections {
					if !<-chan_num_connections {
						return
					}
				}

				var conn, err = listen.Accept()

				if err != nil {
					func(err error) {
						GetDispatchMain().Async(func() {
							if v.OnFail != nil {
								v.OnFail(err)
							}
						})
					}(err)
					return
				}

				if conn == nil {
					return
				}

				func(conn net.Conn) {

					GetDispatchMain().Async(func() {

						num_connections += 1

						log.Printf("connections: %d\n", num_connections)

						var client = NewTCPClientConnection(conn)
						v.clients.PushBack(client)
						client.OnDisconnected = func(err error) {
							if v.OnDisconnected != nil {
								v.OnDisconnected(client)
							}
							var e = v.clients.Front()
							for e != nil {
								var f = e.Value.(*TCPClient)
								if f == client {
									var n = e.Next()
									v.clients.Remove(e)
									e = n
									break
								}
								e = e.Next()
							}
							if num_connections >= maxconnections {
								num_connections -= 1
								chan_num_connections <- true
							} else {
								num_connections -= 1
							}
							log.Printf("connections: %d\n", num_connections)
						}
						client.OnMessage = func(message *Message) {
							if v.OnMessage != nil {
								v.OnMessage(message, client)
							}
						}
						if v.OnAccept != nil {
							v.OnAccept(client)
						}
					})
				}(conn)
			}
		}()

		<-v.chan_break

		GetDispatchMain().Async(func() {
			var e = v.clients.Front()
			for e != nil {
				var f = e.Value.(*TCPClient)
				f.Break()
				e = e.Next()
			}
		})

	}()

	return &v
}
