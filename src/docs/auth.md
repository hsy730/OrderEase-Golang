# OrderEase 认证接口文档

## 概述

OrderEase 使用 JWT（JSON Web Token）进行身份认证，支持管理员、店主、前台用户三种角色。

### 认证流程

1. 客户端调用登录接口获取 Token
2. 在后续请求中通过 Header 携带 Token
3. Token 过期后可通过刷新接口获取新 Token

### Token 规范

- **格式**: `Bearer <token>`
- **有效期**: 2小时
- **请求头**: `Authorization: Bearer <token>`

---

## 1. 统一登录接口

### 1.1 管理员/店主登录

**接口**: POST `/api/order-ease/v1/login`

**描述**: 统一登录接口，支持管理员和店主登录

**请求参数**:

```json
{
  "username": "admin",
  "password": "Admin@123456"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名（管理员或店主）|
| password | string | 是 | 密码 |

**成功响应 - 管理员**:

```json
{
  "code": 200,
  "role": "admin",
  "user_info": {
    "id": 1,
    "username": "admin"
  },
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expiredAt": 1704067200
}
```

**成功响应 - 店主**:

```json
{
  "code": 200,
  "role": "shop",
  "user_info": {
    "id": 1234567890,
    "shop_name": "店铺A",
    "username": "shop_owner"
  },
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expiredAt": 1704067200
}
```

**错误响应**:

```json
{
  "code": 401,
  "error": "用户名或密码错误"
}
```

---

### 1.2 前台用户登录

**接口**: POST `/api/order-ease/v1/user/login`

**描述**: 前台用户登录

**请求参数**:

```json
{
  "username": "testuser",
  "password": "password123"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，2-20位 |
| password | string | 是 | 密码，6-20位，需包含字母和数字 |

**成功响应**:

```json
{
  "code": 200,
  "message": "登录成功",
  "user": {
    "id": 1234567890123456789,
    "name": "testuser",
    "type": "delivery"
  },
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expiredAt": 1704067200
}
```

**错误响应**:

```json
{
  "code": 401,
  "error": "用户名或密码错误"
}
```

---

### 1.3 前台用户注册

**接口**: POST `/api/order-ease/v1/user/register`

**描述**: 前台用户注册

**请求参数**:

```json
{
  "username": "testuser",
  "password": "password123"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，2-20位，唯一 |
| password | string | 是 | 密码，6-20位，需包含字母和数字 |

**成功响应**:

```json
{
  "code": 200,
  "message": "注册成功",
  "user": {
    "id": 1234567890123456789,
    "name": "testuser",
    "type": "pickup"
  }
}
```

**错误响应**:

```json
{
  "code": 409,
  "error": "用户名已存在"
}
```

```json
{
  "code": 400,
  "error": "密码必须为6-20位，且包含字母和数字"
}
```

---

## 2. Token 管理

### 2.1 刷新管理员 Token

**接口**: POST `/api/order-ease/v1/admin/refresh-token`

**描述**: 使用旧 Token 换取新 Token

**请求头**:

```
Authorization: Bearer <old-token>
```

**成功响应**:

```json
{
  "code": 200,
  "message": "token刷新成功",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expiredAt": 1704067200
}
```

**错误响应**:

```json
{
  "code": 401,
  "error": "无效的token"
}
```

---

### 2.2 刷新店主 Token

**接口**: POST `/api/order-ease/v1/shop/refresh-token`

**描述**: 使用旧 Token 换取新 Token，同时会检查店铺有效期

**请求头**:

```
Authorization: Bearer <old-token>
```

**成功响应**:

```json
{
  "code": 200,
  "message": "token刷新成功",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expiredAt": 1704067200
}
```

**错误响应**:

```json
{
  "code": 403,
  "error": "店铺服务已到期"
}
```

---

### 2.3 登出

**接口**: POST `/api/order-ease/v1/admin/logout` 或 `/api/order-ease/v1/shopOwner/logout`

**描述**: 将当前 Token 加入黑名单

**请求头**:

```
Authorization: Bearer <token>
```

**成功响应**:

```json
{
  "code": 200,
  "message": "登出成功"
}
```

---

## 3. 密码管理

### 3.1 修改管理员密码

**接口**: POST `/api/order-ease/v1/admin/change-password`

**描述**: 修改管理员密码

**认证**: 需要

**请求头**:

```
Authorization: Bearer <token>
```

**请求参数**:

```json
{
  "old_password": "Admin@123456",
  "new_password": "NewPass@789"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| old_password | string | 是 | 当前密码 |
| new_password | string | 是 | 新密码，需满足密码强度要求 |

**密码强度要求**:

1. 长度至少8位
2. 必须包含数字
3. 必须包含大写字母
4. 必须包含小写字母
5. 必须包含特殊字符（如：@#$%^&*等）

**成功响应**:

```json
{
  "code": 200,
  "message": "密码修改成功"
}
```

**错误响应**:

```json
{
  "code": 401,
  "error": "旧密码错误"
}
```

```json
{
  "code": 400,
  "error": "密码必须包含大小写字母、数字和特殊字符，且长度至少8位"
}
```

---

### 3.2 修改店主密码

**接口**: POST `/api/order-ease/v1/shopOwner/change-password`

**描述**: 修改店主登录密码

**认证**: 需要

**请求头**:

```
Authorization: Bearer <token>
```

**请求参数**:

```json
{
  "old_password": "oldpass123",
  "new_password": "NewPass@789"
}
```

**密码强度要求**: 与管理员密码要求相同

**成功响应**:

```json
{
  "code": 200,
  "message": "密码修改成功"
}
```

---

## 4. 临时令牌登录

### 4.1 获取临时令牌

**接口**: GET `/api/order-ease/v1/admin/shop/temp-token` 或 `/api/order-ease/v1/shopOwner/shop/temp-token`

**描述**: 获取店铺的6位数字临时令牌，用于临时登录，有效期1小时

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | string | 是 | 店铺ID |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "shop_id": 1234567890,
    "token": "123456",
    "expires_at": "2025-02-06T14:30:00Z"
  }
}
```

### 4.2 临时令牌登录

**接口**: POST `/api/order-ease/v1/shop/temp-login`

**描述**: 使用6位临时令牌登录，获取 JWT Token

**请求参数**:

```json
{
  "shop_id": 1234567890,
  "token": "123456"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 是 | 店铺ID |
| token | string | 是 | 6位数字临时令牌 |

**成功响应**:

```json
{
  "code": 200,
  "role": "user",
  "user_info": {
    "id": 987654321,
    "name": "shop_user",
    "shop_id": 1234567890,
    "shop_name": "店铺A"
  },
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expiredAt": 1704067200
}
```

**错误响应**:

```json
{
  "code": 401,
  "error": "无效的临时令牌"
}
```

---

## 5. 用户名检查

**接口**: GET `/api/order-ease/v1/user/check-username`

**描述**: 检查用户名是否已存在

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 要检查的用户名 |

**成功响应**:

```json
{
  "code": 200,
  "exists": false
}
```

---

## 6. 管理员账户说明

系统内置一个管理员账户：

- **用户名**: `admin`
- **初始密码**: `Admin@123456`

**安全提示**: 请在首次登录后立即修改密码！

---

## 错误码汇总

| 状态码 | 错误信息 | 说明 |
|--------|----------|------|
| 400 | 无效的登录数据 | 请求参数缺失或格式错误 |
| 401 | 用户名或密码错误 | 凭据验证失败 |
| 401 | 无效的token | Token 无效或已过期 |
| 403 | 店铺服务已到期 | 店主账户已过期 |
| 409 | 用户名已存在 | 注册时用户名重复 |
| 500 | 登录失败/服务器错误 | 服务器内部错误 |

---

## 接口限流

| 接口 | 限流规则 |
|------|----------|
| POST /login | 每10秒1次，最大突发3次 |
| POST /user/register | 每10秒1次，最大突发3次 |
| POST /user/login | 每10秒1次，最大突发3次 |
| POST /shop/temp-login | 每10秒1次，最大突发3次 |
| Token刷新接口 | 每秒2次，最大突发5次 |

超出限流将返回 `429 Too Many Requests`。
