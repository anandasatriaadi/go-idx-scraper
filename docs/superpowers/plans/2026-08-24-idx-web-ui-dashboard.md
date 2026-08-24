# IDX Intelligence Terminal UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a sleek, high-contrast, dark-themed financial intelligence terminal UI for `idx-web` (Nuxt 4 / Vue 3) featuring an Executive Overview dashboard, Daily Market Briefings reader, News Intelligence Terminal with instant ticker & sector filters, IDX Announcements feed, Financial Reports archive, Article Reader modal, and Firebase Auth with Watchlist synchronization.

**Architecture:** Component-driven Vue 3 / Nuxt 4 architecture with single-page tabbed navigation. State and API integration are managed via Vue composables (`useAuth`, `useWatchlist`), styled using a Dark Terminal design system (`assets/main.css`) with emerald green bullish indicators, crimson red bearish alerts, and high-legibility monospace tickers.

**Tech Stack:** Nuxt 4 / Vue 3, Firebase Client SDK (`firebase/auth`, `firebase/app`), Nitro Server Engine, Vanilla Dark Terminal CSS (CSS Variables, Flexbox, Grid).

## Global Constraints

- Use Dark Terminal palette (`#080c14` background, `#0f172a` cards, `#1e293b` borders).
- Monospace font for all stock tickers, numbers, and dates (`ui-monospace`, `SFMono-Regular`, `Menlo`, `Monaco`, `Consolas`).
- Support responsive layout (Desktop, Laptop, Tablet, Mobile).
- Clean error handling with fallback states for empty data / loading states.
- Client-only guards for Firebase browser APIs (`useAuth`).

---

### Task 1: Design System & State Composables (CSS, Auth, Watchlist)

**Files:**
- Create: `idx-web/src/assets/main.css`
- Create: `idx-web/src/composables/useAuth.ts`
- Create: `idx-web/src/composables/useWatchlist.ts`
- Modify: `idx-web/nuxt.config.ts`

**Interfaces:**
- Produces: `useAuth()` (`user`, `token`, `loginWithGoogle`, `loginWithEmail`, `signupWithEmail`, `logout`), `useWatchlist()` (`watchlist`, `fetchWatchlist`, `toggleWatchlist`, `isWatched`).

- [ ] **Step 1: Create `idx-web/src/assets/main.css` with Dark Terminal design system**

```css
:root {
  --bg-app: #080c14;
  --bg-card: #0f172a;
  --bg-card-hover: #172033;
  --bg-card-active: #1e293b;
  --border-color: #1e293b;
  --border-subtle: #334155;

  --text-primary: #f8fafc;
  --text-secondary: #94a3b8;
  --text-muted: #64748b;

  --bullish-bg: rgba(16, 185, 129, 0.12);
  --bullish-border: #10b981;
  --bullish-text: #34d399;

  --bearish-bg: rgba(239, 68, 68, 0.12);
  --bearish-border: #ef4444;
  --bearish-text: #f87171;

  --neutral-bg: rgba(56, 189, 248, 0.12);
  --neutral-border: #38bdf8;
  --neutral-text: #7dd3fc;

  --accent-amber: #f59e0b;
  --accent-blue: #2563eb;
  --font-mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  background-color: var(--bg-app);
  color: var(--text-primary);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  min-height: 100vh;
  line-height: 1.5;
  -webkit-font-smoothing: antialiased;
}

/* Monospace Helpers */
.font-mono {
  font-family: var(--font-mono);
}

/* Custom Scrollbars */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
::-webkit-scrollbar-track {
  background: var(--bg-app);
}
::-webkit-scrollbar-thumb {
  background: var(--border-subtle);
  border-radius: 3px;
}
::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}
```

- [ ] **Step 2: Create `idx-web/src/composables/useAuth.ts`**

```typescript
import { ref } from 'vue'
import {
  signInWithPopup,
  GoogleAuthProvider,
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  signOut,
  onAuthStateChanged,
  type User
} from 'firebase/auth'

export const useAuth = () => {
  const nuxtApp = useNuxtApp()
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const loading = ref(true)
  const error = ref<string | null>(null)

  const auth = nuxtApp.$firebaseAuth as any

  if (import.meta.client && auth) {
    onAuthStateChanged(auth, async (u) => {
      user.value = u
      if (u) {
        token.value = await u.getIdToken()
      } else {
        token.value = null
      }
      loading.value = false
    })
  } else {
    loading.value = false
  }

  const loginWithGoogle = async () => {
    error.value = null
    try {
      if (!auth) throw new Error('Firebase Auth not available')
      const provider = new GoogleAuthProvider()
      const result = await signInWithPopup(auth, provider)
      user.value = result.user
      token.value = await result.user.getIdToken()
      return result.user
    } catch (e: any) {
      error.value = e.message || 'Failed to sign in with Google'
      throw e
    }
  }

  const loginWithEmail = async (email: string, pass: string) => {
    error.value = null
    try {
      if (!auth) throw new Error('Firebase Auth not available')
      const result = await signInWithEmailAndPassword(auth, email, pass)
      user.value = result.user
      token.value = await result.user.getIdToken()
      return result.user
    } catch (e: any) {
      error.value = e.message || 'Failed to sign in with email'
      throw e
    }
  }

  const signupWithEmail = async (email: string, pass: string) => {
    error.value = null
    try {
      if (!auth) throw new Error('Firebase Auth not available')
      const result = await createUserWithEmailAndPassword(auth, email, pass)
      user.value = result.user
      token.value = await result.user.getIdToken()
      return result.user
    } catch (e: any) {
      error.value = e.message || 'Failed to create account'
      throw e
    }
  }

  const logout = async () => {
    try {
      if (auth) await signOut(auth)
      user.value = null
      token.value = null
    } catch (e: any) {
      error.value = e.message || 'Failed to sign out'
    }
  }

  return {
    user,
    token,
    loading,
    error,
    loginWithGoogle,
    loginWithEmail,
    signupWithEmail,
    logout
  }
}
```

- [ ] **Step 3: Create `idx-web/src/composables/useWatchlist.ts`**

