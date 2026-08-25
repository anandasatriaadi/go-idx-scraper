<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm p-4" @click.self="$emit('close')">
    <Card class="relative w-full max-w-sm border-border bg-card shadow-2xl overflow-hidden">
      <CardHeader class="border-b border-border/80 pb-4">
        <div class="flex items-center justify-between">
          <CardTitle class="text-base font-bold font-mono tracking-tight text-foreground">
            {{ isSignUp ? 'CREATE TERMINAL ACCOUNT' : 'INVESTOR SIGN IN' }}
          </CardTitle>
          <Button
            variant="ghost"
            size="iconSm"
            class="h-7 w-7 rounded-full text-muted-foreground hover:bg-muted hover:text-foreground"
            @click="$emit('close')"
          >
            ✕
          </Button>
        </div>
      </CardHeader>

      <CardContent class="p-6 space-y-4">
        <div v-if="error" class="rounded-md border border-rose-500/30 bg-rose-500/10 p-3 text-xs text-rose-400 font-mono">
          {{ error }}
        </div>

        <!-- Google Login -->
        <Button
          variant="outline"
          class="w-full font-mono text-xs border-border hover:border-primary/50"
          :disabled="loading"
          @click="handleGoogleLogin"
        >
          Sign in with Google
        </Button>

        <div class="relative my-3 text-center">
          <div class="absolute inset-0 flex items-center">
            <span class="w-full border-t border-border"></span>
          </div>
          <span class="relative bg-card px-2 text-[11px] font-mono text-muted-foreground uppercase">
            or with email
          </span>
        </div>

        <form class="space-y-3" @submit.prevent="handleSubmit">
          <div class="space-y-1">
            <label class="text-[11px] font-mono text-muted-foreground">Email Address</label>
            <Input
              v-model="email"
              type="email"
              required
              placeholder="investor@example.com"
              class="h-9 font-mono text-xs bg-background/80"
            />
          </div>

          <div class="space-y-1">
            <label class="text-[11px] font-mono text-muted-foreground">Password</label>
            <Input
              v-model="password"
              type="password"
              required
              placeholder="••••••••"
              class="h-9 font-mono text-xs bg-background/80"
            />
          </div>

          <Button
            type="submit"
            class="w-full font-mono text-xs mt-2"
            :disabled="loading"
          >
            {{ loading ? 'Processing...' : (isSignUp ? 'Sign Up' : 'Sign In') }}
          </Button>
        </form>

        <div class="pt-2 text-center text-xs text-muted-foreground">
          <span>{{ isSignUp ? 'Already have an account?' : "Don't have an account?" }}</span>
          <button
            type="button"
            class="ml-1.5 font-mono text-primary hover:underline font-semibold"
            @click="isSignUp = !isSignUp"
          >
            {{ isSignUp ? 'Sign In' : 'Create One' }}
          </button>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useAuth } from '@/composables/useAuth'

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
