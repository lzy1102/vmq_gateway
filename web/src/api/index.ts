import type { ApiResponse, Device, Pool, Binding } from '@/types'

async function request<T>(url: string, options?: RequestInit): Promise<ApiResponse<T>> {
  const resp = await fetch(url, options)
  return resp.json()
}

// ========== 支付相关 ==========

export function createOrder(amount: number, serviceId: string, callbackUrl: string) {
  return request<{ order_id: string; amount: number; amount_str: number; device_id: string; qr_url: string }>('/api/order', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ amount, service_id: serviceId, callback_url: callbackUrl })
  })
}

export function queryOrderStatus(orderId: string) {
  return request<{ status: string; amount?: number }>('/api/order/status?order_id=' + encodeURIComponent(orderId))
}

// ========== 管理后台 ==========

export function login(username: string, password: string) {
  return request('/admin/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password })
  })
}

export function logout() {
  return request('/admin/logout', { method: 'POST' })
}

export function addDevice(deviceId: string) {
  return request<{ key: string }>('/admin/device', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ device_id: deviceId })
  })
}

export function deleteDevice(deviceId: string) {
  return request('/admin/device', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ device_id: deviceId })
  })
}

export function updateDevice(deviceId: string, key?: string) {
  const body: Record<string, string> = { device_id: deviceId }
  if (key) body.key = key
  return request('/admin/device', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  })
}

export function uploadQRCode(deviceId: string, type: 'wechat' | 'alipay', file: File) {
  const formData = new FormData()
  formData.append('device_id', deviceId)
  formData.append('type', type)
  formData.append('file', file)
  return request<{ qr_url: string }>('/admin/device/qrcode', {
    method: 'POST',
    body: formData
  })
}

export function listDevices(params?: { keyword?: string; page?: number; page_size?: number }) {
  const query = new URLSearchParams()
  if (params?.keyword) query.set('keyword', params.keyword)
  if (params?.page) query.set('page', String(params.page))
  if (params?.page_size) query.set('page_size', String(params.page_size))
  const qs = query.toString()
  return request<{ items: Device[]; total: number; page: number; page_size: number; total_pages: number }>(
    '/admin/devices' + (qs ? '?' + qs : '')
  )
}

export function addPool(poolId: string, name: string) {
  return request('/admin/pool', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pool_id: poolId, name })
  })
}

export function addDeviceToPool(poolId: string, deviceId: string) {
  return request('/admin/pool/device', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pool_id: poolId, device_id: deviceId })
  })
}

export function removeDeviceFromPool(poolId: string, deviceId: string) {
  return request('/admin/pool/device', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pool_id: poolId, device_id: deviceId })
  })
}

export function deletePool(poolId: string) {
  return request('/admin/pool', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pool_id: poolId })
  })
}

export function listPools(params?: { keyword?: string; page?: number; page_size?: number }) {
  const query = new URLSearchParams()
  if (params?.keyword) query.set('keyword', params.keyword)
  if (params?.page) query.set('page', String(params.page))
  if (params?.page_size) query.set('page_size', String(params.page_size))
  const qs = query.toString()
  return request<{ items: Pool[]; total: number; page: number; page_size: number; total_pages: number }>(
    '/admin/pools' + (qs ? '?' + qs : '')
  )
}

export function addBinding(serviceId: string, callbackUrl: string, deviceId?: string, poolId?: string) {
  const body: Record<string, string> = { service_id: serviceId, callback_url: callbackUrl }
  if (deviceId) body.device_id = deviceId
  if (poolId) body.pool_id = poolId
  return request('/admin/binding', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  })
}

export function updateBinding(serviceId: string, callbackUrl: string, deviceId?: string, poolId?: string) {
  const body: Record<string, string> = { service_id: serviceId, callback_url: callbackUrl }
  if (deviceId !== undefined) body.device_id = deviceId
  if (poolId !== undefined) body.pool_id = poolId
  return request('/admin/binding', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  })
}

export function deleteBinding(serviceId: string) {
  return request('/admin/binding', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ service_id: serviceId })
  })
}

export function listBindings(params?: { keyword?: string; page?: number; page_size?: number }) {
  const query = new URLSearchParams()
  if (params?.keyword) query.set('keyword', params.keyword)
  if (params?.page) query.set('page', String(params.page))
  if (params?.page_size) query.set('page_size', String(params.page_size))
  const qs = query.toString()
  return request<{ items: Binding[]; total: number; page: number; page_size: number; total_pages: number }>(
    '/admin/bindings' + (qs ? '?' + qs : '')
  )
}