```typescript
import { ref } from 'vue'
import { useAuth } from './useAuth'

export const useWatchlist = () => {
  const { token, user } = useAuth()
  const watchlist = ref<string[]>([])
  const loading = ref(false)

  const fetchWatchlist = async () => {
    if (!token.value) {
      watchlist.value = []
      return
    }
    loading.value = true
    try {
      const data = await $fetch<{ watchlist: string[] }>('/api/v1/user/watchlist', {
        headers: {
          Authorization: `Bearer ${token.value}`
        }
      })
      if (data && Array.isArray(data.watchlist)) {
        watchlist.value = data.watchlist
      }
    } catch (e) {
      console.warn('Could not fetch watchlist', e)
    } finally {
      loading.value = false
    }
  }

  const toggleWatchlist = async (ticker: string) => {
    if (!token.value) {
      alert('Please sign in to manage your watchlist')
      return
    }
    const upper = ticker.toUpperCase().trim()
    const index = watchlist.value.indexOf(upper)
    const newWatchlist = [...watchlist.value]
    if (index >= 0) {
      newWatchlist.splice(index, 1)
    } else {
      newWatchlist.push(upper)
    }

    watchlist.value = newWatchlist

    try {
      await $fetch('/api/v1/user/watchlist', {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${token.value}`
        },
        body: {
          watchlist: newWatchlist
        }
      })
    } catch (e) {
      console.error('Failed to sync watchlist', e)
      // Rollback on failure
      fetchWatchlist()
    }
  }

  const isWatched = (ticker: string) => {
    if (!ticker) return false
    return watchlist.value.includes(ticker.toUpperCase().trim())
  }

  return {
    watchlist,
    loading,
    fetchWatchlist,
    toggleWatchlist,
    isWatched
  }
}
```

- [ ] **Step 4: Update `idx-web/nuxt.config.ts` to include `assets/main.css`**

Add `css: ['~/assets/main.css']` to `idx-web/nuxt.config.ts`.

- [ ] **Step 5: Verify build & commit**

```bash
cd idx-web && npm run build
git add src/assets/main.css src/composables/ nuxt.config.ts
git commit -m "feat(web): add dark terminal design system and auth/watchlist composables"
```

---

### Task 2: Navigation & Global Shell Components

**Files:**
- Create: `idx-web/src/components/Navbar.vue`
- Modify: `idx-web/src/app.vue`

**Interfaces:**
- Produces: `Navbar` component with brand header, view tab switcher, global search input, and Auth modal trigger.

- [ ] **Step 1: Create `idx-web/src/components/Navbar.vue`**

```vue
<template>
  <header class="navbar">
    <div class="nav-container">
      <!-- Brand Logo -->
      <div class="brand" @click="$emit('select-tab', 'overview')">
        <span class="brand-badge">IDX</span>
        <span class="brand-title">TERMINAL</span>
        <span class="pulse-indicator" title="Live System Online"></span>
      </div>

      <!-- Navigation Tabs -->
      <nav class="nav-tabs">
        <button
          :class="['tab-btn', { active: activeTab === 'overview' }]"
          @click="$emit('select-tab', 'overview')"
        >
          Overview
        </button>
        <button
          :class="['tab-btn', { active: activeTab === 'briefing' }]"
          @click="$emit('select-tab', 'briefing')"
        >
          Daily Briefing
        </button>
        <button
          :class="['tab-btn', { active: activeTab === 'news' }]"
          @click="$emit('select-tab', 'news')"
        >
          News Terminal
        </button>
        <button
          :class="['tab-btn', { active: activeTab === 'announcements' }]"
          @click="$emit('select-tab', 'announcements')"
        >
          Disclosures
        </button>
        <button
          :class="['tab-btn', { active: activeTab === 'reports' }]"
          @click="$emit('select-tab', 'reports')"
        >
          Financial Reports
        </button>
      </nav>

      <!-- Right Controls -->
      <div class="nav-actions">
        <!-- Quick Search -->
        <div class="search-box">
          <span class="search-icon">🔍</span>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search ticker (e.g. BBRI)..."
            class="search-input font-mono"
            @keyup.enter="handleSearch"
          />
        </div>

        <!-- User Auth -->
        <button v-if="!user" class="btn-auth" @click="$emit('open-auth')">
          Sign In
        </button>
        <div v-else class="user-pill">
          <span class="user-email font-mono">{{ userEmail }}</span>
          <button class="btn-logout" title="Sign Out" @click="logout">✕</button>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuth } from '../composables/useAuth'

defineProps<{
  activeTab: string
}>()

const emit = defineEmits<{
  (e: 'select-tab', tab: string): void
  (e: 'open-auth'): void
  (e: 'search-ticker', ticker: string): void
}>()

const { user, logout } = useAuth()
const searchQuery = ref('')

const userEmail = computed(() => {
  if (!user.value?.email) return 'User'
  return user.value.email.split('@')[0]
})

const handleSearch = () => {
  if (searchQuery.value.trim()) {
    emit('search-ticker', searchQuery.value.trim().toUpperCase())
  }
}
</script>

<style scoped>
.navbar {
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
  position: sticky;
  top: 0;
  z-index: 50;
  backdrop-filter: blur(8px);
}
.nav-container {
  max-width: 1440px;
  margin: 0 auto;
  padding: 0 20px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}
.brand {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}
.brand-badge {
  background: #2563eb;
  color: #fff;
  font-weight: 800;
  font-size: 0.75rem;
  padding: 2px 6px;
  border-radius: 4px;
  letter-spacing: 0.05em;
}
.brand-title {
  font-weight: 700;
  font-size: 1.1rem;
  letter-spacing: 0.08em;
  color: var(--text-primary);
}
.pulse-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #10b981;
  box-shadow: 0 0 8px #10b981;
}
.nav-tabs {
  display: flex;
  gap: 4px;
  background: var(--bg-app);
  padding: 4px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
}
.tab-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}
.tab-btn:hover {
  color: var(--text-primary);
  background: rgba(255, 255, 255, 0.04);
}
.tab-btn.active {
  background: var(--bg-card-active);
  color: #38bdf8;
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}
.nav-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
.search-box {
  display: flex;
  align-items: center;
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 6px 10px;
  gap: 6px;
}
.search-icon {
  font-size: 0.8rem;
  opacity: 0.6;
}
.search-input {
  background: transparent;
  border: none;
  color: var(--text-primary);
  font-size: 0.85rem;
  outline: none;
  width: 180px;
}
.search-input::placeholder {
  color: var(--text-muted);
}
.btn-auth {
  background: #2563eb;
  color: #fff;
  border: none;
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s ease;
}
.btn-auth:hover {
  background: #1d4ed8;
}
.user-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  padding: 4px 10px;
  border-radius: 6px;
}
.user-email {
  font-size: 0.85rem;
  color: #38bdf8;
}
.btn-logout {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 0.75rem;
  cursor: pointer;
}
.btn-logout:hover {
  color: #ef4444;
}
@media (max-width: 900px) {
  .nav-tabs {
    display: none;
  }
}
</style>
```

- [ ] **Step 2: Update `idx-web/src/app.vue` to render main layout**

```vue
<template>
  <div class="app-root">
    <NuxtPage />
  </div>
</template>

