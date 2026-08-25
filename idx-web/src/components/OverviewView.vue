<template>
  <div class="mx-auto max-w-7xl space-y-8 px-4 py-6 sm:px-6">
    <!-- HERO: Latest Daily Market Briefing -->
    <Card v-if="briefing" class="relative overflow-hidden border-border/80 bg-gradient-to-b from-card to-card/60 shadow-md">
      <div class="absolute inset-x-0 top-0 h-1 bg-gradient-to-r from-emerald-500 via-primary to-blue-500"></div>
      
      <CardHeader class="pb-4">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <Badge variant="outline" class="border-emerald-500/30 bg-emerald-500/10 text-emerald-400 font-mono">
            <span class="mr-1.5 h-2 w-2 animate-ping rounded-full bg-emerald-400"></span>
            7:00 AM GMT+8 BRIEFING
          </Badge>
          <span class="font-mono text-xs text-muted-foreground">
            {{ formatDate(briefing.date) }}
          </span>
        </div>
        <CardTitle class="mt-2 text-xl font-bold tracking-tight text-foreground sm:text-2xl">
          {{ briefing.title }}
        </CardTitle>
        <p class="mt-1.5 text-sm leading-relaxed text-muted-foreground">
          {{ briefing.macro_pulse }}
        </p>
      </CardHeader>

      <CardContent class="space-y-6">
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <!-- Bullish Opportunities -->
          <Card class="border-emerald-500/20 bg-emerald-950/10 backdrop-blur-xs">
            <CardHeader class="pb-2">
              <div class="flex items-center gap-2 text-xs font-mono font-bold text-emerald-400">
                <TrendingUp class="h-4 w-4" />
                BUY LOOKOUT / MOAT EXPANSION
              </div>
            </CardHeader>
            <CardContent>
              <div v-if="!briefing.bullish_lookout || briefing.bullish_lookout.length === 0" class="text-xs text-muted-foreground font-mono py-2">
                No high-conviction bullish candidates today.
              </div>
              <div v-else class="space-y-3">
                <div
                  v-for="item in briefing.bullish_lookout"
                  :key="item.ticker"
                  class="rounded-md border border-emerald-500/20 bg-card/60 p-3 transition-colors hover:border-emerald-500/40"
                >
                  <div class="flex items-center justify-between mb-1.5">
                    <Button
                      variant="outline"
                      size="xs"
                      class="h-6 border-emerald-500/30 font-mono text-xs font-bold text-emerald-400 hover:bg-emerald-500/20"
                      @click="$emit('filter-ticker', item.ticker)"
                    >
                      ${{ item.ticker }}
                    </Button>
                    <Badge variant="bullish">
                      +{{ item.value_score }}
                    </Badge>
                  </div>
                  <div class="text-xs font-medium text-foreground mb-1">{{ item.headline }}</div>
                  <div class="text-xs text-emerald-300/80 leading-relaxed">
                    💡 {{ item.investment_takeaway }}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <!-- Bearish Risk Alerts -->
          <Card class="border-rose-500/20 bg-rose-950/10 backdrop-blur-xs">
            <CardHeader class="pb-2">
              <div class="flex items-center gap-2 text-xs font-mono font-bold text-rose-400">
                <AlertTriangle class="h-4 w-4" />
                RISK ALERTS / HEADWINDS
              </div>
            </CardHeader>
            <CardContent>
              <div v-if="!briefing.bearish_lookout || briefing.bearish_lookout.length === 0" class="text-xs text-muted-foreground font-mono py-2">
                No major corporate risk alerts today.
              </div>
              <div v-else class="space-y-3">
                <div
                  v-for="item in briefing.bearish_lookout"
                  :key="item.ticker"
                  class="rounded-md border border-rose-500/20 bg-card/60 p-3 transition-colors hover:border-rose-500/40"
                >
                  <div class="flex items-center justify-between mb-1.5">
                    <Button
                      variant="outline"
                      size="xs"
                      class="h-6 border-rose-500/30 font-mono text-xs font-bold text-rose-400 hover:bg-rose-500/20"
                      @click="$emit('filter-ticker', item.ticker)"
                    >
                      ${{ item.ticker }}
                    </Button>
                    <Badge variant="bearish">
                      {{ item.value_score }}
                    </Badge>
                  </div>
                  <div class="text-xs font-medium text-foreground mb-1">{{ item.headline }}</div>
                  <div class="text-xs text-rose-300/80 leading-relaxed">
                    ⚠️ {{ item.investment_takeaway }}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <!-- Action Plan -->
        <div v-if="briefing.action_plan" class="rounded-lg border border-primary/20 bg-primary/5 p-4">
          <div class="flex items-center gap-2 font-mono text-xs font-bold text-primary mb-1.5">
            <Compass class="h-4 w-4" />
            VALUE INVESTOR ACTION PLAN
          </div>
          <p class="text-xs leading-relaxed text-muted-foreground sm:text-sm">
            {{ briefing.action_plan }}
          </p>
        </div>
      </CardContent>
    </Card>

    <!-- NEWS STREAM SECTION -->
    <section class="space-y-4">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-border/80 pb-3">
        <div class="flex items-center gap-2">
          <Newspaper class="h-5 w-5 text-primary" />
          <h2 class="text-base font-bold tracking-tight text-foreground font-mono">
            REAL-TIME INTELLIGENCE STREAM
          </h2>
        </div>
        
        <!-- Filter Chips -->
        <div class="flex flex-wrap items-center gap-1.5">
          <Button
            v-for="c in categories"
            :key="c.value"
            :variant="activeFilter === c.value ? 'default' : 'outline'"
            size="xs"
            class="font-mono text-xs"
            @click="setFilter(c.value)"
          >
            {{ c.label }}
          </Button>
        </div>
      </div>

      <!-- News Cards Grid -->
      <div v-if="loading" class="flex h-40 items-center justify-center rounded-lg border border-dashed border-border bg-card/40 font-mono text-xs text-muted-foreground">
        Fetching live market intelligence...
      </div>
      <div v-else-if="filteredNews.length === 0" class="flex h-40 items-center justify-center rounded-lg border border-dashed border-border bg-card/40 font-mono text-xs text-muted-foreground">
        No news articles matching current filter.
      </div>
      <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card
          v-for="item in filteredNews"
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
                <Badge v-if="item.industry" variant="secondary" class="text-[10px] px-1.5 py-0">
                  {{ item.industry }}
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

          <CardContent class="pb-3">
            <p class="text-xs text-muted-foreground line-clamp-3 leading-relaxed">
              {{ item.summary }}
            </p>
          </CardContent>

          <CardFooter class="pt-0 border-t border-border/40 mt-auto flex items-center justify-between text-[11px] font-mono text-muted-foreground">
            <span>{{ formatDate(item.date || item.created_at) }}</span>
            <span class="text-primary opacity-0 group-hover:opacity-100 transition-opacity">Read Analysis →</span>
          </CardFooter>
        </Card>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { TrendingUp, AlertTriangle, Compass, Newspaper } from 'lucide-vue-next'
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { Briefing, News } from '@/server/utils/types'

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
  { label: 'ALL', value: 'ALL' },
  { label: 'HIGH CONVICTION', value: 'HIGH_SCORE' },
  { label: 'FINANCIALS', value: 'FINANCIALS' },
  { label: 'ENERGY', value: 'ENERGY' },
  { label: 'MACRO', value: 'MACRO' },
]

const setFilter = (val: string) => {
  activeFilter.value = val
}

const filteredNews = computed(() => {
  if (!props.news) return []
  if (activeFilter.value === 'ALL') return props.news
  if (activeFilter.value === 'HIGH_SCORE') {
    return props.news.filter(n => Math.abs(n.value_score || 0) >= 6)
  }
  if (activeFilter.value === 'FINANCIALS') {
    return props.news.filter(n => n.sector?.includes('Financials') || n.industry?.includes('Banks'))
  }
  if (activeFilter.value === 'ENERGY') {
    return props.news.filter(n => n.sector?.includes('Energy') || n.industry?.includes('Coal') || n.industry?.includes('Oil'))
  }
  if (activeFilter.value === 'MACRO') {
    return props.news.filter(n => n.sector?.includes('Macroeconomics') || n.is_industry_wide)
  }
  return props.news
})

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
</script>
