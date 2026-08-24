<template>
  <div class="screener-container">
    <!-- Header & Pitch -->
    <div class="screener-header">
      <div>
        <h1 class="screener-title">📊 Value Investing Screener</h1>
        <p class="screener-sub">
          Filter 900+ IDX companies by Margin of Safety, Return on Invested Capital (ROIC), and Piotroski F-Score
        </p>
      </div>

      <!-- Quick Preset Buttons -->
      <div class="presets-row">
        <span class="preset-label font-mono">Presets:</span>
        <button class="preset-btn" @click="applyPreset('deep_value')">🎯 Deep Value (MOS ≥ 30%)</button>
        <button class="preset-btn" @click="applyPreset('buffett_moat')">🏰 Buffett Moat (ROIC ≥ 15%)</button>
        <button class="preset-btn" @click="applyPreset('piotroski_strong')">⭐ High F-Score (≥ 7)</button>
        <button class="preset-btn reset" @click="resetFilters">↺ Reset</button>
      </div>
    </div>

    <!-- Filter Controls Bar -->
    <div class="filter-panel">
      <div class="filter-group">
        <label>Min Margin of Safety (%)</label>
        <div class="input-with-unit">
          <input
            v-model.number="filters.minMos"
            type="number"
            placeholder="e.g. 30"
            class="filter-input font-mono"
            @change="fetchScreener"
          />
          <span class="unit">%</span>
        </div>
      </div>

      <div class="filter-group">
        <label>Min ROIC (%)</label>
        <div class="input-with-unit">
          <input
            v-model.number="filters.minRoicPct"
            type="number"
            placeholder="e.g. 15"
            class="filter-input font-mono"
            @change="fetchScreener"
          />
          <span class="unit">%</span>
        </div>
      </div>

      <div class="filter-group">
        <label>Min Piotroski F-Score (0-9)</label>
        <select v-model.number="filters.minFScore" class="filter-select font-mono" @change="fetchScreener">
          <option :value="0">All Scores (0+)</option>
          <option :value="5">5+ (Moderate)</option>
          <option :value="7">7+ (Strong)</option>
          <option :value="8">8+ (Elite)</option>
        </select>
      </div>

      <div class="filter-group">
        <label>Max Debt / Equity</label>
        <input
          v-model.number="filters.maxDe"
          type="number"
          step="0.1"
          placeholder="e.g. 1.0"
          class="filter-input font-mono"
          @change="fetchScreener"
        />
      </div>

      <div class="filter-group">
        <label>Sector</label>
        <select v-model="filters.sector" class="filter-select" @change="fetchScreener">
          <option value="">All Sectors</option>
          <option value="A. Energy">Energy (Oil, Gas & Coal)</option>
          <option value="B. Basic Materials">Basic Materials & Mining</option>
          <option value="C. Industrials">Industrials</option>
          <option value="D. Consumer Non-Cyclicals">Consumer Non-Cyclicals</option>
          <option value="E. Consumer Cyclicals">Consumer Cyclicals</option>
          <option value="F. Healthcare">Healthcare</option>
          <option value="G. Financials">Financials & Banking</option>
          <option value="H. Properties & Real Estate">Properties & Real Estate</option>
          <option value="I. Technology">Technology</option>
          <option value="J. Infrastructures">Infrastructures & Utilities</option>
        </select>
      </div>

      <button class="btn-search font-mono" @click="fetchScreener">
        🔍 Screen ({{ totalResults }})
      </button>
    </div>

    <!-- Results Table -->
    <div v-if="loading" class="loading-state font-mono">
      Screening XBRL financial statements...
    </div>
    <div v-else-if="statements.length === 0" class="empty-state font-mono">
      No companies matched the selected value investing filters.
    </div>
    <div v-else class="table-card">
      <table class="screener-table">
        <thead>
          <tr class="font-mono">
            <th>Ticker</th>
            <th>Company Name</th>
            <th>Sector</th>
            <th>Period</th>
            <th>Price</th>
            <th>Graham No.</th>
            <th>Margin of Safety</th>
            <th>ROIC</th>
            <th>F-Score</th>
            <th>Net Debt</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="s in statements"
            :key="s.id || s._id"
            class="screener-row"
            @click="$emit('open-ticker-financials', s.ticker)"
          >
            <td class="ticker-cell font-mono">${{ s.ticker }}</td>
            <td class="company-name-cell">{{ s.company_name }}</td>
            <td class="sector-cell">{{ s.metadata?.sector || '-' }}</td>
            <td class="period-cell font-mono">{{ s.year }}-{{ s.period }}</td>
            <td class="font-mono price-col">
              {{ s.valuation?.current_price ? formatCurrency(s.valuation.current_price, 'IDR') : '-' }}
            </td>
            <td class="font-mono graham-col">
              {{ s.valuation?.graham_number ? formatCurrency(s.valuation.graham_number, 'IDR') : '-' }}
            </td>
            <td class="font-mono">
              <span
                v-if="s.valuation?.margin_of_safety_pct !== undefined"
                :class="['mos-badge', getMosClass(s.valuation.margin_of_safety_pct)]"
              >
                {{ (s.valuation.margin_of_safety_pct > 0 ? '+' : '') + s.valuation.margin_of_safety_pct.toFixed(1) }}%
              </span>
              <span v-else class="text-muted">-</span>
            </td>
            <td class="font-mono">
              <span
                v-if="s.computed_ratios?.roic !== undefined"
                :class="['roic-badge', s.computed_ratios.roic >= 0.15 ? 'roic-high' : 'roic-normal']"
              >
                {{ (s.computed_ratios.roic * 100).toFixed(1) }}%
              </span>
              <span v-else class="text-muted">-</span>
            </td>
            <td class="font-mono">
              <span :class="['fscore-pill', getFScoreClass(s.computed_ratios?.piotroski_f_score)]">
                {{ s.computed_ratios?.piotroski_f_score || 0 }}/9
              </span>
            </td>
            <td class="font-mono net-debt-col">
              {{ formatNetDebt(s.computed_ratios?.net_debt, s.metadata?.currency) }}
            </td>
            <td>
              <button
                class="btn-inspect font-mono"
                @click.stop="$emit('open-ticker-financials', s.ticker)"
              >
                360° View ↗
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import type { XBRLStatement } from '../server/utils/types'