<style>
.app-root {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
</style>
```

- [ ] **Step 3: Commit Task 2**

```bash
git add src/components/Navbar.vue src/app.vue
git commit -m "feat(web): add dark terminal Navbar and root layout shell"
```

---

### Task 3: Interactive Modals (Article Reader & Auth Dialog)

**Files:**
- Create: `idx-web/src/components/ArticleModal.vue`
- Create: `idx-web/src/components/AuthModal.vue`

**Interfaces:**
- `ArticleModal`: Displays complete cleaned markdown news, value score, impact direction badge, ticker pills, and link.
- `AuthModal`: Google Sign-In and Email password authentication.

- [ ] **Step 1: Create `idx-web/src/components/ArticleModal.vue`**

```vue
<template>
  <div v-if="article" class="modal-backdrop" @click.self="$emit('close')">
    <div class="modal-card">
      <!-- Header -->
      <div class="modal-header">
        <div class="header-badges">
          <span :class="['direction-badge', article.impact_direction?.toLowerCase()]">
            {{ article.impact_direction || 'Neutral' }}
          </span>
          <span :class="['score-badge font-mono', getScoreClass(article.value_score)]">
            Value Score: {{ (article.value_score !== undefined && article.value_score > 0 ? '+' : '') + (article.value_score || 0) }}
          </span>
          <span v-if="article.industry" class="industry-badge">
            {{ article.industry }}
          </span>
        </div>
        <button class="btn-close" @click="$emit('close')">✕</button>
      </div>

      <!-- Content Scroll -->
      <div class="modal-body">
        <h1 class="article-title">{{ article.title }}</h1>
        
        <div class="meta-row font-mono">
          <span>📅 {{ formatDate(article.date || article.created_at) }}</span>
          <div v-if="article.tickers && article.tickers.length > 0" class="tickers-list">
            <span>Tickers:</span>
            <span
              v-for="t in article.tickers"
              :key="t"
              class="ticker-tag"
              @click="$emit('filter-ticker', t)"
            >
              ${{ t }}
            </span>
          </div>
        </div>

        <!-- Value Takeaway Box -->
        <div v-if="article.investment_takeaway" class="takeaway-card">
          <div class="takeaway-title">💡 Value Investor Takeaway</div>
          <p class="takeaway-text">{{ article.investment_takeaway }}</p>
        </div>

        <!-- 3-Sentence Executive Summary -->
        <div v-if="article.summary" class="summary-card">
          <div class="summary-title">📋 Executive Summary</div>
          <p class="summary-text">{{ article.summary }}</p>
        </div>

        <!-- Full Content -->
        <div class="content-view">
          <div class="content-title">Full Story</div>
          <div class="markdown-body">{{ article.content }}</div>
        </div>

        <!-- External Link -->
        <div v-if="article.link" class="source-row">
          <a :href="article.link" target="_blank" rel="noopener" class="source-link">
            Open Original Source on Kontan ↗
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { News } from '../server/utils/types'

defineProps<{
  article: News | null
}>()

defineEmits<{
  (e: 'close'): void
  (e: 'filter-ticker', ticker: string): void
}>()

const getScoreClass = (score?: number) => {
  if (score === undefined || score === 0) return 'score-neutral'
  return score > 0 ? 'score-bullish' : 'score-bearish'
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  })
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
.modal-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  width: 100%;
  max-width: 840px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.6);
}
.modal-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.header-badges {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
.direction-badge {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 4px;
  text-transform: uppercase;
}
.direction-badge.bullish {
  background: var(--bullish-bg);
  color: var(--bullish-text);
  border: 1px solid var(--bullish-border);
}
.direction-badge.bearish {
  background: var(--bearish-bg);
  color: var(--bearish-text);
  border: 1px solid var(--bearish-border);
}
.direction-badge.neutral {
  background: var(--neutral-bg);
  color: var(--neutral-text);
  border: 1px solid var(--neutral-border);
}
.score-badge {
  font-size: 0.8rem;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 4px;
}
.score-bullish {
  background: var(--bullish-bg);
  color: var(--bullish-text);
}
.score-bearish {
  background: var(--bearish-bg);
  color: var(--bearish-text);
}
.score-neutral {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-secondary);
}
.industry-badge {
  font-size: 0.75rem;
  background: #1e293b;
  color: #94a3b8;
  padding: 2px 8px;
  border-radius: 4px;
}
.btn-close {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 1.1rem;
  cursor: pointer;
}
.btn-close:hover {
  color: #fff;
}
.modal-body {
  padding: 20px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.article-title {
  font-size: 1.4rem;
  font-weight: 700;
  line-height: 1.3;
}
.meta-row {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 0.85rem;
  color: var(--text-muted);
}
.tickers-list {
  display: flex;
  align-items: center;
  gap: 6px;
}
.ticker-tag {
  background: #1e293b;
  color: #38bdf8;
  padding: 2px 6px;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
}
.ticker-tag:hover {
  background: #2563eb;
  color: #fff;
}
.takeaway-card {
  background: rgba(245, 158, 11, 0.08);
  border-left: 3px solid var(--accent-amber);
  padding: 12px 16px;
  border-radius: 0 6px 6px 0;
}
.takeaway-title {
  font-weight: 700;
  font-size: 0.85rem;
  color: var(--accent-amber);
  margin-bottom: 4px;
}
.takeaway-text {
  font-size: 0.9rem;
  color: #fde68a;
}
.summary-card {
  background: rgba(56, 189, 248, 0.06);
  border-left: 3px solid #38bdf8;
  padding: 12px 16px;
  border-radius: 0 6px 6px 0;
}
.summary-title {
  font-weight: 700;
  font-size: 0.85rem;
  color: #38bdf8;
  margin-bottom: 4px;
}
.summary-text {
  font-size: 0.9rem;
  color: #e0f2fe;
}
.content-view {
  margin-top: 8px;
}
.content-title {
  font-size: 0.8rem;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 8px;
  letter-spacing: 0.05em;
}
.markdown-body {
  white-space: pre-wrap;
  line-height: 1.6;
  color: var(--text-secondary);
  font-size: 0.95rem;
}
.source-row {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}
.source-link {
  color: #38bdf8;
  text-decoration: none;
  font-size: 0.85rem;
  font-weight: 500;
}
.source-link:hover {
  text-decoration: underline;
}
</style>
```

- [ ] **Step 2: Create `idx-web/src/components/AuthModal.vue`**

```vue
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
```

- [ ] **Step 3: Commit Task 3**

```bash
git add src/components/ArticleModal.vue src/components/AuthModal.vue
git commit -m "feat(web): add Article reader and Firebase Auth modals"
```

---

### Task 4: Overview & Daily Briefing Views

**Files:**
- Create: `idx-web/src/components/OverviewView.vue`
- Create: `idx-web/src/components/BriefingView.vue`

**Interfaces:**
- `OverviewView`: Hero Briefing Pulse banner (Macro, Bullish Watchlist, Bearish Risk Alerts, Action Plan) + Live News Stream Grid with sector filter chips.
- `BriefingView`: Full historical briefing reader with date picker & Markdown view.

- [ ] **Step 1: Create `idx-web/src/components/OverviewView.vue`**

```vue
<template>
  <div class="overview-container">
    <!-- HERO: Latest Daily Market Briefing -->
    <section v-if="briefing" class="hero-briefing">
      <div class="briefing-header">
        <div class="pulse-tag font-mono">
          <span class="live-dot"></span> 7:00 AM GMT+8 BRIEFING
        </div>
        <h1 class="briefing-headline">{{ briefing.title }}</h1>
        <p class="macro-text">{{ briefing.macro_pulse }}</p>
      </div>

      <div class="lookout-grid">
        <!-- Bullish Opportunities -->
        <div class="lookout-col bullish-card">
          <div class="col-title font-mono">
            <span>🟢 BUY LOOKOUT / OPPORTUNITIES</span>
          </div>
          <div v-if="!briefing.bullish_lookout || briefing.bullish_lookout.length === 0" class="empty-text">
            No high-conviction bullish candidates today.
          </div>
          <div v-else class="items-list">
            <div
              v-for="item in briefing.bullish_lookout"
              :key="item.ticker"
              class="lookout-item"
            >
              <div class="item-header">
                <span class="ticker-badge font-mono" @click="$emit('filter-ticker', item.ticker)">
                  ${{ item.ticker }}
                </span>
                <span class="score-pill font-mono">+{{ item.value_score }}</span>
              </div>
              <div class="item-headline">{{ item.headline }}</div>
              <div class="item-takeaway">💡 {{ item.investment_takeaway }}</div>
            </div>
          </div>
        </div>

        <!-- Bearish Risk Alerts -->
        <div class="lookout-col bearish-card">
          <div class="col-title font-mono">
            <span>🔴 RISK ALERTS / HEADWINDS</span>
          </div>
          <div v-if="!briefing.bearish_lookout || briefing.bearish_lookout.length === 0" class="empty-text">
            No major corporate risk alerts today.
          </div>
          <div v-else class="items-list">
            <div
              v-for="item in briefing.bearish_lookout"
              :key="item.ticker"
              class="lookout-item"
            >
              <div class="item-header">
                <span class="ticker-badge font-mono" @click="$emit('filter-ticker', item.ticker)">
                  ${{ item.ticker }}
                </span>
                <span class="score-pill font-mono score-red">{{ item.value_score }}</span>
              </div>
              <div class="item-headline">{{ item.headline }}</div>
              <div class="item-takeaway takeaway-red">⚠️ {{ item.investment_takeaway }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Action Plan -->
      <div v-if="briefing.action_plan" class="action-plan-box">
        <div class="action-title font-mono">🎯 VALUE INVESTOR ACTION PLAN</div>
        <p class="action-text">{{ briefing.action_plan }}</p>
      </div>
    </section>

    <!-- NEWS STREAM SECTION -->
    <section class="news-section">
      <div class="section-bar">
        <h2 class="section-title">📰 Real-Time News Stream</h2>
        
        <!-- Filter Chips -->
        <div class="filter-chips">
          <button
            v-for="c in categories"
            :key="c.value"
            :class="['chip-btn', { active: activeFilter === c.value }]"
            @click="setFilter(c.value)"
          >
            {{ c.label }}
          </button>
        </div>
      </div>

      <!-- News Cards Grid -->
      <div v-if="loading" class="loading-state font-mono">
        Fetching live market intelligence...
      </div>
      <div v-else-if="filteredNews.length === 0" class="empty-state font-mono">
        No news articles matching current filter.
      </div>
      <div v-else class="news-grid">
        <article
          v-for="item in filteredNews"
          :key="item.id"
          class="news-card"
          @click="$emit('read-article', item)"
        >
          <div class="card-top">
            <div class="card-meta font-mono">
              <span v-if="item.tickers && item.tickers.length > 0" class="card-ticker">
                ${{ item.tickers.join(', $') }}
              </span>
              <span v-if="item.industry" class="card-industry">{{ item.industry }}</span>
              <span class="card-date">{{ formatDate(item.date || item.created_at) }}</span>
            </div>
            <span :class="['card-score font-mono', getScoreClass(item.value_score)]">
              {{ (item.value_score && item.value_score > 0 ? '+' : '') + (item.value_score || 0) }}
            </span>
          </div>

          <h3 class="card-title">{{ item.title }}</h3>
          <p class="card-summary">{{ item.summary }}</p>

          <div v-if="item.investment_takeaway" class="card-takeaway">
            💡 {{ item.investment_takeaway }}
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Briefing, News } from '../server/utils/types'

const props = defineProps<{
  briefing: Briefing | null
  news: News[]
  loading: boolean
}>()

defineEmits<{
  (e: 'read-article', article: News): void
  (e: 'filter-ticker', ticker: string): void
}>()

const activeFilter = ref('ALL')

const categories = [
  { label: 'All Stream', value: 'ALL' },
  { label: '🟢 Bullish', value: 'BULLISH' },
  { label: '🔴 Bearish', value: 'BEARISH' },
  { label: 'Banking', value: 'Banking' },
  { label: 'Poultry', value: 'Poultry' },
  { label: 'Mining', value: 'Mining' },
  { label: 'Energy', value: 'Energy' },
  { label: 'Consumer', value: 'Consumer Goods' },
  { label: 'Macro', value: 'Macroeconomics' }
]

const setFilter = (val: string) => {
  activeFilter.value = val
}

const filteredNews = computed(() => {
  if (activeFilter.value === 'ALL') return props.news
  if (activeFilter.value === 'BULLISH') return props.news.filter(n => n.impact_direction === 'Bullish')
  if (activeFilter.value === 'BEARISH') return props.news.filter(n => n.impact_direction === 'Bearish')
  return props.news.filter(n => n.industry?.toLowerCase() === activeFilter.value.toLowerCase())
})

const getScoreClass = (score?: number) => {
  if (score === undefined || score === 0) return 'score-neutral'
  return score > 0 ? 'score-bullish' : 'score-bearish'
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short' })
}
</script>

<style scoped>
.overview-container {
  max-width: 1440px;
  margin: 0 auto;
  padding: 24px 20px;
  display: flex;
  flex-direction: column;
  gap: 32px;
}
.hero-briefing {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
}
.briefing-header {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.pulse-tag {
  font-size: 0.75rem;
  font-weight: 700;
  color: #38bdf8;
  display: flex;
  align-items: center;
  gap: 6px;
  letter-spacing: 0.05em;
}
.live-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #38bdf8;
  box-shadow: 0 0 6px #38bdf8;
}
.briefing-headline {
  font-size: 1.6rem;
  font-weight: 700;
  line-height: 1.25;
}
.macro-text {
  color: var(--text-secondary);
  font-size: 1rem;
  line-height: 1.5;
}
.lookout-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}
.lookout-col {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
}
.bullish-card {
  border-top: 3px solid var(--bullish-border);
}
.bearish-card {
  border-top: 3px solid var(--bearish-border);
}
.col-title {
  font-size: 0.8rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  margin-bottom: 12px;
}
.items-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.lookout-item {
  background: var(--bg-card);
  padding: 12px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
}
.item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}
.ticker-badge {
  background: #1e293b;
  color: #38bdf8;
  font-weight: 700;
  font-size: 0.8rem;
  padding: 2px 6px;
  border-radius: 4px;
  cursor: pointer;
}
.score-pill {
  font-size: 0.75rem;
  font-weight: 700;
  background: var(--bullish-bg);
  color: var(--bullish-text);
  padding: 2px 6px;
  border-radius: 4px;
}
.score-pill.score-red {
  background: var(--bearish-bg);
  color: var(--bearish-text);
}
.item-headline {
  font-size: 0.9rem;
  font-weight: 600;
  margin-bottom: 4px;
}
.item-takeaway {
  font-size: 0.8rem;
  color: #fde68a;
  line-height: 1.4;
}
.item-takeaway.takeaway-red {
  color: #fca5a5;
}
.action-plan-box {
  background: rgba(37, 99, 235, 0.08);
  border: 1px solid rgba(37, 99, 235, 0.3);
  border-radius: 8px;
  padding: 14px 18px;
}
.action-title {
  font-size: 0.75rem;
  font-weight: 700;
  color: #60a5fa;
  margin-bottom: 4px;
}
.action-text {
  font-size: 0.95rem;
  color: #bfdbfe;
  line-height: 1.5;
}
.news-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.section-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}
.section-title {
  font-size: 1.25rem;
  font-weight: 700;
}
.filter-chips {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.chip-btn {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}
.chip-btn:hover {
  background: var(--bg-card-hover);
  color: #fff;
}
.chip-btn.active {
  background: #2563eb;
  color: #fff;
  border-color: #2563eb;
}
.news-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 16px;
}
.news-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.news-card:hover {
  background: var(--bg-card-hover);
  border-color: var(--border-subtle);
  transform: translateY(-2px);
}
.card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.card-meta {
  display: flex;
  gap: 6px;
  align-items: center;
  font-size: 0.75rem;
}
.card-ticker {
  color: #38bdf8;
  font-weight: 700;
}
.card-industry {
  background: #1e293b;
  color: #94a3b8;
  padding: 1px 6px;
  border-radius: 4px;
}
.card-date {
  color: var(--text-muted);
}
.card-score {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}
.score-bullish {
  background: var(--bullish-bg);
  color: var(--bullish-text);
}
.score-bearish {
  background: var(--bearish-bg);
  color: var(--bearish-text);
}
.score-neutral {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-secondary);
}
.card-title {
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.35;
}
.card-summary {
  font-size: 0.85rem;
  color: var(--text-secondary);
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.card-takeaway {
  font-size: 0.8rem;
  color: #fde68a;
  background: rgba(245, 158, 11, 0.06);
  padding: 6px 8px;
  border-radius: 4px;
  margin-top: auto;
}
.loading-state, .empty-state {
  text-align: center;
  padding: 40px;
  color: var(--text-muted);
  font-size: 0.9rem;
}
@media (max-width: 768px) {
  .lookout-grid {
    grid-template-columns: 1fr;
  }
}
</style>
```

- [ ] **Step 2: Create `idx-web/src/components/BriefingView.vue`**

```vue
<template>
  <div class="briefing-view-container">
    <div class="view-header">
      <div>
        <h1 class="view-title">🌅 Daily Market Intelligence Briefings</h1>
        <p class="view-sub">Value Investing Morning Digest ("Today's Summarization of Yesterday")</p>
      </div>
      <button v-if="briefing?.raw_markdown" class="btn-copy font-mono" @click="copyMarkdown">
        {{ copied ? '✓ Copied Markdown' : '📋 Copy Markdown' }}
      </button>
    </div>

    <div v-if="loading" class="loading-box font-mono">Loading briefings...</div>
    <div v-else-if="!briefing" class="empty-box font-mono">No briefing available.</div>
    <div v-else class="briefing-full-card">
      <div class="card-header">
        <h2 class="briefing-title">{{ briefing.title }}</h2>
        <div class="date-tag font-mono">📅 {{ formatDate(briefing.date) }}</div>
      </div>

      <!-- Macro Pulse -->
      <section class="section-block">
        <h3 class="section-heading">🌐 Executive Macro & Market Pulse</h3>
        <p class="section-content">{{ briefing.macro_pulse }}</p>
      </section>

      <!-- Lookouts Grid -->
      <div class="sections-grid">
        <section class="section-block lookout-box bullish">
          <h3 class="section-heading">🟢 Stocks to Watch (Buy Lookout)</h3>
          <div v-if="!briefing.bullish_lookout || briefing.bullish_lookout.length === 0" class="text-empty">
            No high-conviction bullish candidates today.
          </div>
          <div v-else class="cards-stack">
            <div v-for="item in briefing.bullish_lookout" :key="item.ticker" class="mini-card">
              <div class="mini-top">
                <span class="ticker font-mono">${{ item.ticker }}</span>
                <span class="score font-mono">+{{ item.value_score }}</span>
              </div>
              <div class="mini-headline">{{ item.headline }}</div>
              <div class="mini-rationale">{{ item.rationale }}</div>
              <div class="mini-takeaway">💡 {{ item.investment_takeaway }}</div>
            </div>
          </div>
        </section>

        <section class="section-block lookout-box bearish">
          <h3 class="section-heading">🔴 Risk Alerts (Headwinds & Governance)</h3>
          <div v-if="!briefing.bearish_lookout || briefing.bearish_lookout.length === 0" class="text-empty">
            No major corporate risk alerts today.
          </div>
          <div v-else class="cards-stack">
            <div v-for="item in briefing.bearish_lookout" :key="item.ticker" class="mini-card">
              <div class="mini-top">
                <span class="ticker font-mono">${{ item.ticker }}</span>
                <span class="score font-mono score-red">{{ item.value_score }}</span>
              </div>
              <div class="mini-headline">{{ item.headline }}</div>
              <div class="mini-rationale">{{ item.rationale }}</div>
              <div class="mini-takeaway red">⚠️ {{ item.investment_takeaway }}</div>
            </div>
          </div>
        </section>
      </div>

      <!-- Sector Highlights -->
      <section v-if="briefing.sector_highlights && briefing.sector_highlights.length > 0" class="section-block">
        <h3 class="section-heading">🏭 Sector & Industry Breakdown</h3>
        <div class="sector-grid">
          <div v-for="sec in briefing.sector_highlights" :key="sec.sector" class="sector-card">
            <div class="sector-top">
              <span class="sector-name">{{ sec.sector }}</span>
              <span :class="['sentiment-pill font-mono', sec.sentiment?.toLowerCase()]">{{ sec.sentiment }}</span>
            </div>
            <p class="sector-summary">{{ sec.summary }}</p>
          </div>
        </div>
      </section>

      <!-- Action Plan -->
      <section v-if="briefing.action_plan" class="section-block action-box">
        <h3 class="section-heading">🎯 Value Investor Action Plan</h3>
        <p class="section-content">{{ briefing.action_plan }}</p>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { Briefing } from '../server/utils/types'

const props = defineProps<{
  briefing: Briefing | null
  loading: boolean
}>()

const copied = ref(false)

const copyMarkdown = async () => {
  if (!props.briefing?.raw_markdown) return
  await navigator.clipboard.writeText(props.briefing.raw_markdown)
  copied.value = true
  setTimeout(() => (copied.value = false), 2000)
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-GB', {
    weekday: 'long',
    day: '2-digit',
    month: 'long',
    year: 'numeric'
  })
}
</script>

