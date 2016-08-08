package kk

import (
	"log"
	"net"
)

type TCPClient struct {
	Neuron
	chan_break   chan bool
	chan_message chan Message
	isconnected  bool

	OnMessage func(message *Message)

	OnConnected    func()
	OnDisconnected func(err error)
}

func (c *TCPClient) Break() {
	c.chan_break <- true
}

func (c *TCPClient) Send(message *Message, from INeuron) {
	c.chan_message <- *message
}

func (c *TCPClient) onDisconnected(err error) {
	if c.isconnected {
		c.isconnected = false
		if c.OnDisconnected != nil {
			c.OnDisconnected(err)
		}
	}
}

func (c *TCPClient) onConnected() {
	if !c.isconnected {
		c.isconnected = true
		if c.OnConnected != nil {
			c.OnConnected()
		}
	}
}

func NewTCPClient(name string, address string) *TCPClient {

	var v = TCPClient{}

	v.name = name
	v.address = address
	v.chan_message = make(chan Message)
	v.chan_break = make(chan bool)

	go func() {

		var conn, err = net.Dial("tcp", address)

		if err != nil {
			func(err error) {
				GetDispatchMain().Async(func() {
					if v.OnDisconnected != nil {
						v.OnDisconnected(err)
					}
				})
			}(err)
			close(v.chan_break)
			close(v.chan_message)
			return
		} else {
			GetDispatchMain().Async(func() {
				v.onConnected()
			})
		}

		var chan_rd = make(chan bool)
		var chan_wd = make(chan bool)

		go func() {

			var rd = NewMessageReader()

			for {

				var m, err = rd.Read(conn)

				if m != nil {
					func(message Message) {
						GetDispatchMain().Async(func() {
							if v.OnMessage != nil {
								v.OnMessage(&message)
							}
						})
					}(*m)
				} else if err != nil {
					func(err error) {
						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})
					}(err)
					break
				}
			}

			select {
			case v.chan_break <- true:
			default:
			}

			chan_rd <- true

		}()

		go func() {

			var wd = NewMessageWriter()

			{
				var m = Message{"CONNECT", name, "", "", []byte("")}
				wd.Write(&m)
			}

			for {

				var r, err = wd.Done(conn)

				if err != nil {
					func(err error) {
						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})
					}(err)
					break
				}

				if r {

					var m, ok = <-v.chan_message

					if !ok {
						break
					} else {
						wd.Write(&m)
					}

				}

			}

			select {
			case v.chan_break <- true:
			default:
			}

			chan_wd <- true
		}()

		<-v.chan_break

		close(v.chan_message)
		conn.Close()

		<-chan_rd
		<-chan_wd

		close(v.chan_break)
		close(chan_rd)
		close(chan_wd)

	}()

	return &v
}

func NewTCPClientConnection(conn net.Conn) *TCPClient {

	var v = TCPClient{}

	v.name = ""
	v.address = conn.RemoteAddr().String()
	v.chan_message = make(chan Message)
	v.chan_break = make(chan bool)
	v.isconnected = true

	go func() {

		var chan_rd = make(chan bool)
		var chan_wd = make(chan bool)

		go func() {

			var rd = NewMessageReader()

			for {

				var m, err = rd.Read(conn)

				if m != nil {
					func(message Message) {
						GetDispatchMain().Async(func() {
							if message.Method == "CONNECT" {
								v.name = message.From
								log.Println("CONNECT " + v.name + " address: " + v.Address())
							} else if v.OnMessage != nil {
								v.OnMessage(&message)
							}
						})
					}(*m)
				} else if err != nil {
					func(err error) {
						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})
					}(err)
					break
				}

			}

			select {
			case v.chan_break <- true:
			default:
			}

			chan_rd <- true

		}()

		go func() {

			var wd = NewMessageWriter()

			for {

				var r, err = wd.Done(conn)

				if err != nil {
					func(err error) {
						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})
					}(err)
					break
				}

				if r {

					var m, ok = <-v.chan_message

					if !ok {
						break
					} else {
						wd.Write(&m)
					}

				}

			}

			select {
			case v.chan_break <- true:
			default:
			}

			chan_wd <- true
		}()

		<-v.chan_break

		close(v.chan_message)
		conn.Close()

		<-chan_rd
		<-chan_wd

		close(v.chan_break)
		close(chan_rd)
		close(chan_wd)

	}()

	return &v
}
