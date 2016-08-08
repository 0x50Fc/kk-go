package kk

import (
	"log"
	"net"
)

type TCPClient struct {
	Neuron
	chan_break   chan bool
	chan_message chan *Message
	isconnected  bool

	OnMessage func(message *Message)

	OnConnected    func()
	OnDisconnected func(err error)
}

func (c *TCPClient) Break() {
	c.chan_break <- true
}

func (c *TCPClient) Send(message *Message, from INeuron) {
	var m = *message
	c.chan_message <- &m
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
	v.chan_message = make(chan *Message)
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
			return
		} else {
			GetDispatchMain().Async(func() {
				v.onConnected()
			})
		}

		defer close(v.chan_message)
		defer close(v.chan_break)
		defer conn.Close()

		go func() {

			var rd = NewMessageReader()

			for {

				var m, err = rd.Read(conn)

				if err != nil {
					func(err error) {
						v.chan_break <- true
						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})
					}(err)
					break
				}

				if m != nil {
					func(message Message) {
						GetDispatchMain().Async(func() {
							if v.OnMessage != nil {
								v.OnMessage(&message)
							}
						})
					}(*m)
				}
			}

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
						v.chan_break <- true
						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})
					}(err)
					break
				}

				if r {

					var m = <-v.chan_message

					if m == nil {

						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})

						break
					} else {
						wd.Write(m)
						r, err = wd.Done(conn)
						if err != nil {
							func(err error) {
								GetDispatchMain().Async(func() {
									v.onDisconnected(err)
								})
							}(err)
							break
						}
					}

				}

			}
		}()

		<-v.chan_break

	}()

	return &v
}

func NewTCPClientConnection(conn net.Conn) *TCPClient {

	var v = TCPClient{}

	v.name = ""
	v.address = conn.RemoteAddr().String()
	v.chan_message = make(chan *Message)
	v.chan_break = make(chan bool)
	v.isconnected = true

	go func() {

		defer close(v.chan_message)
		defer close(v.chan_break)
		defer conn.Close()

		go func() {

			var rd = NewMessageReader()

			for {

				var m, err = rd.Read(conn)

				if err != nil {
					func(err error) {
						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})
					}(err)
					break
				}

				if m != nil {
					func(message Message) {
						GetDispatchMain().Async(func() {
							if message.Method == "CONNECT" {
								v.name = message.From
								log.Println("CONNECT " + v.name + " address: " + v.Address())
							}
							if v.OnMessage != nil {
								v.OnMessage(&message)
							}
						})
					}(*m)
				}
			}

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

					var m = <-v.chan_message

					if m == nil {
						break
					} else {
						wd.Write(m)
						r, err = wd.Done(conn)
						if err != nil {
							func(err error) {
								GetDispatchMain().Async(func() {
									v.onDisconnected(err)
								})
							}(err)
							break
						}
					}

				}

			}
		}()

		<-v.chan_break

	}()

	return &v
}
