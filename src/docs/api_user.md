# 用户相关 API 文档

## 删除用户
`DELETE /user/delete`

**请求参数:**
```json
{
  "user_id": "雪flake ID"
}
```

**权限要求:**
- 需要店铺管理员权限

**成功响应:**
```json
{
  "code": 200,
  "msg": "用户已删除",
  "data": {
    "deleted_at": "2024-03-20T15:04:05Z"
  }
}
```

### 创建用户
- **方法**: POST
- **路径**: /user/create
- **请求参数**:
  - username (string): 用户名
  - password (string): 密码
  - role (string): 用户角色
- **响应**:
  ```json
  {
    "code": 200,
    "data": {
      "id": "USER123",
      "username": "testuser",
      "role": "admin",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### 获取用户列表
- **方法**: GET
- **路径**: /user/list
- **请求参数**:
  - page (int): 页码
  - page_size (int): 每页数量
- **响应**:
  ```json
  {
    "code": 200,
    "data": {
      "users": [
        {
          "id": "USER123",
          "username": "testuser",
          "role": "admin"
        }
      ],
      "total": 1
    }
  }
  ```

### 其他接口...
