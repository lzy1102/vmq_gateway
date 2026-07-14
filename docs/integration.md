# V免签支付网关 - 接入文档

## 概览

V免签是一个自托管支付网关，通过 Android 手机监听支付宝/微信收款通知，无需第三方支付 API。

**架构：**

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

---

## 快速开始

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

**回调模式**：网关 POST 到你的回调地址
**轮询模式**：你主动查询 `GET /api/order/status?order_id=xxx`

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

**字段说明：**

| 字段 | 说明 |
|------|------|
| order_id | 订单号 |
| request_amount | 你请求的金额（分） |
| pay_amount | 用户实际支付金额（分），含浮动 |
| qr_url | 收款二维码地址 |
| expire_at | 订单过期时间戳 |
| remaining_seconds | 剩余秒数 |

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

---

## 两种通知模式

### 回调模式

适用场景：有公网可访问的服务端

**流程：**
1. 创建订单时填写 `callback_url`
2. 用户付款后，网关 POST 到你的地址

**回调请求格式：**

```
POST 你的回调地址
Content-Type: application/json

{
  "trade_no": "V1783926073_103",
  "amount": 103,
  "pay_type": "alipay",
  "paid_at": 1783926200,
  "service_id": "my_shop"
}
```

**你的服务需返回：**

```json
{"code": 1, "msg": "ok"}
```

### 轮询模式

适用场景：内网服务、小程序、定时任务、无法接收回调的场景

**流程：**
1. 创建订单时**不填** `callback_url`
2. 每隔 3 秒查询一次订单状态
3. 收到 `status: paid` 后处理业务

```bash
curl "http://你的域名/api/order/status?order_id=V1783926073_103"
```

---

## 代码示例

完整示例见 `examples/` 目录：

| 语言 | 文件 | 说明 |
|------|------|------|
| Python | `examples/python/example.py` | Flask 回调 + 轮询等待 |
| Node.js | `examples/node/poll.js` | 轮询模式 |
| Node.js | `examples/node/callback.js` | Express 回调服务 |
| Go | `examples/go/main.go` | 两种模式完整示例 |

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
    "time"
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

---

## 签名算法

签名用于 APP 端与网关通信，业务端接入**不需要**了解签名。

### 心跳签名

```
sign = MD5(t + key)
```

### 收款推送签名

```
sign = MD5(type + price + t + key)
```

```python
# Python
import hashlib
sign = hashlib.md5((t + key).encode()).hexdigest()
```

```javascript
// Node.js
const crypto = require('crypto')
const sign = crypto.createHash('md5').update(t + key).digest('hex')
```

```go
// Go
sign := fmt.Sprintf("%x", md5.Sum([]byte(t + key)))
```

---

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

---

## 常见问题

### Q: 为什么实际支付金额和设置的不一样？

A: V免签通过收款金额匹配订单，会在你请求的金额上加 1~19 分浮动。这是正常现象，浮动金额会在订单过期后释放。

### Q: 订单多长时间过期？

A: 默认 15 分钟。超时后订单自动标记为 expired，浮动金额释放。

### Q: 一个服务可以绑定多个设备吗？

A: 可以。通过支付池管理设备，一个池里可以有多台手机，增加并发处理能力。

### Q: 回调没收到怎么办？

A: 
1. 检查回调地址是否可公网访问
2. 检查服务器防火墙
3. 在「订单管理」查看订单状态
4. 建议用轮询作为兜底

### Q: 支持哪些支付方式？

A: 支持支付宝和微信支付。创建订单时通过 `pay_type` 参数指定。

---

## 示例项目

- [examples/python](../examples/python) - Python 接入示例
- [examples/node](../examples/node) - Node.js 接入示例
- [examples/go](../examples/go) - Go 接入示例
