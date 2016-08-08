package kk

import (
	"container/list"
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

func NewTCPServer(name string, address string) *TCPServer {

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

		go func() {

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
								continue
							}
							e = e.Next()
						}
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