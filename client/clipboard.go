package main

import (
	"bytes"
	"fmt"
	"github.com/atotto/clipboard"
	"os/exec"
	"runtime"
	"strings"
)

// ReadText 读取剪贴板文本内容
func ReadText() (string, error) {
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

// WriteText 写入文本到剪贴板
func WriteText(text string) error {
	switch runtime.GOOS {
	case "darwin": // macOS
		cmd := exec.Command("pbcopy")
		cmd.Stdin = bytes.NewBufferString(text)
		return cmd.Run()

	case "windows":
		//无法处理转义字符
		//cmd := exec.Command("powershell", "-command", "Set-Clipboard", "-Value", text)
		//return cmd.Run()

		return clipboard.WriteAll(text)

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

// WriteFile 将文件写入剪贴板（简化实现，主要处理文本）
func WriteFile(filename string, data []byte) error {
	// 这里简化处理，实际文件剪贴板操作比较复杂
	// 对于小文件，可以尝试作为文本处理，或者使用系统特定的方法
	return WriteText(string(data))
}
