# 第一阶段：构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app/src

# 复制整个src目录（包括go.mod和go.sum）
COPY src/ .

# 设置国内Go模块代理
ENV GOPROXY=https://goproxy.cn,direct

# 执行go mod tidy来确保依赖完整并清理未使用的依赖
RUN go mod tidy

# 编译Go应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -o ../orderease main.go

# 第二阶段：运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/orderease .

# 复制配置文件
COPY src/config/ ./config/

# 创建uploads目录
RUN mkdir -p uploads/products

# 暂时移除时区配置以解决构建问题

# 暴露端口（与配置文件中的port一致）
EXPOSE 8080

# 设置入口点
ENTRYPOINT ["./orderease"]