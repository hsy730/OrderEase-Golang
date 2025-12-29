# OrderEase 合一版容器部署文档

## 概述

OrderEase 合一版容器将三个服务整合到单个 Docker 容器中运行：
- **Go 后端 API** (端口 8080)
- **前台用户 UI** (通过 Go 静态文件服务访问 `/order-ease-iui/`)
- **后台管理 UI** (通过 Go 静态文件服务访问 `/order-ease-adminiui/`)

所有服务由 Go 后端统一提供，无需 Nginx 代理。

## 架构说明

### 容器内部架构
```
┌─────────────────────────────────────────┐
│  OrderEase All-in-One Container         │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │  Go 后端服务 (8080端口)           │ │
│  │  ├─ API 接口: /api/*              │ │
│  │  ├─ 前台UI: /order-ease-iui/*      │ │
│  │  ├─ 后台UI: /order-ease-adminiui/* │ │
│  │  └─ 文件上传: /uploads/*            │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

### 端口映射
- **8080**: Go 后端服务（提供 API + 前后UI + 文件访问）

### 访问地址
- 前台用户界面: `http://localhost:8080/order-ease-iui/`
- 后台管理界面: `http://localhost:8080/order-ease-adminiui/`
- API 接口: `http://localhost:8080/api/`

## 文件说明

### 核心文件
1. **Dockerfile.all-in-one** - 多阶段构建配置
   - 阶段1: 构建后台管理 UI (Node 16)
   - 阶段2: 构建前台 UI (Node 18)
   - 阶段3: 构建 Go 后端
   - 阶段4: 整合运行 (Alpine + Go 服务)

2. **docker-compose.all-in-one.yml** - 容器编排配置
   - 定义应用容器和数据库容器
   - 配置环境变量
   - 数据持久化

3. **main.go** - Go 后端主程序（已添加 UI 静态文件服务）
   - `/order-ease-iui/` - 前台 UI 静态文件
   - `/order-ease-adminiui/` - 后台 UI 静态文件
   - `/uploads/` - 上传文件访问

## 部署步骤

### 前置要求
- Docker 20.10+
- Docker Compose 1.29+
- 至少 2GB 可用内存
- 至少 5GB 可用磁盘空间

### 1. 进入项目目录
```bash
cd d:\local_code_repo\OrderEase-Golang
```

### 2. 构建并启动容器
```bash
# 使用合一版配置启动
docker-compose -f docker-compose.all-in-one.yml up -d

# 查看构建日志
docker-compose -f docker-compose.all-in-one.yml logs -f
```

### 4. 验证服务状态
```bash
# 查看容器状态
docker-compose -f docker-compose.all-in-one.yml ps

# 查看应用日志
docker-compose -f docker-compose.all-in-one.yml logs orderease-all-in-one

# 进入容器检查
docker exec -it orderease-all-in-one sh
```

### 5. 健康检查
```bash
# 检查前台 UI
curl http://localhost:8080/order-ease-iui/

# 检查后台 UI
curl http://localhost:8080/order-ease-adminiui/

# 检查 API
curl http://localhost:8080/api/health
```

## 环境变量配置

可在 `docker-compose.all-in-one.yml` 中修改以下环境变量：

### 数据库配置
- `DB_HOST`: 数据库主机 (默认: db)
- `DB_PORT`: 数据库端口 (默认: 3306)
- `DB_USERNAME`: 数据库用户名 (默认: root)
- `DB_PASSWORD`: 数据库密码 (默认: 123456)
- `DB_NAME`: 数据库名称 (默认: orderease)

### JWT 配置
- `JWT_SECRET`: JWT 签名密钥
- `JWT_EXPIRATION`: Token 过期时间(秒)

### 服务器配置
- `SERVER_HOST`: 服务监听地址 (默认: 0.0.0.0)
- `SERVER_PORT`: 服务监听端口 (默认: 8080)

## 数据持久化

容器使用以下 Volume 持久化数据：

```yaml
volumes:
  - ./uploads:/app/uploads          # 上传文件
  - ./logs:/app/logs                # 应用日志
  - ./src/config/config.yaml:/app/config/config.yaml  # 配置文件
  - mysql-data:/var/lib/mysql       # 数据库数据
```

