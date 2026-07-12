<template>
  <div class="management-page">
    <div class="page-header">
      <h1>支付池管理</h1>
      <p class="description">将多个设备组成支付池，实现负载均衡</p>
    </div>

    <!-- 创建支付池 -->
    <div class="card">
      <h2>创建支付池</h2>
      <div class="form-row">
        <div class="form-group">
          <label>池ID</label>
          <input v-model="newPoolId" placeholder="例如：pool_main" />
        </div>
        <div class="form-group">
          <label>池名称</label>
          <input v-model="newPoolName" placeholder="例如：主支付池" />
        </div>
      </div>
      <div class="form-row">
        <div class="form-group" style="flex: 1">
          <label>选择设备（可选，可多选）</label>
          <div class="device-checkboxes" v-if="devices.length > 0">
            <label v-for="d in devices" :key="d.device_id" class="checkbox-label">
              <input type="checkbox" :value="d.device_id" v-model="selectedDeviceIds" />
              {{ d.device_id }}
            </label>
          </div>
          <p v-else class="hint">暂无设备，请先在设备管理中添加</p>
        </div>
      </div>
      <div class="form-actions">
        <button class="btn-primary" @click="handleAddPool" :disabled="!newPoolId.trim() || !newPoolName.trim()">
          创建池{{ selectedDeviceIds.length > 0 ? '并添加 ' + selectedDeviceIds.length + ' 台设备' : '' }}
        </button>
      </div>
    </div>

    <!-- 支付池列表 -->
    <div class="card">
      <div class="card-header">
        <h2>支付池列表</h2>
        <span class="total-count">共 {{ total }} 个池</span>
      </div>

      <!-- 搜索 -->
      <div class="toolbar">
        <div class="search-box">
          <input v-model="searchKeyword" placeholder="搜索池ID或名称..." @keyup.enter="loadData" />
        </div>
        <button class="btn-secondary" @click="loadData">🔄 刷新</button>
      </div>

      <div v-if="loading" class="loading">加载中...</div>
      <div v-else-if="pools.length === 0" class="empty-state">
        <p v-if="searchKeyword">没有找到匹配的支付池</p>
        <p v-else>暂无支付池</p>
        <p class="hint" v-if="!searchKeyword">创建支付池后，将设备添加到池中实现负载均衡</p>
      </div>
      <div v-else class="pool-list">
        <div v-for="p in pools" :key="p.pool_id" class="pool-item">
          <div class="pool-header">
            <span class="pool-name">{{ p.name }}</span>
            <span class="pool-id">{{ p.pool_id }}</span>
            <button class="delete-pool-btn" @click="handleDeletePool(p.pool_id)" title="删除支付池">🗑️</button>
          </div>
          <div class="pool-devices">
            <span v-if="!p.device_ids || p.device_ids.length === 0" class="no-devices">暂无设备</span>
            <span v-else v-for="did in p.device_ids" :key="did" class="device-tag">
              {{ did }}
              <button class="remove-btn" @click="handleRemoveDeviceFromPool(p.pool_id, did)">×</button>
            </span>
          </div>
        </div>

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
        <li>创建支付池，为池命名</li>
        <li>将已添加的设备分配到池中</li>
        <li>在"服务绑定"中将服务绑定到此池</li>
        <li>支付请求将在池内设备间轮询分配</li>
      </ol>
    </div>

    <!-- Toast -->
    <div v-if="toastMsg" class="toast" :class="toastType">{{ toastMsg }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import type { Pool, Device } from '@/types'
import { listPools, addPool, addDeviceToPool, removeDeviceFromPool, deletePool, listDevices } from '@/api'

const pools = ref<Pool[]>([])
const devices = ref<Device[]>([])
const loading = ref(true)
const toastMsg = ref('')
const toastType = ref<'success' | 'error'>('success')
const total = ref(0)
const totalPages = ref(0)

const newPoolId = ref('')
const newPoolName = ref('')
const selectedDeviceIds = ref<string[]>([])

// 搜索
const searchKeyword = ref('')

// 分页
const currentPage = ref(1)
const pageSize = ref(10)

let toastTimer: ReturnType<typeof setTimeout> | null = null

function toast(msg: string, type: 'success' | 'error' = 'success') {
  toastMsg.value = msg
  toastType.value = type
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastMsg.value = '' }, 3000)
}

