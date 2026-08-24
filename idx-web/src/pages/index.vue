<template>
  <div>
    <!-- Top Navbar -->
    <Navbar
      :active-tab="activeTab"
      @select-tab="setTab"
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

      <ValueScreenerView
        v-else-if="activeTab === 'screener'"
        @open-ticker-financials="openTicker"
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
        @open-ticker-financials="openTicker"
      />
    </main>

    <!-- Modals -->
    <ArticleModal
      :article="selectedArticle"
      @close="selectedArticle = null"
      @filter-ticker="handleTickerClick"
    />

    <TickerFinancialsModal
      :ticker="selectedTickerForFinancials"
      @close="closeTicker"
    />

    <AuthModal
      v-if="showAuthModal"
      @close="showAuthModal = false"
      @authenticated="onAuthenticated"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import Navbar from '../components/Navbar.vue'
import OverviewView from '../components/OverviewView.vue'
import BriefingView from '../components/BriefingView.vue'
import NewsTerminalView from '../components/NewsTerminalView.vue'
import AnnouncementsView from '../components/AnnouncementsView.vue'
import FinReportsView from '../components/FinReportsView.vue'
import ValueScreenerView from '../components/ValueScreenerView.vue'
import TickerFinancialsModal from '../components/TickerFinancialsModal.vue'
import ArticleModal from '../components/ArticleModal.vue'
import AuthModal from '../components/AuthModal.vue'
import { useWatchlist } from '../composables/useWatchlist'
import type { Briefing, News, Announcement, FinancialReport } from '../server/utils/types'

interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  total_pages: number
}

const route = useRoute()
const router = useRouter()

const activeTab = ref((route.query.tab as string) || 'overview')
const showAuthModal = ref(false)
const selectedArticle = ref<News | null>(null)
const selectedTickerForFinancials = ref<string | null>((route.query.ticker as string) || null)
const filteredTicker = ref((route.query.ticker as string) || '')

const briefing = ref<Briefing | null>(null)
const newsList = ref<News[]>([])
const announcementsList = ref<Announcement[]>([])
const reportsList = ref<FinancialReport[]>([])

const loadingBriefing = ref(true)
const loadingNews = ref(true)
const loadingAnnouncements = ref(true)
const loadingReports = ref(true)

const { fetchWatchlist } = useWatchlist()

const setTab = (tab: string) => {
  activeTab.value = tab
  router.push({ query: { ...route.query, tab } })
}

const openTicker = (ticker: string) => {
  const t = ticker.toUpperCase().trim()
  selectedTickerForFinancials.value = t
  filteredTicker.value = t
  router.push({ query: { ...route.query, ticker: t } })
}

const closeTicker = () => {
  selectedTickerForFinancials.value = null
  const q = { ...route.query }
  delete q.ticker
  router.push({ query: q })
}

const handleGlobalSearch = (ticker: string) => {
  openTicker(ticker)
}

const handleTickerClick = (ticker: string) => {
  selectedArticle.value = null
  openTicker(ticker)
}

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
    const n = await $fetch<PaginatedResponse<News> | News[]>('/api/v1/news?limit=50')
    if (n && 'data' in n && Array.isArray(n.data)) {
      newsList.value = n.data
    } else if (Array.isArray(n)) {
      newsList.value = n
    } else {
      newsList.value = []
    }
  } catch (e) {
    console.error('Failed to load news', e)
    newsList.value = []
  } finally {
    loadingNews.value = false
  }

  // Fetch Announcements
  try {
    const a = await $fetch<PaginatedResponse<Announcement> | Announcement[]>('/api/v1/announcements?limit=20')
    if (a && 'data' in a && Array.isArray(a.data)) {
      announcementsList.value = a.data
    } else if (Array.isArray(a)) {
      announcementsList.value = a
    } else {
      announcementsList.value = []
    }
  } catch (e) {
    console.error('Failed to load announcements', e)
    announcementsList.value = []
  } finally {
    loadingAnnouncements.value = false
  }

  // Fetch Reports
  try {
    const r = await $fetch<PaginatedResponse<FinancialReport> | FinancialReport[]>('/api/v1/financial-reports?limit=20')
    if (r && 'data' in r && Array.isArray(r.data)) {
      reportsList.value = r.data
    } else if (Array.isArray(r)) {
      reportsList.value = r
    } else {
      reportsList.value = []
    }
  } catch (e) {
    console.error('Failed to load financial reports', e)
    reportsList.value = []
  } finally {
    loadingReports.value = false
  }
}

const onAuthenticated = () => {
  fetchWatchlist()
}

watch(() => route.query.tab, (newTab) => {
  if (newTab && typeof newTab === 'string') {
    activeTab.value = newTab
  }
})

watch(() => route.query.ticker, (newTicker) => {
  if (newTicker && typeof newTicker === 'string') {
    selectedTickerForFinancials.value = newTicker.toUpperCase().trim()
    filteredTicker.value = newTicker.toUpperCase().trim()
  } else if (!newTicker) {
    selectedTickerForFinancials.value = null
  }
})

onMounted(() => {
  loadData()
  fetchWatchlist()
})
</script>
