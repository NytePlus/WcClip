package main

import (
	"WcClip/protocol"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type WebsocketServer struct {
	port      string
	token     string
	clients   map[*websocket.Conn]bool
	clientsMu sync.Mutex
	upgrader  websocket.Upgrader
}

func NewWebsocketServer(port string, token string) *WebsocketServer {
	return &WebsocketServer{
		port:      port,
		token:     token,
		clients:   make(map[*websocket.Conn]bool),
		clientsMu: sync.Mutex{},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (se *WebsocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if !se.authenticate(r) {
		log.Printf("认证失败: %s", r.RemoteAddr)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := se.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	// 注册新客户端
	se.clientsMu.Lock()
	se.clients[conn] = true
	se.clientsMu.Unlock()

	log.Printf("新客户端连接，当前客户端数: %d", len(se.clients))

	// 监听客户端消息
	for {
		var msg protocol.ClipboardMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("读取消息失败: %v", err)
			break
		}

		log.Printf("收到来自客户端 %s 的消息: %s\n", msg.ClientID, msg.Type)
		if msg.Type == "text" {
			log.Printf("剪贴板文本: %s", msg.Data)
		} else {
			log.Printf("文件名: %s", msg.FileName)
		}
		se.broadcastMessage(msg, conn)
	}

	// 客户端断开连接
	se.clientsMu.Lock()
	delete(se.clients, conn)
	se.clientsMu.Unlock()
	log.Printf("客户端断开连接，剩余客户端数: %d", len(se.clients))
}

func (se *WebsocketServer) authenticate(r *http.Request) bool {
	// 如果服务器没有设置 token，允许所有连接
	if se.token == "" {
		log.Printf("警告: 服务器未设置 token，允许所有连接")
		return true
	}

	// 从 Header 中获取 Token
	clientToken := r.Header.Get("Authorization")
	if clientToken == "" {
		log.Printf("客户端未提供 Token")
		return false
	}

	// 支持多种 Token 格式
	if strings.HasPrefix(clientToken, "Bearer ") {
		clientToken = clientToken[7:] // 去掉 "Bearer " 前缀
	} else if strings.HasPrefix(clientToken, "Token ") {
		clientToken = clientToken[6:] // 去掉 "Token " 前缀
	}

	// 比较 Token
	if clientToken != se.token {
		log.Printf("Token 不匹配: 期望=%s, 实际=%s", se.token, clientToken)
		return false
	}

	log.Printf("Token 验证成功")
	return true
}

func (se *WebsocketServer) broadcastMessage(msg protocol.ClipboardMessage, sender *websocket.Conn) {
	se.clientsMu.Lock()
	defer se.clientsMu.Unlock()

	for client := range se.clients {
		if client != sender { // 不发送回原发送者
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("广播消息失败: %v", err)
				client.Close()
				delete(se.clients, client)
			}
		}
	}
}

func (se *WebsocketServer) start() {
	http.HandleFunc("/ws", se.handleWebSocket)
	port := se.port

	// 健康检查端点
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		se.clientsMu.Lock()
		count := len(se.clients)
		se.clientsMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"clients":   count,
			"timestamp": time.Now().Unix(),
		})
	})

	log.Printf("剪贴板同步服务器启动在 http://localhost:%s", port)
	log.Printf("WebSocket 端点: ws://localhost:%s/ws", port)
	log.Printf("健康检查端点: http://localhost:%s/health", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("服务器启动失败: ", err)
	}
}
