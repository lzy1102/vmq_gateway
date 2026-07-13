<template>
  <div class="management-page">
    <div class="page-header">
      <h1>服务绑定</h1>
      <p class="description">配置回调接口，将支付结果通知到你的业务系统</p>
    </div>

    <!-- 创建绑定 -->
    <div class="card">
      <h2>创建绑定</h2>
      <div class="form-grid">
        <div class="form-group">
          <label>服务ID</label>
          <input v-model="newServiceId" placeholder="例如：my_app" />
        </div>
        <div class="form-group">
          <label>回调URL</label>
          <input v-model="newCallbackUrl" placeholder="https://your-server.com/callback" />
        </div>
        <div class="form-group">
          <label>绑定设备（可选）</label>
          <select v-model="newBindDevice">
            <option value="">不绑定设备</option>
            <option v-for="d in devices" :key="d.device_id" :value="d.device_id">{{ d.device_id }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>绑定池（可选）</label>
          <select v-model="newBindPool">
            <option value="">不绑定池</option>
            <option v-for="p in pools" :key="p.pool_id" :value="p.pool_id">{{ p.name }}</option>
          </select>
        </div>
      </div>
      <div class="form-actions">
        <button class="btn-primary" @click="handleAddBinding" :disabled="!newServiceId.trim()">
          创建绑定
        </button>
      </div>
    </div>

    <!-- 绑定列表 -->
    <div class="card">
      <div class="card-header">
        <h2>绑定列表</h2>
        <span class="total-count">共 {{ total }} 个绑定</span>
      </div>

      <!-- 搜索 -->
      <div class="toolbar">
        <div class="search-box">
          <input v-model="searchKeyword" placeholder="搜索服务ID或回调URL..." @keyup.enter="loadData" />
        </div>
        <button class="btn-secondary" @click="loadData">🔄 刷新</button>
      </div>

      <div v-if="loading" class="loading">加载中...</div>
      <div v-else-if="bindings.length === 0" class="empty-state">
        <p v-if="searchKeyword">没有找到匹配的绑定</p>
        <p v-else>暂无绑定</p>
        <p class="hint" v-if="!searchKeyword">创建绑定后，支付结果将回调到指定URL</p>
      </div>
      <div v-else class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>服务ID</th>
              <th>API Key</th>
              <th>回调URL</th>
              <th>绑定设备</th>
              <th>绑定池</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="b in bindings" :key="b.service_id">
              <td class="service-id">{{ b.service_id }}</td>
              <td>
                <code class="api-key" @click="copyKey(b.api_key)" title="点击复制">{{ b.api_key }}</code>
              </td>
              <td>
                <code class="callback-url">{{ b.callback_url }}</code>
              </td>
              <td>
                <span v-if="b.device_id" class="binding-tag device">{{ b.device_id }}</span>
                <span v-else class="no-binding">-</span>
              </td>
              <td>
                <span v-if="b.pool_id" class="binding-tag pool">{{ b.pool_id }}</span>
                <span v-else class="no-binding">-</span>
              </td>
              <td class="actions">
                <button class="btn-icon" @click="openEdit(b)" title="编辑">✏️</button>
                <button class="btn-icon" @click="handleDeleteBinding(b.service_id)" title="删除">🗑️</button>
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

    <!-- 回调说明 -->
    <div class="card info-card">
      <h2>回调说明</h2>
      <p>当用户支付成功后，系统会向回调URL发送POST请求：</p>
      <div class="code-block">
        <pre>{
  "order_id": "V1234567890_3001",
  "amount": 3001,
  "service_id": "my_app",
  "status": "paid",
  "paid_at": 1699000000
}</pre>
      </div>
      <p class="hint">你的服务需要返回 <code>{"code": 1}</code> 表示处理成功</p>
    </div>

    <!-- 编辑弹窗 -->
    <div v-if="editingBinding" class="modal-overlay" @click.self="closeEdit">
      <div class="modal">
        <h3>编辑绑定 - {{ editingBinding.service_id }}</h3>
        <div class="form-group">
          <label>回调URL</label>
          <input v-model="editCallbackUrl" placeholder="https://your-server/callback" />
        </div>
        <div class="form-group">
          <label>绑定设备</label>
          <select v-model="editBindDevice">
            <option value="">不绑定设备</option>
            <option v-for="d in devices" :key="d.device_id" :value="d.device_id">{{ d.device_id }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>绑定池</label>
          <select v-model="editBindPool">
            <option value="">不绑定池</option>
            <option v-for="p in pools" :key="p.pool_id" :value="p.pool_id">{{ p.name }}</option>
          </select>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="closeEdit">取消</button>
          <button class="btn-primary" @click="handleUpdateBinding">保存</button>
        </div>
      </div>
    </div>

    <!-- Toast -->
    <div v-if="toastMsg" class="toast" :class="toastType">{{ toastMsg }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import type { Binding, Device, Pool } from '@/types'
import { listBindings, addBinding, updateBinding, deleteBinding, listDevices, listPools } from '@/api'

const bindings = ref<Binding[]>([])
const devices = ref<Device[]>([])
const pools = ref<Pool[]>([])
const loading = ref(true)
const toastMsg = ref('')
const toastType = ref<'success' | 'error'>('success')
const total = ref(0)
const totalPages = ref(0)

const newServiceId = ref('')
const newCallbackUrl = ref('')
const newBindDevice = ref('')
const newBindPool = ref('')

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
    const [bindingResp, deviceResp, poolResp] = await Promise.all([
      listBindings({ keyword: searchKeyword.value, page: currentPage.value, page_size: pageSize.value }),
      listDevices({ page_size: 100 }),
      listPools({ page_size: 100 })
    ])
    if (bindingResp.code === 1 && bindingResp.data) {
      bindings.value = bindingResp.data.items
      total.value = bindingResp.data.total
      totalPages.value = bindingResp.data.total_pages
    }
    if (deviceResp.code === 1 && deviceResp.data) {
      devices.value = deviceResp.data.items
    }
    if (poolResp.code === 1 && poolResp.data) {
      pools.value = poolResp.data.items
    }
  } finally {
    loading.value = false
  }
}

async function handleAddBinding() {
  const serviceId = newServiceId.value.trim()
  if (!serviceId) return
  
  const resp = await addBinding(
    serviceId,
    newCallbackUrl.value.trim() || undefined,
    newBindDevice.value.trim() || undefined,
    newBindPool.value.trim() || undefined
  )
  
  if (resp.code === 1) {
    toast('绑定已创建')
    newServiceId.value = ''
    newCallbackUrl.value = ''
    newBindDevice.value = ''
    newBindPool.value = ''
    loadData()
  } else {
    toast(resp.msg || '创建失败', 'error')
  }
}

const editingBinding = ref<Binding | null>(null)
const editCallbackUrl = ref('')
const editBindDevice = ref('')
const editBindPool = ref('')

function openEdit(b: Binding) {
  editingBinding.value = b
  editCallbackUrl.value = b.callback_url
  editBindDevice.value = b.device_id || ''
  editBindPool.value = b.pool_id || ''
}

function closeEdit() {
  editingBinding.value = null
}

async function handleUpdateBinding() {
  if (!editingBinding.value) return
  const resp = await updateBinding(
    editingBinding.value.service_id,
    editCallbackUrl.value.trim(),
    editBindDevice.value.trim() || undefined,
    editBindPool.value.trim() || undefined
  )
  if (resp.code === 1) {
    toast('绑定已更新')
    closeEdit()
    loadData()
  } else {
    toast(resp.msg || '更新失败', 'error')
  }
}

async function handleDeleteBinding(serviceId: string) {
  if (!confirm('确定要删除此绑定吗？')) return
  const resp = await deleteBinding(serviceId)
  if (resp.code === 1) {
    toast('绑定已删除')
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

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
}

.form-group label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.form-group input,
.form-group select {
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

.form-actions {
  display: flex;
  justify-content: flex-end;
}

.btn-primary {
  padding: 10px 20px;
  background: var(--primary);
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 14px;
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

.service-id {
  font-weight: 500;
}

.api-key {
  display: inline-block;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 4px 8px;
  background: #f8f9fa;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
}

.api-key:hover {
  background: #e9ecef;
}

.callback-url {
  display: inline-block;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 4px 8px;
  background: #f8f9fa;
  border-radius: 4px;
  font-size: 12px;
}

.binding-tag {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
}

.binding-tag.device {
  background: #e3f2fd;
  color: #1565c0;
}

.binding-tag.pool {
  background: #e8f5e9;
  color: #2e7d32;
}

.no-binding {
  color: var(--text-secondary);
}

.info-card {
  background: #f8f9fa;
}

.info-card p {
  margin-bottom: 12px;
  line-height: 1.6;
}

.info-card code {
  padding: 2px 6px;
  background: #e9ecef;
  border-radius: 4px;
  font-size: 13px;
}

.code-block {
  background: #1a1a2e;
  color: #a8e6cf;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  margin-bottom: 12px;
}

.code-block pre {
  margin: 0;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 13px;
  line-height: 1.5;
}

.actions {
  display: flex;
  gap: 4px;
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  padding: 4px 6px;
  border-radius: 4px;
}

.btn-icon:hover {
  background: rgba(0, 0, 0, 0.05);
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  width: 100%;
  max-width: 480px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
}

.modal h3 {
  font-size: 16px;
  margin-bottom: 16px;
}

.modal .form-group {
  margin-bottom: 12px;
}

.modal .form-group label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.modal .form-group input,
.modal .form-group select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border-dark);
  border-radius: 6px;
  font-size: 14px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}

.btn-secondary {
  padding: 10px 20px;
  background: #f5f5f5;
  border: 1px solid var(--border-dark);
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
}

.btn-secondary:hover {
  background: #e9ecef;
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
