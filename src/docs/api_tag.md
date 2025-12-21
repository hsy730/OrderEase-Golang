# 标签相关 API 文档

## 获取商品标签列表
- **方法**: GET
- **路径**: /tag/list
- **描述**: 获取所有标签
- **响应**:
  ```json
  {
    "code": 200,
    "data": [
      {
        "id": 1,
        "name": "标签1"
      },
      {
        "id": 2,
        "name": "标签2"
      }
    ]
  }
  ```

### 获取标签详情  
- **方法**: GET
- **路径**: /tag/detail
- **描述**: 获取单个标签详情
- **请求参数**:
  - tag_id (string): 标签ID
- **响应**:
  ```json
  {
    "code": 200,
    "data": {
      "id": "TAG1",
      "name": "标签1",
      "description": "标签描述",
      "product_count": 10
    }
  }
  ```