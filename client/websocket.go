package main

import (
	"WcClip/protocol"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

type ClipboardManager struct {
	conn         *websocket.Conn
	serverURL    string
	clientID     string
	token        string
	lastContent  string
	lastFileHash string
	isConnected  bool
	clipboardMu  sync.Mutex
}

func NewClipboardManager(serverURL, clientID string, token string) *ClipboardManager {
	return &ClipboardManager{
		serverURL:   serverURL,
		clientID:    clientID,
		token:       token,
		clipboardMu: sync.Mutex{},
	}
}

func (cm *ClipboardManager) connect() error {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+cm.token)
	headers.Set("X-Client-ID", cm.clientID)
	headers.Set("X-Client-OS", runtime.GOOS)

	conn, _, err := websocket.DefaultDialer.Dial(cm.serverURL, headers)
	if err != nil {
		return fmt.Errorf("连接服务器失败: %v", err)
	}

	cm.conn = conn
	cm.isConnected = true
	log.Printf("成功连接到服务器: %s", cm.serverURL)

	go cm.listenServer()
	return nil
}

func (cm *ClipboardManager) listenServer() {
	for cm.isConnected {
		var msg protocol.ClipboardMessage
		err := cm.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("从服务器读取消息失败: %v", err)
			cm.isConnected = false
			return
		}

		// 处理接收到的剪贴板更新
		cm.handleReceivedMessage(msg)
	}
}

func (cm *ClipboardManager) handleReceivedMessage(msg protocol.ClipboardMessage) {
	cm.clipboardMu.Lock()
	switch msg.Type {
	case "text":
		currentText, err := ReadText()
		if err == nil && currentText == msg.Data {
			cm.clipboardMu.Unlock()
			return
		}

		log.Printf("接收 %d 字符", len(msg.Data))
		if err := WriteText(msg.Data); err != nil {
			log.Printf("写入剪贴板失败: %v", err)
		} else {
			cm.lastContent = msg.Data
		}

	case "file":
		log.Printf("接收到文件: %s, 大小: %d bytes", msg.FileName, len(msg.Data))

		// 解码base64文件数据
		fileData, err := base64.StdEncoding.DecodeString(msg.Data)
		if err != nil {
			log.Printf("文件数据解码失败: %v", err)
			cm.clipboardMu.Unlock()
			return
		}

		// 将文件写入剪贴板
		if err := WriteFile(msg.FileName, fileData); err != nil {
			log.Printf("写入文件到剪贴板失败: %v", err)
		}
	}
	cm.clipboardMu.Unlock()
}

func (cm *ClipboardManager) sendClipboardUpdate() {
	cm.clipboardMu.Lock()
	text, err := ReadText()
	//log.Printf("|%s|%s|%b", text, cm.lastContent, text == cm.lastContent)
	if err == nil && text != "" && text != cm.lastContent {
		msg := protocol.ClipboardMessage{
			Type:      "text",
			Data:      text,
			Timestamp: time.Now().Unix(),
			ClientID:  cm.clientID,
		}

		if err := cm.conn.WriteJSON(msg); err != nil {
			log.Printf("发送文本更新失败: %v", err)
			cm.isConnected = false
			cm.clipboardMu.Unlock()
			return
		}

		cm.lastContent = text
		log.Printf("发送 %d 字符", len(text))
	}

	// 检查文件内容（这里简化处理，实际可能需要更复杂的文件检测逻辑）
	// 注意：文件剪贴板操作比较复杂，这里主要展示文本同步
	cm.clipboardMu.Unlock()
}

func (cm *ClipboardManager) Start() {
	// 连接重试逻辑
	for {
		if err := cm.connect(); err != nil {
			log.Printf("连接失败，30秒后重试: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}

	// 主循环：监控本地剪贴板变化
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()
	cm.lastContent, _ = ReadText()

	for {
		select {
		case <-ticker.C:
			if !cm.isConnected {
				// 尝试重连
				if err := cm.connect(); err != nil {
					log.Printf("重连失败，30秒后重试: %v", err)
					time.Sleep(30 * time.Second)
					continue
				}
			}

			cm.sendClipboardUpdate()
		}
	}
}

func getClientID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("%s-%s", hostname, runtime.GOOS)
}
