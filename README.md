<div align="center">

# ***WcC***lip: 跨设备剪贴板同步

</div>

> [!CAUTION]
> 凌晨两点，你正在家里用Windows电脑紧急修复线上Bug，突然发现需要用到公司macOS电脑上的密钥配置片段。
>
>- 微信？—— “Windows微信已登录，Mac微信自动退出”
>- 网盘？—— 登录、上传、下载、重命名...繁琐到让人放弃
>- 邮件？—— 凌晨两点真的不想再收验证码了

> [!TIP]
> 就在准备手动输入那段超长的哈希数字时，你想起已经安装了 <span style="color: #FFD700;">Wc</span>Clip。
>
>### 3秒解决战斗
>1. 公司电脑：`Cmd+C` 复制配置文本
>2. 家中电脑：`Ctrl+V` 粘贴使用
>   
>🎉**完成！** 就像在同一台电脑上操作一样自然。

---

**WcClip** 正是为解决这种"设备孤岛"困境而生。当传统的跨设备传输方式都在关键时刻掉链子时，我们回归到最本质的需求：复制、粘贴，就这么简单。无论您是在 Windows 上进行开发，还是在 macOS 上处理设计工作，WcClip 都能实现无缝、实时的剪贴板内容共享，让工作效率跨越设备界限。

## 🚀 Quick Start

需要一个中心化的服务器汇总剪贴板信息，每个设备运行自己的客户端。客户端会实时共享剪贴板内容，并接受服务端发来的更新消息。

源码运行，你需要至少运行一个服务端，一个客户端
```bash
go run ./server
go run ./client
```

## 📦 Build

### Build Binary
首先需要获取必要的项目依赖
```bash
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
go mod tidy
```

可以编译客户端代码，在任何平台下直接运行二进制代码
```powershell
# 编译客户端（Windows）
go build -o bin/wcclip-client.exe ./client

# 编译客户端（macOS arm）
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o bin/wcclip-client-macos-amd64 ./client

# 编译客户端（macOS arm）
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o bin/wcclip-client-macos-arm64 ./client

# 编译服务端
go build -o bin/wcclip-server ./server
```

### Build Docker image
```bash
docker build -t wcclip-server:latest .
```

## 🌐 Deploy Server

### Deploy with Binary
你可以将二进制代码运行在任何服务器上，通过两个参数或环境变量指定配置
```bash
wcclip-server -port 8080 -token 1234
```

### Deploy with Docker
如果你借助在线服务托管平台[Render](https://dashboard.render.com/)部署你的服务，我非常推荐使用镜像。

我提供了服务端镜像`nyteplus/wcclip-server:v1`，只需两个环境变量完成配置，并在容器入口自动运行
- `SERVER_TOKEN` 服务端鉴权所需凭证
- `SERVER_PORT` 服务端使用的网络端口

## 🖥️️ Deploy Client

### Deploy on Win64
> windows的服务与用户交互处于不同session，vc与nssm服务无法成功获取剪贴板内容

使用定时任务部署服务，在用户登录时自动启动客户端。配置定时任务，需要以管理员身份运行CMD
```cmd
set TASK_NAME=WcClipClient
set EXE_PATH=E:\IdeaProjects\WcClip\bin\wcclip-client.exe
set ARG=-url wss://<ip>:<host>/ws -token <token>
schtasks /create /tn "%TASK_NAME%" /tr "%EXE_PATH% %ARG%" /sc onlogon /ru "%USERNAME%" /rl highest /f

schtasks /run /tn WcClipClient
```

### Deploy on MacOS Apple-Mx
> LaunchDaemon服务同样无法获取用户剪贴板，所以使用LaunchAgents

你需要先使用文件配置你的用户服务。首先编辑你的plist文件，样例文件可以参考[plist文件](doc/com.wcclip.client.plist)
```bash
sudo vim ~/Library/LaunchAgents/<your task>.plist
```

加载服务配置并启动服务
```bash
launchctl load ~/Library/LaunchAgents/<your task>.plist
launchctl start <your task>
```
