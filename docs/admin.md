# 管理员接口文档

## 管理员账户说明
系统只有一个管理员账户，初始账户信息：
- 用户名：admin
- 初始密码：Admin@123456

**请在首次登录后立即修改密码！**

## 密码要求
管理员密码必须满足以下所有条件：
1. 长度至少12位
2. 必须包含数字
3. 必须包含大写字母
4. 必须包含小写字母
5. 必须包含特殊字符（如：@#$%^&*等）

## 接口说明

### 1. 管理员登录
- **接口**: POST `/api/v1/admin/login`
- **描述**: 管理员账户登录
- **请求参数 (Body)**:
```json
{
    "username": "admin",
    "password": "your_password"
}
```
- **响应**:
```json
{
    "message": "登录成功",
    "admin": {
        "id": 1,
        "username": "admin"
    }
}
```

### 2. 修改密码
- **接口**: PUT `/api/v1/admin/change-password`
- **描述**: 修改管理员密码
- **请求参数 (Body)**:
```json
{
    "old_password": "当前密码",
    "new_password": "新密码"
}
```
- **响应**:
```json
{
    "message": "密码修改成功"
}
```

### 使用示例

1. 登录管理员账户：
```bash
curl -X POST "http://your-domain/api/v1/admin/login" \
-H "Content-Type: application/json" \
-d '{
    "username": "admin",
    "password": "Admin@123456"
}'
```

2. 修改管理员密码：
```bash
curl -X PUT "http://your-domain/api/v1/admin/change-password" \
-H "Content-Type: application/json" \
-d '{
    "old_password": "Admin@123456",
    "new_password": "NewPassword@2024"
}'
```

### 错误响应
- 400 Bad Request: 请求参数错误或密码不符合要求
- 401 Unauthorized: 用户名或密码错误
- 404 Not Found: 管理员账户不存在
- 500 Internal Server Error: 服务器内部错误

错误响应格式：
```json
{
    "error": "错误信息描述"
}
```

### 注意事项
1. 首次使用系统时请立即修改默认密码
2. 密码不要使用常见组合（如：123456）
3. 不要在不安全的环境下保存密码
4. 定期更换密码以提高安全性
5. 密码修改后需要重新登录
6. 如果忘记密码，需要联系系统管理员重置
