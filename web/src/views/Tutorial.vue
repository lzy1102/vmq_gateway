<template>
  <div class="tutorial">
    <div class="card">
      <div class="card-header">
        <h2>📚 接入教程</h2>
      </div>

      <div class="toc">
        <span class="toc-title">目录</span>
        <a href="#step1" @click.prevent="scrollTo('step1')">1. 创建服务</a>
        <a href="#step2" @click.prevent="scrollTo('step2')">2. 创建订单</a>
        <a href="#step3" @click.prevent="scrollTo('step3')">3. 收款通知</a>
        <a href="#step4" @click.prevent="scrollTo('step4')">4. 签名算法</a>
        <a href="#step5" @click.prevent="scrollTo('step5')">5. 完整示例</a>
      </div>

      <!-- 第一步 -->
      <section id="step1" class="section">
        <h3>第一步：创建服务绑定</h3>
        <p>进入 <strong>服务绑定</strong> 页面，点击添加，填写：</p>
        <ul>
          <li><strong>服务ID</strong>：你的应用标识，如 <code>my_shop</code></li>
          <li><strong>回调地址</strong>（可选）：用户付款后网关会 POST 到此地址。不填则用轮询模式</li>
        </ul>
        <p>创建后会自动生成 <strong>API Key</strong>，请妥善保存。</p>
        <div class="tip">
          💡 <strong>回调模式 vs 轮询模式</strong><br>
          回调模式：适合有公网地址的服务，实时性好<br>
          轮询模式：适合内网服务、小程序、定时任务等无法接收回调的场景
        </div>
      </section>

      <!-- 第二步 -->
      <section id="step2" class="section">
        <h3>第二步：创建订单</h3>
        <p>向网关发送 POST 请求创建支付订单：</p>
        <div class="code-block">
          <div class="code-header">
            <span>POST /api/order</span>
            <button class="copy-btn" @click="copyCode('create-order')">📋 复制</button>
          </div>
          <pre id="create-order"><code>{
  "amount": 100,
  "service_id": "my_shop",
  "api_key": "你的API Key",
  "callback_url": "https://你的服务/callback",
  "pay_type": "alipay"
}</code></pre>
        </div>

        <p class="param-title">请求参数</p>
        <table class="param-table">
          <thead>
            <tr><th>参数</th><th>必填</th><th>说明</th></tr>
          </thead>
          <tbody>
            <tr><td><code>amount</code></td><td>✅</td><td>金额，单位<strong>分</strong>（100 = 1元）</td></tr>
            <tr><td><code>service_id</code></td><td>✅</td><td>服务ID，绑定时创建的</td></tr>
            <tr><td><code>api_key</code></td><td>✅</td><td>绑定时生成的 API Key</td></tr>
            <tr><td><code>callback_url</code></td><td>❌</td><td>回调地址，不填则轮询模式</td></tr>
            <tr><td><code>pay_type</code></td><td>❌</td><td>wechat 或 alipay，默认 alipay</td></tr>
          </tbody>
        </table>

        <p class="param-title">返回示例</p>
        <div class="code-block">
          <div class="code-header"><span>200 OK</span></div>
          <pre><code>{
  "code": 1,
  "data": {
    "order_id": "V1783926073_103",
    "request_amount": 100,
    "request_str": "1.00",
    "pay_amount": 103,
    "pay_str": "1.03",
    "device_id": "phone_01",
    "qr_url": "/qr/phone_01_alipay.jpg",
    "expire_at": 1783926973,
    "remaining_seconds": 900
  }
}</code></pre>
        </div>

        <div class="tip">
          💡 <strong>金额说明</strong><br>
          由于 V免签只能通过金额匹配订单，实际支付金额会在你请求的基础上加 1~19 分浮动。<br>
          <code>request_amount</code> 是你请求的金额，<code>pay_amount</code> 是用户实际需要支付的金额。
        </div>
      </section>

      <!-- 第三步 -->
      <section id="step3" class="section">
        <h3>第三步：接收收款通知</h3>

        <h4>方式A：回调模式</h4>
        <p>网关会 POST JSON 到你填写的回调地址：</p>
        <div class="code-block">
          <div class="code-header"><span>POST 你的回调地址</span></div>
          <pre><code>{
  "trade_no": "V1783926073_103",
  "amount": 103,
  "pay_type": "alipay",
  "paid_at": 1783926200,
  "service_id": "my_shop"
}</code></pre>
        </div>
        <p>收到后返回 <code>{"code":1,"msg":"ok"}</code> 表示处理成功。</p>

        <h4>方式B：轮询模式</h4>
        <p>主动查询订单状态：</p>
        <div class="code-block">
          <div class="code-header"><span>GET /api/order/status?order_id=V1783926073_103</span></div>
          <pre><code>{
  "code": 1,
  "data": {
    "order_id": "V1783926073_103",
    "amount": 103,
    "status": "paid",
    "paid_at": 1783926200
  }
}</code></pre>
        </div>
        <p><code>status</code> 可能的值：<code>pending</code>（待支付）、<code>paid</code>（已支付）、<code>expired</code>（已过期）、<code>cancelled</code>（已取消）</p>
      </section>

      <!-- 第四步 -->
      <section id="step4" class="section">
        <h3>第四步：签名算法</h3>
        <p>如果你要接入 APP 端（自己开发 APP 监听通知），需要了解签名：</p>

        <h4>心跳签名</h4>
        <div class="code-block">
          <div class="code-header"><span>sign = MD5(t + key)</span></div>
          <pre><code># Python