defineEmits<{
  (e: 'open-ticker-financials', ticker: string): void
}>()

const loading = ref(true)
const statements = ref<XBRLStatement[]>([])
const totalResults = ref(0)

const filters = reactive({
  minMos: undefined as number | undefined,
  minRoicPct: undefined as number | undefined,
  minFScore: 0,
  maxDe: undefined as number | undefined,
  sector: ''
})

const fetchScreener = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (filters.minMos !== undefined && !isNaN(filters.minMos)) {
      params.append('min_mos', filters.minMos.toString())
    }
    if (filters.minRoicPct !== undefined && !isNaN(filters.minRoicPct)) {
      params.append('min_roic', (filters.minRoicPct / 100).toString())
    }
    if (filters.minFScore > 0) {
      params.append('min_f_score', filters.minFScore.toString())
    }
    if (filters.maxDe !== undefined && !isNaN(filters.maxDe)) {
      params.append('max_de', filters.maxDe.toString())
    }
    if (filters.sector) {
      params.append('sector', filters.sector)
    }

    const res = await $fetch<{ statements: XBRLStatement[]; total: number }>(
      `/api/v1/screener/value?${params.toString()}`
    )
    statements.value = res.statements || []
    totalResults.value = res.total || 0
  } catch (e) {
    console.error('Failed to fetch screener results', e)
  } finally {
    loading.value = false
  }
}

const applyPreset = (preset: string) => {
  resetFilters()
  if (preset === 'deep_value') {
    filters.minMos = 30
    filters.minFScore = 5
  } else if (preset === 'buffett_moat') {
    filters.minRoicPct = 15
    filters.maxDe = 1.0
    filters.minFScore = 7
  } else if (preset === 'piotroski_strong') {
    filters.minFScore = 7
  }
  fetchScreener()
}

const resetFilters = () => {
  filters.minMos = undefined
  filters.minRoicPct = undefined
  filters.minFScore = 0
  filters.maxDe = undefined
  filters.sector = ''
}

const getMosClass = (mos: number) => {
  if (mos >= 30) return 'mos-elite'
  if (mos > 0) return 'mos-positive'
  return 'mos-negative'
}

const getFScoreClass = (score?: number) => {
  if (!score) return 'fscore-low'
  if (score >= 7) return 'fscore-high'
  if (score >= 5) return 'fscore-mid'
  return 'fscore-low'
}

