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
