<template>
  <div class="modal-backdrop" @click.self="$emit('close')">
    <div class="auth-card">
      <div class="auth-header">
        <h2 class="auth-title">{{ isSignUp ? 'Create Account' : 'Sign In' }}</h2>
        <button class="btn-close" @click="$emit('close')">✕</button>
      </div>

      <div class="auth-body">
        <div v-if="error" class="error-banner">
          {{ error }}
        </div>

        <!-- Google Login Button -->
        <button class="btn-google" @click="handleGoogleLogin">
          <span>Sign in with Google</span>
        </button>

        <div class="divider">
          <span>or with email</span>
        </div>

        <form class="auth-form" @submit.prevent="handleSubmit">
          <div class="form-group">
            <label>Email</label>
            <input v-model="email" type="email" required placeholder="investor@example.com" />
          </div>

          <div class="form-group">
            <label>Password</label>
            <input v-model="password" type="password" required placeholder="••••••••" />
          </div>

          <button type="submit" class="btn-submit" :disabled="loading">
            {{ loading ? 'Processing...' : (isSignUp ? 'Sign Up' : 'Sign In') }}
          </button>
        </form>

        <div class="auth-toggle">
          <span>{{ isSignUp ? 'Already have an account?' : "Don't have an account?" }}</span>
          <button class="btn-toggle" @click="isSignUp = !isSignUp">
            {{ isSignUp ? 'Sign In' : 'Create One' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAuth } from '../composables/useAuth'

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'authenticated'): void
}>()

const { loginWithGoogle, loginWithEmail, signupWithEmail, error } = useAuth()

const isSignUp = ref(false)
const email = ref('')
const password = ref('')
const loading = ref(false)

const handleGoogleLogin = async () => {
  loading.value = true
  try {
    await loginWithGoogle()
    emit('authenticated')
    emit('close')
  } catch (e) {
    // error is handled by useAuth
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  loading.value = true
  try {
    if (isSignUp.value) {
      await signupWithEmail(email.value, password.value)
    } else {
      await loginWithEmail(email.value, password.value)
    }
    emit('authenticated')
    emit('close')
  } catch (e) {
    // error handled by useAuth
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 20px;
}
.auth-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.6);
}
.auth-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.auth-title {
  font-size: 1.1rem;
  font-weight: 700;
}
.btn-close {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 1.1rem;
  cursor: pointer;
}
.auth-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.error-banner {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid #ef4444;
  color: #fca5a5;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 0.85rem;
}
.btn-google {
  background: #fff;
  color: #0f172a;
  border: none;
  padding: 10px;
  border-radius: 6px;
  font-weight: 600;
  font-size: 0.9rem;
  cursor: pointer;
  transition: opacity 0.15s ease;
}
.btn-google:hover {
  opacity: 0.9;
}
.divider {
  display: flex;
  align-items: center;
  text-align: center;
  color: var(--text-muted);
  font-size: 0.75rem;
  text-transform: uppercase;
}
.divider::before, .divider::after {
  content: '';
  flex: 1;
  border-bottom: 1px solid var(--border-color);
}
.divider span {
  padding: 0 8px;
}
.auth-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.form-group label {
  font-size: 0.8rem;
  color: var(--text-secondary);
}
.form-group input {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  color: #fff;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 0.9rem;
  outline: none;
}
.form-group input:focus {
  border-color: #38bdf8;
}
.btn-submit {
  background: #2563eb;
  color: #fff;
  border: none;
  padding: 10px;
  border-radius: 6px;
  font-weight: 600;
  font-size: 0.9rem;
  cursor: pointer;
  margin-top: 4px;
}
.btn-submit:hover:not(:disabled) {
  background: #1d4ed8;
}
.auth-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 0.8rem;
  color: var(--text-secondary);
}
.btn-toggle {
  background: transparent;
  border: none;
  color: #38bdf8;
  cursor: pointer;
  font-weight: 600;
}
</style>
