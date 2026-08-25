<template>
  <div class="mx-auto max-w-7xl space-y-6 px-4 py-6 sm:px-6">
    <!-- Header & Presets -->
    <div class="flex flex-col gap-4 border-b border-border/80 pb-6 lg:flex-row lg:items-center lg:justify-between">
      <div>
        <div class="flex items-center gap-2">
          <SlidersHorizontal class="h-5 w-5 text-primary" />
          <h1 class="text-xl font-bold tracking-tight text-foreground font-mono">
            QUANTITATIVE VALUE INVESTING SCREENER
          </h1>
        </div>
        <p class="text-xs text-muted-foreground mt-1">
          Screen 900+ IDX listed companies by Benjamin Graham Fair Value, MOS %, ROIC %, Piotroski F-Score & Smart Timing VSA Signals
        </p>
      </div>

      <!-- Quick Presets -->
      <div class="flex flex-wrap items-center gap-1.5">
        <span class="text-xs font-mono text-muted-foreground mr-1">Presets:</span>
        <Button
          variant="outline"
          size="xs"
          class="font-mono text-xs border-emerald-500/40 text-emerald-400 bg-emerald-500/10 hover:bg-emerald-500/20"
          @click="applyPreset('actionable_buy')"
        >
          ⚡ Actionable Buy (Score ≥ 70 & MOS ≥ 20%)
        </Button>
        <Button
          variant="outline"
          size="xs"
          class="font-mono text-xs hover:border-primary/40"
          @click="applyPreset('deep_value')"
        >
          🎯 Deep Value (MOS ≥ 30%)
        </Button>
        <Button
          variant="outline"
          size="xs"
          class="font-mono text-xs hover:border-primary/40"
          @click="applyPreset('buffett_moat')"
        >
          🏰 Buffett Moat (ROIC ≥ 15%)
        </Button>
        <Button
          variant="outline"
          size="xs"
          class="font-mono text-xs hover:border-primary/40"
          @click="applyPreset('piotroski_strong')"
        >
          ⭐ High F-Score (≥ 7)
        </Button>
        <Button
          variant="ghost"
          size="xs"
          class="font-mono text-xs text-muted-foreground hover:text-foreground"
          @click="resetFilters"
        >
          <RotateCcw class="mr-1 h-3 w-3" /> Reset
        </Button>
      </div>
    </div>

    <!-- Filter Control Panel -->
    <Card class="border-border bg-card/60 p-4">
      <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-7">
        <div class="space-y-1">
          <label class="text-[11px] font-mono text-muted-foreground">Min MOS (%)</label>
          <Input
            v-model.number="filters.minMos"
            type="number"
            placeholder="e.g. 30"
            class="h-8 font-mono text-xs bg-background/80"
            @change="fetchScreener"
          />
        </div>

        <div class="space-y-1">
          <label class="text-[11px] font-mono text-muted-foreground">Min ROIC (%)</label>
          <Input
            v-model.number="filters.minRoicPct"
            type="number"
            placeholder="e.g. 15"
            class="h-8 font-mono text-xs bg-background/80"
            @change="fetchScreener"
          />
        </div>

        <div class="space-y-1">
          <label class="text-[11px] font-mono text-muted-foreground">Timing Score</label>
          <select
            v-model.number="filters.minTimingScore"
            class="flex h-8 w-full rounded-md border border-input bg-background/80 px-2 py-1 text-xs font-mono text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            @change="fetchScreener"
          >
            <option :value="0">All Scores</option>
            <option :value="50">≥ 50 (Accumulation)</option>
            <option :value="70">≥ 70 (Actionable Buy)</option>
            <option :value="80">≥ 80 (Elite Setup)</option>
          </select>
        </div>

        <div class="space-y-1">
          <label class="text-[11px] font-mono text-muted-foreground">P/E Band</label>
          <select
            v-model="filters.peBand"
            class="flex h-8 w-full rounded-md border border-input bg-background/80 px-2 py-1 text-xs font-mono text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            @change="fetchScreener"
          >
            <option value="">All Levels</option>
            <option value="minus1sd">≤ -1σ (Discount)</option>
            <option value="minus2sd">≤ -2σ (Deep Value)</option>
            <option value="mean">≤ Mean (Fair / Cheap)</option>
          </select>
        </div>

        <div class="space-y-1">
          <label class="text-[11px] font-mono text-muted-foreground">Min F-Score</label>
          <select
            v-model.number="filters.minFScore"
            class="flex h-8 w-full rounded-md border border-input bg-background/80 px-2 py-1 text-xs font-mono text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            @change="fetchScreener"
          >
            <option :value="0">All Scores (0+)</option>
            <option :value="5">5+ (Moderate)</option>
            <option :value="7">7+ (Strong)</option>
            <option :value="8">8+ (Elite)</option>
          </select>
        </div>

        <div class="space-y-1">
          <label class="text-[11px] font-mono text-muted-foreground">Max D/E Ratio</label>
          <Input
            v-model.number="filters.maxDe"
            type="number"
            step="0.1"
            placeholder="e.g. 1.0"
            class="h-8 font-mono text-xs bg-background/80"
            @change="fetchScreener"
          />
        </div>

        <div class="space-y-1">
          <label class="text-[11px] font-mono text-muted-foreground">Sector</label>
          <select
            v-model="filters.sector"
            class="flex h-8 w-full rounded-md border border-input bg-background/80 px-2 py-1 text-xs text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring truncate"
            @change="fetchScreener"
          >
            <option value="">All Sectors</option>
            <option value="A. Energy">Energy</option>
            <option value="B. Basic Materials">Basic Materials</option>
            <option value="C. Industrials">Industrials</option>
            <option value="D. Consumer Non-Cyclicals">Consumer Non-Cyclicals</option>
            <option value="E. Consumer Cyclicals">Consumer Cyclicals</option>
            <option value="F. Healthcare">Healthcare</option>
            <option value="G. Financials">Financials & Banking</option>
            <option value="H. Properties & Real Estate">Properties</option>
            <option value="I. Technology">Technology</option>
            <option value="J. Infrastructures">Infrastructures</option>
          </select>
        </div>
      </div>
    </Card>

    <!-- Results Table -->
    <div v-if="loading" class="flex h-64 items-center justify-center rounded-lg border border-dashed border-border bg-card/40 font-mono text-xs text-muted-foreground">
      Screening XBRL financial statements...
    </div>
    <div v-else-if="statements.length === 0" class="flex h-64 items-center justify-center rounded-lg border border-dashed border-border bg-card/40 font-mono text-xs text-muted-foreground">
      No companies matched the selected value investing filters.
    </div>
    <Card v-else class="overflow-hidden border-border bg-card">
      <div class="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow class="border-border hover:bg-transparent font-mono text-[11px]">
              <TableHead class="font-bold text-foreground">Ticker</TableHead>
              <TableHead class="font-bold text-foreground">Company</TableHead>
              <TableHead class="font-bold text-foreground">Sector</TableHead>
              <TableHead class="font-bold text-foreground">Period</TableHead>
              <TableHead class="text-right font-bold text-foreground">Price</TableHead>
              <TableHead class="text-right font-bold text-foreground">Graham No.</TableHead>
              <TableHead class="text-center font-bold text-foreground">Margin of Safety</TableHead>
              <TableHead class="text-center font-bold text-foreground">ROIC</TableHead>
              <TableHead class="text-center font-bold text-foreground">F-Score</TableHead>
              <TableHead class="text-center font-bold text-foreground">Timing Score</TableHead>
              <TableHead class="text-right font-bold text-foreground">Net Debt</TableHead>
              <TableHead class="text-center font-bold text-foreground">Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="s in statements"
              :key="s.id || s._id"
              class="cursor-pointer border-border hover:bg-muted/40 transition-colors font-mono text-xs"
              @click="$emit('open-ticker-financials', s.ticker)"
            >
              <TableCell class="font-bold text-primary">
                ${{ s.ticker }}
              </TableCell>
              <TableCell class="max-w-[180px] truncate font-sans text-xs text-foreground">
                {{ s.company_name }}
              </TableCell>
              <TableCell class="font-sans text-[11px] text-muted-foreground">
                {{ s.metadata?.sector || '-' }}
              </TableCell>
              <TableCell class="text-muted-foreground">
                {{ s.year }}-{{ s.period }}
              </TableCell>
              <TableCell class="text-right font-semibold text-foreground">
                {{ s.valuation?.current_price ? formatCurrency(s.valuation.current_price, 'IDR') : '-' }}
              </TableCell>
              <TableCell class="text-right text-muted-foreground">
                {{ s.valuation?.graham_number ? formatCurrency(s.valuation.graham_number, 'IDR') : '-' }}
              </TableCell>
              <TableCell class="text-center">
                <Badge
                  v-if="s.valuation?.margin_of_safety_pct !== undefined"
                  :variant="getMosVariant(s.valuation.margin_of_safety_pct)"
                >
                  {{ (s.valuation.margin_of_safety_pct > 0 ? '+' : '') + s.valuation.margin_of_safety_pct.toFixed(1) }}%
                </Badge>
                <span v-else class="text-muted-foreground">-</span>
              </TableCell>
              <TableCell class="text-center">
                <Badge
                  v-if="s.computed_ratios?.roic !== undefined"
                  :variant="s.computed_ratios.roic >= 0.15 ? 'bullish' : 'secondary'"
                >
                  {{ (s.computed_ratios.roic * 100).toFixed(1) }}%
                </Badge>
                <span v-else class="text-muted-foreground">-</span>
              </TableCell>
              <TableCell class="text-center">
                <Badge :variant="getFScoreVariant(s.computed_ratios?.piotroski_f_score)">
                  {{ s.computed_ratios?.piotroski_f_score || 0 }}/9
                </Badge>
              </TableCell>
              <TableCell class="text-center">
                <div v-if="s.timing_signal || s.valuation?.timing_signal" class="inline-flex flex-col items-center gap-0.5">
                  <Badge :variant="getTimingVariant((s.timing_signal || s.valuation?.timing_signal)?.score || 0)">
                    {{ (s.timing_signal || s.valuation?.timing_signal)?.score || 0 }}/100
                  </Badge>
                </div>
                <span v-else class="text-muted-foreground">-</span>
              </TableCell>
              <TableCell class="text-right text-muted-foreground">
                {{ formatNetDebt(s.computed_ratios?.net_debt, s.metadata?.currency) }}
              </TableCell>
              <TableCell class="text-center" @click.stop>
                <Button
                  variant="outline"
                  size="xs"
                  class="font-mono text-[11px] hover:border-primary hover:text-primary"
                  @click="$emit('open-ticker-financials', s.ticker)"
                >
                  360° View ↗
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { SlidersHorizontal, RotateCcw } from 'lucide-vue-next'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from '@/components/ui/table'
import type { XBRLStatement } from '@/server/utils/types'

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
  sector: '',
  minTimingScore: 0,
  peBand: '',
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
    if (filters.minTimingScore > 0) {
      params.append('min_timing_score', filters.minTimingScore.toString())
    }
    if (filters.peBand) {
      params.append('pe_band', filters.peBand)
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
  if (preset === 'actionable_buy') {
    filters.minTimingScore = 70
    filters.minMos = 20
  } else if (preset === 'deep_value') {
    filters.minMos = 30
  } else if (preset === 'buffett_moat') {
    filters.minRoicPct = 15
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
  filters.minTimingScore = 0
  filters.peBand = ''
  fetchScreener()
}

const getMosVariant = (mos?: number) => {
  if (mos === undefined) return 'secondary'
  if (mos >= 30) return 'safe'
  if (mos >= 0) return 'bullish'
  return 'bearish'
}

const getFScoreVariant = (score?: number) => {
  if (score === undefined) return 'secondary'
  if (score >= 7) return 'safe'
  if (score >= 5) return 'warning'
  return 'danger'
}

const getTimingVariant = (score: number) => {
  if (score >= 70) return 'safe'
  if (score >= 50) return 'warning'
  return 'danger'
}

const formatCurrency = (val?: number, currency = 'IDR') => {
  if (val === undefined || isNaN(val)) return '-'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: currency,
    maximumFractionDigits: 0,
  }).format(val)
}

const formatNetDebt = (val?: number, currency = 'IDR') => {
  if (val === undefined || isNaN(val)) return '-'
  if (val < 0) return `Net Cash (${formatCurrency(Math.abs(val), currency)})`
  return formatCurrency(val, currency)
}

onMounted(() => {
  fetchScreener()
})
</script>
