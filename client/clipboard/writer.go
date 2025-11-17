package clipboard

import (
	"bytes"
	"fmt"
	aclipboard "github.com/atotto/clipboard"
	gclipboard "golang.design/x/clipboard"
	"os/exec"
	"runtime"
)

func GoogleWriteText(text string) error {
	return GoogleWriteData([]byte(text))
}

func GoogleWriteData(data []byte) error {
	gclipboard.Write(gclipboard.FmtText, data)
	return nil
}

func AtWriteText(text string) error {
	return aclipboard.WriteAll(text)
}

func AtWriteFile(filename string, data []byte) error {
	return AtWriteText(string(data))
}

func NaiveWriteText(text string) error {
	switch runtime.GOOS {
	case "darwin": // macOS
		cmd := exec.Command("pbcopy")
		cmd.Stdin = bytes.NewBufferString(text)
		return cmd.Run()

	case "windows":
		//无法处理转义字符
		cmd := exec.Command("powershell", "-command", "Set-Clipboard", "-Value", text)
		return cmd.Run()

	case "linux":
		// 尝试不同的Linux剪贴板工具
		for _, cmdName := range []string{"xclip", "xsel"} {
			if path, _ := exec.LookPath(cmdName); path != "" {
				var cmd *exec.Cmd
				if cmdName == "xclip" {
					cmd = exec.Command("xclip", "-selection", "clipboard", "-in")
				} else {
					cmd = exec.Command("xsel", "--clipboard", "--input")
				}

				cmd.Stdin = bytes.NewBufferString(text)
				if err := cmd.Run(); err == nil {
					return nil
				}
			}
		}
		return fmt.Errorf("未找到可用的剪贴板工具")

	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}
