# 商品管理接口文档

## 接口列表

| 接口 | 方法 | 路径 | 权限 | 描述 |
|------|------|------|------|------|
| 创建商品 | POST | `/product/create` | 管理员/店主 | 创建新商品 |
| 更新商品 | PUT | `/product/update` | 管理员/店主 | 更新商品信息 |
| 获取商品列表 | GET | `/product/list` | 管理员/店主/前台 | 分页获取商品列表 |
| 获取商品详情 | GET | `/product/detail` | 管理员/店主/前台 | 获取单个商品详情 |
| 删除商品 | DELETE | `/product/delete` | 管理员/店主 | 删除商品 |
| 切换状态 | PUT | `/product/toggle-status` | 管理员/店主 | 切换商品状态 |
| 上传图片 | POST | `/product/upload-image` | 管理员/店主 | 上传商品图片 |
| 获取图片 | GET | `/product/image` | 公开 | 获取商品图片 |

**路径前缀**:
- 管理员: `/api/order-ease/v1/admin`
- 店主: `/api/order-ease/v1/shopOwner`
- 前台: `/api/order-ease/v1`

---

## 商品状态说明

| 状态 | 值 | 说明 | 可修改字段 |
|------|-----|------|-----------|
| pending | 待上架 | 初始状态，未上架销售 | 所有字段 |
| online | 已上架 | 正在销售中 | 库存、描述、图片（不可改名称和价格）|
| offline | 已下架 | 已停止销售 | 无（不可修改）|

**状态流转规则**:
1. 新建商品默认为"待上架"状态
2. "待上架" → "已上架"（上架商品）
3. "已上架" → "已下架"（下架商品）
4. "已下架" 不可再次上架，需要创建新商品

---

## 1. 创建商品

**接口**: POST `/product/create`

**描述**: 创建新商品，支持配置参数类别和选项

**认证**: 需要

**请求参数**:

```json
{
  "shop_id": 1234567890,
  "name": "拿铁咖啡",
  "description": "香浓拿铁，选用优质咖啡豆",
  "price": 28.00,
  "stock": 100,
  "image_url": "",
  "option_categories": [
    {
      "name": "杯型",
      "is_required": true,
      "is_multiple": false,
      "display_order": 1,
      "options": [
        {
          "name": "中杯",
          "price_adjustment": 0,
          "display_order": 1
        },
        {
          "name": "大杯",
          "price_adjustment": 5,
          "display_order": 2
        }
      ]
    },
    {
      "name": "温度",
      "is_required": true,
      "is_multiple": false,
      "display_order": 2,
      "options": [
        {
          "name": "热饮",
          "price_adjustment": 0,
          "display_order": 1
        },
        {
          "name": "去冰",
          "price_adjustment": 0,
          "display_order": 2
        },
        {
          "name": "正常冰",
          "price_adjustment": 0,
          "display_order": 3
        }
      ]
    }
  ]
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 是 | 店铺ID |
| name | string | 是 | 商品名称，1-200字符 |
| description | string | 否 | 商品描述，最大5000字符 |
| price | float64 | 是 | 基础价格，必须大于0 |
| stock | int | 是 | 库存数量，必须大于等于0 |
| image_url | string | 否 | 商品图片URL |
| option_categories | array | 否 | 参数类别列表 |

**OptionCategory 字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 类别名称 |
| is_required | boolean | 否 | 是否必填，默认false |
| is_multiple | boolean | 否 | 是否可多选，默认false |
| display_order | int | 否 | 显示顺序，默认0 |
| options | array | 是 | 选项列表 |

**Option 字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 选项名称 |
| price_adjustment | float64 | 否 | 价格调整值，可为正负 |
| display_order | int | 否 | 显示顺序，默认0 |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "id": 1234567890123456789,
    "shop_id": 1234567890,
    "name": "拿铁咖啡",
    "description": "香浓拿铁，选用优质咖啡豆",
    "price": 28.00,
    "stock": 100,
    "image_url": "",
    "status": "pending",
    "created_at": "2025-02-06T10:00:00Z",
    "updated_at": "2025-02-06T10:00:00Z",
    "option_categories": [
      {
        "id": 1234567890123456790,
        "product_id": 1234567890123456789,
        "name": "杯型",
        "is_required": true,
        "is_multiple": false,
        "display_order": 1,
        "options": [
          {
            "id": 1234567890123456791,
            "category_id": 1234567890123456790,
            "name": "中杯",
            "price_adjustment": 0,
            "display_order": 1
          },
          {
            "id": 1234567890123456792,
            "category_id": 1234567890123456790,
            "name": "大杯",
            "price_adjustment": 5,
            "display_order": 2
          }
        ]
      }
    ]
  }
}
```

