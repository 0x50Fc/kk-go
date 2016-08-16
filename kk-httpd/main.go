package main

import (
	"fmt"
	"github.com/hailongz/kk-go/kk"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func help() {
	fmt.Println("kk-httpd <name> <127.0.0.1:8700> <:8900> /kk/")
}

func main() {

	var args = os.Args
	var name string = ""
	var address string = ""
	var httpaddress string = ""
	var alias string = ""

	if len(args) > 4 {
		name = args[1]
		address = args[2]
		httpaddress = args[3]
		alias = args[4]
	} else {
		help()
		return
	}

	var https = make(map[int64]chan kk.Message)
	var cli *kk.TCPClient = nil
	var cli_connect func() = nil

	cli_connect = func() {
		log.Println("connect " + address + " ...")
		cli = kk.NewTCPClient(name, address, map[string]interface{}{"exclusive": true})
		cli.OnConnected = func() {
			log.Println(cli.Address())
		}
		cli.OnDisconnected = func(err error) {
			log.Println("disconnected: " + cli.Address() + " error:" + err.Error())
			kk.GetDispatchMain().AsyncDelay(cli_connect, time.Second)
		}
		cli.OnMessage = func(message *kk.Message) {

			log.Println(message)

			if strings.HasPrefix(message.Type, "text") {
				log.Println(string(message.Content))
			}

			var i = strings.LastIndex(message.To, ".")
			var id, _ = strconv.ParseInt(message.To[i+1:], 10, 64)
			var ch = https[id]

			if ch != nil {
				if message.Method == "REQUEST" {
					ch <- *message
					delete(https, id)
				} else {
					var m = kk.Message{"UNAVAILABLE", "", "", "", []byte("")}
					ch <- m
					delete(https, id)
				}
			}

		}
	}

	cli_connect()

	var http_handler = func(w http.ResponseWriter, r *http.Request) {

		var id = kk.UUID()
		var ch = make(chan kk.Message)
		defer close(ch)

		var body = make([]byte, r.ContentLength)
		var contentType = r.Header.Get("Content-Type")
		var to = r.RequestURI[len(alias):]
		var n, err = r.Body.Read(body)
		defer r.Body.Close()

		if err != nil && err != io.EOF {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		} else if int64(n) != r.ContentLength {
			log.Printf("%d %d\n", n, r.ContentLength)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		kk.GetDispatchMain().Async(func() {
			if cli != nil {
				https[id] = ch
				var m = kk.Message{"REQUEST", fmt.Sprintf("%s%d", cli.Name(), id), to, contentType, body}
				cli.Send(&m, nil)
			} else {
				var m = kk.Message{"TIMEOUT", "", "", "", []byte("")}
				ch <- m
				delete(https, id)
			}
		})

		kk.GetDispatchMain().AsyncDelay(func() {

			var ch = https[id]

			if ch != nil {
				var m = kk.Message{"TIMEOUT", "", "", "", []byte("")}
				ch <- m
				delete(https, id)
			}

		}, time.Second)

		var m, ok = <-ch

		if !ok {
			w.WriteHeader(http.StatusGatewayTimeout)
		} else {
			if m.Method == "TIMEOUT" {
				w.WriteHeader(http.StatusGatewayTimeout)
			} else if m.Method == "UNAVAILABLE" {
				w.WriteHeader(http.StatusServiceUnavailable)
			} else if m.Method == "REQUEST" {
				w.Header().Add("From", m.From)
				if strings.HasPrefix(m.Type, "text") {
					w.Header().Add("Content-Type", m.Type+"; charset=utf-8")
				} else {
					w.Header().Add("Content-Type", m.Type)
				}
				w.Header().Add("Content-Length", strconv.Itoa(len(m.Content)))
				w.WriteHeader(http.StatusOK)
				w.Write(m.Content)
			} else {
				w.WriteHeader(http.StatusUnsupportedMediaType)
			}
		}
	}

	go func() {

		http.HandleFunc(alias, http_handler)

		log.Println("httpd " + httpaddress)

		log.Fatal(http.ListenAndServe(httpaddress, nil))

	}()

	kk.DispatchMain()

}
