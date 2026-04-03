---
name: "orderease-api-test"
description: "OrderEase API自动化测试方法指南。当需要编写、修改或理解 orderease-deploy 项目中的 Python API 测试用例时调用。涵盖测试框架、Fixture模式、业务流程测试、边界条件测试、安全测试等全部测试模式。"
---

# OrderEase API 自动化测试方法指南

## 1. 技术栈概览

| 组件 | 版本/说明 |
|------|----------|
| 测试框架 | pytest 8.3.2 |
| HTTP客户端 | requests 2.32.3 |
| 环境变量 | python-dotenv 1.0.1 |
| 性能测试 | locust 2.24.0 |
| API文档测试 | schemathesis 3.30.2 |
| HTML报告 | pytest-html 4.1.1 |
| API基础路径 | `http://localhost:8080/api/order-ease/v1` |

## 2. 项目目录结构

```
test/
├── conftest.py                    # 全局配置：fixtures、重试机制、测试排序、计时钩子
├── pytest.ini                     # pytest 配置
├── requirements.txt               # 依赖
├── config/
│   └── test_data.py              # 测试数据工厂（单例模式）
├── utils/
│   ├── base_test.py              # 混入基类（Boundary/Auth/Pagination/Validation/Concurrency）
│   ├── response_validator.py     # 链式响应验证器
│   └── field_resolver.py         # 多格式字段解析器
├── admin/                        # 管理员接口测试 + actions 操作工具类
├── shop_owner/                   # 商家接口测试 + actions 操作工具类
├── frontend/                     # 前端用户接口测试
└── auth/                         # 认证/安全测试（未授权访问、越权等）
```

## 3. 核心设计模式

### 3.1 Fixture 分层体系（conftest.py）

**Session 级别 Fixtures** — 整个测试会话共享一次：

```python
@pytest.fixture(scope="session")
def api_base_url():
    return API_BASE_URL

@pytest.fixture(scope="session")
def admin_token():
    """管理员令牌 - 通过登录获取真实token"""
    url = f"{API_BASE_URL}/login"
    payload = {"username": "admin", "password": "Admin@123456"}
    response = make_request_with_retry(lambda: requests.post(url, json=payload))
    return response.json().get("token", "") if response.status_code == 200 else ""

@pytest.fixture(scope="session")
def shop_owner_token(admin_token):
    """商家令牌 - 动态创建店铺并获取token"""
    # 1. 用admin_token创建店铺(同时创建店主用户)
    # 2. 用店主账号登录获取token
    ...

@pytest.fixture(scope="session")
def test_shop_id(admin_token):
    """从数据库获取第一个店铺ID"""

@pytest.fixture(scope="session")
def test_product_id(admin_token, test_shop_id):
    """获取或创建商品ID"""

@pytest.fixture(scope="session")
def test_order_id(admin_token, test_shop_id):
    """从数据库获取第一个订单ID"""
```

**Function 级别 Fixtures** — 每个测试函数独立创建：

```python
@pytest.fixture(scope="function")
def shop_owner_order_id(shop_owner_token, shop_owner_shop_id, ...):
    """每次测试创建新订单，避免状态污染"""
```

### 3.2 重试与容错机制（make_request_with_retry）

```python
def make_request_with_retry(request_func, max_retries=10, initial_wait=1, backoff_factor=2):
    """指数退避重试，专门处理429速率限制"""
    retry_count = 0
    wait_time = initial_wait
    while True:
        response = request_func()
        if response.status_code == 429 and retry_count < max_retries:
            time.sleep(wait_time)
            retry_count += 1
            wait_time *= backoff_factor  # 指数退避
        else:
            return response
```

**使用方式**：所有HTTP请求都通过 `make_request_with_retry` 包裹：
```python
def request_func():
    return requests.post(url, json=payload, headers=headers)
response = make_request_with_retry(request_func)
```

### 3.3 链式响应验证器（ResponseValidator）

```python
from utils.response_validator import ResponseValidator, validate_response, assert_success_response, assert_error_response

# 链式调用风格
ResponseValidator(response)\
    .status(200)\
    .has_data()\
    .has_field("data.id", expected_type=int)\
    .field_equals("data.name", "Expected Name")\
    .list_length("data.items", min_length=1)

# 便捷函数
validate_response(response, 200)           # 验证状态码
assert_success_response(response)          # 断言成功响应
assert_error_response(response, 400, "error message")  # 断言错误响应
```

