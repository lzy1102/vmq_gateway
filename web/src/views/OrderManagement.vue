<template>
  <div class="management-page">
    <div class="page-header">
      <h1>订单管理</h1>
      <p class="description">查看和管理所有支付订单</p>
    </div>

    <div class="card">
      <div class="card-header">
        <h2>订单列表</h2>
        <span class="total-count">共 {{ total }} 个订单</span>
      </div>

      <div class="toolbar">
        <div class="search-box">
          <input v-model="searchKeyword" placeholder="搜索订单号或设备ID..." @keyup.enter="loadOrders" />
        </div>
        <div class="filter-group">
          <select v-model="statusFilter">
            <option value="">全部状态</option>
            <option value="pending">待支付</option>
            <option value="paid">已支付</option>
            <option value="expired">已过期</option>
            <option value="cancelled">已取消</option>
          </select>
        </div>
        <button class="btn-secondary" @click="loadOrders">🔄 刷新</button>
      </div>

      <div v-if="loading" class="loading">加载中...</div>
      <div v-else-if="orders.length === 0" class="empty-state">
        <p v-if="searchKeyword || statusFilter">没有找到匹配的订单</p>
        <p v-else>暂无订单</p>
      </div>
      <div v-else class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>订单号</th>
              <th>申请金额</th>
              <th>支付金额</th>
              <th>状态</th>
              <th>设备</th>
              <th>创建时间</th>
              <th>有效期</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="o in orders" :key="o.trade_no">
              <td class="order-id">
                <code @click="copyKey(o.trade_no)" :title="o.trade_no">{{ o.trade_no }}</code>
              </td>
              <td>{{ formatAmount(o.amount) }}</td>
              <td>{{ formatAmount(o.amount) }}</td>
              <td>
                <span :class="['status-badge', o.status]">{{ statusText(o.status) }}</span>
              </td>
              <td>{{ o.device_id || '-' }}</td>
              <td>{{ formatTime(o.created_at) }}</td>
              <td>
                <span v-if="o.status === 'pending'" class="countdown">
                  {{ formatCountdown(o.expire_at) }}
                </span>
                <span v-else-if="o.status === 'paid'">{{ formatTime(o.paid_at) }}</span>
                <span v-else>-</span>
              </td>
            </tr>
          </tbody>
        </table>

        <div class="pagination" v-if="totalPages > 1">
          <button :disabled="currentPage === 1" @click="currentPage = 1">首页</button>
          <button :disabled="currentPage === 1" @click="currentPage--">上一页</button>
          <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
          <button :disabled="currentPage === totalPages" @click="currentPage++">下一页</button>
          <button :disabled="currentPage === totalPages" @click="currentPage = totalPages">末页</button>
          <select v-model.number="pageSize" class="page-size">
            <option :value="10">10条/页</option>
            <option :value="20">20条/页</option>
            <option :value="50">50条/页</option>
          </select>
        </div>
      </div>
    </div>

    <div v-if="toastMsg" class="toast" :class="toastType">{{ toastMsg }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { listOrders } from '@/api'
import type { Order } from '@/api'

const orders = ref<Order[]>([])
const total = ref(0)
const totalPages = ref(0)
const loading = ref(true)
const toastMsg = ref('')
const toastType = ref<'success' | 'error'>('success')

const searchKeyword = ref('')
const statusFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(10)

let toastTimer: ReturnType<typeof setTimeout> | null = null

function toast(msg: string, type: 'success' | 'error' = 'success') {
  toastMsg.value = msg
  toastType.value = type
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastMsg.value = '' }, 3000)
}

function copyKey(key: string) {
  navigator.clipboard?.writeText(key).then(
    () => toast('已复制'),
    () => {}
  )
}

function formatAmount(cents: number) {
  return (cents / 100).toFixed(2) + ' 元'
}

function formatTime(ts: number) {
  if (!ts) return '-'
  const d = new Date(ts * 1000)
  return d.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function formatCountdown(expireAt: number) {
  const remaining = expireAt - Math.floor(Date.now() / 1000)
  if (remaining <= 0) return '已过期'
  const m = Math.floor(remaining / 60)
  const s = remaining % 60
  return `${m}分${s < 10 ? '0' : ''}${s}秒`
}

function statusText(s: string) {
  const map: Record<string, string> = { pending: '待支付', paid: '已支付', expired: '已过期', cancelled: '已取消' }
  return map[s] || s
}

async function loadOrders() {
  loading.value = true
  try {
    const resp = await listOrders({
      keyword: searchKeyword.value,
      status: statusFilter.value,
      page: currentPage.value,
      page_size: pageSize.value
    })
    if (resp.code === 1 && resp.data) {
      orders.value = resp.data.items || []
      total.value = resp.data.total
      totalPages.value = resp.data.total_pages
    }
  } finally {
    loading.value = false
  }
}

watch([searchKeyword, statusFilter, pageSize], () => {
  currentPage.value = 1
  loadOrders()
})

watch(currentPage, () => {
  loadOrders()
})

onMounted(loadOrders)
</script>

<style scoped>
.management-page {
  max-width: 1200px;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 24px;
  margin-bottom: 8px;
}

.description {
  color: var(--text-secondary);
  font-size: 14px;
}

.card {
  background: var(--card-bg);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.card h2 {
  font-size: 16px;
  margin-bottom: 16px;
  color: var(--text);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.card-header h2 {
  margin-bottom: 0;
}

.total-count {
  font-size: 13px;
  color: var(--text-secondary);
}

.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.search-box {
  flex: 1;
  min-width: 200px;
}

.search-box input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border-dark);
  border-radius: 6px;
  font-size: 14px;
}

.filter-group select {
  padding: 10px 12px;
  border: 1px solid var(--border-dark);
  border-radius: 6px;
  font-size: 14px;
  background: #fff;
}

.btn-secondary {
  padding: 10px 16px;
  background: #f8f9fa;
  border: 1px solid var(--border-dark);
  border-radius: 6px;
  font-size: 14px;
  white-space: nowrap;
}

.btn-secondary:hover {
  background: #e9ecef;
}

.loading {
  padding: 40px;
  text-align: center;
  color: var(--text-secondary);
}

.empty-state {
  padding: 40px;
  text-align: center;
  color: var(--text-secondary);
}

.table-wrap {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th, td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid var(--border);
}

th {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 500;
}

.order-id code {
  display: inline-block;
  padding: 4px 8px;
  background: #f8f9fa;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
}

.order-id code:hover {
  background: #e9ecef;
}

.status-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.status-badge.pending {
  background: #fff3cd;
  color: #856404;
}

.status-badge.paid {
  background: #d4edda;
  color: #155724;
}

.status-badge.expired {
  background: #f8d7da;
  color: #721c24;
}

.status-badge.cancelled {
  background: #e2e3e5;
  color: #383d41;
}

.countdown {
  color: #e67e22;
  font-weight: 500;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--border);
}

.pagination button {
  padding: 8px 12px;
  border: 1px solid var(--border-dark);
  border-radius: 4px;
  background: #fff;
  cursor: pointer;
  font-size: 13px;
}

.pagination button:hover:not(:disabled) {
  background: #f8f9fa;
}

.pagination button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  font-size: 13px;
  color: var(--text-secondary);
  padding: 0 8px;
}

.page-size {
  padding: 8px;
  border: 1px solid var(--border-dark);
  border-radius: 4px;
  font-size: 13px;
}

.toast {
  position: fixed;
  top: 24px;
  right: 24px;
  padding: 12px 20px;
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  z-index: 1000;
}

.toast.success {
  background: var(--success);
}

.toast.error {
  background: var(--danger);
}
</style>