<style scoped>
.briefing-view-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px 20px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.view-title {
  font-size: 1.5rem;
  font-weight: 700;
}
.view-sub {
  color: var(--text-secondary);
  font-size: 0.9rem;
}
.btn-copy {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  padding: 8px 14px;
  border-radius: 6px;
  font-size: 0.85rem;
  cursor: pointer;
}
.btn-copy:hover {
  background: var(--bg-card-hover);
}
.briefing-full-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 28px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}
.card-header {
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 16px;
}
.briefing-title {
  font-size: 1.6rem;
  font-weight: 700;
  line-height: 1.25;
  margin-bottom: 6px;
}
.date-tag {
  color: #38bdf8;
  font-size: 0.85rem;
}
.section-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.section-heading {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--text-primary);
}
.section-content {
  color: var(--text-secondary);
  font-size: 0.95rem;
  line-height: 1.6;
}
.sections-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}
.lookout-box {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
}
.lookout-box.bullish {
  border-top: 3px solid var(--bullish-border);
}
.lookout-box.bearish {
  border-top: 3px solid var(--bearish-border);
}
.cards-stack {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 8px;
}
.mini-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  padding: 12px;
  border-radius: 6px;
}
.mini-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}
.ticker {
  color: #38bdf8;
  font-weight: 700;
  font-size: 0.85rem;
}
.score {
  font-size: 0.75rem;
  font-weight: 700;
  background: var(--bullish-bg);
  color: var(--bullish-text);
  padding: 1px 6px;
  border-radius: 4px;
}
.score.score-red {
  background: var(--bearish-bg);
  color: var(--bearish-text);
}
.mini-headline {
  font-size: 0.9rem;
  font-weight: 600;
  margin-bottom: 4px;
}
.mini-rationale {
  font-size: 0.8rem;
  color: var(--text-secondary);
  margin-bottom: 4px;
}
.mini-takeaway {
  font-size: 0.8rem;
  color: #fde68a;
}
.mini-takeaway.red {
  color: #fca5a5;
}
.sector-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
  margin-top: 8px;
}
.sector-card {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 12px;
}
.sector-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}
.sector-name {
  font-weight: 600;
  font-size: 0.9rem;
}
.sentiment-pill {
  font-size: 0.7rem;
  padding: 1px 6px;
  border-radius: 4px;
  text-transform: uppercase;
}
.sentiment-pill.bullish {
  background: var(--bullish-bg);
  color: var(--bullish-text);
}
.sentiment-pill.bearish {
  background: var(--bearish-bg);
  color: var(--bearish-text);
}
.sentiment-pill.neutral {
  background: var(--neutral-bg);
  color: var(--neutral-text);
}
.sector-summary {
  font-size: 0.8rem;
  color: var(--text-secondary);
  line-height: 1.4;
}
.action-box {
  background: rgba(37, 99, 235, 0.08);
  border: 1px solid rgba(37, 99, 235, 0.3);
  border-radius: 8px;
  padding: 16px;
}
@media (max-width: 768px) {
  .sections-grid {
    grid-template-columns: 1fr;
  }
}
</style>
```

- [ ] **Step 3: Commit Task 4**

```bash
git add src/components/OverviewView.vue src/components/BriefingView.vue
git commit -m "feat(web): add Overview dashboard and Daily Briefing reader views"
```

---

### Task 5: News Terminal, Announcements & Financial Reports Views

**Files:**
- Create: `idx-web/src/components/NewsTerminalView.vue`
- Create: `idx-web/src/components/AnnouncementsView.vue`
- Create: `idx-web/src/components/FinReportsView.vue`
- Modify: `idx-web/src/pages/index.vue`

**Interfaces:**
- `NewsTerminalView`: Filterable news stream by ticker, sector, impact direction, and value score.
- `AnnouncementsView`: Searchable IDX official disclosures.
- `FinReportsView`: Searchable quarterly financial reports archive.
- `index.vue`: Master page orchestrating views, modals, and reactive state.

- [ ] **Step 1: Create `idx-web/src/components/NewsTerminalView.vue`**

```vue
<template>
  <div class="terminal-container">
    <div class="terminal-header">
      <div>
        <h1 class="terminal-title">📰 News Intelligence Terminal</h1>
        <p class="terminal-sub">Real-time multi-channel news classified with Value Investing metrics</p>
      </div>

      <!-- Controls Bar -->
      <div class="controls-bar">
        <input
          v-model="tickerFilter"
          type="text"
          placeholder="Filter by Ticker..."
          class="control-input font-mono"
        />
        <select v-model="industryFilter" class="control-select">
          <option value="">All Industries</option>
          <option value="Banking">Banking</option>
          <option value="Poultry">Poultry</option>
          <option value="Mining">Mining</option>
          <option value="Energy">Energy</option>
          <option value="Consumer Goods">Consumer Goods</option>
          <option value="Technology">Technology</option>
          <option value="Macroeconomics">Macroeconomics</option>
        </select>
        <select v-model="directionFilter" class="control-select">
          <option value="">All Directions</option>
          <option value="Bullish">🟢 Bullish</option>
          <option value="Bearish">🔴 Bearish</option>
          <option value="Neutral">🔵 Neutral</option>
        </select>
      </div>
    </div>

    <!-- News List -->
    <div v-if="loading" class="loading-state font-mono">Loading news terminal...</div>
    <div v-else-if="filteredList.length === 0" class="empty-state font-mono">No matching articles found.</div>
    <div v-else class="terminal-grid">
      <article
        v-for="item in filteredList"
        :key="item.id"
        class="terminal-card"
        @click="$emit('read-article', item)"
      >
        <div class="card-top">
          <div class="meta font-mono">
            <span v-if="item.tickers && item.tickers.length > 0" class="ticker">
              ${{ item.tickers.join(', $') }}
            </span>
            <span v-if="item.industry" class="industry">{{ item.industry }}</span>
            <span class="date">{{ formatDate(item.date || item.created_at) }}</span>
          </div>
          <span :class="['score font-mono', getScoreClass(item.value_score)]">
            {{ (item.value_score && item.value_score > 0 ? '+' : '') + (item.value_score || 0) }}
          </span>
        </div>

        <h3 class="title">{{ item.title }}</h3>
        <p class="summary">{{ item.summary }}</p>

        <div v-if="item.investment_takeaway" class="takeaway">
          💡 {{ item.investment_takeaway }}
        </div>
      </article>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { News } from '../server/utils/types'