**错误响应**:

```json
{
  "code": 400,
  "error": "无效的商品数据: Key: 'CreateProductRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag"
}
```

---

## 2. 更新商品

**接口**: PUT `/product/update`

**描述**: 更新商品信息，支持修改参数类别

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 商品ID |
| shop_id | string | 是 | 店铺ID |

**请求参数**:

```json
{
  "name": "拿铁咖啡（升级款）",
  "description": "更新后的描述",
  "price": 32.00,
  "stock": 150,
  "image_url": "product_xxx.jpg",
  "option_categories": [
    {
      "name": "杯型",
      "is_required": true,
      "is_multiple": false,
      "display_order": 1,
      "options": [
        {
          "name": "中杯",
          "price_adjustment": 0,
          "display_order": 1
        },
        {
          "name": "大杯",
          "price_adjustment": 5,
          "display_order": 2
        },
        {
          "name": "超大杯",
          "price_adjustment": 8,
          "display_order": 3
        }
      ]
    }
  ]
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 商品名称（已上架商品不可修改）|
| description | string | 否 | 商品描述 |
| price | float64 | 否 | 基础价格（已上架商品不可修改）|
| stock | int | 否 | 库存数量 |
| image_url | string | 否 | 商品图片URL |
| option_categories | array | 否 | 参数类别列表（提供则替换所有现有类别）|

**注意事项**:
- 已上架商品不可修改名称和价格
- option_categories 如果提供，会替换所有现有参数类别
- stock 字段使用指针类型，传0表示将库存设为0

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "id": 1234567890123456789,
    "shop_id": 1234567890,
    "name": "拿铁咖啡（升级款）",
    "description": "更新后的描述",
    "price": 32.00,
    "stock": 150,
    "image_url": "product_xxx.jpg",
    "status": "pending",
    "created_at": "2025-02-06T10:00:00Z",
    "updated_at": "2025-02-06T12:00:00Z",
    "option_categories": [ ... ]
  }
}
```

**错误响应**:

```json
{
  "code": 400,
  "error": "至少需要提供一个要更新的字段"
}
```

```json
{
  "code": 403,
  "error": "无权操作此商品"
}
```

---

## 3. 获取商品列表

**接口**: GET `/product/list`

**描述**: 分页获取商品列表，前台只返回已上架商品

**认证**: 需要（管理员/店主），可选（前台）

**查询参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| shop_id | string | 是 | - | 店铺ID |
| page | int | 否 | 1 | 页码 |
| pageSize | int | 否 | 10 | 每页数量 |
| search | string | 否 | - | 搜索关键词（商品名称）|

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "total": 100,
    "page": 1,
    "pageSize": 10,
    "data": [
      {
        "id": 1234567890123456789,
        "shop_id": 1234567890,
        "name": "拿铁咖啡",
        "description": "香浓拿铁...",
        "price": 28.00,
        "stock": 100,
        "image_url": "product_xxx.jpg",
        "status": "online",
        "created_at": "2025-02-06T10:00:00Z",
        "updated_at": "2025-02-06T10:00:00Z",
        "option_categories": [ ... ]
      }
    ]
  }
}
```

**说明**:
- 管理员/店主接口返回所有状态商品
- 前台接口只返回 `online` 状态商品

---

## 4. 获取商品详情

**接口**: GET `/product/detail`

**描述**: 获取单个商品详情

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 商品ID |
| shop_id | string | 是 | 店铺ID |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "id": 1234567890123456789,
    "shop_id": 1234567890,
    "name": "拿铁咖啡",
    "description": "香浓拿铁，选用优质咖啡豆",
    "price": 28.00,
    "stock": 100,
    "image_url": "product_xxx.jpg",
    "status": "online",
    "created_at": "2025-02-06T10:00:00Z",
    "updated_at": "2025-02-06T10:00:00Z",
    "option_categories": [
      {
        "id": 1234567890123456790,
        "product_id": 1234567890123456789,
        "name": "杯型",
        "is_required": true,
        "is_multiple": false,
        "display_order": 1,
        "options": [
          {
            "id": 1234567890123456791,
            "category_id": 1234567890123456790,
            "name": "中杯",
            "price_adjustment": 0,
            "display_order": 1
          },
          {
            "id": 1234567890123456792,
            "category_id": 1234567890123456790,
            "name": "大杯",
            "price_adjustment": 5,
            "display_order": 2
          }
        ]
      }
    ]
  }
}
```

