<template>
  <div v-if="article" class="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm p-4" @click.self="$emit('close')">
    <Card class="relative flex max-h-[90vh] w-full max-w-3xl flex-col border-border bg-card shadow-2xl overflow-hidden">
      <!-- Header -->
      <CardHeader class="border-b border-border/80 pb-4">
        <div class="flex items-center justify-between gap-2">
          <div class="flex flex-wrap items-center gap-2">
            <Badge :variant="getDirectionVariant(article.impact_direction)">
              {{ article.impact_direction || 'Neutral' }}
            </Badge>
            <Badge :variant="getValueScoreVariant(article.value_score)" class="font-mono">
              Value Score: {{ (article.value_score !== undefined && article.value_score > 0 ? '+' : '') + (article.value_score || 0) }}
            </Badge>
            <Badge v-if="article.industry" variant="outline" class="text-xs">
              {{ article.industry }}
            </Badge>
          </div>
          <Button
            variant="ghost"
            size="iconSm"
            class="h-8 w-8 rounded-full text-muted-foreground hover:bg-muted hover:text-foreground"
            @click="$emit('close')"
          >
            ✕
          </Button>
        </div>

        <CardTitle class="text-lg font-bold leading-snug text-foreground sm:text-xl mt-3">
          {{ article.title }}
        </CardTitle>

        <div class="flex flex-wrap items-center gap-3 pt-2 text-xs font-mono text-muted-foreground">
          <span>📅 {{ formatDate(article.date || article.created_at) }}</span>
          <div v-if="article.tickers && article.tickers.length > 0" class="flex items-center gap-1.5">
            <span>Tickers:</span>
            <Button
              v-for="t in article.tickers"
              :key="t"
              variant="outline"
              size="xs"
              class="h-5 font-mono text-[11px] font-bold text-primary border-primary/30 hover:bg-primary/10"
              @click="$emit('filter-ticker', t)"
            >
              ${{ t }}
            </Button>
          </div>
        </div>
      </CardHeader>

      <!-- Scrollable Body -->
      <div class="overflow-y-auto p-6 space-y-5">
        <!-- Value Takeaway Box -->
        <div v-if="article.investment_takeaway" class="rounded-lg border border-emerald-500/30 bg-emerald-950/20 p-4 space-y-1">
          <div class="flex items-center gap-1.5 text-xs font-mono font-bold text-emerald-400">
            💡 VALUE INVESTOR TAKEAWAY
          </div>
          <p class="text-xs sm:text-sm leading-relaxed text-emerald-300">
            {{ article.investment_takeaway }}
          </p>
        </div>

        <!-- 3-Sentence Executive Summary -->
        <div v-if="article.summary" class="rounded-lg border border-border bg-muted/30 p-4 space-y-1">
          <div class="text-xs font-mono font-bold text-foreground">
            📋 EXECUTIVE SUMMARY
          </div>
          <p class="text-xs sm:text-sm leading-relaxed text-muted-foreground">
            {{ article.summary }}
          </p>
        </div>

        <!-- Full Story Content -->
        <div class="space-y-2">
          <div class="text-xs font-mono font-bold text-foreground">
            FULL STORY
          </div>
          <div class="rounded-lg border border-border bg-background/50 p-4 text-xs sm:text-sm leading-relaxed text-muted-foreground whitespace-pre-line">
            {{ article.content }}
          </div>
        </div>

        <!-- External Link -->
        <div v-if="article.link" class="pt-2">
          <a
            :href="article.link"
            target="_blank"
            rel="noopener"
            class="inline-flex items-center gap-1.5 font-mono text-xs font-semibold text-primary hover:underline"
          >
            Open Original Source on Kontan ↗
          </a>
        </div>
      </div>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { Card, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { News } from '@/server/utils/types'

defineProps<{
  article: News | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'filter-ticker', ticker: string): void
}>()

const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    emit('close')
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
  if (typeof document !== 'undefined') {
    document.body.style.overflow = 'hidden'
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
  if (typeof document !== 'undefined') {
    document.body.style.overflow = ''
  }
})

const getDirectionVariant = (dir?: string) => {
  const d = (dir || '').toLowerCase()
  if (d === 'bullish') return 'bullish'
  if (d === 'bearish') return 'bearish'
  return 'neutral'
}

const getValueScoreVariant = (score?: number) => {
  const s = score || 0
  if (s >= 5) return 'bullish'
  if (s <= -5) return 'bearish'
  return 'neutral'
}

const formatDate = (dateStr?: string | Date) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}
</script>
