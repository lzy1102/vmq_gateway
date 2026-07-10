# V免签 支付网关参考

基于 Android APP 监听支付宝/微信收款通知，HTTP 回调服务端完成订单匹配。无第三方支付 API 依赖，只需一台安卓手机。

## 服务端接口

### 1. 心跳 `/appHeart`

```
GET /appHeart?t={timestamp_ms}&sign={md5_hex}
```

APP 每 50 秒发送一次，证明在线。

| 参数 | 说明 |
|------|------|
| t | 毫秒时间戳 |
| sign | `MD5(t + key)` |

返回：`{"code":1,"msg":"成功"}` / `{"code":0,"msg":"签名错误"}`

### 2. 收款回调 `/appPush`

```
GET /appPush?t={timestamp_ms}&type={1|2}&price={金额_元}&sign={md5_hex}
```

APP 检测到收款时自动调用。

| 参数 | 说明 |
|------|------|
| t | 毫秒时间戳 |
| type | 1=微信，2=支付宝 |
| price | 金额，单位元（如 `10.01`） |
| sign | `MD5(type + price + t + key)` |

返回：`{"code":1,"msg":"成功"}` / `{"code":0,"msg":"未找到匹配订单"}`

---

## 签名算法

```python
import hashlib

key = "your_vmq_key"

def sign_heartbeat(ts: str) -> str:
    return hashlib.md5((ts + key).encode()).hexdigest()

def sign_push(pay_type: str, price: str, ts: str) -> str:
    return hashlib.md5((pay_type + price + ts + key).encode()).hexdigest()
```

```go
func signHeartbeat(ts string) string {
    return fmt.Sprintf("%x", md5.Sum([]byte(ts + key)))
}

func signPush(payType, price, ts string) string {
    return fmt.Sprintf("%x", md5.Sum([]byte(payType + price + ts + key)))
}
```

---

## 订单匹配策略

V免签只能拿到**金额**，拿不到订单号。所以采用浮动金额匹配：

1. 创建订单时生成唯一金额：`基础价 + 浮动分（1~19分）`
2. 确保同一时刻没有两个 pending 订单金额相同（互斥锁 + 原子抢单）
3. APP 回调时按金额匹配：找到第一个 `status=pending` 且 `amount` 匹配的订单
4. 原子操作：`FindOneAndUpdate` 把 `pending` 改为 `processing`，防止并发重复发货

并发安全要点：
- 创建订单：`mu.Lock()` 包裹查金额 + 写入，串行化
- 支付回调：`FindOneAndUpdate` 原子抢单，只改一条

---

## 回调处理流程

```
1. APP 收到收款通知
2. APP → GET /appPush?type=2&price=30.01&sign=xxx
3. 服务端验证签名
4. 金额解析：元 → 分（priceFloat * 100）
5. 查 MongoDB：RechargeOrder{amount: 3001, status: "pending"}
6. 原子更新：pending → processing
7. 给用户充值：User.stream_number += order.stream_number
8. 更新订单：processing → paid
9. 返回 {"code": 1, "msg": "成功"}
```

---

## APP 配置

APP 内部存储：`SharedPreferences("shinian")`

| Key | 值 | 说明 |
|------|------|------|
| host | `服务器IP:端口` | 不含 http:// |
| key | 通讯密钥 | 与服务端一致 |

---

## 前端交互流程

```
用户选套餐 → POST /recharge/vmpay → 创建浮动金额订单
→ 显示金额 + 收款码（静态图片 /qr/alipay.png）
→ 用户扫码支付
→ APP 检测到账 → /appPush 回调
→ 前端轮询 /recharge/vmpay-status 检测到 paid → 弹窗成功
```

前端关键点：
- 提示用户必须按**精确金额**支付（红色加粗）
- 10 分钟倒计时，超时自动取消
- 每 3 秒轮询订单状态

---

## 关键代码片段

### 创建浮动金额订单 (Go)

```go
func nextFloatAmount(baseAmountCents int64) int64 {
    for i := int64(1); i <= 19; i++ {
        candidate := baseAmountCents + i
        var existing Order
        if err := db.FindOne(bson.M{"amount": candidate, "status": "pending"}, &existing); err != nil {
            return candidate // 未占用，返回
        }
    }
    return baseAmountCents // fallback
}

func CreateOrder(userName string, pkg Package) (*Order, error) {
    mu.Lock()
    defer mu.Unlock()
    
    floatAmount := nextFloatAmount(pkg.Amount)
    order := &Order{
        OutTradeNo:  fmt.Sprintf("V%d%d", time.Now().Unix(), floatAmount),
        UserName:    userName,
        Amount:      floatAmount,
        StreamNumber: pkg.StreamNumber,
        Status:      "pending",
        CreatedAt:   time.Now().Unix(),
    }
    return order, db.Insert(order)
}
```

### 回调匹配 (Go)

```go
func HandleCallback(amountCents int64, payType string) error {
    filter := bson.M{"amount": amountCents, "status": "pending"}
    update := bson.M{"$set": bson.M{"status": "processing"}}
    
    var order Order
    err := coll.FindOneAndUpdate(ctx, filter, update,
        options.FindOneAndUpdate().SetReturnDocument(options.Before),
    ).Decode(&order)
    if err != nil {
        return fmt.Errorf("未找到匹配订单: amount=%d", amountCents)
    }
    
    // 发货
    db.UpdateOne(User{}, bson.M{"user_name": order.UserName},
        bson.M{"$inc": bson.M{"stream_number": order.StreamNumber}})
    db.UpdateOne(Order{}, bson.M{"_id": order.ID},
        bson.M{"$set": bson.M{"status": "paid", "paid_at": time.Now().Unix()}})
    
    return nil
}
```

---

## 部署清单

| 组件 | 说明 |
|------|------|
| 安卓手机 | 安装 V免签 APK，已登录支付宝/微信 |
| 服务端 | 实现上述 7 个接口，需公网可达 |
| 收款码 | 静态图片放到服务端 `/qr/alipay.png` |
| nginx | 代理 `/appHeart`、`/appPush` 到后端 |
| 时钟同步 | APP 和服务端时间误差不宜过大（签名含时间戳） |
