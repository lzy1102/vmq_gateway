# V免签支付网关

基于 Android APP 监听支付宝/微信收款通知，HTTP 回调服务端完成订单匹配。无第三方支付 API 依赖，只需一台安卓手机。

📖 **[接入文档](docs/integration.md)** | 📚 **[管理后台 - 接入教程](http://186.241.107.44/dashboard/tutorial)**

## 架构

```
Nginx(:80) → Go(:8080) → SQLite
    ↑
  V免签 APP（安卓手机）
```

- **后端**：Go + Gin + SQLite（纯 Go 驱动，无需 CGO）
- **前端**：Vue 3 + Vite + TypeScript
- **部署**：Docker 多阶段编译

## 快速开始

```bash
git clone https://github.com/lzy1102/vmq_gateway.git
cd vmq_gateway/deploy
docker-compose up -d
```

访问 `http://你的IP`，默认账号 `admin` / `vmq_gateway`。

## API 接口

### APP 端

#### 心跳

```
GET /appHeart?t={timestamp_ms}&sign={md5_hex}
```

| 参数 | 说明 |
|------|------|
| t | 毫秒时间戳 |
| sign | `MD5(t + key)` |

#### 收款回调

```
GET /appPush?t={timestamp_ms}&type={1|2}&price={金额_元}&sign={md5_hex}
```

| 参数 | 说明 |
|------|------|
| t | 毫秒时间戳 |
| type | 1=微信，2=支付宝 |
| price | 金额，单位元（如 `10.01`） |
| sign | `MD5(type + price + t + key)` |

### 业务端

#### 创建订单

```
POST /api/order
Content-Type: application/json

{
  "amount": 100,
  "service_id": "my_app",
  "api_key": "绑定时生成的API Key",
  "callback_url": "https://your-server/callback",
  "pay_type": "alipay"
}
```

| 参数 | 必填 | 说明 |
|------|------|------|
| amount | 是 | 申请金额，单位分 |
| service_id | 是 | 服务ID |
| api_key | 是 | 绑定的 API Key |
| callback_url | 否 | 回调地址，不填则轮询模式 |
| pay_type | 否 | wechat/alipay，默认 alipay |

返回：

```json
{
  "code": 1,
  "data": {
    "order_id": "V1783926073_103",
    "request_amount": 100,
    "request_str": "1.00",
    "pay_amount": 103,
    "pay_str": "1.03",
    "device_id": "phone_01",
    "pool_id": "",
    "qr_url": "/qr/phone_01_alipay.jpg",
    "expire_at": 1783926973,
    "remaining_seconds": 900
  }
}
```

#### 查询订单状态

```
GET /api/order/status?order_id={订单号}
```

返回：

```json
{
  "code": 1,
  "data": {
    "order_id": "V1783926073_103",
    "amount": 103,
    "status": "pending",
    "paid_at": 0,
    "created_at": 1783926073,
    "expire_at": 1783926973,
    "remaining_seconds": 899
  }
}
```

### 管理后台

| 接口 | 方法 | 说明 |
|------|------|------|
| `/admin/login` | POST | 登录 |
| `/admin/logout` | POST | 退出 |
| `/admin/devices` | GET | 设备列表（分页） |
| `/admin/device` | POST | 添加设备 |
| `/admin/device` | PUT | 更新设备 |
| `/admin/device` | DELETE | 删除设备 |
| `/admin/device/qrcode` | POST | 上传收款码（type=wechat/alipay） |
| `/admin/pools` | GET | 池列表（分页） |
| `/admin/pool` | POST | 添加池 |
| `/admin/pool` | DELETE | 删除池 |
| `/admin/pool/device` | POST | 添加设备到池 |
| `/admin/pool/device` | DELETE | 从池移除设备 |
| `/admin/bindings` | GET | 绑定列表（分页） |
| `/admin/binding` | POST | 创建绑定 |
| `/admin/binding` | PUT | 更新绑定 |
| `/admin/binding` | DELETE | 删除绑定 |
| `/admin/orders` | GET | 订单列表（分页+搜索+状态筛选） |

## 签名算法

```go
func signHeartbeat(ts, key string) string {
    return fmt.Sprintf("%x", md5.Sum([]byte(ts + key)))
}

func signPush(payType, price, ts, key string) string {
    return fmt.Sprintf("%x", md5.Sum([]byte(payType + price + ts + key)))
}
```

```python
import hashlib

def sign_heartbeat(ts: str, key: str) -> str:
    return hashlib.md5((ts + key).encode()).hexdigest()

def sign_push(pay_type: str, price: str, ts: str, key: str) -> str:
    return hashlib.md5((pay_type + price + ts + key).encode()).hexdigest()
```

## 订单匹配策略

浮动金额匹配（V免签只能拿到金额，拿不到订单号）：

1. 创建订单时：`基础价 + 浮动分（1~19分）`
2. 从小到大依次尝试，优先复用已释放的金额
3. 同一时刻没有两个 pending 订单金额相同
4. APP 回调时按金额匹配 pending 订单
5. 原子操作防止并发重复匹配

## 订单有效期

- 默认 15 分钟
- 超时自动过期，浮动金额释放
- 后台每 30 秒扫描清理过期订单

## 设备状态

- **上线**：APP 每 50 秒发心跳，服务端标记 `online`
- **离线**：后台每 30 秒检查，超过 60 秒没心跳自动标 `offline`

## 管理后台功能

- **设备管理**：添加/删除设备，上传微信/支付宝收款码，查看在线状态
- **支付池**：将设备分组管理
- **服务绑定**：配置回调地址，生成 API Key
- **订单管理**：查看所有订单，按状态筛选，倒计时显示

## 部署

### Docker 部署（推荐）

```bash
cd deploy
docker-compose up -d
```

### 开发环境

```bash
# 后端
cd server
go run .

# 前端
cd web
npm install
npm run dev
```

## 测试脚本

```bash
# 完整流程测试（自动创建设备和绑定）
bash test.sh http://186.241.107.44

# 使用已有设备测试
bash test_flow.sh http://186.241.107.44 <设备Key> <API Key> <服务ID>

# 模拟设备心跳
./simulate_device.sh http://186.241.107.44 <设备Key>
```

## 项目结构

```
vmq_gateway/
├── deploy/
│   ├── Dockerfile          # Go 多阶段编译
│   ├── Dockerfile.nginx    # Vue 多阶段编译
│   ├── docker-compose.yml
│   └── nginx.conf
├── server/
│   ├── main.go             # 路由 + 启动
│   ├── config/             # 配置
│   ├── handler/            # HTTP 处理器
│   ├── middleware/          # 认证中间件
│   ├── model/              # 数据模型
│   ├── service/            # 业务逻辑
│   └── store/              # 数据库（GORM/MongoDB）
├── web/
│   ├── src/
│   │   ├── api/            # API 调用
│   │   ├── router/         # 路由
│   │   ├── types/          # 类型定义
│   │   └── views/          # 页面组件
│   └── public/
│       └── icons/          # 微信/支付宝图标
├── test.sh                 # 完整测试
├── test_flow.sh            # 已有设备测试
└── simulate_device.sh      # 心跳模拟器
```
