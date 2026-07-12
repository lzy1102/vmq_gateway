<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-header">
        <h1>V免签</h1>
        <p>管理后台登录</p>
      </div>
      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label>用户名</label>
          <input 
            v-model="username" 
            type="text" 
            placeholder="请输入用户名"
            :disabled="loading"
            autofocus
          />
        </div>
        <div class="form-group">
          <label>密码</label>
          <div class="password-input">
            <input 
              v-model="password" 
              :type="showPassword ? 'text' : 'password'" 
              placeholder="请输入密码"
              :disabled="loading"
              @keyup.enter="handleLogin"
            />
            <button type="button" class="toggle-password" @click="showPassword = !showPassword" tabindex="-1">
              {{ showPassword ? '🙈' : '👁️' }}
            </button>
          </div>
        </div>
        <div v-if="errorMsg" class="error-message">{{ errorMsg }}</div>
        <button type="submit" class="login-btn" :disabled="loading || !username || !password">
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '@/api'

const router = useRouter()
const username = ref('')
const password = ref('')
const showPassword = ref(false)
const loading = ref(false)
const errorMsg = ref('')

async function handleLogin() {
  if (!username.value || !password.value) return
  
  loading.value = true
  errorMsg.value = ''
  
  try {
    const resp = await login(username.value, password.value)
    if (resp.code === 1) {
      router.push('/dashboard')
    } else {
      errorMsg.value = resp.msg || '登录失败'
    }
  } catch {
    errorMsg.value = '网络错误，请重试'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg);
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: var(--card-bg);
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  padding: 40px;
  margin: 20px;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h1 {
  font-size: 32px;
  color: var(--primary);
  margin-bottom: 8px;
}

.login-header p {
  color: var(--text-secondary);
  font-size: 14px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  font-size: 14px;
  color: var(--text);
  margin-bottom: 8px;
  font-weight: 500;
}

.form-group input {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid var(--border-dark);
  border-radius: 8px;
  font-size: 14px;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
}

.form-group input:disabled {
  background: #f5f5f5;
  cursor: not-allowed;
}

.password-input {
  position: relative;
}

.password-input input {
  padding-right: 44px;
}

.toggle-password {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  cursor: pointer;
  font-size: 18px;
  padding: 4px 6px;
  border-radius: 4px;
  line-height: 1;
}

.toggle-password:hover {
  background: rgba(0, 0, 0, 0.05);
}

.error-message {
  color: var(--danger);
  font-size: 13px;
  margin-bottom: 16px;
  padding: 10px;
  background: #fef2f2;
  border-radius: 6px;
}

.login-btn {
  width: 100%;
  padding: 12px;
  background: var(--primary);
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 500;
  transition: background-color 0.2s;
}

.login-btn:hover:not(:disabled) {
  background: var(--primary-dark);
}

.login-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
