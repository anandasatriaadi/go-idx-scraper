<template>
  <div class="mx-auto max-w-7xl space-y-6 px-4 py-6 sm:px-6">
    <!-- Header & Filter Controls -->
    <div class="flex flex-col gap-4 border-b border-border/80 pb-6 lg:flex-row lg:items-center lg:justify-between">
      <div>
        <div class="flex items-center gap-2">
          <Newspaper class="h-5 w-5 text-primary" />
          <h1 class="text-xl font-bold tracking-tight text-foreground font-mono">
            NEWS INTELLIGENCE TERMINAL
          </h1>
        </div>
        <p class="text-xs text-muted-foreground mt-1">
          Real-time multi-channel news classified with official IDX-IC taxonomy and Value Investing metrics
        </p>
      </div>

      <!-- Controls Bar -->
      <div class="flex flex-wrap items-center gap-2">
        <!-- Ticker Search -->
        <div class="relative w-40 sm:w-48">
          <Search class="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            v-model="tickerFilter"
            type="text"
            placeholder="Ticker (e.g. BBRI)..."
            class="h-8 pl-8 font-mono text-xs bg-background/80 uppercase placeholder:normal-case"
            @input="onFilterChanged"
          />
          <button
            v-if="tickerFilter"
            class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-muted-foreground hover:text-foreground"
            @click="clearTicker"
          >
            ✕
          </button>
        </div>

        <!-- Sector Filter -->
        <select
          v-model="sectorFilter"
          class="h-8 rounded-md border border-input bg-background/80 px-2.5 py-1 text-xs text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          @change="onSectorChange"
        >
          <option value="">All Sectors</option>
          <option v-for="s in sectorList" :key="s" :value="s">{{ s }}</option>
        </select>

        <!-- Subsector Filter -->
        <select
          v-model="subsectorFilter"
          class="h-8 rounded-md border border-input bg-background/80 px-2.5 py-1 text-xs text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          :disabled="!availableSubsectors.length"
          @change="onFilterChanged"
        >
          <option value="">{{ availableSubsectors.length ? 'All Subsectors' : 'Select Sector first' }}</option>
          <option v-for="sub in availableSubsectors" :key="sub" :value="sub">{{ sub }}</option>
        </select>

        <!-- Impact Direction -->
        <select
          v-model="directionFilter"
          class="h-8 rounded-md border border-input bg-background/80 px-2.5 py-1 text-xs text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          @change="onFilterChanged"
        >
          <option value="">All Directions</option>
          <option value="Bullish">🟢 Bullish</option>
          <option value="Bearish">🔴 Bearish</option>
          <option value="Neutral">🔵 Neutral</option>
        </select>
      </div>
    </div>

    <!-- News Grid -->
    <div v-if="loading && items.length === 0" class="flex h-64 items-center justify-center rounded-lg border border-dashed border-border bg-card/40 font-mono text-xs text-muted-foreground">
      Loading news terminal...
    </div>
    <div v-else-if="filteredList.length === 0" class="flex h-64 items-center justify-center rounded-lg border border-dashed border-border bg-card/40 font-mono text-xs text-muted-foreground">
      No matching news articles found for selected filters.
    </div>
    <div v-else class="space-y-6">
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card
          v-for="item in filteredList"
          :key="item.id"
          class="group cursor-pointer border-border bg-card transition-all hover:border-primary/50 hover:bg-card/80 hover:shadow-md flex flex-col justify-between"
          @click="$emit('read-article', item)"
        >
          <CardHeader class="pb-2">
            <div class="flex items-center justify-between gap-2 mb-2">
              <div class="flex flex-wrap items-center gap-1.5">
                <span v-if="item.tickers && item.tickers.length > 0" class="font-mono text-xs font-bold text-primary">
                  ${{ item.tickers.join(', $') }}
                </span>
                <Badge v-if="item.sector" variant="outline" class="text-[10px] px-1.5 py-0">
                  {{ item.sector }}
                </Badge>
              </div>
              <Badge :variant="getBadgeVariant(item.value_score)">
                {{ (item.value_score && item.value_score > 0 ? '+' : '') + (item.value_score || 0) }}
              </Badge>
            </div>

            <CardTitle class="text-sm font-semibold leading-snug line-clamp-2 text-foreground group-hover:text-primary transition-colors">
              {{ item.title }}
            </CardTitle>
          </CardHeader>

          <CardContent class="space-y-2 pb-3">
            <p class="text-xs text-muted-foreground line-clamp-3 leading-relaxed">
              {{ item.summary }}
            </p>
            <div v-if="item.investment_takeaway" class="text-[11px] text-emerald-400/90 font-medium bg-emerald-500/10 p-2 rounded border border-emerald-500/20 line-clamp-2">
              💡 {{ item.investment_takeaway }}
            </div>
          </CardContent>

          <CardFooter class="pt-0 border-t border-border/40 mt-auto flex items-center justify-between text-[11px] font-mono text-muted-foreground">
            <span>{{ formatDate(item.date || item.created_at) }}</span>
            <span class="text-primary opacity-0 group-hover:opacity-100 transition-opacity">Read Full →</span>
          </CardFooter>
        </Card>
      </div>

      <!-- Pagination Controls -->
      <div v-if="totalPages > 1 || total > limit" class="flex items-center justify-between border-t border-border/80 pt-4 font-mono text-xs">
        <Button
          variant="outline"
          size="sm"
          :disabled="page <= 1 || loading"
          class="font-mono text-xs"
          @click="changePage(page - 1)"
        >
          ◀ Previous
        </Button>

        <span class="text-muted-foreground">
          Page <strong class="text-foreground">{{ page }}</strong> of {{ totalPages || 1 }} ({{ total }} Total)
        </span>

        <Button
          variant="outline"
          size="sm"
          :disabled="page >= totalPages || loading"
          class="font-mono text-xs"
          @click="changePage(page + 1)"
        >
          Next ▶
        </Button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Newspaper, Search } from 'lucide-vue-next'
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { News } from '@/server/utils/types'

