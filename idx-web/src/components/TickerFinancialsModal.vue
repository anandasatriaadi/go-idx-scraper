<template>
  <div v-if="ticker" class="modal-backdrop" @click.self="$emit('close')">
    <div class="modal-card">
      <!-- Modal Header -->
      <div class="modal-header">
        <div class="header-left">
          <div class="title-row">
            <span class="ticker-badge font-mono">${{ ticker }}</span>
            <h1 class="company-name">{{ latestStatement?.company_name || ticker }}</h1>
          </div>
          <div class="meta-tags font-mono">
            <span v-if="latestStatement?.metadata?.sector" class="tag sector-tag">{{ latestStatement.metadata.sector }}</span>
            <span v-if="latestStatement?.metadata?.industry" class="tag">{{ latestStatement.metadata.industry }}</span>
            <span v-if="latestStatement?.metadata?.currency" class="tag">Currency: {{ latestStatement.metadata.currency }}</span>
            <span v-if="latestStatement?.metadata?.conversion_rate" class="tag">Rate: Rp {{ latestStatement.metadata.conversion_rate }}</span>
            <span v-if="latestStatement?.year" class="tag">Latest: {{ latestStatement.year }}-{{ latestStatement.period }}</span>
          </div>
        </div>
        <button class="btn-close" @click="$emit('close')">✕</button>
      </div>

      <!-- Top Navigation Tabs -->
      <div class="modal-tab-bar font-mono">
        <button
          :class="['modal-tab-btn', { active: activeModalTab === 'financials' }]"
          @click="activeModalTab = 'financials'"
        >
          📊 Fundamentals & Valuation
        </button>
        <button
          :class="['modal-tab-btn', { active: activeModalTab === 'news' }]"
          @click="activeModalTab = 'news'"
        >
          📰 Related Sector News
          <span v-if="latestStatement?.metadata?.sector" class="tab-badge">
            {{ latestStatement.metadata.sector }}
          </span>
        </button>
      </div>

      <!-- Modal Body Scroll -->
      <div class="modal-body">
        <div v-if="loading" class="loading-state font-mono">
          <div class="spinner"></div>
          <span>Loading 360° financial intelligence for ${{ ticker }}...</span>
        </div>
        <div v-else-if="statements.length === 0" class="empty-state font-mono">
          No XBRL financial statements found for ticker ${{ ticker }}.
        </div>

        <!-- 1. Fundamentals & Valuation Tab Content -->
        <div v-else-if="activeModalTab === 'financials'" class="content-stack">
          <!-- 1. Valuation & Margin of Safety Gauge -->
          <div class="card-section valuation-card">
            <h2 class="section-title font-mono">🎯 INTRINSIC VALUE & MARGIN OF SAFETY</h2>
            <div class="valuation-grid">
              <div class="metric-box">
                <span class="metric-label">Current Stock Price</span>
                <span class="metric-value font-mono">
                  {{ latestStatement?.valuation?.current_price ? 'Rp ' + formatNum(latestStatement.valuation.current_price) : '-' }}
                </span>
              </div>
              <div class="metric-box">
                <span class="metric-label">Graham Fair Value</span>
                <span class="metric-value font-mono text-green">
                  {{ latestStatement?.valuation?.graham_number ? 'Rp ' + formatNum(latestStatement.valuation.graham_number) : '-' }}
                </span>
              </div>
              <div class="metric-box">
                <span class="metric-label">Margin of Safety</span>
                <span
                  v-if="latestStatement?.valuation?.margin_of_safety_pct !== undefined"
                  :class="['metric-value font-mono', getMosColor(latestStatement.valuation.margin_of_safety_pct)]"
                >
                  {{ (latestStatement.valuation.margin_of_safety_pct > 0 ? '+' : '') + latestStatement.valuation.margin_of_safety_pct.toFixed(1) }}%
                </span>
                <span v-else class="metric-value font-mono">-</span>
              </div>
              <div class="metric-box">
                <span class="metric-label">Normalized P/E (TTM)</span>
                <span class="metric-value font-mono">
                  {{ latestStatement?.valuation?.pe_ratio ? latestStatement.valuation.pe_ratio.toFixed(2) + 'x' : '-' }}
                </span>
              </div>
              <div class="metric-box">
                <span class="metric-label">Normalized P/B</span>
                <span class="metric-value font-mono">
                  {{ latestStatement?.valuation?.pb_ratio ? latestStatement.valuation.pb_ratio.toFixed(2) + 'x' : '-' }}
                </span>
              </div>
            </div>
          </div>

          <!-- 2. Interactive Historical Price & Benjamin Graham Valuation Bands Chart -->
          <PriceValuationChart
            :ticker="ticker"
            :graham-number="latestStatement?.valuation?.graham_number"
            :current-price="latestStatement?.valuation?.current_price"
          />

          <!-- 3. Forensic Quality & Solvency Badges -->
          <div class="card-section forensics-card">
            <h2 class="section-title font-mono">🛡️ FORENSIC HEALTH & MOAT QUALITY</h2>
            <div class="forensics-grid">
              <!-- Piotroski F-Score -->
              <div class="score-card">
                <div class="score-top">
                  <span class="score-name">Piotroski F-Score</span>
                  <span :class="['score-digit font-mono', getFScoreClass(latestStatement?.computed_ratios?.piotroski_f_score)]">
                    {{ latestStatement?.computed_ratios?.piotroski_f_score || 0 }} / 9
                  </span>
                </div>
                <div class="score-bar-bg">
                  <div
                    class="score-bar-fill"
                    :style="{ width: ((latestStatement?.computed_ratios?.piotroski_f_score || 0) / 9 * 100) + '%' }"
                  ></div>
                </div>
                <div class="score-desc">
                  {{ getFScoreDesc(latestStatement?.computed_ratios?.piotroski_f_score) }}
                </div>
              </div>

              <!-- Altman Z-Score -->
              <div class="score-card">
                <div class="score-top">
                  <span class="score-name">Altman Z''-Score (Bankruptcy Risk)</span>
                  <span :class="['score-digit font-mono', getZScoreClass(latestStatement?.computed_ratios?.altman_z_score)]">
                    {{ latestStatement?.computed_ratios?.altman_z_score?.toFixed(2) || '0.00' }}
                  </span>
                </div>
                <div class="score-desc">
                  {{ getZScoreDesc(latestStatement?.computed_ratios?.altman_z_score) }}
                </div>
              </div>

              <!-- ROIC Moat Gauge -->
              <div class="score-card">
                <div class="score-top">
                  <span class="score-name">ROIC (Return on Invested Capital)</span>
                  <span :class="['score-digit font-mono', (latestStatement?.computed_ratios?.roic || 0) >= 0.15 ? 'text-green' : '']">
                    {{ latestStatement?.computed_ratios?.roic ? (latestStatement.computed_ratios.roic * 100).toFixed(1) + '%' : '-' }}
                  </span>
                </div>
                <div class="score-desc">
                  {{ (latestStatement?.computed_ratios?.roic || 0) >= 0.15 ? 'Wide economic moat and compounding power' : 'Moderate capital returns' }}
                </div>
              </div>

              <!-- Debt to Equity -->
              <div class="score-card">
                <div class="score-top">
                  <span class="score-name">Debt / Equity</span>
                  <span class="score-digit font-mono">
                    {{ latestStatement?.computed_ratios?.debt_to_equity ? latestStatement.computed_ratios.debt_to_equity.toFixed(2) : '0.00' }}
                  </span>
                </div>
                <div class="score-desc">
                  {{ (latestStatement?.computed_ratios?.debt_to_equity || 0) <= 0.5 ? 'Conservative fortress balance sheet' : 'Moderate leverage' }}
                </div>
              </div>
            </div>
          </div>

          <!-- 3. Multi-Period Statement Trends -->
          <div class="card-section statement-trends-card">
            <div class="table-nav">
              <h2 class="section-title font-mono">📈 FINANCIAL STATEMENT HISTORY</h2>
              <div class="tab-pills">
                <button :class="['pill-btn', { active: statementTab === 'income' }]" @click="statementTab = 'income'">Income Statement</button>
                <button :class="['pill-btn', { active: statementTab === 'balance' }]" @click="statementTab = 'balance'">Balance Sheet</button>
                <button :class="['pill-btn', { active: statementTab === 'cashflow' }]" @click="statementTab = 'cashflow'">Cash Flow</button>
              </div>
            </div>

            <div class="table-wrap">
              <table class="trend-table">
                <thead>
                  <tr class="font-mono">
                    <th class="line-item-head">Line Item</th>
                    <th v-for="s in statements" :key="s.id" class="period-head">
                      {{ s.year }}-{{ s.period }}
                    </th>
                  </tr>
                </thead>
                <tbody v-if="statementTab === 'income'">
                  <tr>
                    <td class="item-name">Revenue / Sales</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono">{{ formatCompact(s.core?.revenue) }}</td>
                  </tr>
                  <tr>
                    <td class="item-name">Gross Profit</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono">{{ formatCompact(s.core?.gross_profit) }}</td>
                  </tr>
                  <tr>
                    <td class="item-name">Operating Income (EBIT)</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono">{{ formatCompact(s.core?.operating_income) }}</td>
                  </tr>
                  <tr>
                    <td class="item-name">Finance Costs</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono">{{ formatCompact(s.core?.finance_costs) }}</td>
                  </tr>
                  <tr class="highlight-row">
                    <td class="item-name">Net Income</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono text-green">{{ formatCompact(s.core?.net_income) }}</td>
                  </tr>
                </tbody>

                <tbody v-else-if="statementTab === 'balance'">
                  <tr>
                    <td class="item-name">Cash & Cash Equivalents</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono">{{ formatCompact(s.core?.cash_and_equivalents) }}</td>
                  </tr>
                  <tr>
                    <td class="item-name">Current Assets</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono">{{ formatCompact(s.core?.current_assets) }}</td>
                  </tr>
                  <tr class="highlight-row">
                    <td class="item-name">Total Assets</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono">{{ formatCompact(s.core?.total_assets) }}</td>
                  </tr>
                  <tr>
                    <td class="item-name">Total Debt (Short + Long)</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono">{{ formatCompact(s.core?.total_debt) }}</td>
                  </tr>
                  <tr>
                    <td class="item-name">Total Liabilities</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono">{{ formatCompact(s.core?.total_liabilities) }}</td>
                  </tr>
                  <tr class="highlight-row">
                    <td class="item-name">Total Equity</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono text-green">{{ formatCompact(s.core?.total_equity) }}</td>
                  </tr>
                </tbody>

                <tbody v-else-if="statementTab === 'cashflow'">
                  <tr>
                    <td class="item-name">Operating Cash Flow (CFO)</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono">{{ formatCompact(s.core?.operating_cash_flow) }}</td>
                  </tr>
                  <tr>
                    <td class="item-name">Capital Expenditures (CapEx)</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono">{{ formatCompact(s.core?.capex) }}</td>
                  </tr>
                  <tr class="highlight-row">
                    <td class="item-name">Free Cash Flow (FCF)</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono text-green">{{ formatCompact(s.core?.free_cash_flow) }}</td>
                  </tr>
                  <tr>
                    <td class="item-name">FCF Conversion %</td>
                    <td v-for="s in statements" :key="s.id" class="font-mono">
                      {{ s.computed_ratios?.fcf_conversion_pct ? s.computed_ratios.fcf_conversion_pct.toFixed(1) + '%' : '-' }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- 2. Related Sector News Tab Content -->
        <div v-else-if="activeModalTab === 'news'" class="content-stack">
          <div class="card-section sector-news-card">
            <div class="news-header-row">
              <div>
                <h2 class="section-title font-mono">
                  📰 INDUSTRY & SECTOR DEVELOPMENTS: {{ latestStatement?.metadata?.sector || 'GENERAL' }}
                </h2>
                <p class="section-sub">
                  Real-time intelligence on macro developments and peer activities affecting {{ latestStatement?.company_name || ticker }}
                </p>
              </div>
              <button
                class="refresh-btn font-mono"
                :disabled="loadingSectorNews"
                @click="fetchSectorNews(latestStatement?.metadata?.sector, ticker || undefined)"
              >
                ↻ Refresh Stream
              </button>
            </div>

            <!-- News List -->
            <div v-if="loadingSectorNews" class="loading-state font-mono">
              <div class="spinner"></div>
              <span>Fetching sector news for {{ latestStatement?.metadata?.sector }}...</span>
            </div>
            <div v-else-if="sectorNews.length === 0" class="empty-state font-mono">
              No recent sector news found for {{ latestStatement?.metadata?.sector || 'this company' }}.
            </div>
            <div v-else class="sector-news-grid">
              <article
                v-for="item in sectorNews"
                :key="item.id"
                class="sector-news-item"
              >
                <div class="item-top">
                  <div class="item-meta font-mono">
                    <span v-if="item.tickers && item.tickers.length > 0" class="ticker-pill">
                      ${{ item.tickers.join(', $') }}
                    </span>
                    <span v-if="item.subsector || item.industry" class="industry-pill">
                      {{ item.subsector || item.industry }}
                    </span>
                    <span class="date-pill">{{ formatDate(item.date || item.created_at) }}</span>
                  </div>
                  <span :class="['score-pill font-mono', getScoreClass(item.value_score)]">
                    {{ (item.value_score && item.value_score > 0 ? '+' : '') + (item.value_score || 0) }}
                  </span>
                </div>

                <h3 class="news-item-title">
                  <a
                    v-if="item.link"
                    :href="item.link"
                    target="_blank"
                    rel="noopener"
                    class="news-link"
                  >
                    {{ item.title }} ↗
                  </a>
                  <span v-else>{{ item.title }}</span>
                </h3>

                <p class="news-item-summary">{{ item.summary }}</p>

                <div v-if="item.investment_takeaway" class="item-takeaway">
                  💡 {{ item.investment_takeaway }}
                </div>
              </article>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import PriceValuationChart from './PriceValuationChart.vue'
import type { XBRLStatement, News } from '../server/utils/types'

const props = defineProps<{
  ticker: string | null
}>()

defineEmits<{
  (e: 'close'): void
}>()

const loading = ref(false)
const statements = ref<XBRLStatement[]>([])
const statementTab = ref<'income' | 'balance' | 'cashflow'>('income')
const activeModalTab = ref<'financials' | 'news'>('financials')

const sectorNews = ref<News[]>([])
const loadingSectorNews = ref(false)

const latestStatement = computed(() => {
  if (statements.value.length === 0) return null
  return statements.value[0]
})

const fetchSectorNews = async (sectorName?: string, tickerCode?: string) => {
  loadingSectorNews.value = true
  try {
    const params = new URLSearchParams()
    if (sectorName) {
      params.append('sector', sectorName)
    } else if (tickerCode) {
      params.append('ticker', tickerCode)
    }
    params.append('limit', '20')

    const res = await $fetch<{
      data: News[]
      total: number
      page: number
      total_pages: number
    } | News[]>(`/api/v1/news?${params.toString()}`)

    if (res && 'data' in res && Array.isArray(res.data)) {
      sectorNews.value = res.data
    } else if (Array.isArray(res)) {
      sectorNews.value = res
    } else {
      sectorNews.value = []
    }
  } catch (err) {
    console.error('Failed to fetch sector news', err)
    sectorNews.value = []
  } finally {
    loadingSectorNews.value = false
  }
}

const fetchFinancials = async (t: string) => {
  loading.value = true
  try {
    const data = await $fetch<XBRLStatement[]>(`/api/v1/stocks/${t}/financials`)
    statements.value = data || []
    if (statements.value.length > 0 && statements.value[0].metadata?.sector) {
      fetchSectorNews(statements.value[0].metadata.sector, t)
    } else {
      fetchSectorNews(undefined, t)
    }
  } catch (e) {
    console.error('Failed to fetch ticker financials', e)
    statements.value = []
  } finally {
    loading.value = false
  }
}

watch(
  () => props.ticker,
  (newTicker) => {
    if (newTicker) {
      activeModalTab.value = 'financials'
      fetchFinancials(newTicker)
    }
  },
  { immediate: true }
)

const formatNum = (val: number) => {
  return new Intl.NumberFormat('id-ID', { maximumFractionDigits: 0 }).format(val)
}

const formatCompact = (val?: number) => {
  if (val === undefined || isNaN(val)) return '-'
  if (Math.abs(val) >= 1e12) return (val / 1e12).toFixed(2) + ' T'
  if (Math.abs(val) >= 1e9) return (val / 1e9).toFixed(2) + ' B'
  if (Math.abs(val) >= 1e6) return (val / 1e6).toFixed(2) + ' M'
  return val.toFixed(0)
}

const getMosColor = (mos: number) => {
  if (mos >= 30) return 'text-green'
  if (mos > 0) return 'text-amber'
  return 'text-red'
}

const getFScoreClass = (score?: number) => {
  if (!score) return 'score-red'
  if (score >= 7) return 'score-green'
  if (score >= 5) return 'score-amber'
  return 'score-red'
}

const getFScoreDesc = (score?: number) => {
  if (!score) return 'Weak accounting fundamentals'
  if (score >= 7) return 'Elite financial health & capital discipline'
  if (score >= 5) return 'Moderate financial stability'
  return 'Potential accounting/leverage risks'
}

const getZScoreClass = (score?: number) => {
  if (!score) return 'score-red'
  if (score >= 2.6) return 'score-green'
  if (score >= 1.1) return 'score-amber'
  return 'score-red'
}

const getZScoreDesc = (score?: number) => {
  if (!score) return 'Distress risk'
  if (score >= 2.6) return 'Safe Zone (Minimal bankruptcy/default risk)'
  if (score >= 1.1) return 'Grey Zone (Monitor debt covenants)'
  return 'Distress Zone (High bankruptcy/default probability)'
}

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
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(6px);
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
  max-width: 1050px;
  max-height: 92vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.7);
}
.modal-header {
  padding: 20px 24px 16px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}
