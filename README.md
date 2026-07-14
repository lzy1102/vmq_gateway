# V免签支付网关

基于 Android APP 监听支付宝/微信收款通知，HTTP 回调服务端完成订单匹配。无第三方支付 API 依赖，只需一台安卓手机。

📖 **[接入文档](docs/integration.md)** | 📚 **[管理后台 - 接入教程](http://186.241.107.44/dashboard/tutorial)** | 💻 **[示例代码](examples/)**

## 架构

```
你的服务 ──POST──> V免签网关 ──> 创建订单
                                    │
                              Android 手机监听
                                    │
                          收到支付宝/微信收款通知
                                    │
                        ┌───────────┴───────────┐
                   回调模式                  轮询模式
              POST 到你的地址          你主动查询状态
```

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

---

## 接入指南

### 1. 创建服务绑定

登录管理后台 `http://你的域名/dashboard/bindings`，点击添加：

| 字段 | 说明 |
|------|------|
| 服务ID | 你的应用标识，如 `my_shop` |
| 回调地址 | 可选，不填则轮询模式 |

创建后保存生成的 **API Key**。

### 2. 创建订单

```bash
curl -X POST http://你的域名/api/order \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100,
    "service_id": "my_shop",
    "api_key": "你的API Key",
    "pay_type": "alipay"
  }'
```

### 3. 用户付款后收到通知

- **回调模式**：网关 POST 到你的回调地址
- **轮询模式**：你主动查询 `GET /api/order/status?order_id=xxx`

---

## API 接口

### 创建订单

```
POST /api/order
Content-Type: application/json
```

**请求参数：**

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| amount | ✅ | int | 金额，单位**分**（100 = 1元） |
| service_id | ✅ | string | 服务ID |
| api_key | ✅ | string | 绑定时生成的 API Key |
| callback_url | ❌ | string | 回调地址，不填则轮询模式 |
| pay_type | ❌ | string | `wechat` 或 `alipay`，默认 `alipay` |

**返回示例：**

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

> **金额浮动说明**：V免签通过金额匹配订单，实际支付金额会在你请求的基础上加 1~19 分浮动。这是正常现象，浮动金额会在订单过期后释放。

### 查询订单状态

```
GET /api/order/status?order_id={订单号}
```

**返回示例：**

```json
{
  "code": 1,
  "data": {
    "order_id": "V1783926073_103",
    "amount": 103,
    "status": "paid",
    "paid_at": 1783926200
  }
}
```

**状态值：**

| 状态 | 说明 |
|------|------|
| pending | 待支付 |
| paid | 已支付 |
| expired | 已过期（15分钟超时） |
| cancelled | 已取消 |

### APP 端接口

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

### 管理后台接口

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

---

## 代码示例

### Python 快速开始

```python
import requests, time

# 创建订单
resp = requests.post("http://你的域名/api/order", json={
    "amount": 100,
    "service_id": "my_shop",
    "api_key": "你的API Key",
    "pay_type": "alipay"
})
order = resp.json()["data"]
print(f"请支付 {order['pay_str']} 元")

# 轮询等待
while True:
    s = requests.get(
        "http://你的域名/api/order/status",
        params={"order_id": order["order_id"]}
    ).json()["data"]["status"]
    if s == "paid":
        print("✅ 收款成功")
        break
    elif s in ("expired", "cancelled"):
        print("❌ 订单已失效")
        break
    time.sleep(3)
```

### Node.js 快速开始

```javascript
import http from 'node:http'

const GATEWAY = 'http://你的域名'

function request(method, path, body) {
  return new Promise((resolve, reject) => {
    const url = new URL(path, GATEWAY)
    const req = http.request({
      hostname: url.hostname,
      path: url.pathname,
      method,
      headers: { 'Content-Type': 'application/json' }
    }, res => {
      let data = ''
      res.on('data', chunk => data += chunk)
      res.on('end', () => resolve(JSON.parse(data)))
    })
    req.on('error', reject)
    if (body) req.write(JSON.stringify(body))
    req.end()
  })
}

// 创建订单
const resp = await request('POST', '/api/order', {
  amount: 100,
  service_id: 'my_shop',
  api_key: '你的API Key',
  pay_type: 'alipay'
})
console.log(`请支付 ${resp.data.pay_str} 元`)

// 轮询
while (true) {
  const s = await request('GET', `/api/order/status?order_id=${resp.data.order_id}`)
  if (s.data.status === 'paid') { console.log('✅ 收款成功'); break }
  if (['expired','cancelled'].includes(s.data.status)) { console.log('❌ 失效'); break }
  await new Promise(r => setTimeout(r, 3000))
}
```

### Go 快速开始

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type OrderReq struct {
    Amount    int    `json:"amount"`
    ServiceID string `json:"service_id"`
    APIKey    string `json:"api_key"`
    PayType   string `json:"pay_type"`
}

type OrderResp struct {
    Code int         `json:"code"`
    Data *OrderData  `json:"data"`
}

type OrderData struct {
    OrderID string `json:"order_id"`
    PayStr  string `json:"pay_str"`
}

func main() {
    body, _ := json.Marshal(OrderReq{
        Amount:    100,
        ServiceID: "my_shop",
        APIKey:    "你的API Key",
        PayType:   "alipay",
    })

    resp, _ := http.Post("http://你的域名/api/order", "application/json", bytes.NewReader(body))
    var result OrderResp
    json.NewDecoder(resp.Body).Decode(&result)
    fmt.Printf("请支付 %s 元\n", result.Data.PayStr)
}
```

更多示例见 [`examples/`](examples/) 目录。

---

## 两种通知模式

### 回调模式

适用场景：有公网可访问的服务端

1. 创建订单时填写 `callback_url`
2. 用户付款后，网关 POST 到你的地址

**回调请求格式：**

```json
{
  "trade_no": "V1783926073_103",
  "amount": 103,
  "pay_type": "alipay",
  "paid_at": 1783926200,
  "service_id": "my_shop"
}
```

你的服务需返回 `{"code":1,"msg":"ok"}`。

### 轮询模式

适用场景：内网服务、小程序、定时任务、无法接收回调的场景

1. 创建订单时**不填** `callback_url`
2. 每隔 3 秒查询一次订单状态
3. 收到 `status: paid` 后处理业务

```bash
curl "http://你的域名/api/order/status?order_id=V1783926073_103"
```

---

## 签名算法

签名用于 APP 端与网关通信，业务端接入**不需要**了解签名。

```go
// Go
sign := fmt.Sprintf("%x", md5.Sum([]byte(ts + key)))
```

```python
# Python
sign = hashlib.md5((ts + key).encode()).hexdigest()
```

```javascript
// Node.js
const sign = crypto.createHash('md5').update(ts + key).digest('hex')
```

## 订单生命周期

```
创建订单 (pending)
    │
    ├── 用户支付成功 → paid（触发回调或轮询返回）
    │
    ├── 15分钟超时 → expired（浮动金额释放）
    │
    └── 用户取消 → cancelled
```

- 默认有效期：15 分钟
- 后台每 30 秒自动清理过期订单
- 同一时刻没有两个 pending 订单金额相同

## 设备状态

- **上线**：APP 每 50 秒发心跳，服务端标记 `online`
- **离线**：后台每 30 秒检查，超过 60 秒没心跳自动标 `offline`

## 管理后台功能

- **设备管理**：添加/删除设备，上传微信/支付宝收款码，查看在线状态
- **支付池**：将设备分组管理
- **服务绑定**：配置回调地址，生成 API Key
- **订单管理**：查看所有订单，按状态筛选，倒计时显示
- **接入教程**：内置接入指南和请求示例

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
├── docs/
│   └── integration.md      # 接入文档
├── examples/
│   ├── python/             # Python 示例
│   ├── node/               # Node.js 示例
│   └── go/                 # Go 示例
├── test.sh                 # 完整测试
├── test_flow.sh            # 已有设备测试
└── simulate_device.sh      # 心跳模拟器
```

---

📖 **详细接入文档：[docs/integration.md](docs/integration.md)**