const props = defineProps<{
  news: News[]
  loading: boolean
  initialTicker?: string
}>()

defineEmits<{
  (e: 'read-article', article: News): void
}>()

const tickerFilter = ref(props.initialTicker || '')
const industryFilter = ref('')
const directionFilter = ref('')

const filteredList = computed(() => {
  return props.news.filter(n => {
    if (tickerFilter.value) {
      const t = tickerFilter.value.toUpperCase().trim()
      const match = n.tickers?.some(x => x.toUpperCase().includes(t)) || n.title?.toUpperCase().includes(t)
      if (!match) return false
    }
    if (industryFilter.value && n.industry?.toLowerCase() !== industryFilter.value.toLowerCase()) {
      return false
    }
    if (directionFilter.value && n.impact_direction?.toLowerCase() !== directionFilter.value.toLowerCase()) {
      return false
    }
    return true
  })
})

const getScoreClass = (score?: number) => {
  if (score === undefined || score === 0) return 'neutral'
  return score > 0 ? 'bullish' : 'bearish'
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}
</script>

<style scoped>
.terminal-container {
  max-width: 1440px;
  margin: 0 auto;
  padding: 24px 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}
.terminal-title {
  font-size: 1.5rem;
  font-weight: 700;
}
.terminal-sub {
  color: var(--text-secondary);
  font-size: 0.9rem;
}
.controls-bar {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}
.control-input, .control-select {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: #fff;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 0.85rem;
  outline: none;
}
.control-input:focus, .control-select:focus {
  border-color: #38bdf8;
}
.terminal-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 16px;
}
.terminal-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.terminal-card:hover {
  background: var(--bg-card-hover);
  border-color: var(--border-subtle);
}
.card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.meta {
  display: flex;
  gap: 6px;
  align-items: center;
  font-size: 0.75rem;
}
.ticker {
  color: #38bdf8;
  font-weight: 700;
}
.industry {
  background: #1e293b;
  color: #94a3b8;
  padding: 1px 6px;
  border-radius: 4px;
}
.date {
  color: var(--text-muted);
}
.score {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}
.score.bullish {
  background: var(--bullish-bg);
  color: var(--bullish-text);
}
.score.bearish {
  background: var(--bearish-bg);
  color: var(--bearish-text);
}
.score.neutral {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-secondary);
}
.title {
  font-size: 0.95rem;
  font-weight: 600;
  line-height: 1.35;
}
.summary {
  font-size: 0.85rem;
  color: var(--text-secondary);
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.takeaway {
  font-size: 0.8rem;
  color: #fde68a;
  background: rgba(245, 158, 11, 0.06);
  padding: 6px 8px;
  border-radius: 4px;
  margin-top: auto;
}
.loading-state, .empty-state {
  text-align: center;
  padding: 40px;
  color: var(--text-muted);
  font-size: 0.9rem;
}
</style>
```

- [ ] **Step 2: Create `idx-web/src/components/AnnouncementsView.vue`**

```vue
<template>
  <div class="view-container">
    <div class="view-header">
      <div>
        <h1 class="title">📢 IDX Official Disclosures</h1>
        <p class="sub">Corporate action announcements, material disclosures, and shareholder reports</p>
      </div>

      <input
        v-model="search"
        type="text"
        placeholder="Filter announcements (e.g. BBRI)..."
        class="search-box font-mono"
      />
    </div>

    <div v-if="loading" class="loading-box font-mono">Loading announcements...</div>
    <div v-else-if="filteredList.length === 0" class="empty-box font-mono">No announcements found.</div>
    <div v-else class="table-card">
      <table class="data-table">
        <thead>
          <tr class="font-mono">
            <th>Date</th>
            <th>No. Pengumuman</th>
            <th>Issuer / Title</th>
            <th>Attachment</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in filteredList" :key="item.id || item._id">
            <td class="font-mono date-col">{{ formatDate(item.tgl_pengumuman || item.created_at) }}</td>
            <td class="font-mono no-col">{{ item.no_pengumuman || '-' }}</td>
            <td>
              <div class="title-cell">
                <span class="announcement-title">{{ item.judul_pengumuman || item.title || 'Official Announcement' }}</span>
              </div>
            </td>
            <td>
              <div v-if="item.attachments && item.attachments.length > 0" class="attachments-row">
                <a
                  v-for="(att, i) in item.attachments"
                  :key="i"
                  :href="att.url || '#'"
                  target="_blank"
                  rel="noopener"
                  class="att-link font-mono"
                >
                  📥 {{ att.file_name || 'PDF' }}
                </a>
              </div>
              <span v-else class="text-muted">-</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Announcement } from '../server/utils/types'

