# 使用官方 Golang 镜像
FROM golang:1.22

# 设置工作目录
WORKDIR /app

# 将代码复制到容器中
COPY . .

# 编译应用程序
RUN go build -o app .

# 设置运行时环境变量（可在部署时覆盖）
ENV POLL_INTERVAL=5

# 容器启动命令
CMD ["./app"]