### 3.4 多格式字段解析器（FieldResolver）

自动处理 Go 后端返回的 **PascalCase / camelCase / snake_case** 混合格式：

```python
from utils.field_resolver import FieldResolver

# 获取嵌套字段（自动尝试多种命名变体）
FieldResolver.get_nested_field(data, "data.shop_id")       # 匹配 shopId/shop_id/ShopID
FieldResolver.extract_id(data)                              # 从任意位置提取 ID
FieldResolver.get_list(data, "products")                    # 获取列表
FieldResolver.normalize_keys(data, to_snake_case=True)      # 键名规范化
```

### 3.5 测试数据工厂（TestDataConfig）

单例模式，统一管理测试数据生成，使用随机后缀避免冲突：

```python
from config.test_data import test_data

# 生成唯一后缀的测试数据
shop_data = test_data.generate_shop_data()                  # 店铺
product_data = test_data.generate_product_data(shop_id)     # 商品
order_data = test_data.generate_order_data(shop_id, user_id, product_id)  # 订单
user_data = test_data.generate_user_data()                  # 用户
tag_data = test_data.generate_tag_data(shop_id)             # 标签
```

## 4. Actions 操作工具类模式

将每个业务实体的 CRUD 操作封装为独立的 **actions 模块**，测试文件只负责编排流程：

```
admin/
├── shop_actions.py      # create_shop, update_shop, delete_shop, get_shop_list, ...
├── product_actions.py   # create_product, update_product, delete_product, ...
├── order_actions.py     # create_order, update_order, delete_order, ...
├── user_actions.py      # create_user, update_user, delete_user, ...
└── tag_actions.py       # create_tag, update_tag, delete_tag, ...
```

**actions 函数签名规范**：
```python
def create_shop(admin_token, name=None, description=None, address=None,
                owner_username=None, owner_password=None, contact_phone=None,
                contact_email=None):
    """
    Returns:
        shop_id: 成功返回ID，失败返回None
    """
    # 1. 如果没传参数，使用 test_data 工厂生成默认值
    # 2. 使用 make_request_with_retry 发送请求
    # 3. 使用 ResponseValidator 提取 ID
    # 4. 打印 [OK]/[FAIL]/[WARN] 前缀日志
    ...
```

## 5. 测试类型与编写规范

### 5.1 业务流程测试（Business Flow Test）

按正确的业务顺序编排端到端测试，是**核心测试类型**：

```python
class TestBusinessFlow:
    """业务流程测试类"""

    @pytest.fixture(scope="function", autouse=True)
    def setup_and_teardown(self, admin_token):
        """每个测试函数前后自动执行 setup/teardown"""
        self.admin_token = admin_token
        self.cleanup_resources = []  # 注册清理函数
        yield
        # Teardown: 反序清理测试数据
        for cleanup_func in reversed(self.cleanup_resources):
            try:
                cleanup_func()
            except Exception as e:
                print(f"清理资源时出错: {e}")

    def test_complete_business_flow(self):
        """完整业务流程：创建店铺 → 创建商品 → 创建订单 → 删除订单 → 删除商品 → 删除店铺"""
        shop_id = self._test_create_shop()
        product_id = self._test_create_product(shop_id)
        user_id = self._test_get_user_id()
        order_id = self._test_create_order(shop_id, user_id, product_id)

        self._test_delete_order(order_id, shop_id)
        self._test_delete_product(product_id, shop_id)
        self._test_delete_shop(shop_id)

    def _test_create_shop(self):
        """辅助方法：创建资源并注册清理函数"""
        shop_id = shop_actions.create_shop(self.admin_token, ...)
        if shop_id:
            self.cleanup_resources.append(lambda: shop_actions.delete_shop(self.admin_token, shop_id))
        return shop_id
```

**关键要点**：
- 使用 `autouse=True` fixture 实现 setup/teardown
- 通过 `cleanup_resources` 列表注册清理函数（lambda），确保测试失败也能清理
- 辅助方法以 `_test_` 前缀命名
- 每步都 assert 并打印 ✓/✗ 日志