const props = defineProps<{
  announcements: Announcement[]
  loading: boolean
}>()

const search = ref('')

const filteredList = computed(() => {
  if (!search.value) return props.announcements
  const q = search.value.toLowerCase()
  return props.announcements.filter(a => {
    return (
      a.judul_pengumuman?.toLowerCase().includes(q) ||
      a.no_pengumuman?.toLowerCase().includes(q) ||
      a.title?.toLowerCase().includes(q)
    )
  })
})

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}
</script>

<style scoped>
.view-container {
  max-width: 1440px;
  margin: 0 auto;
  padding: 24px 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}
.title {
  font-size: 1.5rem;
  font-weight: 700;
}
.sub {
  color: var(--text-secondary);
  font-size: 0.9rem;
}
.search-box {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: #fff;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 0.85rem;
  width: 260px;
  outline: none;
}
.search-box:focus {
  border-color: #38bdf8;
}
.table-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow-x: auto;
}
.data-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}
.data-table th {
  background: var(--bg-app);
  padding: 12px 16px;
  color: var(--text-muted);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid var(--border-color);
}
.data-table td {
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-color);
  font-size: 0.9rem;
}
.data-table tr:hover td {
  background: var(--bg-card-hover);
}
.date-col {
  color: var(--text-muted);
  font-size: 0.8rem;
  white-space: nowrap;
}
.no-col {
  color: #38bdf8;
  font-size: 0.8rem;
  white-space: nowrap;
}
.announcement-title {
  font-weight: 500;
  color: var(--text-primary);
}
.attachments-row {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.att-link {
  background: #1e293b;
  color: #38bdf8;
  padding: 2px 6px;
  border-radius: 4px;
  text-decoration: none;
  font-size: 0.75rem;
}
.att-link:hover {
  background: #2563eb;
  color: #fff;
}
.loading-box, .empty-box {
  text-align: center;
  padding: 40px;
  color: var(--text-muted);
}
</style>
```

- [ ] **Step 3: Create `idx-web/src/components/FinReportsView.vue`**

```vue
<template>
  <div class="view-container">
    <div class="view-header">
      <div>
        <h1 class="title">📊 Financial Reports Archive</h1>
        <p class="sub">Quarterly and annual financial statements downloaded directly from IDX</p>
      </div>

      <input
        v-model="search"
        type="text"
        placeholder="Filter by Ticker (e.g. BBRI)..."
        class="search-box font-mono"
      />
    </div>

    <div v-if="loading" class="loading-box font-mono">Loading financial reports...</div>
    <div v-else-if="filteredList.length === 0" class="empty-box font-mono">No financial reports found.</div>
    <div v-else class="table-card">
      <table class="data-table">
        <thead>
          <tr class="font-mono">
            <th>Issuer Code</th>
            <th>Year</th>
            <th>Quarter / Period</th>
            <th>Report File</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in filteredList" :key="item.id || item._id">
            <td class="font-mono issuer-col">${{ item.issuer_code }}</td>
            <td class="font-mono">{{ item.year || '-' }}</td>
            <td class="font-mono">{{ item.quarter ? 'Q' + item.quarter : (item.period_string || 'Annual') }}</td>
            <td>
              <a
                v-if="item.report_url"
                :href="item.report_url"
                target="_blank"
                rel="noopener"
                class="download-btn font-mono"
              >
                📥 Download Excel
              </a>
              <span v-else class="text-muted">-</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  reports: any[]
  loading: boolean
}>()

