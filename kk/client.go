package kk

import (
	"errors"
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

func NewTCPClient(name string, address string) *TCPClient {

	var v = TCPClient{}

	v.name = name
	v.address = address
	v.chan_message = make(chan *Message)
	v.chan_break = make(chan bool)

	var onconnect = func() {
		if !v.isconnected {
			v.isconnected = true
			if v.OnConnected != nil {
				v.OnConnected()
			}
		}
	}

	var ondisconnect = func(err error) {
		if v.isconnected {
			v.isconnected = false
			if v.OnDisconnected != nil {
				v.OnDisconnected(err)
			}
		}
	}

	go func() {

		var conn, err = net.Dial("tcp", address)

		if err != nil {
			func(err error) {
				GetDispatchMain().Async(func() {
					ondisconnect(err)
				})
			}(err)
			return
		} else {
			GetDispatchMain().Async(func() {
				onconnect()
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
							ondisconnect(err)
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

			for {

				var r, err = wd.Done(conn)

				if err != nil {
					func(err error) {
						v.chan_break <- true
						GetDispatchMain().Async(func() {
							ondisconnect(err)
						})
					}(err)
					break
				}

				if r {

					var m = <-v.chan_message

					if m == nil {

						GetDispatchMain().Async(func() {
							ondisconnect(errors.New("break"))
						})

						break
					} else {
						r, err = wd.Write(conn, m)
						if err != nil {
							func(err error) {
								GetDispatchMain().Async(func() {
									ondisconnect(err)
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

	var ondisconnect = func(err error) {
		if v.isconnected {
			v.isconnected = false
			if v.OnDisconnected != nil {
				v.OnDisconnected(err)
			}
		}
	}

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
							ondisconnect(err)
						})
					}(err)
					break
				}

				if m != nil {
					func(message Message) {
						GetDispatchMain().Async(func() {
							if message.Method == "CONNECT" {
								v.name = message.From
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
							ondisconnect(err)
						})
					}(err)
					break
				}

				if r {

					var m = <-v.chan_message

					if m == nil {

						GetDispatchMain().Async(func() {
							ondisconnect(errors.New("break"))
						})

						break
					} else {
						r, err = wd.Write(conn, m)
						if err != nil {
							func(err error) {
								GetDispatchMain().Async(func() {
									ondisconnect(err)
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
