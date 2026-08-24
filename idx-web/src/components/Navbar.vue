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
