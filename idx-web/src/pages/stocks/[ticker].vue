<template>
  <div class="stock-page-root">
    <Navbar
      active-tab="screener"
      @select-tab="handleNavTab"
      @open-auth="showAuthModal = true"
      @search-ticker="handleSearch"
    />

    <main class="stock-main-content">
      <TickerFinancialsModal
        v-if="tickerParam"
        :ticker="tickerParam"
        @close="handleClose"
      />
    </main>

    <AuthModal
      v-if="showAuthModal"
      @close="showAuthModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import Navbar from '../../components/Navbar.vue'
import TickerFinancialsModal from '../../components/TickerFinancialsModal.vue'
import AuthModal from '../../components/AuthModal.vue'

const route = useRoute()
const router = useRouter()
const showAuthModal = ref(false)

const tickerParam = computed(() => {
  const t = route.params.ticker
  return Array.isArray(t) ? t[0] : (t ? String(t).toUpperCase() : '')
})

const handleNavTab = (tab: string) => {
  router.push({ path: '/', query: { tab } })
}

const handleSearch = (ticker: string) => {
  router.push({ path: `/stocks/${ticker.toUpperCase().trim()}` })
}

const handleClose = () => {
  router.push({ path: '/' })
}
</script>

<style scoped>
.stock-page-root {
  min-height: 100vh;
  background: var(--bg-app);
}
.stock-main-content {
  max-width: 1440px;
  margin: 0 auto;
}
</style>
