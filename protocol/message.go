package protocol

type ClipboardMessage struct {
	Type      string `json:"type"`                // "text" or "file"
	Data      string `json:"data"`                // 文本内容或文件的base64编码
	FileName  string `json:"file_name,omitempty"` // 文件名（仅文件类型）
	Timestamp int64  `json:"timestamp"`           // 时间戳，用于去重
	ClientID  string `json:"client_id,omitempty"` // 客户端标识
}
