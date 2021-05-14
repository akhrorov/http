package server

import (
	"bytes"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"sync"
)

const CRLF = "\r\n"

type HandleFunc func(req *Request)

type Server struct {
	addr     string
	mu       *sync.RWMutex
	handlers map[string]HandleFunc
}

func NewServer(addr string) *Server {
	return &Server{addr: addr, mu: &sync.RWMutex{}, handlers: map[string]HandleFunc{}}
}

type Request struct {
	Conn        net.Conn
	QueryParams url.Values
	PathParams  map[string]string
	Headers 	map[string]string
}

func (s *Server) Register(path string, handler HandleFunc) {
	log.Println(path)
	s.mu.Lock()
	s.handlers[path] = handler
	s.mu.Unlock()
}

func (s *Server) Start() error {
	address := s.addr
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		if cerr := listener.Close(); cerr != nil {
			if err == nil {
				err = err
				return
			}
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go s.handle(conn)
	}
	return nil
}

func (s *Server) handle(conn net.Conn) {
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err == io.EOF {
		log.Printf("%s", buf[:n])
		return
	}
	if err != nil {
		log.Print(err)
		return
	}

	data := buf[:n]
	requestLineDelim := []byte{'\r', '\n'}
	requestLineEnd := bytes.Index(data, requestLineDelim)
	headerPart := strings.Split(string(data),"\r\n")
	headers := headerPart[1:]
	result := make(map[string]string)
	for _, header := range headers {
		if header == "" {
			continue
		}
		keyIndex := strings.Index(header, ":")
		key := header[:keyIndex]
		value := header[keyIndex+2:]
		result[key] = value
	}
	if requestLineEnd == -1 {
		log.Print("not found")
		return
	}

	requestLine := string(data[:requestLineEnd])
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		log.Print("not equal")
		return
	}

	path, version := parts[1], parts[2]

	uri, err := url.ParseRequestURI(path)
	if err != nil {
		log.Print(err)
		return
	}

	if version != "HTTP/1.1" {
		log.Print("version error")
		return
	}

	query := uri.Query()

	pathParams := map[string]string{}
	handler, ok := route(path, s.mu, s.handlers, pathParams)
	if !ok {
		_, _ = conn.Write([]byte(
			"HTTP/1.1 Not Found" + CRLF + CRLF,
		))
		log.Println(path)
		log.Print("not found handler")
		return
	}

	request := &Request{Conn: conn, QueryParams: query, PathParams: pathParams, Headers: result}
	handler(request)
}

func route(path string, mutex *sync.RWMutex, handlers map[string]HandleFunc, params map[string]string) (HandleFunc, bool) {
	mutex.RLock()
	handler, ok := handlers[path]
	mutex.RUnlock()
	if ok {
		return handler, true
	}

	// /payments/123
	// /payments/{id}
	handleFunc, ok := findHandler(path, handlers, params)
	if ok {
		return handleFunc, true
	}
	return nil, false
}

func findHandler(path string, handlers map[string]HandleFunc, params map[string]string) (HandleFunc, bool) {
	for registeredPath, handleFunc := range handlers {
		ok := isHandlerSuitable(path, registeredPath, params)
		if ok {
			return handleFunc, true
		}
	}
	return nil, false
}

func isHandlerSuitable(path, registeredPath string, params map[string]string) (ok bool) {
	pathParts := strings.Split(path, "/")
	registeredParts := strings.Split(registeredPath, "/")
	if len(pathParts) != len(registeredParts) {
		return false
	}
	for i, registeredPart := range registeredParts {
		placeholder, ok := takePlaceholder(registeredPart)
		if ok {
			pathPart := pathParts[i]
			index := strings.Index(registeredPart, "{")
			value := pathPart[index:]
			params[placeholder] = value
			continue
		}

		if registeredPart != pathParts[i] {
			return false
		}
	}
	return true
}

func takePlaceholder(part string) (string, bool) {
	if len(part) == 0 {
		return "", false
	}
	index := strings.Index(part, "{")
	if index == -1 {
		return "", false
	}
	lastIndex := strings.Index(part, "}")
	if lastIndex == -1 {
		return "", false
	}

	return part[index+1 : lastIndex], true
}
