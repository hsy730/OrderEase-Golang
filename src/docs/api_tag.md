# 标签管理接口文档

## 接口列表

| 接口 | 方法 | 路径 | 权限 | 描述 |
|------|------|------|------|------|
| 创建标签 | POST | `/tag/create` | 管理员/店主 | 创建新标签 |
| 更新标签 | PUT | `/tag/update` | 管理员/店主 | 更新标签信息 |
| 获取标签列表 | GET | `/tag/list` | 管理员/店主/前台 | 获取标签列表 |
| 获取标签详情 | GET | `/tag/detail` | 管理员/店主/前台 | 获取单个标签详情 |
| 删除标签 | DELETE | `/tag/delete` | 管理员/店主 | 删除标签 |
| 批量打标签 | POST | `/tag/batch-tag` | 管理员/店主 | 为多个商品打标签 |
| 批量解绑 | DELETE | `/tag/batch-untag` | 管理员/店主 | 批量解绑商品标签 |
| 批量设置标签 | POST | `/tag/batch-tag-product` | 管理员/店主 | 为单个商品设置多个标签 |
| 获取已绑定标签 | GET | `/tag/bound-tags` | 管理员/店主 | 获取商品已绑定的标签 |
| 获取未绑定标签 | GET | `/tag/unbound-tags` | 管理员/店主 | 获取商品未绑定的标签 |
| 获取标签商品 | GET | `/tag/bound-products` | 管理员/店主/前台 | 获取标签绑定的商品列表 |
| 获取未绑定商品 | GET | `/tag/unbound-products` | 管理员/店主 | 获取标签未绑定的商品列表 |
| 获取未绑定标签列表 | GET | `/tag/unbound-list` | 管理员/店主 | 获取没有绑定商品的标签 |
| 获取上架商品 | GET | `/tag/online-products` | 管理员/店主 | 获取标签关联的已上架商品 |

**路径前缀**:
- 管理员: `/api/order-ease/v1/admin`
- 店主: `/api/order-ease/v1/shopOwner`
- 前台: `/api/order-ease/v1`

---

## 1. 创建标签

**接口**: POST `/tag/create`

**描述**: 创建新标签

**认证**: 需要

**请求参数**:

```json
{
  "shop_id": 1234567890,
  "name": "新品",
  "description": "新上架商品标签"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 是 | 店铺ID |
| name | string | 是 | 标签名称 |
| description | string | 否 | 标签描述 |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "shop_id": 1234567890,
    "name": "新品",
    "description": "新上架商品标签",
    "created_at": "2025-02-06T10:00:00Z",
    "updated_at": "2025-02-06T10:00:00Z"
  }
}
```

---

## 2. 更新标签

**接口**: PUT `/tag/update`

**描述**: 更新标签信息

**认证**: 需要

**请求参数**:

```json
{
  "id": 1,
  "shop_id": 1234567890,
  "name": "新品推荐",
  "description": "更新后的描述"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 标签ID |
| shop_id | uint64 | 是 | 店铺ID |
| name | string | 是 | 标签名称 |
| description | string | 否 | 标签描述 |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "shop_id": 1234567890,
    "name": "新品推荐",
    "description": "更新后的描述",
    "created_at": "2025-02-06T10:00:00Z",
    "updated_at": "2025-02-06T12:00:00Z"
  }
}
```

---

## 3. 获取标签列表

**接口**: GET `/tag/list`

