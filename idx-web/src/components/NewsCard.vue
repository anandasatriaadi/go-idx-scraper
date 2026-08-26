<template>
  <Card
    class="group cursor-pointer border-border bg-card transition-all hover:border-primary/50 hover:bg-card/80 hover:shadow-lg flex flex-col justify-between overflow-hidden"
    @click="$emit('read', news)"
  >
    <CardHeader :class="compact ? 'p-3 pb-1.5' : 'p-4 pb-2'">
      <div class="flex items-center justify-between gap-2 mb-2">
        <div class="flex flex-wrap items-center gap-1.5">
          <!-- Ticker Pills -->
          <div v-if="news.tickers && news.tickers.length > 0" class="flex flex-wrap items-center gap-1">
            <Badge
              v-for="t in news.tickers"
              :key="t"
              variant="outline"
              class="font-mono text-[11px] font-bold text-primary border-primary/30 hover:bg-primary/20 transition-colors"
              @click.stop="$emit('filter-ticker', t)"
            >
              ${{ t }}
            </Badge>
          </div>

          <!-- Sector / Industry Pill -->
          <Badge v-if="news.subsector || news.industry || news.sector" variant="secondary" class="text-[10px] px-1.5 py-0 truncate max-w-[140px]">
            {{ news.subsector || news.industry || news.sector }}
          </Badge>
        </div>

        <!-- Value Score Badge -->
        <Badge :variant="getBadgeVariant(news.value_score)" class="font-mono font-bold shrink-0">
          {{ (news.value_score && news.value_score > 0 ? '+' : '') + (news.value_score || 0) }}
        </Badge>
      </div>

      <CardTitle :class="[compact ? 'text-xs' : 'text-sm', 'font-semibold leading-snug line-clamp-2 text-foreground group-hover:text-primary transition-colors']">
        {{ news.title }}
      </CardTitle>
    </CardHeader>

    <CardContent :class="[compact ? 'p-3 pt-0 pb-2 space-y-1.5' : 'p-4 pt-0 pb-3 space-y-2.5']">
      <p :class="[compact ? 'text-[11px] line-clamp-2' : 'text-xs line-clamp-3', 'text-muted-foreground leading-relaxed']">
        {{ news.summary }}
      </p>

      <!-- Investment Takeaway Pill -->
      <div
        v-if="news.investment_takeaway"
        :class="[
          'rounded p-2.5 text-[11px] font-medium leading-relaxed border',
          (news.value_score && news.value_score < 0) || (news.impact_direction && news.impact_direction.toLowerCase() === 'bearish')
            ? 'border-rose-500/30 bg-rose-500/10 text-rose-300'
            : (news.value_score && news.value_score > 0) || (news.impact_direction && news.impact_direction.toLowerCase() === 'bullish')
            ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300'
            : 'border-sky-500/30 bg-sky-500/10 text-sky-300'
        ]"
      >
        <span class="mr-1">{{ (news.value_score && news.value_score < 0) || (news.impact_direction && news.impact_direction.toLowerCase() === 'bearish') ? '⚠️' : '💡' }}</span>
        <span>{{ news.investment_takeaway }}</span>
      </div>
    </CardContent>

    <CardFooter :class="[compact ? 'p-3 pt-1.5' : 'p-4 pt-2', 'border-t border-border/40 mt-auto flex items-center justify-between text-[11px] font-mono text-muted-foreground']">
      <div class="flex items-center gap-2">
        <span>{{ formatDate(news.date || news.created_at) }}</span>
        <span
          v-if="news.impact_direction"
          :class="[
            'px-1.5 py-0.5 rounded text-[10px] font-bold uppercase',
            news.impact_direction.toLowerCase() === 'bullish' ? 'text-emerald-400 bg-emerald-500/10' :
            news.impact_direction.toLowerCase() === 'bearish' ? 'text-rose-400 bg-rose-500/10' :
            'text-sky-400 bg-sky-500/10'
          ]"
        >
          {{ news.impact_direction }}
        </span>
      </div>
      <span class="text-primary font-semibold opacity-0 group-hover:opacity-100 transition-opacity">
        Read Full →
      </span>
    </CardFooter>
  </Card>
</template>

<script setup lang="ts">
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import type { News } from '@/server/utils/types'

withDefaults(
  defineProps<{
    news: News
    compact?: boolean
  }>(),
  {
    compact: false,
  }
)

defineEmits<{
  (e: 'read', news: News): void
  (e: 'filter-ticker', ticker: string): void
}>()

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