const formatCurrency = (val: number, cur: string) => {
  return new Intl.NumberFormat('id-ID', {
    maximumFractionDigits: 0
  }).format(val)
}

const formatNetDebt = (netDebt?: number, cur?: string) => {
  if (netDebt === undefined) return '-'
  if (netDebt < 0) {
    return 'Net Cash ' + formatCompact(Math.abs(netDebt))
  }
  return formatCompact(netDebt)
}

const formatCompact = (val: number) => {
  if (Math.abs(val) >= 1e12) return (val / 1e12).toFixed(1) + 'T'
  if (Math.abs(val) >= 1e9) return (val / 1e9).toFixed(1) + 'B'
  if (Math.abs(val) >= 1e6) return (val / 1e6).toFixed(1) + 'M'
  return val.toFixed(0)
}

onMounted(() => {
  fetchScreener()
})
</script>

<style scoped>
.screener-container {
  max-width: 1440px;
  margin: 0 auto;
  padding: 24px 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.screener-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}
.screener-title {
  font-size: 1.5rem;
  font-weight: 700;
}
.screener-sub {
  color: var(--text-secondary);
  font-size: 0.9rem;
}
.presets-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.preset-label {
  font-size: 0.8rem;
  color: var(--text-muted);
}
.preset-btn {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}
.preset-btn:hover {
  background: var(--bg-card-hover);
  color: #38bdf8;
  border-color: #38bdf8;
}
.preset-btn.reset {
  color: var(--text-muted);
}
.filter-panel {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 16px 20px;
  display: flex;
  align-items: flex-end;
  gap: 16px;
  flex-wrap: wrap;
}
.filter-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
  min-width: 150px;
}
.filter-group label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-secondary);
}
.input-with-unit {
  position: relative;
  display: flex;
  align-items: center;
}
.filter-input, .filter-select {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  color: #fff;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 0.85rem;
  width: 100%;
  outline: none;
}
.filter-input:focus, .filter-select:focus {
  border-color: #38bdf8;
}
.unit {
  position: absolute;
  right: 12px;
  font-size: 0.8rem;
  color: var(--text-muted);
}
.btn-search {
  background: #2563eb;
  color: #fff;
  border: none;
  padding: 9px 18px;
  border-radius: 6px;
  font-weight: 700;
  font-size: 0.85rem;
  cursor: pointer;
  height: 38px;
}
.btn-search:hover {
  background: #1d4ed8;
}
.table-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  overflow-x: auto;
}
.screener-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}
.screener-table th {
  background: var(--bg-app);
  padding: 12px 16px;
  color: var(--text-muted);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid var(--border-color);
}
.screener-table td {
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-color);
  font-size: 0.9rem;
}
.screener-row {
  cursor: pointer;
  transition: background 0.15s ease;
}
.screener-row:hover td {
  background: var(--bg-card-hover);
}
.ticker-cell {
  color: #38bdf8;
  font-weight: 700;
  font-size: 0.95rem;
}
.company-name-cell {
  font-weight: 500;
  color: var(--text-primary);
  max-width: 220px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sector-cell {
  font-size: 0.8rem;
  color: var(--text-secondary);
}
.period-cell {
  font-size: 0.8rem;
  color: var(--text-muted);
}
.price-col {
  font-weight: 600;
}
.graham-col {
  color: #a7f3d0;
}
.mos-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 700;
  font-size: 0.8rem;
}
.mos-elite {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
  border: 1px solid #10b981;
}
.mos-positive {
  background: rgba(16, 185, 129, 0.1);
  color: #6ee7b7;
}
.mos-negative {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
}
.roic-badge {
  font-size: 0.8rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}
.roic-high {
  background: rgba(16, 185, 129, 0.15);
  color: #34d399;
}
.roic-normal {
  color: var(--text-secondary);
}
.fscore-pill {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}
.fscore-high {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
}
.fscore-mid {
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
}
.fscore-low {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
}
.net-debt-col {
  font-size: 0.8rem;
  color: var(--text-secondary);
}
.btn-inspect {
  background: #1e293b;
  color: #38bdf8;
  border: 1px solid #38bdf8;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}
.btn-inspect:hover {
  background: #38bdf8;
  color: #080c14;
}
.loading-state, .empty-state {
  text-align: center;
  padding: 60px;
  color: var(--text-muted);
  font-size: 0.95rem;
}
</style>
