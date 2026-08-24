<template>
  <div class="terminal-container">
    <div class="terminal-header">
      <div>
        <h1 class="terminal-title">📰 News Intelligence Terminal</h1>
        <p class="terminal-sub">Real-time multi-channel news classified with official IDX-IC taxonomy and Value Investing metrics</p>
      </div>

      <!-- Controls Bar -->
      <div class="controls-bar">
        <!-- Ticker Search -->
        <div class="input-wrap">
          <input
            v-model="tickerFilter"
            type="text"
            placeholder="Filter by Ticker (e.g. BBRI)..."
            class="control-input font-mono"
            @input="onFilterChanged"
          />
          <button v-if="tickerFilter" class="clear-icon" @click="clearTicker">✕</button>
        </div>

        <!-- 11 IDX-IC Sectors + Macroeconomics -->
        <select v-model="sectorFilter" class="control-select" @change="onSectorChange">
          <option value="">All Sectors</option>
          <option v-for="s in sectorList" :key="s" :value="s">{{ s }}</option>
        </select>

        <!-- Dynamic Subsector Dropdown -->
        <select
          v-model="subsectorFilter"
          class="control-select"
          :disabled="!availableSubsectors.length"
          @change="onFilterChanged"
        >
          <option value="">{{ availableSubsectors.length ? 'All Subsectors' : 'Select Sector first' }}</option>
          <option v-for="sub in availableSubsectors" :key="sub" :value="sub">{{ sub }}</option>
        </select>

        <!-- Impact Direction -->
        <select v-model="directionFilter" class="control-select" @change="onFilterChanged">
          <option value="">All Directions</option>
          <option value="Bullish">🟢 Bullish</option>
          <option value="Bearish">🔴 Bearish</option>
          <option value="Neutral">🔵 Neutral</option>
        </select>
      </div>
    </div>

    <!-- News List -->
    <div v-if="loading && items.length === 0" class="loading-state font-mono">
      <div class="spinner"></div>
      <span>Loading news terminal...</span>
    </div>
    <div v-else-if="filteredList.length === 0" class="empty-state font-mono">
      No matching news articles found for selected filters.
    </div>
    <div v-else class="terminal-grid-wrap">
      <div v-if="loading" class="grid-overlay">
        <span class="overlay-text font-mono">Refreshing terminal...</span>
      </div>

      <div class="terminal-grid">
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
              <span v-if="item.sector" class="sector-tag">{{ item.sector }}</span>
              <span v-if="item.subsector || item.industry" class="industry">
                {{ item.subsector || item.industry }}
              </span>
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

    <!-- Pagination Bar -->
    <div v-if="totalPages > 1 || total > limit" class="pagination-bar font-mono">
      <button
        class="pagination-btn"
        :disabled="page <= 1 || loading"
        @click="changePage(page - 1)"
      >
        ◀ Previous
      </button>

      <div class="pagination-info">
        Page <span class="page-current">{{ page }}</span> of <span class="page-total">{{ totalPages || 1 }}</span>
        <span class="items-total">({{ total }} Total)</span>
      </div>

      <button
        class="pagination-btn"
        :disabled="page >= totalPages || loading"
        @click="changePage(page + 1)"
      >
        Next ▶
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import type { News } from '../server/utils/types'

const IDX_IC_TAXONOMY: Record<string, string[]> = {
  'A. Energy': ['A1. Oil, Gas, and Coal', 'A2. Alternative Energy'],
  'B. Basic Materials': ['B1. Basic Materials'],
  'C. Industrials': ['C1. Industrial Goods', 'C2. Industrial Services', 'C3. Multi-sector Holdings'],
  'D. Consumer Non-Cyclicals': [
    'D1. Food and Staples Retailing',
    'D2. Food and Beverage',
    'D3. Tobacco',
    'D4. Nondurable Household Products'
  ],
  'E. Consumer Cyclicals': [
    'E1. Automobiles and Components',
    'E2. Household Goods',
    'E3. Leisure Goods',
    'E4. Apparel and Luxury Goods',
    'E5. Consumer Services',
    'E6. Media and Entertainment',
    'E7. Retailing'
  ],
  'F. Healthcare': [
    'F1. Healthcare Equipment & Providers',
    'F2. Pharmaceuticals & Health Care Research'
  ],
  'G. Financials': [
    'G1. Banks',
    'G2. Financing Service',
    'G3. Investment Service',
    'G4. Insurance',
    'G5. Holding and Investment Companies'
  ],
  'H. Properties and Real Estate': ['H1. Properties & Real Estate'],
  'I. Technology': ['I1. Software & IT Services', 'I2. Technology Hardware & Equipment'],
  'J. Infrastructures': [
    'J1. Transportation Infrastructure',
    'J2. Heavy Constructions & Civil Engineering',
    'J3. Telecommunication',
    'J4. Utilities'
  ],
  'K. Transportation and Logistic': ['K1. Transportation', 'K2. Logistics & Deliveries'],
  'Macroeconomics': ['General Market & Policy']
}

