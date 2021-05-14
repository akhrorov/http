package main

import (
	"github.com/akhrorov/http/pkg/server"
	"log"
	"net"
	"os"
	"strconv"
)

func main() {
	host := "0.0.0.0"
	port := "9999"

	if err := execute(host, port); err != nil {
		os.Exit(1)
	}
}

func execute(host string, port string) (err error) {
	crlf := "\r\n"
	srv := server.NewServer(net.JoinHostPort(host, port))
	srv.Register("/api/categkjhgory{category}fg/{id}", func(req *server.Request) {

		id := req.PathParams["id"]
		category := req.PathParams["category"]
		body := id + " " + category
			_, err := req.Conn.Write([]byte(
				"HTTP/1.1 200 OK" + crlf +
					"Content-Length: " + strconv.Itoa(len(body)) + crlf +
					"Content-Type: text/html" + crlf +
					"Connection: close" + crlf +
					crlf +
					body,
			))
			if err != nil {
				log.Print(err)
			}
	})
	return srv.Start()
}