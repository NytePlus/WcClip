package main

import (
	"flag"
	"log"
	"os"
)

var (
	port  string
	token string
)

func main() {
	flag.StringVar(&port, "port", "8080", "服务器端口")
	flag.StringVar(&token, "token", "", "连接凭证")
	flag.Parse()

	if envPort := os.Getenv("SERVER_PORT"); envPort != "" {
		port = envPort
		log.Printf("从环境变量读取并覆盖端口: %s", port)
	}

	if envToken := os.Getenv("SERVER_TOKEN"); envToken != "" {
		token = envToken
		log.Printf("从环境变量读取并覆盖 Token")
	}

	server := NewWebsocketServer(port, token)
	server.start()
}
