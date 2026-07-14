export interface ApiResponse<T = any> {
  code: number
  msg: string
  data?: T
}

export interface Order {
  trade_no: string
  service_id: string
  callback_url: string
  amount: number
  device_id?: string
  status: 'pending' | 'paid' | 'cancelled' | 'expired'
  created_at: number
  paid_at?: number
}

export interface Device {
  device_id: string
  key: string
  status: string
  last_heartbeat: number
  wechat_qr: string
  alipay_qr: string
}

export interface Pool {
  pool_id: string
  name: string
  device_ids: string[]
}

export interface Binding {
  service_id: string
  callback_url: string
  device_id: string
  pool_id: string
  api_key: string
  ip_whitelist: string
}