const IDX_IC_TAXONOMY: Record<string, string[]> = {
  'A. Energy': ['A1. Oil, Gas, and Coal', 'A2. Alternative Energy'],
  'B. Basic Materials': ['B1. Basic Materials'],
  'C. Industrials': ['C1. Industrial Goods', 'C2. Industrial Services', 'C3. Multi-sector Holdings'],
  'D. Consumer Non-Cyclicals': [
    'D1. Food and Staples Retailing',
    'D2. Food and Beverage',
    'D3. Tobacco',
    'D4. Nondurable Household Products',
  ],
  'E. Consumer Cyclicals': [
    'E1. Automobiles and Components',
    'E2. Household Goods',
    'E3. Leisure Goods',
    'E4. Apparel and Luxury Goods',
    'E5. Consumer Services',
    'E6. Media and Entertainment',
    'E7. Retailing',
  ],
  'F. Healthcare': [
    'F1. Healthcare Equipment & Providers',
    'F2. Pharmaceuticals & Health Care Research',
  ],
  'G. Financials': [
    'G1. Banks',
    'G2. Financing Service',
    'G3. Investment Service',
    'G4. Insurance',
    'G5. Holding and Investment Companies',
  ],
  'H. Properties and Real Estate': ['H1. Properties & Real Estate'],
  'I. Technology': ['I1. Software & IT Services', 'I2. Technology Hardware & Equipment'],
  'J. Infrastructures': [
    'J1. Transportation Infrastructure',
    'J2. Heavy Constructions & Civil Engineering',
    'J3. Telecommunication',
    'J4. Utilities',
  ],
  'K. Transportation and Logistic': ['K1. Transportation', 'K2. Logistics & Deliveries'],
  'Macroeconomics': ['General Market & Policy'],
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
    if (tickerFilter.value) params.append('ticker', tickerFilter.value.trim().toUpperCase())
    if (sectorFilter.value) params.append('sector', sectorFilter.value)
    if (subsectorFilter.value) params.append('subsector', subsectorFilter.value)
    if (directionFilter.value) params.append('direction', directionFilter.value)

    const res = await $fetch<{ data: News[]; total: number; page: number; total_pages: number }>(
      `/api/v1/news?${params.toString()}`
    )
    items.value = res.data || []
    total.value = res.total || 0
    page.value = res.page || 1
    totalPages.value = res.total_pages || 1
  } catch (e) {
    console.error('Failed to fetch news', e)
  } finally {
    loading.value = false
  }
}

const changePage = (p: number) => {
  page.value = p
  fetchNews(p)
}

const filteredList = computed(() => items.value)

const getBadgeVariant = (score?: number) => {
  const s = score || 0
  if (s >= 5) return 'bullish'
  if (s <= -5) return 'bearish'
  return 'neutral'
}

const formatDate = (d?: string | Date) => {
  if (!d) return ''
  return new Date(d).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

onMounted(() => {
  fetchNews(1)
})
</script>
