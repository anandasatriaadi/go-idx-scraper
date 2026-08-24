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
            <span v-if="latestStatement?.year" class="tag">Latest: {{ latestStatement.year }} {{ latestStatement.period }}</span>
          </div>
        </div>
        <button class="btn-close" @click="$emit('close')">✕</button>
      </div>

      <!-- Top Navigation Tabs -->
      <div class="modal-tab-bar font-mono">
        <button
          :class="['modal-tab-btn', { active: activeModalTab === 'terminal' }]"
          @click="activeModalTab = 'terminal'"
        >
          📊 Institutional Matrix
        </button>
        <button
          :class="['modal-tab-btn', { active: activeModalTab === 'chart' }]"
          @click="activeModalTab = 'chart'"
        >
          📈 Price & Graham Valuation Chart
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
          <span>Loading financial intelligence for ${{ ticker }}...</span>
        </div>
        <div v-else-if="statements.length === 0" class="empty-state font-mono">
          No XBRL financial statements found for ticker ${{ ticker }}.
        </div>

        <!-- 1. INSTITUTIONAL 3-COLUMN MATRIX (Stockbit Style) -->
        <div v-else-if="activeModalTab === 'terminal'" class="matrix-layout">
          <!-- LEFT COLUMN: Valuation & Solvency Multiples -->
          <div class="matrix-col">
            <!-- Current Valuation Card -->
            <div class="metric-card">
              <h3 class="card-heading">Current Valuation</h3>
              <div class="data-rows">
                <div class="data-row">
                  <span class="label">Current PE Ratio (TTM)</span>
                  <span class="val font-mono">{{ formatMultiple(latestStatement?.valuation?.pe_ratio) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Price to Book Value</span>
                  <span class="val font-mono">{{ formatMultiple(latestStatement?.valuation?.pb_ratio) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Price to Sales (TTM)</span>
                  <span class="val font-mono">{{ formatMultiple(latestStatement?.valuation?.ps_ratio) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Price to Free Cashflow (TTM)</span>
                  <span class="val font-mono">{{ formatMultiple(latestStatement?.valuation?.p_fcf_ratio) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Earnings Yield (TTM)</span>
                  <span class="val font-mono text-green">{{ formatPct(latestStatement?.valuation?.earnings_yield_pct) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">EV to EBIT (TTM)</span>
                  <span class="val font-mono">{{ formatMultiple(latestStatement?.valuation?.ev_to_ebit) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">EV to EBITDA (TTM)</span>
                  <span class="val font-mono">{{ formatMultiple(latestStatement?.valuation?.ev_to_ebitda) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Graham Number</span>
                  <span class="val font-mono text-green">{{ formatIDRPrice(latestStatement?.valuation?.graham_number) }}</span>
                </div>
                <div class="data-row highlight-row">
                  <span class="label">Margin of Safety</span>
                  <span :class="['val font-mono', (latestStatement?.valuation?.margin_of_safety_pct || 0) > 0 ? 'text-green' : 'text-red']">
                    {{ formatSignedPct(latestStatement?.valuation?.margin_of_safety_pct) }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Per Share Metrics Card -->
            <div class="metric-card">
              <h3 class="card-heading">Per Share</h3>
              <div class="data-rows">
                <div class="data-row">
                  <span class="label">Current EPS (IDR)</span>
                  <span class="val font-mono">{{ formatIDRPrice(latestStatement?.valuation?.normalized_eps) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Book Value Per Share</span>
                  <span class="val font-mono">{{ formatIDRPrice(latestStatement?.valuation?.normalized_bvps) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Revenue Per Share</span>
                  <span class="val font-mono">{{ formatIDRPrice(latestStatement?.valuation?.revenue_per_share) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Cash Per Share</span>
                  <span class="val font-mono">{{ formatIDRPrice(latestStatement?.valuation?.cash_per_share) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Free Cashflow Per Share</span>
                  <span class="val font-mono">{{ formatIDRPrice(latestStatement?.valuation?.free_cash_flow_per_share) }}</span>
                </div>
              </div>
            </div>

            <!-- Solvency & Health Card -->
            <div class="metric-card">
              <h3 class="card-heading">Solvency & Health</h3>
              <div class="data-rows">
                <div class="data-row">
                  <span class="label">Piotroski F-Score</span>
                  <span class="val font-mono text-cyan">{{ latestStatement?.computed_ratios?.piotroski_f_score || 0 }}/9</span>
                </div>
                <div class="data-row">
                  <span class="label">Altman Z''-Score</span>
                  <span :class="['val font-mono', (latestStatement?.computed_ratios?.altman_z_score || 0) > 2.6 ? 'text-green' : 'text-amber']">
                    {{ (latestStatement?.computed_ratios?.altman_z_score || 0).toFixed(2) }}
                  </span>
                </div>
                <div class="data-row">
                  <span class="label">Current Ratio</span>
                  <span class="val font-mono">{{ (latestStatement?.computed_ratios?.current_ratio || 0).toFixed(2) }}x</span>
                </div>
                <div class="data-row">
                  <span class="label">Debt to Equity</span>
                  <span class="val font-mono">{{ (latestStatement?.computed_ratios?.debt_to_equity || 0).toFixed(2) }}x</span>
                </div>
                <div class="data-row">
                  <span class="label">Interest Coverage</span>
                  <span class="val font-mono">{{ (latestStatement?.computed_ratios?.interest_coverage_ratio || 0).toFixed(2) }}x</span>
                </div>
              </div>
            </div>
          </div>

          <!-- MIDDLE COLUMN: Multi-Year Historical Performance Matrix -->
          <div class="matrix-col col-span-2">
            <!-- Multi-Year Historical Matrix -->
            <div class="metric-card matrix-table-card">
              <div class="table-header-bar">
                <div class="matrix-tabs font-mono">
                  <button
                    :class="['m-tab', { active: matrixMetric === 'net_income' }]"
                    @click="matrixMetric = 'net_income'"
                  >
                    Net Income
                  </button>
                  <button
                    :class="['m-tab', { active: matrixMetric === 'revenue' }]"
                    @click="matrixMetric = 'revenue'"
                  >
                    Revenue
                  </button>
                  <button
                    :class="['m-tab', { active: matrixMetric === 'fcf' }]"
                    @click="matrixMetric = 'fcf'"
                  >
                    Free Cash Flow
                  </button>
                </div>
                <span class="matrix-currency font-mono">Unit: {{ latestStatement?.metadata?.currency || 'IDR' }}</span>
              </div>

              <!-- Historical Table Grid -->
              <div class="table-responsive">
                <table class="matrix-table font-mono">
                  <thead>
                    <tr>
                      <th>Period</th>
                      <th v-for="y in uniqueYears" :key="y">{{ y }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="p in ['Q1', 'Q2', 'Q3', 'FY']" :key="p">
                      <td class="period-cell">{{ p }}</td>
                      <td v-for="y in uniqueYears" :key="y + p">
                        {{ getMatrixValue(y, p, matrixMetric) }}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <!-- Key Company Stats Bar -->
              <div class="key-stats-grid font-mono">
                <div class="stat-item">
                  <span class="stat-label">Market Cap</span>
                  <span class="stat-val">{{ formatCompact(latestStatement?.valuation?.market_cap) }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">Enterprise Value</span>
                  <span class="stat-val">{{ formatCompact(latestStatement?.valuation?.enterprise_value) }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">Shares Outstanding</span>
                  <span class="stat-val">{{ formatCompact(latestStatement?.core?.shares_outstanding) }}</span>
                </div>
              </div>
            </div>

            <!-- Profitability & Growth Cards -->
            <div class="dual-cards-grid">
              <div class="metric-card">
                <h3 class="card-heading">Profitability</h3>
                <div class="data-rows">
                  <div class="data-row">
                    <span class="label">Return on Invested Capital (ROIC)</span>
                    <span class="val font-mono text-green">{{ formatPct((latestStatement?.computed_ratios?.roic || 0) * 100) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Return on Equity (ROE)</span>
                    <span class="val font-mono text-green">{{ formatPct((latestStatement?.computed_ratios?.roe || 0) * 100) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Gross Profit Margin</span>
                    <span class="val font-mono">{{ formatPct(latestStatement?.computed_ratios?.gross_margin_pct) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Operating Profit Margin</span>
                    <span class="val font-mono">{{ formatPct(latestStatement?.computed_ratios?.operating_margin_pct) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Net Profit Margin</span>
                    <span class="val font-mono">{{ formatPct(latestStatement?.computed_ratios?.net_margin_pct) }}</span>
                  </div>
                </div>
              </div>

              <div class="metric-card">
                <h3 class="card-heading">Cash Flow & Capital Allocation</h3>
                <div class="data-rows">
                  <div class="data-row">
                    <span class="label">Cash from Operations (CFO)</span>
                    <span class="val font-mono text-green">{{ formatCompact(latestStatement?.core?.operating_cash_flow) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Capital Expenditure (CapEx)</span>
                    <span class="val font-mono text-amber">{{ formatCompact(latestStatement?.core?.capex) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Free Cash Flow (FCF)</span>
                    <span class="val font-mono text-green">{{ formatCompact(latestStatement?.core?.free_cash_flow) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Cash from Investing (CFI)</span>
                    <span class="val font-mono">{{ formatCompact(latestStatement?.core?.investing_cash_flow) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Cash from Financing (CFF)</span>
                    <span class="val font-mono">{{ formatCompact(latestStatement?.core?.financing_cash_flow) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- RIGHT COLUMN: 3-Statement Snapshot & Balance Sheet -->
          <div class="matrix-col">
            <!-- Balance Sheet Card -->
            <div class="metric-card">
              <h3 class="card-heading">Balance Sheet</h3>
              <div class="data-rows">
                <div class="data-row">
                  <span class="label">Cash & Equivalents</span>
                  <span class="val font-mono text-cyan">{{ formatCompact(latestStatement?.core?.cash_and_equivalents) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Current Assets</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.current_assets) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Total Assets</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.total_assets) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Current Liabilities</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.current_liabilities) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Total Liabilities</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.total_liabilities) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Working Capital</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.working_capital) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Total Debt</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.total_debt) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Net Debt (Cash)</span>
                  <span :class="['val font-mono', (latestStatement?.computed_ratios?.net_debt || 0) <= 0 ? 'text-green' : 'text-red']">
                    {{ (latestStatement?.computed_ratios?.net_debt || 0) <= 0 ? 'Net Cash ' + formatCompact(Math.abs(latestStatement?.computed_ratios?.net_debt || 0)) : formatCompact(latestStatement?.computed_ratios?.net_debt) }}
                  </span>
                </div>
                <div class="data-row">
                  <span class="label">Total Equity</span>
                  <span class="val font-mono text-cyan">{{ formatCompact(latestStatement?.core?.total_equity) }}</span>
                </div>
              </div>
            </div>

            <!-- Income Statement Card -->
            <div class="metric-card">
              <h3 class="card-heading">Income Statement</h3>
              <div class="data-rows">
                <div class="data-row">
                  <span class="label">Revenue / Sales</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.revenue) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Cost of Revenue</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.cost_of_revenue) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Gross Profit</span>
                  <span class="val font-mono text-green">{{ formatCompact(latestStatement?.core?.gross_profit) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Operating Profit (EBIT)</span>
                  <span class="val font-mono text-green">{{ formatCompact(latestStatement?.core?.operating_income) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Finance Costs</span>
                  <span class="val font-mono text-amber">{{ formatCompact(latestStatement?.core?.finance_costs) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Net Income</span>
                  <span class="val font-mono text-green">{{ formatCompact(latestStatement?.core?.net_income) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 2. CHART VIEW TAB -->
        <div v-else-if="activeModalTab === 'chart'" class="chart-tab-content">
          <PriceValuationChart
            :ticker="ticker"
            :graham-number="latestStatement?.valuation?.graham_number || 0"
            :current-price="latestStatement?.valuation?.current_price || 0"
          />
        </div>

        <!-- 3. RELATED SECTOR NEWS TAB -->
        <div v-else-if="activeModalTab === 'news'" class="sector-news-tab">
          <div class="news-tab-header">
            <div>
              <h3 class="news-tab-title font-mono">Related Industry & Sector Intelligence</h3>
              <p class="news-tab-sub">Stream of news developments affecting {{ latestStatement?.metadata?.sector || 'the sector' }}</p>
            </div>
            <button class="btn-refresh font-mono" @click="fetchSectorNews">🔄 Refresh Stream</button>
          </div>

          <div v-if="loadingNews" class="loading-state font-mono">Loading related news...</div>
          <div v-else-if="sectorNews.length === 0" class="empty-state font-mono">No news articles found for this sector.</div>
          <div v-else class="news-cards-list">
            <article v-for="item in sectorNews" :key="item.id || item._id" class="sector-news-card">
              <div class="card-meta font-mono">
                <span :class="['sentiment-tag', item.impact_direction?.toLowerCase()]">{{ item.impact_direction || 'Neutral' }}</span>
                <span class="score-tag font-mono">Score: {{ (item.value_score && item.value_score > 0 ? '+' : '') + (item.value_score || 0) }}</span>
                <span class="date">{{ formatDate(item.date || item.created_at) }}</span>
              </div>
              <h4 class="card-headline">{{ item.title }}</h4>
              <p class="card-summary">{{ item.summary }}</p>
              <div v-if="item.investment_takeaway" class="card-takeaway">
                💡 {{ item.investment_takeaway }}
              </div>
            </article>
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
  ticker: string
}>()

defineEmits<{
  (e: 'close'): void
}>()

const activeModalTab = ref<'terminal' | 'chart' | 'news'>('terminal')
const matrixMetric = ref<'net_income' | 'revenue' | 'fcf'>('net_income')

const statements = ref<XBRLStatement[]>([])
const sectorNews = ref<News[]>([])
const loading = ref(true)
const loadingNews = ref(false)

const latestStatement = computed(() => {
  if (statements.value.length === 0) return null
  return statements.value[0]
})

const uniqueYears = computed(() => {
  const years = Array.from(new Set(statements.value.map(s => s.year))).sort((a, b) => b - a)
  return years.slice(0, 5)
})

const fetchFinancials = async () => {
  if (!props.ticker) return
  loading.value = true
  try {
    const data = await $fetch<XBRLStatement[]>(`/api/v1/stocks/${props.ticker}/financials`)
    statements.value = data || []
  } catch (e) {
    console.error('Failed to load financials', e)
    statements.value = []
  } finally {
    loading.value = false
  }
}

const fetchSectorNews = async () => {
  const sector = latestStatement.value?.metadata?.sector
  if (!sector) return
  loadingNews.value = true
  try {
    const res = await $fetch<any>(`/api/v1/news?sector=${encodeURIComponent(sector)}&limit=20`)
    sectorNews.value = Array.isArray(res) ? res : (res.data || [])
  } catch (e) {
    console.error('Failed to fetch sector news', e)
  } finally {
    loadingNews.value = false
  }
}

const getMatrixValue = (year: number, period: string, metric: 'net_income' | 'revenue' | 'fcf') => {
  const stmt = statements.value.find(s => s.year === year && s.period.toUpperCase() === period.toUpperCase())
  if (!stmt) return '-'
  let val = 0
  if (metric === 'net_income') val = stmt.core?.net_income || 0
  if (metric === 'revenue') val = stmt.core?.revenue || 0
  if (metric === 'fcf') val = stmt.core?.free_cash_flow || 0
  return formatCompact(val)
}

const formatMultiple = (val?: number) => {
  if (!val || val <= 0) return '-'
  return val.toFixed(2) + 'x'
}

const formatPct = (val?: number) => {
  if (!val || isNaN(val)) return '-'
  return val.toFixed(2) + '%'
}

const formatSignedPct = (val?: number) => {
  if (val === undefined || isNaN(val)) return '-'
  return (val > 0 ? '+' : '') + val.toFixed(1) + '%'
}

const formatIDRPrice = (val?: number) => {
  if (!val || isNaN(val) || val <= 0) return '-'
  return 'Rp ' + Math.round(val).toLocaleString('en-US')
}

const formatCompact = (val?: number) => {
  if (val === undefined || val === null || isNaN(val) || val === 0) return '-'
  const abs = Math.abs(val)
  if (abs >= 1e12) return (val / 1e12).toFixed(2) + ' T'
  if (abs >= 1e9) return (val / 1e9).toFixed(2) + ' B'
  if (abs >= 1e6) return (val / 1e6).toFixed(2) + ' M'
  return Math.round(val).toLocaleString('en-US')
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

watch(() => props.ticker, () => {
  fetchFinancials()
}, { immediate: true })

watch(activeModalTab, (tab) => {
  if (tab === 'news' && sectorNews.value.length === 0) {
    fetchSectorNews()
  }
})
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.85);
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
  max-width: 1400px;
  max-height: 94vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.7);
}
.modal-header {
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.title-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
.ticker-badge {
  background: #2563eb;
  color: #fff;
  font-weight: 800;
  font-size: 0.95rem;
  padding: 3px 8px;
  border-radius: 4px;
}
.company-name {
  font-size: 1.35rem;
  font-weight: 700;
}
.meta-tags {
  display: flex;
  gap: 8px;
  margin-top: 6px;
  flex-wrap: wrap;
}
.tag {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  font-size: 0.75rem;
  padding: 2px 8px;
  border-radius: 4px;
}
.sector-tag {
  color: #38bdf8;
  border-color: rgba(56, 189, 248, 0.3);
}
.btn-close {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 1.25rem;
  cursor: pointer;
}
.btn-close:hover {
  color: #fff;
}
.modal-tab-bar {
  display: flex;
  background: var(--bg-app);
  padding: 6px 24px;
  gap: 8px;
  border-bottom: 1px solid var(--border-color);
}
.modal-tab-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 6px 14px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
}
.modal-tab-btn:hover {
  color: #fff;
  background: var(--bg-card-hover);
}
.modal-tab-btn.active {
  background: var(--bg-card);
  color: #38bdf8;
  border: 1px solid var(--border-color);
}
.tab-badge {
  background: #1e293b;
  color: #38bdf8;
  font-size: 0.7rem;
  padding: 1px 6px;
  border-radius: 4px;
}
.modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}
.matrix-layout {
  display: grid;
  grid-template-columns: 320px 1fr 340px;
  gap: 20px;
  align-items: start;
}
.matrix-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.metric-card {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
}
.card-heading {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  padding-bottom: 6px;
}
.data-rows {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.data-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
}
.data-row .label {
  color: var(--text-secondary);
}
.data-row .val {
  font-weight: 600;
  color: #f8fafc;
}
.highlight-row {
  background: rgba(255, 255, 255, 0.03);
  padding: 4px 6px;
  border-radius: 4px;
}
.text-green {
  color: #34d399 !important;
}
.text-red {
  color: #f87171 !important;
}
.text-amber {
  color: #fbbf24 !important;
}
.text-cyan {
  color: #38bdf8 !important;
}
.matrix-table-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.table-header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.matrix-tabs {
  display: flex;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  padding: 2px;
  border-radius: 6px;
  gap: 2px;
}
.m-tab {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 4px 10px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  border-radius: 4px;
}
.m-tab.active {
  background: #2563eb;
  color: #fff;
}
.matrix-currency {
  font-size: 0.75rem;
  color: var(--text-muted);
}
.matrix-table {
  width: 100%;
  border-collapse: collapse;
  text-align: right;
  font-size: 0.85rem;
}
.matrix-table th {
  background: rgba(255, 255, 255, 0.02);
  padding: 8px 12px;
  color: var(--text-muted);
  border-bottom: 1px solid var(--border-color);
}
.matrix-table th:first-child, .matrix-table td:first-child {
  text-align: left;
}
.matrix-table td {
  padding: 10px 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}
.period-cell {
  color: #38bdf8;
  font-weight: 700;
}
.key-stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  background: var(--bg-card);
  padding: 10px 14px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
}
.stat-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.stat-label {
  font-size: 0.7rem;
  color: var(--text-muted);
}
.stat-val {
  font-size: 0.9rem;
  font-weight: 700;
  color: #f8fafc;
}
.dual-cards-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.chart-tab-content {
  background: var(--bg-app);
  border-radius: 8px;
  padding: 20px;
}
.sector-news-tab {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.news-tab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.news-tab-title {
  font-size: 1.1rem;
  font-weight: 700;
}
.news-tab-sub {
  color: var(--text-secondary);
  font-size: 0.85rem;
}
.btn-refresh {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: #38bdf8;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 0.8rem;
  cursor: pointer;
}
.news-cards-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.sector-news-card {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.card-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.75rem;
}
.sentiment-tag {
  font-size: 0.7rem;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 4px;
  text-transform: uppercase;
}
.sentiment-tag.bullish { background: var(--bullish-bg); color: var(--bullish-text); }
.sentiment-tag.bearish { background: var(--bearish-bg); color: var(--bearish-text); }
.sentiment-tag.neutral { background: var(--neutral-bg); color: var(--neutral-text); }
.score-tag {
  color: var(--text-secondary);
}
.card-headline {
  font-size: 1rem;
  font-weight: 600;
}
.card-summary {
  font-size: 0.85rem;
  color: var(--text-secondary);
  line-height: 1.4;
}
.card-takeaway {
  font-size: 0.8rem;
  color: #fde68a;
  background: rgba(245, 158, 11, 0.06);
  padding: 6px 8px;
  border-radius: 4px;
}
.loading-state, .empty-state {
  text-align: center;
  padding: 60px;
  color: var(--text-muted);
}
@media (max-width: 1200px) {
  .matrix-layout {
    grid-template-columns: 1fr 1fr;
  }
}
@media (max-width: 768px) {
  .matrix-layout {
    grid-template-columns: 1fr;
  }
  .dual-cards-grid {
    grid-template-columns: 1fr;
  }
}
</style>
