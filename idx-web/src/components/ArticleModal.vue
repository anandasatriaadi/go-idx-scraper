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
