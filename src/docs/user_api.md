# 前端用户API文档

## 概述

前端用户API提供用户注册和登录功能，专为前端应用设计。支持简单的用户名和密码认证，密码要求为6位字母或数字。

## 基础信息

- **基础URL**: `http://127.0.0.1:8080/api/order-ease/v1/`
- **认证方式**: JWT Token (登录后使用)
- **数据格式**: JSON

## 接口列表

### 1. 前端用户注册

- **路径**: `/user/register`
- **方法**: POST
- **描述**: 前端用户注册接口，密码为6位字母或数字
- **认证**: 无需认证

**请求参数**:
```json
{
  "username": "string", // 用户名，必填，不能重复
  "password": "string"  // 密码，必填，6位字母或数字
}
```

**请求示例**:
```json
{
  "username": "testuser",
  "password": "abc123"
}
```

**响应成功**:
```json
{
  "message": "注册成功",
  "user": {
    "id": "1234567890",
    "name": "testuser",
    "type": "delivery"
  }
}
```

**错误响应**:
```json
{
  "error": "用户名已存在",
  "code": 409
}
```

**错误码**:
- `400`: 请求参数错误（密码格式不正确、参数缺失等）
- `409`: 用户名已存在
- `500`: 服务器内部错误

### 2. 前端用户登录

- **路径**: `/user/login`
- **方法**: POST
- **描述**: 前端用户登录接口
- **认证**: 无需认证

**请求参数**:
```json
{
  "username": "string", // 用户名，必填
  "password": "string"  // 密码，必填
}
```

**请求示例**:
```json
{
  "username": "testuser",
  "password": "abc123"
}
```

**响应成功**:
```json
{
  "message": "登录成功",
  "user": {
    "id": "1234567890",
    "name": "testuser",
    "type": "delivery"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiredAt": 1700000000
}
```

**错误响应**:
```json
{
  "error": "用户名或密码错误",
  "code": 401
}
```

**错误码**:
- `400`: 请求参数错误
- `401`: 用户名或密码错误
- `500`: 服务器内部错误

## 数据结构

### 用户对象
```json
{
  "id": "string",           // 用户ID（雪花算法生成）
  "name": "string",         // 用户名
  "type": "string"          // 用户类型：delivery(邮寄) / pickup(自提)
}
```

### 登录响应对象
```json
{
  "message": "string",      // 操作消息
  "user": {},               // 用户信息对象
  "token": "string",        // JWT令牌
  "expiredAt": 1234567890    // 令牌过期时间戳
}
```

## 响应格式

### 成功响应
```json
{
  "message": "操作成功",
  "data": {}
}
```

### 错误响应
```json
{
  "error": "错误描述",
  "code": 400
}
```

## 使用流程

### 注册流程
1. 调用 `/user/register` 接口注册新用户
2. 检查用户名是否已存在
3. 验证密码格式（6位字母或数字）
4. 创建用户并返回用户信息

### 登录流程
1. 调用 `/user/login` 接口进行登录
2. 验证用户名和密码
3. 生成JWT令牌
4. 返回用户信息和令牌

## 注意事项

1. **密码安全**: 用户密码使用bcrypt算法加密存储，确保安全性
2. **密码格式**: 前端用户注册时密码必须为6位字母或数字
3. **用户名唯一性**: 用户名在系统中必须唯一
4. **令牌使用**: 登录成功后，需要在后续请求的Header中携带token：`Authorization: Bearer {token}`
5. **令牌过期**: JWT令牌有有效期，过期后需要重新登录

## 示例代码

### JavaScript 注册示例
```javascript
fetch('/api/order-ease/v1/user/register', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    username: 'testuser',
    password: 'abc123'
  })
})
.then(response => response.json())
.then(data => {
  if (data.message === '注册成功') {
    console.log('注册成功:', data.user);
  } else {
    console.error('注册失败:', data.error);
  }
});
```

### JavaScript 登录示例
```javascript
fetch('/api/order-ease/v1/user/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    username: 'testuser',
    password: 'abc123'
  })
})
.then(response => response.json())
.then(data => {
  if (data.message === '登录成功') {
    localStorage.setItem('token', data.token);
    localStorage.setItem('user', JSON.stringify(data.user));
    console.log('登录成功:', data.user);
  } else {
    console.error('登录失败:', data.error);
  }
});
```

## 版本信息

- **版本**: v1.0
- **更新日期**: 2024-01-01
- **维护者**: OrderEase开发团队