<template>
  <div class="admin-page">
    <!-- 登录 -->
    <div v-if="!isLoggedIn" class="login-box">
      <h2>管理后台登录</h2>
      <div v-if="loginError" class="error">{{ loginError }}</div>
      <input v-model="loginUser" placeholder="用户名" value="admin" />
      <input v-model="loginPass" type="password" placeholder="密码" @keyup.enter="doLogin" />
      <button @click="doLogin" :disabled="loggingIn">{{ loggingIn ? '登录中...' : '登录' }}</button>
    </div>

    <!-- 管理面板 -->
    <div v-else class="container">
      <div class="header">
        <h1>V免签管理后台</h1>
        <button class="logout" @click="doLogout">退出登录</button>
      </div>

      <div class="tabs">
        <div :class="['tab', { active: activeTab === 'devices' }]" @click="switchTab('devices')">设备管理</div>
        <div :class="['tab', { active: activeTab === 'pools' }]" @click="switchTab('pools')">支付池</div>
        <div :class="['tab', { active: activeTab === 'bindings' }]" @click="switchTab('bindings')">服务绑定</div>
      </div>

      <!-- 设备管理 -->
      <div v-show="activeTab === 'devices'" class="panel">
        <div class="form-row">
          <input v-model="newDeviceId" placeholder="设备ID（如 phone_a）" />
          <button @click="handleAddDevice">添加设备</button>
        </div>
        <div class="table-wrap">
          <table>
            <thead>
              <tr><th>设备ID</th><th>Key（填入APP）</th><th>状态</th><th>上次心跳</th></tr>
            </thead>
            <tbody>
              <tr v-if="devices.length === 0">
                <td colspan="4" class="empty">暂无设备</td>
              </tr>
              <tr v-for="d in devices" :key="d.device_id">
                <td>{{ d.device_id }}</td>
                <td><span class="key-cell">{{ d.vmq_key }}</span></td>
                <td><span :class="['status', d.status]">{{ d.status }}</span></td>
                <td>{{ d.last_heartbeat ? formatTime(d.last_heartbeat) : '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 支付池 -->
      <div v-show="activeTab === 'pools'" class="panel">
        <div class="form-row">
          <input v-model="newPoolId" placeholder="池子ID" />
          <input v-model="newPoolName" placeholder="池子名称" />
          <button @click="handleAddPool">创建池子</button>
        </div>
        <div class="table-wrap">
          <table>
            <thead>
              <tr><th>池子ID</th><th>名称</th><th>包含设备</th></tr>
            </thead>
            <tbody>
              <tr v-if="pools.length === 0">
                <td colspan="3" class="empty">暂无池子</td>
              </tr>
              <tr v-for="p in pools" :key="p.pool_id">
                <td>{{ p.pool_id }}</td>
                <td>{{ p.name }}</td>
                <td>{{ p.device_ids?.join(', ') || '空' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 服务绑定 -->
      <div v-show="activeTab === 'bindings'" class="panel">
        <div class="form-row">
          <input v-model="newServiceId" placeholder="服务ID" />
          <input v-model="newCallbackUrl" placeholder="回调URL" />
          <input v-model="newBindDevice" placeholder="绑定设备ID（可选）" />
          <input v-model="newBindPool" placeholder="绑定池子ID（可选）" />
          <button @click="handleAddBinding">创建绑定</button>
        </div>
        <div class="table-wrap">
          <table>
            <thead>
              <tr><th>服务ID</th><th>回调URL</th><th>绑定设备</th><th>绑定池子</th></tr>
            </thead>
            <tbody>
              <tr v-if="bindings.length === 0">
                <td colspan="4" class="empty">暂无绑定</td>
              </tr>
              <tr v-for="b in bindings" :key="b.service_id">
                <td>{{ b.service_id }}</td>
                <td>{{ b.callback_url }}</td>
                <td>{{ b.device_id || '-' }}</td>
                <td>{{ b.pool_id || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Toast -->
    <div v-if="toastMsg" class="toast">{{ toastMsg }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Device, Pool, Binding } from '@/types'
import * as api from '@/api'

// 登录状态
const isLoggedIn = ref(false)
const loginUser = ref('admin')
const loginPass = ref('')
const loginError = ref('')
const loggingIn = ref(false)

// Tab
const activeTab = ref<'devices' | 'pools' | 'bindings'>('devices')

// 设备
const devices = ref<Device[]>([])
const newDeviceId = ref('')

// 池子
const pools = ref<Pool[]>([])
const newPoolId = ref('')
const newPoolName = ref('')

// 绑定
const bindings = ref<Binding[]>([])
const newServiceId = ref('')
const newCallbackUrl = ref('')
const newBindDevice = ref('')
const newBindPool = ref('')

// Toast
const toastMsg = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null

function toast(msg: string) {
  toastMsg.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastMsg.value = '' }, 2000)
}

function formatTime(ts: number) {
  return new Date(ts * 1000).toLocaleString()
}

// ========== 登录 ==========

async function checkAuth() {
  try {
    const resp = await fetch('/admin/devices')
    if (resp.status === 401) return false
    return true
  } catch { return false }
}

async function doLogin() {
  if (!loginUser.value || !loginPass.value) return
  loggingIn.value = true
  loginError.value = ''
  try {
    const resp = await api.login(loginUser.value, loginPass.value)
    if (resp.code === 1) {
      isLoggedIn.value = true
      loadDevices()
    } else {
      loginError.value = resp.msg
    }
  } catch {
    loginError.value = '网络错误'
  } finally {
    loggingIn.value = false
  }
}

async function doLogout() {
  await api.logout()
  isLoggedIn.value = false
  devices.value = []
  pools.value = []
  bindings.value = []
}

// ========== Tab ==========

function switchTab(name: 'devices' | 'pools' | 'bindings') {
  activeTab.value = name
  if (name === 'devices') loadDevices()
  if (name === 'pools') loadPools()
  if (name === 'bindings') loadBindings()
}

// ========== 设备 ==========

async function loadDevices() {
  const resp = await api.listDevices()
  if (resp.code === 1 && resp.data) {
    devices.value = resp.data
  }
}

async function handleAddDevice() {
  const id = newDeviceId.value.trim()
  if (!id) return
  const resp = await api.addDevice(id)
  if (resp.code === 1) {
    toast('Key: ' + resp.data?.key)
    newDeviceId.value = ''
    loadDevices()
  }
}

// ========== 池子 ==========

async function loadPools() {
  const resp = await api.listPools()
  if (resp.code === 1 && resp.data) {
    pools.value = resp.data
  }
}

async function handleAddPool() {
  const id = newPoolId.value.trim()
  const name = newPoolName.value.trim()
  if (!id || !name) return
  await api.addPool(id, name)
  newPoolId.value = ''
  newPoolName.value = ''
  loadPools()
}

// ========== 绑定 ==========

async function loadBindings() {
  const resp = await api.listBindings()
  if (resp.code === 1 && resp.data) {
    bindings.value = resp.data
  }
}

async function handleAddBinding() {
  const serviceId = newServiceId.value.trim()
  const callbackUrl = newCallbackUrl.value.trim()
  if (!serviceId || !callbackUrl) return
  await api.addBinding(serviceId, callbackUrl, newBindDevice.value.trim() || undefined, newBindPool.value.trim() || undefined)
  newServiceId.value = ''
  newCallbackUrl.value = ''
  newBindDevice.value = ''
  newBindPool.value = ''
  loadBindings()
}

// ========== 初始化 ==========

onMounted(async () => {
  const ok = await checkAuth()
  if (ok) {
    isLoggedIn.value = true
    loadDevices()
  }
})
</script>

<style scoped>
.admin-page {
  padding: 20px;
}

/* 登录 */
.login-box {
  max-width: 360px;
  margin: 80px auto;
  background: var(--card-bg);
  padding: 32px;
  border-radius: 8px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.08);
}

.login-box h2 {
  margin-bottom: 20px;
  text-align: center;
}

.login-box input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border-dark);
  border-radius: 4px;
  margin-bottom: 12px;
  font-size: 14px;
}

.login-box button {
  width: 100%;
  padding: 10px;
  background: var(--primary);
  color: #fff;
  border: none;
  border-radius: 4px;
  font-size: 14px;
}

.login-box button:hover:not(:disabled) {
  background: var(--primary-dark);
}

.login-box button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.error {
  color: var(--danger);
  font-size: 13px;
  margin-bottom: 12px;
}

/* 管理面板 */
.container {
  max-width: 900px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h1 {
  font-size: 24px;
}

.logout {
  background: none;
  border: 1px solid var(--border-dark);
  padding: 6px 12px;
  border-radius: 4px;
  color: #666;
  font-size: 13px;
}

.logout:hover {
  border-color: var(--danger);
  color: var(--danger);
}

/* Tabs */
.tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
}

.tab {
  padding: 10px 20px;
  border: 1px solid var(--border-dark);
  border-radius: 6px;
  cursor: pointer;
  background: var(--card-bg);
}

.tab.active {
  background: var(--primary);
  color: #fff;
  border-color: var(--primary);
}

/* Panel */
.panel {
  background: var(--card-bg);
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.08);
}

.form-row {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.form-row input {
  flex: 1;
  min-width: 120px;
  padding: 8px 12px;
  border: 1px solid var(--border-dark);
  border-radius: 4px;
}

.form-row button {
  padding: 8px 16px;
  background: var(--primary);
  color: #fff;
  border: none;
  border-radius: 4px;
  white-space: nowrap;
}

.form-row button:hover {
  background: var(--primary-dark);
}

/* Table */
.table-wrap {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 16px;
}

th, td {
  padding: 10px 12px;
  text-align: left;
  border-bottom: 1px solid var(--border);
}

th {
  background: #f8f9fa;
  font-weight: 600;
}

.empty {
  color: #999;
  padding: 20px;
  text-align: center;
}

.status {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.status.online { background: #d4edda; color: #155724; }
.status.offline { background: #f8d7da; color: #721c24; }

.key-cell {
  font-family: monospace;
  font-size: 13px;
  background: #f8f9fa;
  padding: 4px 8px;
  border-radius: 4px;
  user-select: all;
}

/* Toast */
.toast {
  position: fixed;
  top: 20px;
  right: 20px;
  background: var(--success);
  color: #fff;
  padding: 12px 20px;
  border-radius: 6px;
  z-index: 100;
}
</style>
