package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/rs/xid"
)

func ENV(name string) string {
	result := ""
	if s, ok := os.LookupEnv(name); ok {
		result = s
	} else {
	}
	return result
}

func main() {
	socketHost := ENV("SOCKET_HOST")
	httpHost := ENV("HTTP_HOST")
	channels := make(map[string](chan string))
	term := make(chan string)

	go func(channels map[string](chan string)) {
		http.HandleFunc("/send", httpHandler(channels))
		http.ListenAndServe(httpHost, nil)
	}(channels)

	listener, _ := net.Listen("tcp", socketHost)
	go terminator(channels, term)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		id := getToken()
		channels[id] = make(chan string)
		go socketHandler(conn, channels[id], term, id)
	}
}

func terminator(channelsHub map[string](chan string), term chan string) {
	for {
		select {
		case userId := <-term:
			fmt.Println("User disconnect")
			delete(channelsHub, userId)
		}
	}
}

func getToken() string {
	guid := xid.New()
	return guid.String()
}

func getJsonToken(token string) string {
	obj := gabs.New()
	obj.Set("token", "message_type")
	obj.Set(token, "data", "token")
	return obj.String()
}

func httpHandler(channelsHub map[string](chan string)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		token := r.Form.Get("token")
		data := r.Form.Get("data")
		if token != "" {
			if channel, ok := channelsHub[token]; ok {
				channel <- data
			} else {
				w.WriteHeader(500)
				w.Write([]byte("User is not conneted"))
			}

		}

	}
}

const CLOSE_KW = "<CLOSE>"
const EOM_KW = "<EOF>"

func socketHandler(conn net.Conn, ch chan string, term chan string, id string) {
	fmt.Println("User connected")
	inputBuffer := make([]byte, 256)
	conn.Write([]byte(getJsonToken(id)))
	conn.Write([]byte(EOM_KW))
	connectionCheckTimer := time.NewTicker(time.Second * 2)
	defer conn.Close()
	for {
		select {
		case message := <-ch:
			_, err := conn.Write([]byte(message))
			if err == nil {
				conn.Write([]byte(EOM_KW))
			} else {
				term <- id
				return
			}
		case <-connectionCheckTimer.C:
			closed := false
			var err error = nil
			var n int = 0
			n, err = conn.Read(inputBuffer)
			inputStr := string(inputBuffer[:n])
			if inputStr == CLOSE_KW { //"disconnect") {
				closed = true
			}

			_, err = conn.Write([]byte{})
			if err != nil {
				closed = true
			}

			if closed == true {
				term <- id
				return
			}
			time.Sleep(time.Second)
		}
	}
}

func checkError(err error) {
	if err != nil {
		os.Exit(1)
	}
}
