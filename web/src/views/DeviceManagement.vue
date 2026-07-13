<template>
  <div class="management-page">
    <div class="page-header">
      <h1>设备管理</h1>
      <p class="description">管理收款监控设备，配置V免签APP连接</p>
    </div>

    <!-- 添加设备表单 -->
    <div class="card">
      <h2>添加设备</h2>
      <div class="form-row">
        <div class="form-group">
          <label>设备ID</label>
          <input v-model="newDeviceId" placeholder="例如：phone_01" />
        </div>
        <button class="btn-primary" @click="handleAddDevice" :disabled="!newDeviceId.trim()">
          添加设备
        </button>
      </div>
    </div>

    <!-- 设备列表 -->
    <div class="card">
      <div class="card-header">
        <h2>设备列表</h2>
        <span class="total-count">共 {{ total }} 台设备</span>
      </div>

      <!-- 搜索和筛选 -->
      <div class="toolbar">
        <div class="search-box">
          <input v-model="searchKeyword" placeholder="搜索设备ID或Key..." @keyup.enter="loadDevices" />
        </div>
        <div class="filter-group">
          <select v-model="statusFilter">
            <option value="">全部状态</option>
            <option value="online">在线</option>
            <option value="offline">离线</option>
          </select>
        </div>
        <button class="btn-secondary" @click="loadDevices">🔄 刷新</button>
      </div>

      <div v-if="loading" class="loading">加载中...</div>
      <div v-else-if="devices.length === 0" class="empty-state">
        <p v-if="searchKeyword">没有找到匹配的设备</p>
        <p v-else>暂无设备</p>
        <p class="hint" v-if="!searchKeyword">请先添加设备，然后将Key填入V免签APP</p>
      </div>
      <div v-else class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>设备ID</th>
              <th>VMQ Key</th>
              <th>状态</th>
              <th>上次心跳</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="d in devices" :key="d.device_id">
              <td class="device-id">{{ d.device_id }}</td>
              <td>
                <code class="vmq-key" @click="copyKey(d.key)" :title="d.key">
                  {{ d.key }}
                </code>
                <span class="copy-hint">点击复制</span>
              </td>
              <td>
                <span :class="['status-badge', d.status]">
                  {{ d.status === 'online' ? '在线' : '离线' }}
                </span>
              </td>
              <td>{{ d.last_heartbeat ? formatTime(d.last_heartbeat) : '从未' }}</td>
              <td class="actions">
                <button class="btn-small btn-qr btn-wechat" @click="triggerQRUpload(d.device_id, 'wechat')" title="上传微信收款码">
                  <img src="/icons/wechat.svg" alt="微信" width="16" height="16" />
                </button>
                <button class="btn-small btn-qr btn-alipay" @click="triggerQRUpload(d.device_id, 'alipay')" title="上传支付宝收款码">
                  <img src="/icons/alipay.svg" alt="支付宝" width="16" height="16" />
                </button>
                <button class="btn-small btn-warning" @click="handleRegenerateKey(d)" title="重新生成Key">
                  🔄
                </button>
                <button class="btn-small btn-danger" @click="handleDeleteDevice(d)" title="删除设备">
                  🗑️
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- 分页 -->
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

    <!-- 使用说明 -->
    <div class="card info-card">
      <h2>使用说明</h2>
      <ol class="instructions">
        <li>添加设备后，复制生成的 <strong>VMQ Key</strong></li>
        <li>在V免签APP中填入服务器地址：<code>{{ serverAddr }}</code></li>
        <li>在V免签APP中填入复制的 <strong>VMQ Key</strong></li>
        <li>APP启动后会自动发送心跳，状态将变为"在线"</li>
      </ol>
    </div>

    <input ref="fileInput" type="file" accept=".png,.jpg,.jpeg" style="display:none" @change="handleQRUpload" />

    <!-- Toast -->
    <div v-if="toastMsg" class="toast" :class="toastType">{{ toastMsg }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import type { Device } from '@/types'
import { listDevices, addDevice, deleteDevice, updateDevice, uploadQRCode } from '@/api'

const devices = ref<Device[]>([])
const total = ref(0)
const totalPages = ref(0)
const newDeviceId = ref('')
const loading = ref(true)
const toastMsg = ref('')
const toastType = ref<'success' | 'error'>('success')

// 搜索和筛选
const searchKeyword = ref('')
const statusFilter = ref('')

// 分页
const currentPage = ref(1)
const pageSize = ref(10)

const serverAddr = window.location.hostname + ':8080'
const fileInput = ref<HTMLInputElement>()
const uploadingDeviceId = ref('')
const uploadingType = ref<'wechat' | 'alipay'>('wechat')

let toastTimer: ReturnType<typeof setTimeout> | null = null

function toast(msg: string, type: 'success' | 'error' = 'success') {
  toastMsg.value = msg
  toastType.value = type
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastMsg.value = '' }, 3000)
}

function formatTime(ts: number) {
  const date = new Date(ts * 1000)
  const now = new Date()
  const diff = (now.getTime() - date.getTime()) / 1000
  
  if (diff < 60) return '刚刚'
  if (diff < 3600) return Math.floor(diff / 60) + '分钟前'
  if (diff < 86400) return Math.floor(diff / 3600) + '小时前'
  return date.toLocaleDateString()
}

