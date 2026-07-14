/**
 * Node.js 轮询模式示例 - V免签支付网关
 *
 * 用法: node poll.js
 */

import http from 'node:http'

// ========== 配置 ==========
const GATEWAY_URL = 'http://186.241.107.44'
const SERVICE_ID = 'my_node_service'   // 你的服务ID
const API_KEY = 'xxxxxxxxxxxx'         // 绑定时生成的 API Key

// ========== HTTP 请求封装 ==========
function request(method, path, body) {
  return new Promise((resolve, reject) => {
    const url = new URL(path, GATEWAY_URL)
    const options = {
      hostname: url.hostname,
      port: url.port,
      path: url.pathname + url.search,
      method,
      headers: { 'Content-Type': 'application/json' },
    }

    const req = http.request(options, (res) => {
      let data = ''
      res.on('data', (chunk) => (data += chunk))
      res.on('end', () => resolve(JSON.parse(data)))
    })
    req.on('error', reject)
    if (body) req.write(JSON.stringify(body))
    req.end()
  })
}

// ========== 1. 创建订单 ==========
async function createOrder(amount, payType = 'alipay', callbackUrl = '') {
  const payload = {
    amount,
    service_id: SERVICE_ID,
    api_key: API_KEY,
    pay_type: payType,
  }
  if (callbackUrl) payload.callback_url = callbackUrl

  const resp = await request('POST', '/api/order', payload)
  if (resp.code === 1) {
    const order = resp.data
    console.log(`订单创建成功: ${order.order_id}`)
    console.log(`  实付金额: ${order.pay_str} 元`)
    console.log(`  二维码: ${order.qr_url}`)
    console.log(`  有效期: ${order.remaining_seconds} 秒`)
    return order
  }
  console.error(`创建失败: ${resp.msg}`)
  return null
}

// ========== 2. 查询订单状态 ==========
async function queryOrder(orderId) {
  const resp = await request('GET', `/api/order/status?order_id=${orderId}`)
  return resp.code === 1 ? resp.data : null
}

// ========== 3. 轮询等待支付 ==========
async function pollOrder(orderId, timeout = 900) {
  const start = Date.now()
  console.log(`开始轮询订单 ${orderId}...`)

  while (Date.now() - start < timeout * 1000) {
    const result = await queryOrder(orderId)
    if (result) {
      if (result.status === 'paid') {
        console.log(`✅ 收款成功！金额: ${result.amount / 100} 元`)
        return 'paid'
      }
      if (result.status === 'expired' || result.status === 'cancelled') {
        console.log(`❌ 订单已${result.status}`)
        return result.status
      }
    }
    await new Promise((r) => setTimeout(r, 3000))
  }

  console.log('⏰ 轮询超时')
  return 'timeout'
}

// ========== 主流程 ==========
async function main() {
  // 创建 1 元订单（轮询模式）
  const order = await createOrder(100, 'alipay')
  if (order) {
    await pollOrder(order.order_id)
  }
}

main().catch(console.error)