### 5.2 边界条件测试（Boundary Condition Test）

使用 **Mixin 模式** 复用边界测试逻辑：

```python
from utils.base_test import BoundaryTestMixin, ValidationTestMixin

class TestProductBoundaryConditions(BoundaryTestMixin, ValidationTestMixin):
    """继承混入类获得标准边界测试能力"""

    @pytest.fixture(autouse=True)
    def setup(self, admin_token):
        self.admin_token = admin_token
        self.shop_id = self._create_test_shop()

    def make_request_func(self, payload):
        """必须实现：返回请求函数"""
        url = f"{API_BASE_URL}/shopOwner/product/create"
        headers = {"Authorization": f"Bearer {self.admin_token}"}
        def request_func():
            return requests.post(url, json=payload, headers=headers)
        return request_func

    def _get_valid_payload(self):
        """提供有效载荷模板"""
        return test_data.generate_product_data(self.shop_id)

    # ===== 直接使用 Mixin 提供的方法 =====
    def test_empty_product_name(self):
        time.sleep(0.5)  # 避免速率限制
        self._check_empty_string("name", self._get_valid_payload())

    def test_negative_product_price(self):
        self._check_negative_value("price", self._get_valid_payload())

    def test_very_long_product_name(self):
        self._check_very_long_string("name", self._get_valid_payload(), length=500)
```

**BoundaryTestMixin 提供的内置方法**：

| 方法 | 测试场景 |
|------|---------|
| `test_empty_payload()` | 空载荷 |
| `_check_missing_required_field(field, payload)` | 缺少必填字段 |
| `_check_invalid_field_type(field, value, payload)` | 无效字段类型 |
| `_check_negative_value(field, payload)` | 负数值 |
| `_check_zero_value(field, payload)` | 零值 |
| `_check_empty_string(field, payload)` | 空字符串 |
| `_check_very_long_string(field, payload, length)` | 超长字符串 |
| `_check_sql_injection_attempt(field, payload)` | SQL注入 |
| `_check_invalid_token(make_request_func)` | 无效token |

**ValidationTestMixin 提供的内置方法**：

| 方法 | 测试场景 |
|------|---------|
| `_check_invalid_phone_number(field, payload)` | 无效电话号码 |
| `_check_invalid_email(field, payload)` | 无效邮箱 |
| `_check_invalid_password(field, payload)` | 无效密码 |

### 5.3 安全测试（Security Test）

**未授权访问测试**：
```python
class TestUnauthorizedAccess:
    def test_admin_api_without_token(self):
        """无token访问 → 期望401"""
        response = make_request_with_retry(lambda: requests.get(url))
        assert response.status_code == 401

    def test_admin_api_with_invalid_token(self):
        """无效token访问 → 期望401"""
        headers = {"Authorization": "Bearer invalid_token"}
        response = make_request_with_retry(lambda: requests.get(url, headers=headers))
        assert response.status_code == 401

    def test_empty_token(self):
        """空token → 期望401"""
        headers = {"Authorization": "Bearer "}
        ...
```

**特殊字符/XSS测试**：
```python
special_names = [
    "<script>alert('xss')</script>",
    "'; DROP TABLE products; --",
    "../../etc/passwd",
    "test\x00null",
]
for name in special_names:
    payload["name"] = name
    response = make_request_func(payload)()
    assert response.status_code in [200, 400, 422, 429]
    if response.status_code == 200:
        assert "<script>" not in str(response.json())  # 验证转义
```

### 5.4 认证流程测试（Auth Flow Test）

测试间有依赖关系时，使用**类变量共享数据** + **测试排序控制**：

```python
class TestAuthFlow:
    shop_owner_token_value = None     # 类变量在测试间共享
    shop_owner_shop_id_value = None

    def test_universal_login_admin(self):        # 优先级 0
        """管理员登录"""

    def test_universal_login_shop_owner(self, admin_token):  # 优先级 1
        """商家登录（依赖admin_token）"""
        # 保存到类变量供后续测试使用
        TestAuthFlow.shop_owner_token_value = login_data.get("token")

    def test_refresh_admin_token(self, admin_token):  # 优先级 2
        """刷新管理员令牌"""

    def test_z_admin_logout(self, admin_token):      # 优先级 7（最后）
        """管理员登出（z_前缀确保最后执行）"""
```