**描述**: 获取标签列表

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| shop_id | string | 是 | - | 店铺ID |
| page | int | 否 | 1 | 页码 |
| pageSize | int | 否 | 10 | 每页数量 |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "total": 10,
    "page": 1,
    "pageSize": 10,
    "tags": [
      {
        "id": 1,
        "shop_id": 1234567890,
        "name": "新品",
        "description": "新上架商品标签",
        "created_at": "2025-02-06T10:00:00Z",
        "updated_at": "2025-02-06T10:00:00Z"
      },
      {
        "id": 2,
        "shop_id": 1234567890,
        "name": "热销",
        "description": "热销商品标签",
        "created_at": "2025-02-06T10:00:00Z",
        "updated_at": "2025-02-06T10:00:00Z"
      }
    ]
  }
}
```

**前台接口特殊说明**:

前台接口（`/tag/list`）如果存在未绑定任何标签的商品，会自动添加一个虚拟标签：

```json
{
  "id": -1,
  "name": "其他"
}
```

---

## 4. 获取标签详情

**接口**: GET `/tag/detail`

**描述**: 获取单个标签详情

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 标签ID |
| shop_id | string | 是 | 店铺ID |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "shop_id": 1234567890,
    "name": "新品",
    "description": "新上架商品标签",
    "created_at": "2025-02-06T10:00:00Z",
    "updated_at": "2025-02-06T10:00:00Z"
  }
}
```

**错误响应**:

```json
{
  "code": 404,
  "error": "标签不存在"
}
```

---

## 5. 删除标签

**接口**: DELETE `/tag/delete`

**描述**: 删除标签，有关联商品的标签不能删除

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 标签ID |
| shop_id | string | 是 | 店铺ID |

**成功响应**:

```json
{
  "code": 200,
  "message": "标签删除成功"
}
```

**错误响应**:

```json
{
  "code": 404,
  "error": "标签不存在"
}
```

```json
{
  "code": 400,
  "error": "该标签已关联 5 个商品，请先解除关联后再删除"
}
```

---

## 6. 批量打标签

**接口**: POST `/tag/batch-tag`

**描述**: 为多个商品批量打标签

**认证**: 需要

**请求参数**:

```json
{
  "shop_id": 1234567890,
  "product_ids": [1234567890123456789, 1234567890123456790, 1234567890123456791],
  "tag_id": 1
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 是 | 店铺ID |
| product_ids | array | 是 | 商品ID列表 |
| tag_id | int | 是 | 标签ID |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "message": "批量打标签成功",
    "total": 3,
    "successful": 3
  }
}
```

**错误响应**:

```json
{
  "code": 404,
  "error": "标签不存在"
}
```

---

## 7. 批量解绑标签

**接口**: DELETE `/tag/batch-untag`

**描述**: 批量解绑商品的标签

**认证**: 需要

**请求参数**:

```json
{
  "shop_id": 1234567890,
  "product_ids": [1234567890123456789, 1234567890123456790],
  "tag_id": 1
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 是 | 店铺ID |
| product_ids | array | 是 | 商品ID列表 |
| tag_id | uint | 是 | 标签ID |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "message": "批量解绑标签成功",
    "total": 2,
    "successful": 2
  }
}
```

---

## 8. 批量设置商品标签

**接口**: POST `/tag/batch-tag-product`

**描述**: 为单个商品批量设置标签（会替换该商品的所有标签）

**认证**: 需要

**请求参数**:

```json
{
  "shop_id": 1234567890,
  "product_id": 1234567890123456789,
  "tag_ids": [1, 2, 3]
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 是 | 店铺ID |
| product_id | uint64 | 是 | 商品ID |
| tag_ids | array | 是 | 标签ID列表 |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "message": "批量更新标签成功",
    "added_count": 2,
    "deleted_count": 1
  }
}
```

---

## 9. 获取商品已绑定的标签

**接口**: GET `/tag/bound-tags`

**描述**: 获取指定商品已绑定的标签列表

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| product_id | string | 是 | 商品ID |
| shop_id | string | 是 | 店铺ID |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "product_id": "1234567890123456789",
    "tags": [
      {
        "id": 1,
        "shop_id": 1234567890,
        "name": "新品",
        "description": "新上架商品标签",
        "created_at": "2025-02-06T10:00:00Z",
        "updated_at": "2025-02-06T10:00:00Z"
      }
    ]
  }
}
```

---

## 10. 获取商品未绑定的标签

**接口**: GET `/tag/unbound-tags`

**描述**: 获取指定商品未绑定的标签列表

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| product_id | string | 是 | 商品ID |
| shop_id | string | 是 | 店铺ID |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "product_id": "1234567890123456789",
    "tags": [
      {
        "id": 2,
        "shop_id": 1234567890,
        "name": "热销",
        "description": "热销商品标签",
        "created_at": "2025-02-06T10:00:00Z",
        "updated_at": "2025-02-06T10:00:00Z"
      }
    ]
  }
}
```