.header-left {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.ticker-badge {
  background: #2563eb;
  color: #fff;
  font-size: 1.1rem;
  font-weight: 800;
  padding: 2px 8px;
  border-radius: 6px;
}
.company-name {
  font-size: 1.4rem;
  font-weight: 700;
  line-height: 1.2;
}
.meta-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.tag {
  background: #1e293b;
  color: #94a3b8;
  font-size: 0.75rem;
  padding: 2px 8px;
  border-radius: 4px;
}
.sector-tag {
  background: rgba(56, 189, 248, 0.15);
  color: #38bdf8;
  font-weight: 600;
}
.btn-close {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 1.3rem;
  cursor: pointer;
}
.btn-close:hover {
  color: #fff;
}
.modal-tab-bar {
  display: flex;
  gap: 4px;
  padding: 10px 24px;
  background: var(--bg-app);
  border-bottom: 1px solid var(--border-color);
}
.modal-tab-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.15s ease;
}
.modal-tab-btn:hover {
  background: var(--bg-card);
  color: #fff;
}
.modal-tab-btn.active {
  background: #2563eb;
  color: #fff;
}
.tab-badge {
  background: rgba(0, 0, 0, 0.3);
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 0.7rem;
}
.modal-body {
  padding: 24px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.content-stack {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.card-section {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 18px 20px;
}
.section-title {
  font-size: 0.85rem;
  font-weight: 700;
  color: #38bdf8;
  letter-spacing: 0.05em;
  margin-bottom: 4px;
}
.section-sub {
  color: var(--text-secondary);
  font-size: 0.8rem;
  margin-bottom: 16px;
}
.news-header-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 12px;
}
.refresh-btn {
  background: #1e293b;
  border: 1px solid var(--border-subtle);
  color: #38bdf8;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s ease;
}
.refresh-btn:hover:not(:disabled) {
  background: #2563eb;
  color: #fff;
}
.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.valuation-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 16px;
  margin-top: 10px;
}
.metric-box {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.metric-label {
  font-size: 0.75rem;
  color: var(--text-muted);
}
.metric-value {
  font-size: 1.15rem;
  font-weight: 700;
}
.forensics-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-top: 10px;
}
.score-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.score-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.score-name {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-secondary);
}
.score-digit {
  font-size: 1.1rem;
  font-weight: 800;
}
.score-bar-bg {
  height: 6px;
  background: #1e293b;
  border-radius: 3px;
  overflow: hidden;
}
.score-bar-fill {
  height: 100%;
  background: #10b981;
  border-radius: 3px;
}
.score-desc {
  font-size: 0.75rem;
  color: var(--text-muted);
}
.table-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 12px;
}
.tab-pills {
  display: flex;
  gap: 4px;
  background: var(--bg-card);
  padding: 3px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
}
.pill-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 0.75rem;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 4px;
  cursor: pointer;
}
.pill-btn.active {
  background: #2563eb;
  color: #fff;
}
.table-wrap {
  overflow-x: auto;
}
.trend-table {
  width: 100%;
  border-collapse: collapse;
  text-align: right;
  font-size: 0.85rem;
}
.trend-table th {
  background: var(--bg-card);
  padding: 10px 14px;
  color: var(--text-muted);
  font-size: 0.75rem;
  border-bottom: 1px solid var(--border-color);
}
.trend-table td {
  padding: 10px 14px;
  border-bottom: 1px solid var(--border-color);
}
.line-item-head, .item-name {
  text-align: left;
  font-weight: 500;
  color: var(--text-secondary);
}
.highlight-row td {
  font-weight: 700;
  background: rgba(255, 255, 255, 0.02);
}
.sector-news-grid {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.sector-news-item {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.item-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.item-meta {
  display: flex;
  gap: 6px;
  align-items: center;
  font-size: 0.75rem;
  flex-wrap: wrap;
}
.ticker-pill {
  color: #38bdf8;
  font-weight: 700;
}
.industry-pill {
  background: #1e293b;
  color: #94a3b8;
  padding: 1px 6px;
  border-radius: 4px;
}
.date-pill {
  color: var(--text-muted);
}
.score-pill {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}
.score-pill.bullish {
  background: var(--bullish-bg);
  color: var(--bullish-text);
}
.score-pill.bearish {
  background: var(--bearish-bg);
  color: var(--bearish-text);
}
.score-pill.neutral {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-secondary);
}
.news-item-title {
  font-size: 0.95rem;
  font-weight: 600;
  line-height: 1.35;
}
.news-link {
  color: #f8fafc;
  text-decoration: none;
}
.news-link:hover {
  color: #38bdf8;
}
.news-item-summary {
  font-size: 0.85rem;
  color: var(--text-secondary);
  line-height: 1.45;
}
.item-takeaway {
  font-size: 0.8rem;
  color: #fde68a;
  background: rgba(245, 158, 11, 0.06);
  padding: 6px 8px;
  border-radius: 4px;
}
.text-green { color: #34d399; }
.text-amber { color: #fbbf24; }
.text-red { color: #f87171; }
.score-green { color: #34d399; }
.score-amber { color: #fbbf24; }
.score-red { color: #f87171; }
.loading-state, .empty-state {
  text-align: center;
  padding: 60px;
  color: var(--text-muted);
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
@media (max-width: 768px) {
  .forensics-grid {
    grid-template-columns: 1fr;
  }
}
</style>
