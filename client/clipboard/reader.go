package clipboard

import (
	"context"
	"fmt"
	aclipboard "github.com/atotto/clipboard"
	gclipboard "golang.design/x/clipboard"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// FakePushReadDataIter 注意，底层依然是轮询
func FakePushReadDataIter() <-chan []byte {
	ch := gclipboard.Watch(context.TODO(), gclipboard.FmtText)
	return ch
}

// PullReadTextIter 轮询检查剪贴板
func PullReadTextIter() <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		interval := 2 * time.Second
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			current, err := AtReadText()
			if err == nil {
				ch <- current
			} else {
				log.Printf("Read text error: %s", err)
			}
		}
	}()

	return ch
}

func AtReadText() (string, error) {
	return aclipboard.ReadAll()
}

func GoogleReadText() (string, error) {
	data := gclipboard.Read(gclipboard.FmtText)
	return string(data), nil
}

func GoogleReadData() ([]byte, error) {
	data := gclipboard.Read(gclipboard.FmtText)
	return data, nil
}

func NaiveReadText() (string, error) {
	switch runtime.GOOS {
	case "darwin": // macOS
		cmd := exec.Command("pbpaste")
		output, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return string(output), nil

	case "windows":
		cmd := exec.Command("powershell", "-command", "[Console]::OutputEncoding = [Text.Encoding]::UTF8; Get-Clipboard")
		output, err := cmd.Output()
		if err != nil {
			return "", err
		}
		text := (strings.TrimRight(string(output), "\r\n"))

		return text, nil

	case "linux":
		// 尝试不同的Linux剪贴板工具
		for _, cmdName := range []string{"xclip", "xsel"} {
			if path, _ := exec.LookPath(cmdName); path != "" {
				var cmd *exec.Cmd
				if cmdName == "xclip" {
					cmd = exec.Command("xclip", "-selection", "clipboard", "-out")
				} else {
					cmd = exec.Command("xsel", "--clipboard", "--output")
				}

				output, err := cmd.Output()
				if err == nil {
					return string(output), nil
				}
			}
		}
		return "", fmt.Errorf("未找到可用的剪贴板工具")

	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}
