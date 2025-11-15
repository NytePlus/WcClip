# 构建阶段
FROM golang:1.25.4-alpine3.22 AS builder

RUN apk add --no-cache git

WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /app/wcclip-server ./server

# 运行阶段
FROM scratch

# RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/wcclip-server ./

CMD ["./wcclip-server"]