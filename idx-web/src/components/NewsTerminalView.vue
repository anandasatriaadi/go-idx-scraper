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
