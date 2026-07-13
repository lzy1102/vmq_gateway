<template>
  <div class="dashboard">
    <aside class="sidebar">
      <div class="sidebar-header">
        <h2>V免签</h2>
        <span class="subtitle">管理后台</span>
      </div>
      <nav class="sidebar-nav">
        <router-link to="/dashboard/devices" class="nav-item" :class="{ active: currentRoute === 'devices' }">
          <span class="nav-icon">📱</span>
          <span class="nav-text">设备管理</span>
        </router-link>
        <router-link to="/dashboard/pools" class="nav-item" :class="{ active: currentRoute === 'pools' }">
          <span class="nav-icon">💳</span>
          <span class="nav-text">支付池</span>
        </router-link>
        <router-link to="/dashboard/bindings" class="nav-item" :class="{ active: currentRoute === 'bindings' }">
          <span class="nav-icon">🔗</span>
          <span class="nav-text">服务绑定</span>
        </router-link>
        <router-link to="/dashboard/orders" class="nav-item" :class="{ active: currentRoute === 'orders' }">
          <span class="nav-icon">📋</span>
          <span class="nav-text">订单管理</span>
        </router-link>
      </nav>
      <div class="sidebar-footer">
        <button class="logout-btn" @click="handleLogout">
          <span class="nav-icon">🚪</span>
          <span class="nav-text">退出登录</span>
        </button>
      </div>
    </aside>
    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { logout } from '@/api'

const route = useRoute()
const router = useRouter()

const currentRoute = computed(() => {
  const path = route.path
  if (path.includes('/devices')) return 'devices'
  if (path.includes('/pools')) return 'pools'
  if (path.includes('/bindings')) return 'bindings'
  if (path.includes('/orders')) return 'orders'
  return ''
})

async function handleLogout() {
  await logout()
  router.push('/login')
}
</script>

<style scoped>
.dashboard {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: 240px;
  background: #1a1a2e;
  color: #fff;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-header {
  padding: 24px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.sidebar-header h2 {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 4px;
}

.sidebar-header .subtitle {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
}

.sidebar-nav {
  flex: 1;
  padding: 16px 12px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  color: rgba(255, 255, 255, 0.7);
  text-decoration: none;
  transition: all 0.2s;
  margin-bottom: 4px;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.nav-item.active {
  background: var(--primary);
  color: #fff;
}

.nav-icon {
  font-size: 18px;
  width: 24px;
  text-align: center;
}

.nav-text {
  font-size: 14px;
}

.sidebar-footer {
  padding: 16px 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.logout-btn {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px 16px;
  border-radius: 8px;
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  transition: all 0.2s;
}

.logout-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.main-content {
  flex: 1;
  padding: 32px;
  background: var(--bg);
  overflow-y: auto;
}
</style>
