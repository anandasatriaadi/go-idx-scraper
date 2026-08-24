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
