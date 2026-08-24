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