const IDX_IC_SECTORS = Object.keys(IDX_IC_TAXONOMY)

const props = defineProps<{
  news?: News[]
  loading?: boolean
  initialTicker?: string
}>()

defineEmits<{
  (e: 'read-article', article: News): void
}>()

const sectorList = IDX_IC_SECTORS
const items = ref<News[]>(props.news || [])
const loading = ref(props.loading || false)

const tickerFilter = ref(props.initialTicker || '')
const sectorFilter = ref('')
const subsectorFilter = ref('')
const directionFilter = ref('')

const page = ref(1)
const limit = ref(24)
const total = ref(props.news?.length || 0)
const totalPages = ref(1)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

const availableSubsectors = computed(() => {
  if (!sectorFilter.value) return []
  return IDX_IC_TAXONOMY[sectorFilter.value] || []
})

const onSectorChange = () => {
  subsectorFilter.value = ''
  onFilterChanged()
}

const onFilterChanged = () => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    page.value = 1
    fetchNews(1)
  }, 300)
}

const clearTicker = () => {
  tickerFilter.value = ''
  onFilterChanged()
}

const fetchNews = async (p = page.value) => {
  loading.value = true
  try {
    const params = new URLSearchParams()
    params.append('page', p.toString())
    params.append('limit', limit.value.toString())

    if (tickerFilter.value.trim()) {
      params.append('ticker', tickerFilter.value.trim())
    }
    if (sectorFilter.value) {
      params.append('sector', sectorFilter.value)
    }
    if (subsectorFilter.value) {
      params.append('subsector', subsectorFilter.value)
    }

    const res = await $fetch<{
      data: News[]
      total: number
      page: number
      total_pages: number
    }>(`/api/v1/news?${params.toString()}`)

    items.value = res.data || []
    total.value = res.total || 0
    totalPages.value = res.total_pages || 1
    page.value = res.page || p
  } catch (err) {
    console.error('Failed to fetch news terminal data', err)
  } finally {
    loading.value = false
  }
}

const changePage = (newPage: number) => {
  if (newPage < 1 || newPage > totalPages.value || newPage === page.value) return
  page.value = newPage
  fetchNews(newPage)
}

watch(() => props.initialTicker, (val) => {
  if (val !== undefined && val !== tickerFilter.value) {
    tickerFilter.value = val
    page.value = 1
    fetchNews(1)
  }
})

watch(() => props.news, (newVal) => {
  if (newVal && items.value.length === 0 && !tickerFilter.value && !sectorFilter.value) {
    items.value = newVal
    total.value = newVal.length
  }
})

const filteredList = computed(() => {
  // If direction filter is active on top of server query, filter client-side
  if (!directionFilter.value) return items.value
  return items.value.filter(n => n.impact_direction?.toLowerCase() === directionFilter.value.toLowerCase())
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

onMounted(() => {
  fetchNews(1)
})
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
  align-items: center;
}
.input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}
.control-input, .control-select {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: #fff;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 0.85rem;
  outline: none;
  transition: border-color 0.15s ease;
}
.control-input {
  width: 220px;
  padding-right: 28px;
}
.control-input:focus, .control-select:focus {
  border-color: #38bdf8;
}
.control-select:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.clear-icon {
  position: absolute;
  right: 8px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 0.8rem;
}
.clear-icon:hover {
  color: #fff;
}
.terminal-grid-wrap {
  position: relative;
  min-height: 200px;
}
.grid-overlay {
  position: absolute;
  inset: 0;
  background: rgba(8, 12, 20, 0.6);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  border-radius: 8px;
}
.overlay-text {
  color: #38bdf8;
  font-size: 0.85rem;
  background: var(--bg-card);
  padding: 6px 14px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
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
  transform: translateY(-2px);
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
  flex-wrap: wrap;
}
.ticker {
  color: #38bdf8;
  font-weight: 700;
}
.sector-tag {
  background: rgba(56, 189, 248, 0.12);
  color: #38bdf8;
  padding: 1px 6px;
  border-radius: 4px;
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
.pagination-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  gap: 12px;
}
.pagination-btn {
  background: #1e293b;
  border: 1px solid var(--border-subtle);
  color: #38bdf8;
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}
.pagination-btn:hover:not(:disabled) {
  background: #2563eb;
  color: #fff;
  border-color: #2563eb;
}
.pagination-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.pagination-info {
  color: var(--text-secondary);
  font-size: 0.85rem;
  display: flex;
  align-items: center;
  gap: 6px;
}
.page-current {
  color: #38bdf8;
  font-weight: 700;
}
.page-total {
  color: var(--text-primary);
  font-weight: 600;
}
.items-total {
  color: var(--text-muted);
  font-size: 0.8rem;
}
.loading-state, .empty-state {
  text-align: center;
  padding: 60px;
  color: var(--text-muted);
  font-size: 0.9rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}
.spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--border-color);
  border-top-color: #38bdf8;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
