import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/dashboard'
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/Login.vue'),
      meta: { guest: true }
    },
    {
      path: '/dashboard',
      component: () => import('@/views/Dashboard.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: '/dashboard/devices'
        },
        {
          path: 'devices',
          name: 'admin-devices',
          component: () => import('@/views/DeviceManagement.vue')
        },
        {
          path: 'pools',
          name: 'admin-pools',
          component: () => import('@/views/PoolManagement.vue')
        },
        {
          path: 'bindings',
          name: 'admin-bindings',
          component: () => import('@/views/BindingManagement.vue')
        },
        {
          path: 'orders',
          name: 'admin-orders',
          component: () => import('@/views/OrderManagement.vue')
        },
        {
          path: 'tutorial',
          name: 'admin-tutorial',
          component: () => import('@/views/Tutorial.vue')
        }
      ]
    }
  ]
})

async function checkAuth(): Promise<boolean> {
  try {
    const resp = await fetch('/admin/devices', { credentials: 'same-origin' })
    if (!resp.ok) return false
    const data = await resp.json()
    return data.code === 1
  } catch {
    return false
  }
}

// 路由守卫
router.beforeEach(async (to, from, next) => {
  // 已登录用户访问登录页时跳转到管理页
  if (to.meta.guest) {
    const loggedIn = await checkAuth()
    if (loggedIn) {
      next('/dashboard')
      return
    }
    next()
    return
  }

  // 需要登录的页面
  if (to.meta.requiresAuth) {
    const loggedIn = await checkAuth()
    if (!loggedIn) {
      next('/login')
      return
    }
  }

  next()
})

export default router
