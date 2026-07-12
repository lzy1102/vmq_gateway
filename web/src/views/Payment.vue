<template>
  <div class="payment-page">
    <div class="card">
      <h2>扫码支付</h2>
      <div class="pkg-name">{{ pkgName }}</div>
      <div class="amount">¥{{ amountDisplay }}</div>
      <div class="warning">
        请使用支付宝扫描下方二维码支付<br>
        务必支付<strong>精确金额</strong>，否则无法到账
      </div>
      <div class="qr-wrapper">
        <img class="qr-code" :src="qrUrl" alt="支付宝收款码" />
      </div>
      <div class="countdown">{{ countdownText }}</div>
      <div :class="['status', statusClass]">{{ statusText }}</div>
      <button v-if="!settled" class="cancel-btn" @click="handleCancel" :disabled="cancelling">
        {{ cancelling ? '取消中...' : '取消订单' }}
      </button>
      <div v-if="settled" class="btn-group">
        <button class="retry-btn" @click="reload">重新支付</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { queryOrderStatus } from '@/api'

const route = useRoute()

const orderId = ref(route.query.order_id as string || '')
const amountCents = ref(parseInt(route.query.amount as string) || 0)
const qrUrl = ref(route.query.qr as string || '/qr/alipay.png')

const amountDisplay = computed(() => (amountCents.value / 100).toFixed(2))

const remaining = ref(600)
const settled = ref(false)
const cancelling = ref(false)
const status = ref<'pending' | 'paid' | 'expired' | 'cancelled'>('pending')

const countdownText = computed(() => {
  const mins = Math.floor(remaining.value / 60)
  const secs = remaining.value % 60
  return String(mins).padStart(2, '0') + ':' + String(secs).padStart(2, '0')
})

const statusClass = computed(() => status.value)
const statusText = computed(() => {
  const map: Record<string, string> = {
    pending: '等待支付...',
    paid: '支付成功！',
    expired: '订单已过期',
    cancelled: '订单已取消'
  }
  return map[status.value]
})

let countdownTimer: ReturnType<typeof setInterval> | null = null
let pollTimer: ReturnType<typeof setInterval> | null = null

function updateStatus(s: string) {
  if (settled.value) return
  if (s === 'paid') {
    status.value = 'paid'
    settled.value = true
    stopTimers()
  } else if (s === 'cancelled') {
    status.value = 'cancelled'
    settled.value = true
    stopTimers()
  }
}

function handleCancel() {
  cancelling.value = true
  status.value = 'cancelled'
  settled.value = true
  stopTimers()
}

function reload() {
  window.location.reload()
}

function stopTimers() {
  if (countdownTimer) clearInterval(countdownTimer)
  if (pollTimer) clearInterval(pollTimer)
}

function startCountdown() {
  countdownTimer = setInterval(() => {
    if (settled.value) return
    if (remaining.value <= 0) {
      status.value = 'expired'
      settled.value = true
      stopTimers()
      return
    }
    remaining.value--
  }, 1000)
}

function startPolling() {
  if (!orderId.value) return
  pollTimer = setInterval(async () => {
    if (settled.value) return
    try {
      const resp = await queryOrderStatus(orderId.value)
      if (resp.code === 1 && resp.data) {
        updateStatus(resp.data.status)
      }
    } catch {
      // 忽略网络错误
    }
  }, 3000)
}

onMounted(() => {
  if (!tradeNo.value) return
  startCountdown()
  startPolling()
})

onUnmounted(() => {
  stopTimers()
})
</script>

<style scoped>
.payment-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.card {
  background: var(--card-bg);
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.08);
  padding: 32px 24px;
  max-width: 400px;
  width: 100%;
  text-align: center;
}

.card h2 {
  font-size: 20px;
  margin-bottom: 8px;
}

.pkg-name {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 20px;
}

.amount {
  font-size: 36px;
  font-weight: 700;
  color: var(--danger);
  margin-bottom: 8px;
  font-family: monospace;
}

.warning {
  font-size: 13px;
  color: var(--warning);
  background: #fef5e7;
  padding: 8px 12px;
  border-radius: 6px;
  margin-bottom: 24px;
  line-height: 1.5;
}

.qr-wrapper {
  margin-bottom: 20px;
}

.qr-code {
  max-width: 200px;
  width: 100%;
  border: 1px solid var(--border);
  border-radius: 8px;
}

.countdown {
  font-size: 28px;
  font-family: "Courier New", monospace;
  margin-bottom: 12px;
}

.status {
  font-size: 15px;
  font-weight: 500;
  margin-bottom: 20px;
  padding: 10px;
  border-radius: 6px;
}

.status.pending { color: var(--primary); background: #eaf2fd; }
.status.paid { color: var(--success); background: #eafaf1; }
.status.expired { color: #999; background: #f0f0f0; }
.status.cancelled { color: #999; background: #f0f0f0; }

.cancel-btn {
  background: var(--card-bg);
  color: #999;
  border: 1px solid var(--border-dark);
  padding: 10px 32px;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.2s;
}

.cancel-btn:hover:not(:disabled) {
  border-color: var(--danger);
  color: var(--danger);
}

.cancel-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-group {
  margin-top: 16px;
}

.retry-btn {
  background: var(--primary);
  color: #fff;
  border: none;
  padding: 12px 40px;
  border-radius: 6px;
  font-size: 15px;
}

.retry-btn:hover {
  background: var(--primary-dark);
}
</style>
