package connection

import (
	"WcClip/client/clipboard"
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

type Connection struct {
	conn         *websocket.Conn
	serverURL    string
	clientID     string
	token        string
	lastContent  string
	lastFileHash string
	isConnected  bool
	clipboardMu  sync.Mutex
}

func NewConnection(serverURL, clientID string, token string) *Connection {
	return &Connection{
		serverURL:   serverURL,
		clientID:    clientID,
		token:       token,
		clipboardMu: sync.Mutex{},
	}
}

func (cm *Connection) connect() error {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+cm.token)
	headers.Set("X-Connection-ID", cm.clientID)
	headers.Set("X-Connection-OS", runtime.GOOS)

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

func (cm *Connection) listenServer() {
	for cm.isConnected {
		var msg protocol.ClipboardMessage
		err := cm.conn.ReadJSON(&msg) // 会将[]byte统一认为是base64
		if err != nil {
			log.Printf("从服务器读取消息失败: %v", err)
			cm.isConnected = false
			return
		}

		// 处理接收到的剪贴板更新
		cm.handleReceivedMessage(msg)
	}
}

func (cm *Connection) handleReceivedMessage(msg protocol.ClipboardMessage) {
	cm.clipboardMu.Lock()
	log.Print(msg.DataString)
	switch msg.Type {
	case "text":
		log.Printf("接收 %d 字符", len(msg.DataString))
		if err := clipboard.AtWriteText(msg.DataString); err != nil {
			log.Printf("写入剪贴板失败: %v", err)
		} else {
			cm.lastContent = msg.DataString
		}

	case "file":
		log.Printf("接收到文件: %s, 大小: %d bytes", msg.FileName, len(msg.DataString))

		fileData, err := base64.StdEncoding.DecodeString(msg.DataString)
		if err != nil {
			log.Printf("文件数据解码失败: %v", err)
			cm.clipboardMu.Unlock()
			return
		}

		if err := clipboard.AtWriteFile(msg.FileName, fileData); err != nil {
			log.Printf("写入文件到剪贴板失败: %v", err)
		}
	}
	cm.clipboardMu.Unlock()
}

func (cm *Connection) sendClipboardUpdate(dataString string) {
	cm.clipboardMu.Lock()
	if dataString != "" && dataString != cm.lastContent {
		msg := protocol.ClipboardMessage{
			Type:       "text",
			DataString: dataString,
			Timestamp:  time.Now().Unix(),
			ClientID:   cm.clientID,
		}

		if err := cm.conn.WriteJSON(msg); err != nil {
			log.Printf("发送文本更新失败: %v", err)
			cm.isConnected = false
			cm.clipboardMu.Unlock()
			return
		}

		cm.lastContent = dataString
		log.Printf("发送 %d 字符", len(dataString))
	}
	cm.clipboardMu.Unlock()
}

func (cm *Connection) Start() {
	// 连接重试逻辑
	for {
		if err := cm.connect(); err != nil {
			log.Printf("连接失败，30秒后重试: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}

	iter := clipboard.PullReadTextIter()
	for text := range iter {
		if !cm.isConnected {
			// 尝试重连
			if err := cm.connect(); err != nil {
				log.Printf("重连失败，30秒后重试: %v", err)
				time.Sleep(30 * time.Second)
				continue
			}
		}

		cm.sendClipboardUpdate(text)
	}
}

func GetClientID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("%s-%s", hostname, runtime.GOOS)
}