## 运维管理

### 启动/停止服务
```bash
# 启动
docker-compose -f docker-compose.all-in-one.yml up -d

# 停止
docker-compose -f docker-compose.all-in-one.yml down

# 重启
docker-compose -f docker-compose.all-in-one.yml restart
```

### 查看日志
```bash
# 所有服务日志
docker-compose -f docker-compose.all-in-one.yml logs -f

# 仅查看应用日志
docker-compose -f docker-compose.all-in-one.yml logs -f orderease-all-in-one

# 仅查看数据库日志
docker-compose -f docker-compose.all-in-one.yml logs -f db

# 查看 Go 应用日志（容器内）
docker exec orderease-all-in-one cat /app/logs/app.log
```

### 更新部署
```bash
# 重新构建镜像
docker-compose -f docker-compose.all-in-one.yml build --no-cache

# 重新启动服务
docker-compose -f docker-compose.all-in-one.yml up -d
```

### 备份数据
```bash
# 备份数据库
docker exec orderease-mysql mysqldump -u root -p123456 orderease > backup_$(date +%Y%m%d).sql

# 备份上传文件
tar -czf uploads_backup_$(date +%Y%m%d).tar.gz ./uploads
```

### 恢复数据
```bash
# 恢复数据库
docker exec -i orderease-mysql mysql -u root -p123456 orderease < backup_20231229.sql

# 恢复上传文件
tar -xzf uploads_backup_20231229.tar.gz
```

## 故障排查

### 容器无法启动
1. 检查端口是否被占用
   ```bash
   netstat -ano | findstr ":8080"
   ```

2. 查看容器日志
   ```bash
   docker-compose -f docker-compose.all-in-one.yml logs
   ```

### 前端页面无法访问
1. 检查 Go 后端是否运行
   ```bash
   docker exec orderease-all-in-one ps aux | grep orderease
   ```

2. 检查静态文件目录是否存在
   ```bash
   docker exec orderease-all-in-one ls -la /app/static/
   ```

### API 无法访问
1. 检查 Go 后端是否运行
   ```bash
   docker exec orderease-all-in-one ps aux | grep orderease
   ```

2. 检查后端日志
   ```bash
   docker exec orderease-all-in-one cat /app/logs/app.log
   ```

### 数据库连接失败
1. 检查数据库容器状态
   ```bash
   docker-compose -f docker-compose.all-in-one.yml ps db
   ```

2. 测试数据库连接
   ```bash
   docker exec orderease-mysql mysql -u root -p123456 -e "SELECT 1"
   ```

## 性能优化建议

1. **生产环境配置**
   - 调整 Go 程序 GOMAXPROCS
   - 调整数据库连接池大小
   - 启用 Redis 缓存（需额外配置）

2. **资源限制**
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '2'
         memory: 2G
       reservations:
         cpus: '1'
         memory: 1G
   ```

3. **日志轮转**
   配置 logrotate 避免日志文件过大

## 安全建议

1. 修改默认密码
   - 数据库 root 密码
   - JWT Secret 密钥

2. 使用 HTTPS
   - 配置 SSL 证书
   - 启用 HTTPS 重定向

3. 网络隔离
   - 数据库端口不对外暴露
   - 使用防火墙限制访问

## 与独立部署的对比

### 合一版优势
- ✅ 部署简单，一个容器包含所有服务
- ✅ 资源占用更少（无需 Nginx 进程）
- ✅ 架构简单，由 Go 统一处理所有请求
- ✅ 适合小规模部署和开发环境

### 独立版优势
- ✅ 服务隔离更好
- ✅ 可独立扩展各个服务
- ✅ 故障隔离更彻底
- ✅ 适合生产环境和大规模部署

## 技术支持

如遇问题，请检查：
1. Docker 和 Docker Compose 版本
2. 系统资源是否充足
3. 网络端口是否可用
4. 查看详细日志定位问题

---

**最后更新**: 2025-12-29
**版本**: 1.0.0