function copyKey(key: string) {
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(key).then(
      () => toast('已复制到剪贴板'),
      () => fallbackCopy(key)
    )
  } else {
    fallbackCopy(key)
  }
}

function fallbackCopy(text: string) {
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  document.body.appendChild(textarea)
  textarea.select()
  try {
    document.execCommand('copy')
    toast('已复制到剪贴板')
  } catch {
    toast('复制失败: ' + text, 'error')
  }
  document.body.removeChild(textarea)
}

// 过滤后的设备列表（状态过滤在前端做）
const filteredDevices = computed(() => {
  if (!statusFilter.value) return devices.value
  return devices.value.filter(d => d.status === statusFilter.value)
})

async function loadDevices() {
  loading.value = true
  try {
    const resp = await listDevices({
      keyword: searchKeyword.value,
      page: currentPage.value,
      page_size: pageSize.value
    })
    if (resp.code === 1 && resp.data) {
      devices.value = resp.data.items
      total.value = resp.data.total
      totalPages.value = resp.data.total_pages
    }
  } finally {
    loading.value = false
  }
}

watch([searchKeyword, pageSize], () => {
  currentPage.value = 1
  loadDevices()
})

watch(currentPage, () => {
  loadDevices()
})

async function handleAddDevice() {
  const id = newDeviceId.value.trim()
  if (!id) return
  
  const resp = await addDevice(id)
  if (resp.code === 1 && resp.data) {
    toast('设备已添加，Key: ' + resp.data.key)
    newDeviceId.value = ''
    loadDevices()
  } else {
    toast(resp.msg || '添加失败', 'error')
  }
}

async function handleDeleteDevice(device: Device) {
  if (!confirm(`确定要删除设备 "${device.device_id}" 吗？`)) return
  
  const resp = await deleteDevice(device.device_id)
  if (resp.code === 1) {
    toast('设备已删除')
    loadDevices()
  } else {
    toast(resp.msg || '删除失败', 'error')
  }
}

async function handleRegenerateKey(device: Device) {
  if (!confirm(`确定要重新生成设备 "${device.device_id}" 的Key吗？旧Key将失效。`)) return
  
  const newKey = Math.random().toString(36).substring(2) + Date.now().toString(36)
  const resp = await updateDevice(device.device_id, newKey)
  if (resp.code === 1) {
    toast('Key已更新: ' + newKey)
    loadDevices()
  } else {
    toast(resp.msg || '更新失败', 'error')
  }
}

function triggerQRUpload(deviceId: string, type: 'wechat' | 'alipay') {
  uploadingDeviceId.value = deviceId
  uploadingType.value = type
  fileInput.value?.click()
}

async function handleQRUpload(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file || !uploadingDeviceId.value) return

  const resp = await uploadQRCode(uploadingDeviceId.value, uploadingType.value, file)
  if (resp.code === 1) {
    toast('收款码已上传')
    loadDevices()
  } else {
    toast(resp.msg || '上传失败', 'error')
  }
  input.value = ''
  uploadingDeviceId.value = ''
}

onMounted(loadDevices)
</script>

<style scoped>
.management-page {
  max-width: 1000px;
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

.form-row {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.form-group {
  flex: 1;
}

.form-group label {
  display: block;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.form-group input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border-dark);
  border-radius: 6px;
  font-size: 14px;
}

.form-group input:focus {
  outline: none;
  border-color: var(--primary);
}

.btn-primary {
  padding: 10px 20px;
  background: var(--primary);
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  white-space: nowrap;
}

.btn-primary:hover:not(:disabled) {
  background: var(--primary-dark);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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

.empty-state p {
  margin-bottom: 8px;
}

.hint {
  font-size: 13px;
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

.device-id {
  font-weight: 500;
}

.vmq-key {
  display: inline-block;
  padding: 4px 8px;
  background: #f8f9fa;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  transition: background 0.2s;
}

.vmq-key:hover {
  background: #e9ecef;
}

.copy-hint {
  font-size: 11px;
  color: var(--text-secondary);
  margin-left: 8px;
}

.status-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.status-badge.online {
  background: #d4edda;
  color: #155724;
}

.status-badge.offline {
  background: #f8d7da;
  color: #721c24;
}

.actions {
  display: flex;
  gap: 8px;
}

.btn-small {
  padding: 6px 10px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: opacity 0.2s;
}

.btn-small:hover {
  opacity: 0.8;
}

.btn-warning {
  background: #ffc107;
  color: #212529;
}

.btn-danger {
  background: var(--danger);
  color: #fff;
}

.btn-info {
  background: #17a2b8;
  color: #fff;
}

.btn-qr {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  background: transparent;
  border: 1px solid var(--border-dark);
}

.btn-wechat {
  color: #07c160;
}

.btn-alipay {
  color: #1677ff;
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

.info-card {
  background: #f8f9fa;
}

.instructions {
  padding-left: 20px;
  margin: 0;
}

.instructions li {
  margin-bottom: 8px;
  line-height: 1.6;
}

.instructions code {
  padding: 2px 6px;
  background: #e9ecef;
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
  animation: slideIn 0.3s ease;
}

.toast.success {
  background: var(--success);
}

.toast.error {
  background: var(--danger);
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}
</style>