import hashlib
sign = hashlib.md5((t + key).encode()).hexdigest()

# JavaScript
const crypto = require('crypto')
const sign = crypto.createHash('md5').update(t + key).digest('hex')

// Go
sign := fmt.Sprintf("%x", md5.Sum([]byte(t + key)))</code></pre>
        </div>

        <h4>收款推送签名</h4>
        <div class="code-block">
          <div class="code-header"><span>sign = MD5(type + price + t + key)</span></div>
          <pre><code># Python
sign = hashlib.md5((pay_type + price + t + key).encode()).hexdigest()

# JavaScript
const sign = crypto.createHash('md5').update(pay_type + price + t + key).digest('hex')</code></pre>
        </div>
      </section>

      <!-- 第五步 -->
      <section id="step5" class="section">
        <h3>第五步：完整代码示例</h3>
        <p>完整示例代码见 GitHub 仓库 <code>examples/</code> 目录：</p>
        <div class="example-grid">
          <div class="example-card" v-for="lang in examples" :key="lang.name">
            <div class="example-icon">{{ lang.icon }}</div>
            <div class="example-info">
              <strong>{{ lang.name }}</strong>
              <span>{{ lang.desc }}</span>
              <code class="example-file">{{ lang.file }}</code>
            </div>
          </div>
        </div>

        <h4>最简 Python 示例</h4>
        <div class="code-block">
          <div class="code-header"><span>quick_start.py</span>
            <button class="copy-btn" @click="copyCode('python-quick')">📋 复制</button>
          </div>
          <pre id="python-quick"><code>import requests

resp = requests.post("http://你的网关地址/api/order", json={
    "amount": 100,           # 1 元
    "service_id": "my_shop",
    "api_key": "你的API Key",
    "pay_type": "alipay"
})
order = resp.json()["data"]
print(f"请支付 {order['pay_str']} 元")
print(f"二维码: {order['qr_url']}")

# 轮询等待支付
import time
while True:
    status = requests.get(
        f"http://你的网关地址/api/order/status",
        params={"order_id": order["order_id"]}
    ).json()["data"]["status"]
    if status == "paid":
        print("✅ 收款成功")
        break
    elif status in ("expired", "cancelled"):
        print("❌ 订单已失效")
        break
    time.sleep(3)</code></pre>
        </div>

        <h4>最简 Node.js 示例</h4>
        <div class="code-block">
          <div class="code-header"><span>quick_start.js</span>
            <button class="copy-btn" @click="copyCode('node-quick')">📋 复制</button>
          </div>
          <pre id="node-quick"><code>import http from 'node:http'

const GATEWAY = 'http://你的网关地址'

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
}</code></pre>
        </div>
      </section>

      <!-- FAQ -->
      <section class="section">
        <h3>❓ 常见问题</h3>
        <div class="faq-list">
          <div class="faq-item" v-for="faq in faqs" :key="faq.q">
            <div class="faq-q" @click="faq.open = !faq.open">
              <span>{{ faq.q }}</span>
              <span class="faq-toggle">{{ faq.open ? '−' : '+' }}</span>
            </div>
            <div v-if="faq.open" class="faq-a" v-html="faq.a"></div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'

const examples = [
  { name: 'Python', icon: '🐍', desc: 'Flask 回调 + 轮询', file: 'examples/python/example.py' },
  { name: 'Node.js', icon: '🟢', desc: 'Express 回调 + 轮询', file: 'examples/node/poll.js' },
  { name: 'Go', icon: '🔵', desc: '两种模式完整示例', file: 'examples/go/main.go' },
]

