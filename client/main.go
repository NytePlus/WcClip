package main

import (
	"flag"
	"log"
)

var (
	host  string
	port  string
	url   string
	token string
)

func main() {
	flag.StringVar(&host, "host", "localhost", "服务器主机")
	flag.StringVar(&port, "port", "8080", "服务器端口")
	flag.StringVar(&url, "url", "", "服务地址")
	flag.StringVar(&token, "token", "", "连接凭证")
	flag.Parse()
	serverURL := "ws://" + host + ":" + port + "/ws"
	if url != "" {
		serverURL = url
		log.Printf("配置url覆盖主机和端口 %s", url)
	}
	clientID := getClientID()

	log.Printf("启动剪贴板客户端 (ID: %s)", clientID)
	log.Printf("连接服务器: %s", serverURL)

	manager := NewClipboardManager(serverURL, clientID, token)
	manager.Start()
}