const search = ref('')

const filteredList = computed(() => {
  if (!search.value) return props.reports
  const q = search.value.toLowerCase()
  return props.reports.filter(r => r.issuer_code?.toLowerCase().includes(q))
})
</script>

<style scoped>
.view-container {
  max-width: 1440px;
  margin: 0 auto;
  padding: 24px 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}
.title {
  font-size: 1.5rem;
  font-weight: 700;
}
.sub {
  color: var(--text-secondary);
  font-size: 0.9rem;
}
.search-box {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: #fff;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 0.85rem;
  width: 260px;
  outline: none;
}
.search-box:focus {
  border-color: #38bdf8;
}
.table-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow-x: auto;
}
.data-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}
.data-table th {
  background: var(--bg-app);
  padding: 12px 16px;
  color: var(--text-muted);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid var(--border-color);
}
.data-table td {
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-color);
  font-size: 0.9rem;
}
.data-table tr:hover td {
  background: var(--bg-card-hover);
}
.issuer-col {
  color: #38bdf8;
  font-weight: 700;
}
.download-btn {
  background: #1e293b;
  color: #34d399;
  padding: 4px 10px;
  border-radius: 4px;
  text-decoration: none;
  font-size: 0.8rem;
  font-weight: 600;
  display: inline-block;
}
.download-btn:hover {
  background: #059669;
  color: #fff;
}
.loading-box, .empty-box {
  text-align: center;
  padding: 40px;
  color: var(--text-muted);
}
</style>
```

- [ ] **Step 4: Update `idx-web/src/pages/index.vue` to orchestrate views and modals**

```vue
<template>
  <div>
    <!-- Top Navbar -->
    <Navbar
      :active-tab="activeTab"
      @select-tab="activeTab = $event"
      @open-auth="showAuthModal = true"
      @search-ticker="handleGlobalSearch"
    />

    <!-- Main View Switcher -->
    <main>
      <OverviewView
        v-if="activeTab === 'overview'"
        :briefing="briefing"
        :news="newsList"
        :loading="loadingNews"
        @read-article="selectedArticle = $event"
        @filter-ticker="handleTickerClick"
      />

      <BriefingView
        v-else-if="activeTab === 'briefing'"
        :briefing="briefing"
        :loading="loadingBriefing"
      />

      <NewsTerminalView
        v-else-if="activeTab === 'news'"
        :news="newsList"
        :loading="loadingNews"
        :initial-ticker="filteredTicker"
        @read-article="selectedArticle = $event"
      />

      <AnnouncementsView
        v-else-if="activeTab === 'announcements'"
        :announcements="announcementsList"
        :loading="loadingAnnouncements"
      />

      <FinReportsView
        v-else-if="activeTab === 'reports'"
        :reports="reportsList"
        :loading="loadingReports"
      />
    </main>

    <!-- Modals -->
    <ArticleModal
      :article="selectedArticle"
      @close="selectedArticle = null"
      @filter-ticker="handleTickerClick"
    />

    <AuthModal
      v-if="showAuthModal"
      @close="showAuthModal = false"
      @authenticated="onAuthenticated"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Navbar from '../components/Navbar.vue'