---

## 11. 获取标签绑定的商品列表

**接口**: GET `/tag/bound-products`

**描述**: 获取指定标签绑定的商品列表（分页）

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| tag_id | string | 是 | - | 标签ID，传-1表示获取未绑定任何标签的商品 |
| shop_id | string | 是 | - | 店铺ID |
| page | int | 否 | 1 | 页码 |
| pageSize | int | 否 | 10 | 每页数量 |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "total": 20,
    "page": 1,
    "pageSize": 10,
    "data": [
      {
        "id": 1234567890123456789,
        "shop_id": 1234567890,
        "name": "拿铁咖啡",
        "description": "香浓拿铁",
        "price": 28.00,
        "stock": 100,
        "status": "online",
        "image_url": "product_xxx.jpg"
      }
    ]
  }
}
```

**说明**:
- 管理员/店主接口返回所有状态商品
- 前台接口只返回 `online` 状态商品
- `tag_id` 传 `-1` 可获取未绑定任何标签的商品列表

---

## 12. 获取标签未绑定的商品列表

**接口**: GET `/tag/unbound-products`

**描述**: 获取指定标签未绑定的商品列表（分页）

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| tag_id | string | 是 | - | 标签ID |
| shop_id | string | 是 | - | 店铺ID |
| page | int | 否 | 1 | 页码 |
| pageSize | int | 否 | 10 | 每页数量 |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "total": 50,
    "page": 1,
    "pageSize": 10,
    "products": [
      {
        "id": 1234567890123456789,
        "shop_id": 1234567890,
        "name": "美式咖啡",
        "description": "经典美式",
        "price": 22.00,
        "stock": 80,
        "status": "online",
        "image_url": "product_yyy.jpg"
      }
    ]
  }
}
```

---

## 13. 获取未绑定商品的标签列表

**接口**: GET `/tag/unbound-list`

**描述**: 获取没有绑定任何商品的标签列表

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| shop_id | string | 是 | - | 店铺ID |
| page | int | 否 | 1 | 页码 |
| pageSize | int | 否 | 10 | 每页数量 |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "total": 5,
    "page": 1,
    "pageSize": 10,
    "tags": [
      {
        "id": 3,
        "shop_id": 1234567890,
        "name": "限时特惠",
        "description": "限时特惠商品",
        "created_at": "2025-02-06T10:00:00Z",
        "updated_at": "2025-02-06T10:00:00Z"
      }
    ]
  }
}
```

---

## 14. 获取标签关联的已上架商品

**接口**: GET `/tag/online-products`

**描述**: 获取指定标签关联的所有已上架商品

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| tag_id | string | 是 | 标签ID |
| shop_id | string | 是 | 店铺ID |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "tag_id": "1",
    "products": [
      {
        "id": 1234567890123456789,
        "shop_id": 1234567890,
        "name": "拿铁咖啡",
        "description": "香浓拿铁",
        "price": 28.00,
        "stock": 100,
        "status": "online",
        "image_url": "product_xxx.jpg"
      }
    ]
  }
}
```

**错误响应**:

```json
{
  "code": 400,
  "error": "缺少标签ID"
}
```

---

## 错误码汇总

| 状态码 | 错误信息 | 说明 |
|--------|----------|------|
| 400 | 无效的标签数据 | 请求参数缺失或格式错误 |
| 400 | 缺少标签ID | 未提供标签ID |
| 400 | 缺少商品ID | 未提供商品ID |
| 400 | 无效的店铺ID | 店铺ID格式错误 |
| 400 | 该标签已关联 N 个商品，请先解除关联后再删除 | 标签有关联商品 |
| 404 | 标签不存在 | 标签ID不存在 |
| 500 | 创建标签失败 | 服务器内部错误 |