const faqs = reactive([
  {
    q: '为什么实际支付金额和我设置的不一样？',
    a: 'V免签通过收款金额匹配订单，所以会在你请求的金额上加 1~19 分浮动，确保同一时刻没有相同金额的订单。这是正常现象，浮动金额会在订单过期后释放。',
    open: false,
  },
  {
    q: '订单多长时间过期？',
    a: '默认 15 分钟。超时后订单自动标记为 expired，浮动金额释放。',
    open: false,
  },
  {
    q: '一个服务可以绑定多个设备吗？',
    a: '可以。通过支付池管理设备，一个池里可以有多台手机，增加并发处理能力。',
    open: false,
  },
  {
    q: '回调没收到怎么办？',
    a: '1. 检查回调地址是否可公网访问\n2. 检查服务器防火墙\n3. 在「订单管理」查看订单状态，也可以用轮询接口查询\n4. 回调失败不会重试，建议用轮询作为兜底',
    open: false,
  },
  {
    q: '支持哪些支付方式？',
    a: '支持支付宝和微信支付。创建订单时通过 pay_type 参数指定（alipay/wechat）。',
    open: false,
  },
])

function scrollTo(id: string) {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' })
}

function copyCode(id: string) {
  const el = document.getElementById(id)
  if (el) {
    navigator.clipboard?.writeText(el.textContent || '')
  }
}
</script>

<style scoped>
.tutorial {
  max-width: 900px;
}

.card {
  background: var(--bg-card);
  border-radius: 12px;
  border: 1px solid var(--border);
  padding: 24px;
}

.card-header h2 {
  margin: 0 0 20px;
}

.toc {
  display: flex;
  gap: 16px;
  padding: 12px 16px;
  background: #f8f9fa;
  border-radius: 8px;
  margin-bottom: 24px;
  flex-wrap: wrap;
  align-items: center;
}

.toc-title {
  font-weight: 600;
  color: var(--text-secondary);
  font-size: 13px;
}

.toc a {
  color: var(--primary);
  text-decoration: none;
  font-size: 13px;
}

.toc a:hover {
  text-decoration: underline;
}

.section {
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--border);
}

.section:last-child {
  border-bottom: none;
}

.section h3 {
  margin: 0 0 16px;
  font-size: 18px;
}

.section h4 {
  margin: 16px 0 8px;
  font-size: 15px;
  color: var(--text-primary);
}

.section p {
  color: var(--text-secondary);
  line-height: 1.6;
  margin: 8px 0;
}

.section ul {
  padding-left: 20px;
  color: var(--text-secondary);
  line-height: 1.8;
}

.section code {
  background: #f1f3f5;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
  color: #e83e8c;
}

.tip {
  padding: 12px 16px;
  background: #eff6ff;
  border-left: 3px solid #3b82f6;
  border-radius: 4px;
  font-size: 13px;
  line-height: 1.6;
  margin: 12px 0;
  color: #1e40af;
}

.code-block {
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
  margin: 12px 0;
}

.code-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #f8f9fa;
  border-bottom: 1px solid var(--border);
  font-size: 12px;
  color: var(--text-secondary);
}

.copy-btn {
  background: none;
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 2px 8px;
  font-size: 12px;
  cursor: pointer;
}

.copy-btn:hover {
  background: #e9ecef;
}

.code-block pre {
  margin: 0;
  padding: 16px;
  overflow-x: auto;
  font-size: 13px;
  line-height: 1.5;
}

.code-block code {
  background: none;
  padding: 0;
  color: inherit;
  font-family: 'Fira Code', 'Consolas', monospace;
}

.param-title {
  font-weight: 600;
  color: var(--text-primary) !important;
  margin-top: 16px !important;
}

.param-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  margin: 8px 0;
}

.param-table th, .param-table td {
  padding: 8px 12px;
  border: 1px solid var(--border);
  text-align: left;
}

.param-table th {
  background: #f8f9fa;
  font-weight: 500;
}

.example-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
  margin: 12px 0;
}

.example-card {
  display: flex;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--border);
  border-radius: 8px;
}

.example-icon {
  font-size: 28px;
}

.example-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.example-info strong {
  font-size: 14px;
}

.example-info span {
  font-size: 12px;
  color: var(--text-secondary);
}

.example-file {
  font-size: 11px !important;
  color: var(--text-tertiary) !important;
}

.faq-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.faq-item {
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}

.faq-q {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  background: #f8f9fa;
}

.faq-q:hover {
  background: #e9ecef;
}

.faq-toggle {
  font-size: 18px;
  color: var(--text-secondary);
}

.faq-a {
  padding: 12px 16px;
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
  white-space: pre-line;
}
</style>