import OverviewView from '../components/OverviewView.vue'
import BriefingView from '../components/BriefingView.vue'
import NewsTerminalView from '../components/NewsTerminalView.vue'
import AnnouncementsView from '../components/AnnouncementsView.vue'
import FinReportsView from '../components/FinReportsView.vue'
import ArticleModal from '../components/ArticleModal.vue'
import AuthModal from '../components/AuthModal.vue'
import { useWatchlist } from '../composables/useWatchlist'
import type { Briefing, News, Announcement } from '../server/utils/types'

const activeTab = ref('overview')
const showAuthModal = ref(false)
const selectedArticle = ref<News | null>(null)
const filteredTicker = ref('')

const briefing = ref<Briefing | null>(null)
const newsList = ref<News[]>([])
const announcementsList = ref<Announcement[]>([])
const reportsList = ref<any[]>([])

const loadingBriefing = ref(true)
const loadingNews = ref(true)
const loadingAnnouncements = ref(true)
const loadingReports = ref(true)

const { fetchWatchlist } = useWatchlist()

const loadData = async () => {
  // Fetch Latest Briefing
  try {
    const b = await $fetch<Briefing>('/api/v1/briefings/latest')
    briefing.value = b
  } catch (e) {
    console.warn('No briefing found', e)
  } finally {
    loadingBriefing.value = false
  }

  // Fetch News
  try {
    const n = await $fetch<News[]>('/api/v1/news?limit=50')
    newsList.value = n || []
  } catch (e) {
    console.error('Failed to load news', e)
  } finally {
    loadingNews.value = false
  }

  // Fetch Announcements
  try {
    const a = await $fetch<Announcement[]>('/api/v1/announcements?limit=50')
    announcementsList.value = a || []
  } catch (e) {
    console.error('Failed to load announcements', e)
  } finally {
    loadingAnnouncements.value = false
  }

  // Fetch Reports
  try {
    const r = await $fetch<any[]>('/api/v1/financial-reports?limit=50')
    reportsList.value = r || []
  } catch (e) {
    console.error('Failed to load financial reports', e)
  } finally {
    loadingReports.value = false
  }
}

const handleGlobalSearch = (ticker: string) => {
  filteredTicker.value = ticker
  activeTab.value = 'news'
}

const handleTickerClick = (ticker: string) => {
  selectedArticle.value = null
  filteredTicker.value = ticker
  activeTab.value = 'news'
}

const onAuthenticated = () => {
  fetchWatchlist()
}

onMounted(() => {
  loadData()
  fetchWatchlist()
})
</script>
```

- [ ] **Step 5: Verify build & commit**

```bash
cd idx-web && npm run build
git add src/components/ src/pages/index.vue
git commit -m "feat(web): add News Terminal, Announcements, Reports views and integrate master dashboard"
```

---

### Task 6: Verification & End-to-End Build

**Files:**
- Full frontend compilation & server verification.

- [ ] **Step 1: Run production build in `idx-web`**

Run: `cd idx-web && npm run build`  
Expected: Build passes with 0 errors.

- [ ] **Step 2: Verify dev server responds cleanly**

Run: `node -e 'const { spawn } = require("child_process"); ...'` testing `http://localhost:3000/`.  
Expected: Status 200, HTML contains IDX Intelligence Terminal header and components.

- [ ] **Step 3: Verify git status is clean**

Run: `git status`  
Expected: Clean working tree.