watch([searchKeyword, pageSize], () => {
  currentPage.value = 1
  loadData()
})

watch(currentPage, () => {
  loadData()
})

async function loadData() {
  loading.value = true
  try {
    const [poolResp, deviceResp] = await Promise.all([
      listPools({ keyword: searchKeyword.value, page: currentPage.value, page_size: pageSize.value }),
      listDevices({ page_size: 100 })
    ])
    if (poolResp.code === 1 && poolResp.data) {
      pools.value = poolResp.data.items
      total.value = poolResp.data.total
      totalPages.value = poolResp.data.total_pages
    }
    if (deviceResp.code === 1 && deviceResp.data) {
      devices.value = deviceResp.data.items
    }
  } finally {
    loading.value = false
  }
}

async function handleAddPool() {
  const id = newPoolId.value.trim()
  const name = newPoolName.value.trim()
  if (!id || !name) return
  
  const resp = await addPool(id, name)
  if (resp.code === 1) {
    for (const deviceId of selectedDeviceIds.value) {
      await addDeviceToPool(id, deviceId)
    }
    toast('支付池已创建' + (selectedDeviceIds.value.length > 0 ? '，已添加 ' + selectedDeviceIds.value.length + ' 台设备' : ''))
    newPoolId.value = ''
    newPoolName.value = ''
    selectedDeviceIds.value = []
    loadData()
  } else {
    toast(resp.msg || '创建失败', 'error')
  }
}

async function handleAddDeviceToPool() {
  if (!selectedPoolId.value || !selectedDeviceId.value) return
  
  const resp = await addDeviceToPool(selectedPoolId.value, selectedDeviceId.value)
  if (resp.code === 1) {
    toast('设备已添加到池')
    selectedDeviceId.value = ''
    loadData()
  } else {
    toast(resp.msg || '添加失败', 'error')
  }
}

async function handleRemoveDeviceFromPool(poolId: string, deviceId: string) {
  const resp = await removeDeviceFromPool(poolId, deviceId)
  if (resp.code === 1) {
    toast('设备已移除')
    loadData()
  } else {
    toast(resp.msg || '移除失败', 'error')
  }
}

async function handleDeletePool(poolId: string) {
  if (!confirm('确定要删除此支付池吗？')) return
  const resp = await deletePool(poolId)
  if (resp.code === 1) {
    toast('支付池已删除')
    loadData()
  } else {
    toast(resp.msg || '删除失败', 'error')
  }
}

onMounted(loadData)
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
}

.search-box {
  flex: 1;
}

.search-box input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border-dark);
  border-radius: 6px;
  font-size: 14px;
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

.form-row {
  display: flex;
  gap: 12px;
  align-items: flex-end;
  flex-wrap: wrap;
}

.form-group {
  flex: 1;
  min-width: 150px;
}

.form-group label {
  display: block;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border-dark);
  border-radius: 6px;
  font-size: 14px;
  background: #fff;
}

.form-group input:focus,
.form-group select:focus {
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

.pool-list {
  display: grid;
  gap: 16px;
}

.pool-item {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 16px;
}

.pool-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.delete-pool-btn {
  margin-left: auto;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  padding: 2px 6px;
  border-radius: 4px;
}

.delete-pool-btn:hover {
  background: #fee2e2;
}

.device-checkboxes {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  padding: 8px 0;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  width: auto;
}

.pool-name {
  font-weight: 600;
  font-size: 15px;
}

.pool-id {
  font-size: 12px;
  color: var(--text-secondary);
  background: #f0f0f0;
  padding: 2px 8px;
  border-radius: 4px;
}

.pool-devices {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.no-devices {
  color: var(--text-secondary);
  font-size: 13px;
}

.device-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  background: #e3f2fd;
  border-radius: 6px;
  font-size: 13px;
  color: #1565c0;
}

.remove-btn {
  background: none;
  border: none;
  color: #999;
  font-size: 16px;
  padding: 0;
  line-height: 1;
}

.remove-btn:hover {
  color: var(--danger);
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
