# OrderEase-Golang Docker部署指南

本指南将帮助您使用Docker容器化部署OrderEase-Golang应用程序。

## 前置条件

- 已安装[Docker](https://docs.docker.com/get-docker/)
- 已安装[Docker Compose](https://docs.docker.com/compose/install/)（可选，用于多容器部署）

## 构建Docker镜像

在项目根目录（包含Dockerfile的目录）执行以下命令构建Docker镜像：

```bash
docker build -t orderease:latest .
```

## 运行Docker容器

### 基本运行

```bash
docker run -d -p 8080:8080 --name orderease orderease:latest
```

### 带环境变量运行

```bash
docker run -d \
  -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e DB_PORT=3306 \
  -e DB_USERNAME=your-db-username \
  -e DB_PASSWORD=your-db-password \
  -e DB_NAME=your-db-name \
  --name orderease \
  orderease:latest
```

### Windows开发机特别说明

在Windows开发机上运行时，由于Docker容器网络模式的限制，您需要指定数据库主机IP为本机的实际IP地址（而不是使用127.0.0.1或localhost），以便容器能够正确连接到主机上运行的数据库。

例如，如果您的Windows开发机IP地址是192.168.1.3，使用以下命令运行Docker容器：

```bash
docker run -p 8080:8080 -e DB_HOST=192.168.1.3 --name orderease orderease:latest
```

### 持久化数据

如果您希望持久化存储上传的文件，可以映射uploads目录：

```bash
docker run -d \
  -p 8080:8080 \
  -v /path/to/your/uploads:/app/uploads \
  --name orderease \
  orderease:latest
```

## Docker Compose部署

创建一个`docker-compose.yml`文件在项目根目录：

```yaml
version: '3'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - db
    environment:
      - DB_HOST=db
      - DB_PORT=3306
      - DB_USERNAME=root
      - DB_PASSWORD=123456
      - DB_NAME=orderease
    volumes:
      - ./uploads:/app/uploads

  db:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=123456
      - MYSQL_DATABASE=orderease
    volumes:
      - mysql-data:/var/lib/mysql
    ports:
      - "3306:3306"

volumes:
  mysql-data:
```

然后使用以下命令启动所有服务：

```bash
docker-compose up -d
```

## 环境变量配置

以下是可配置的环境变量：

| 环境变量 | 描述 | 默认值 |
|---------|------|--------|
| DB_HOST | 数据库主机地址 | 127.0.0.1 |
| DB_PORT | 数据库端口 | 3306 |
| DB_USERNAME | 数据库用户名 | root |
| DB_PASSWORD | 数据库密码 | 123456 |
| DB_NAME | 数据库名称 | mysql |
| JWT_SECRET | JWT密钥 | e6jf493kdhbms9ew6mv2v1a4dx2 |
| JWT_EXPIRATION | JWT过期时间（秒） | 7200 |
| SERVER_PORT | 服务器端口 | 8080 |
| SERVER_HOST | 服务器主机地址 | 0.0.0.0 |

## 访问应用程序

容器启动后，您可以通过以下URL访问应用程序：

```
http://localhost:8080
```

## 查看容器日志

```bash
docker logs orderease
```

## 进入容器内部

```bash
docker exec -it orderease /bin/sh
```

## 停止和删除容器

```bash
docker stop orderease
docker rm orderease
```

## 常见问题排查

1. **数据库连接问题**：确保数据库主机、端口、用户名和密码正确，并确保数据库服务正在运行且允许远程连接。

2. **端口冲突**：如果端口8080已被占用，可以使用`-p 8081:8080`（将容器的8080端口映射到主机的8081端口）来更改映射。

3. **权限问题**：确保挂载的卷具有正确的读写权限。

## 注意事项

- 生产环境中，请使用强密码并妥善保管配置文件。
- 建议在生产环境中使用HTTPS。
- 定期更新Docker镜像以获取安全更新。