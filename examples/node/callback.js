/**
 * Node.js 回调模式示例 - V免签支付网关
 *
 * 用法: node callback.js
 * 然后在管理后台绑定回调地址为 http://你的IP:3000/callback
 */

import express from 'express'

// ========== 配置 ==========
const PORT = 3000

const app = express()
app.use(express.json())

// ========== 回调接口 ==========
app.post('/callback', (req, res) => {
  /*
   * 回调 POST body:
   * {
   *   "trade_no": "V1783926073_103",
   *   "amount": 103,
   *   "pay_type": "alipay",
   *   "paid_at": 1783926200,
   *   "service_id": "my_service"
   * }
   */
  const { trade_no, amount, pay_type, paid_at, service_id } = req.body

  console.log(`收到回调:`)
  console.log(`  订单号: ${trade_no}`)
  console.log(`  金额: ${amount / 100} 元`)
  console.log(`  支付方式: ${pay_type}`)
  console.log(`  服务: ${service_id}`)

  // TODO: 在这里处理业务逻辑
  // - 更新订单状态
  // - 发货/开通服务
  // - 记录日志

  res.json({ code: 1, msg: 'ok' })
})

// ========== 健康检查 ==========
app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: Date.now() })
})

// ========== 启动 ==========
app.listen(PORT, () => {
  console.log(`回调服务已启动: http://localhost:${PORT}`)
  console.log(`回调地址: http://你的公网IP:${PORT}/callback`)
})