**测试顺序控制**（conftest.py 中 `pytest_collection_modifyitems`）：
- 前端流程测试 > 管理员业务流程 > 商家业务流程 > 认证流程 > 未授权测试
- 认证内部：登录 → 刷新Token → 登出

## 6. 类级别资源共享模式

当多个测试需要相同的基础资源时，使用 `setup_class` / `teardown_class`：

```python
class TestShopOwnerBusinessFlow:
    @classmethod
    def setup_class(cls):
        """一次性创建所有共享资源：店铺→商品→用户→标签→订单"""
        cls.resources = {}
        cls.resources['shop_id'] = create_shop(...)
        cls.resources['product_id'] = create_product(...)
        cls.resources['user_id'] = get_or_create_user(...)
        cls.resources['tag_id'] = create_tag(...)
        cls.resources['order_id'] = create_order(...)

    @classmethod
    def teardown_class(cls):
        """反序清理所有共享资源"""
        delete_order(cls.resources['order_id'])
        delete_tag(cls.resources['tag_id'])
        delete_product(cls.resources['product_id'])
        delete_shop(cls.resources['shop_id'])

    def test_get_shop_detail(self):
        result = shop_actions.get_shop_detail(token, self.resources['shop_id'])

    def test_complete_business_flow(self):
        assert 'shop_id' in self.resources  # 验证资源存在
```

## 7. 编写新测试用例的 Checklist

当需要为 OrderEase 项目编写新的 API 测试时，按以下步骤操作：

### Step 1: 确定测试位置
- **管理员接口** → `test/admin/` 目录
- **商家接口** → `test/shop_owner/` 目录
- **前端用户接口** → `test/frontend/` 目录
- **认证/安全** → `test/auth/` 目录

### Step 2: 选择测试类型
- **CRUD 功能验证** → 在对应目录的 `test_business_flow.py` 中添加方法
- **边界条件** → 新建 `test_xxx_boundary.py`，继承 `BoundaryTestMixin`
- **安全测试** → 在 `auth/test_unauthorized.py` 中添加

### Step 3: 编写代码模板

**功能测试模板**：
```python
def test_xxx_feature(self):
    print("\n========== 测试名称 ==========")
    # 1. 准备数据（使用 test_data 工厂）
    # 2. 调用 actions 函数
    # 3. assert 验证结果
    # 4. 打印 ✓ 日志
    # 5. 清理资源（注册到 cleanup_resources）
```

**边界测试模板**：
```python
class TestXxxBoundaryConditions(BoundaryTestMixin):
    def make_request_func(self, payload):  # 必须实现
        ...
    def _get_valid_payload(self):          # 推荐
        return test_data.generate_xxx_data(...)
    # 然后直接调用 Mixin 方法
```

### Step 4: 必须遵守的规则
1. ✅ 所有 HTTP 请求必须通过 `make_request_with_retry` 包裹
2. ✅ 使用 `time.sleep(0.5~1.5)` 在测试间添加延迟避免 429
3. ✅ 使用 `test_data` 工厂生成唯一测试数据（避免冲突）
4. ✅ 使用 `actions` 模块的函数而非直接发请求（复用逻辑）
5. ✅ 使用 `ResponseValidator` 或 `assert_response_status` 验证响应
6. ✅ 测试数据必须在 teardown 中清理（cleanup_resources 模式）
7. ✅ 打印 `[OK]`/`[FAIL]`/`[WARN]` 前缀日志
8. ✅ 状态码断言要宽容：`assert status in [200, 400, 422, 429]`
9. ✅ 列表元素验证要检查 `isinstance(x, dict)` 和必需字段存在性

## 8. 运行命令速查

```bash
# 运行所有测试
pytest -v

# 运行特定模块
pytest admin/ -v              # 管理员测试
pytest shop_owner/ -v         # 商家测试
pytest frontend/ -v           # 前端测试
pytest auth/ -v               # 认证测试

# 运行特定文件/类/方法
pytest admin/test_business_flow.py::TestBusinessFlow::test_complete_business_flow -v

# 生成报告
pytest -v --html=report.html
pytest -v --junitxml=report.xml

# 性能测试
locust -f locustfile.py --host=http://localhost:8080
```