---

## 5. 删除商品

**接口**: DELETE `/product/delete`

**描述**: 删除商品，会同时删除关联的参数类别和图片

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 商品ID |
| shop_id | string | 是 | 店铺ID |

**成功响应**:

```json
{
  "code": 200,
  "message": "商品删除成功"
}
```

**错误响应**:

```json
{
  "code": 400,
  "error": "商品已有关联订单，无法删除"
}
```

---

## 6. 切换商品状态

**接口**: PUT `/product/toggle-status`

**描述**: 切换商品状态（pending → online → offline）

**认证**: 需要

**请求参数**:

```json
{
  "id": 1234567890123456789,
  "status": "online",
  "shop_id": 1234567890
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint64 | 是 | 商品ID |
| status | string | 是 | 目标状态: pending/online/offline |
| shop_id | uint64 | 是 | 店铺ID |

**状态流转规则**:

| 当前状态 | 可切换到 | 说明 |
|----------|----------|------|
| pending | online | 上架商品 |
| online | offline | 下架商品 |
| offline | - | 不可再次上架 |

**成功响应**:

```json
{
  "code": 200,
  "message": "商品状态更新成功",
  "product": {
    "id": 1234567890123456789,
    "status": "online",
    "updated_at": "2025-02-06T12:00:00Z"
  }
}
```

**错误响应**:

```json
{
  "code": 400,
  "error": "无效的状态变更"
}
```

---

## 7. 上传商品图片

**接口**: POST `/product/upload-image`

**描述**: 上传商品图片，自动压缩至512KB以内

**认证**: 需要

**请求类型**: multipart/form-data

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 商品ID |
| shop_id | string | 是 | 店铺ID |

**请求体**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| image | file | 是 | 图片文件（jpg/png，最大2MB）|

**成功响应**:

```json
{
  "message": "图片上传成功",
  "url": "product_1234567890123456789_abc123.jpg",
  "type": "upload"
}
```

**错误响应**:

```json
{
  "code": 400,
  "error": "只允许上传jpg/jpeg/png格式的图片"
}
```

```json
{
  "code": 400,
  "error": "图片大小不能超过2MB"
}
```

---

## 8. 获取商品图片

**接口**: GET `/product/image`

**描述**: 获取商品图片

**认证**: 公开（前台无需认证，管理端需要）

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| path | string | 是 | 图片文件名 |

**响应**: 图片二进制数据

---

## 错误码汇总

| 状态码 | 错误信息 | 说明 |
|--------|----------|------|
| 400 | 无效的商品数据 | 请求参数缺失或格式错误 |
| 400 | 缺少商品ID | 未提供商品ID |
| 400 | 无效的店铺ID | 店铺ID格式错误 |
| 400 | 至少需要提供一个要更新的字段 | 更新时未提供任何字段 |
| 400 | 无效的状态变更 | 状态流转不符合规则 |
| 400 | 只允许上传jpg/jpeg/png格式的图片 | 图片格式错误 |
| 400 | 图片大小不能超过2MB | 图片过大 |
| 403 | 无权操作此商品 | 商品不属于当前店铺 |
| 404 | 商品不存在 | 商品ID不存在 |
| 500 | 创建商品失败 | 服务器内部错误 |
