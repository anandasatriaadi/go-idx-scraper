<template>
  <div class="mx-auto max-w-5xl space-y-6 px-4 py-6 sm:px-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/80 pb-4">
      <div>
        <div class="flex items-center gap-2">
          <Radio class="h-5 w-5 text-primary animate-pulse" />
          <h1 class="text-xl font-bold tracking-tight text-foreground font-mono">
            DAILY MARKET INTELLIGENCE BRIEFING
          </h1>
        </div>
        <p class="text-xs text-muted-foreground mt-1">
          Value Investing Morning Digest ("Today's Summarization of Yesterday")
        </p>
      </div>

      <Button
        v-if="briefing?.raw_markdown"
        variant="outline"
        size="sm"
        class="font-mono text-xs border-border self-start sm:self-auto"
        @click="copyMarkdown"
      >
        <Check v-if="copied" class="mr-1.5 h-3.5 w-3.5 text-emerald-400" />
        <Copy v-else class="mr-1.5 h-3.5 w-3.5" />
        {{ copied ? 'Copied Markdown' : 'Copy Markdown' }}
      </Button>
    </div>

    <div v-if="loading" class="flex h-64 items-center justify-center rounded-lg border border-dashed border-border bg-card/40 font-mono text-xs text-muted-foreground">
      Loading intelligence briefing...
    </div>
    <div v-else-if="!briefing" class="flex h-64 items-center justify-center rounded-lg border border-dashed border-border bg-card/40 font-mono text-xs text-muted-foreground">
      No market briefing available.
    </div>
    <div v-else class="space-y-6">
      <!-- Main Briefing Card -->
      <Card class="border-border/80 bg-card shadow-sm">
        <CardHeader class="border-b border-border/60 pb-4">
          <div class="flex items-center justify-between gap-2">
            <Badge variant="outline" class="font-mono text-xs border-primary/30 text-primary bg-primary/5">
              📅 {{ formatDate(briefing.date) }}
            </Badge>
          </div>
          <CardTitle class="text-2xl font-bold tracking-tight text-foreground mt-2">
            {{ briefing.title }}
          </CardTitle>
        </CardHeader>

        <CardContent class="space-y-6 pt-6">
          <!-- Macro Pulse -->
          <div class="space-y-2">
            <div class="flex items-center gap-2 text-xs font-mono font-bold text-primary">
              <Globe class="h-4 w-4" />
              EXECUTIVE MACRO & MARKET PULSE
            </div>
            <p class="text-sm leading-relaxed text-muted-foreground bg-muted/30 p-4 rounded-lg border border-border/40">
              {{ briefing.macro_pulse }}
            </p>
          </div>

          <!-- Lookouts Grid -->
          <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
            <!-- Bullish -->
            <Card class="border-emerald-500/20 bg-emerald-950/10">
              <CardHeader class="pb-3 border-b border-emerald-500/20">
                <div class="flex items-center gap-2 text-xs font-mono font-bold text-emerald-400">
                  <TrendingUp class="h-4 w-4" />
                  STOCKS TO WATCH (BUY LOOKOUT)
                </div>
              </CardHeader>
              <CardContent class="pt-4">
                <div v-if="!briefing.bullish_lookout || briefing.bullish_lookout.length === 0" class="text-xs text-muted-foreground font-mono">
                  No high-conviction bullish candidates today.
                </div>
                <div v-else class="space-y-3">
                  <div
                    v-for="item in briefing.bullish_lookout"
                    :key="item.ticker"
                    class="rounded-lg border border-emerald-500/20 bg-card/80 p-3.5 space-y-2"
                  >
                    <div class="flex items-center justify-between">
                      <span class="font-mono text-sm font-bold text-emerald-400">
                        ${{ item.ticker }}
                      </span>
                      <Badge variant="bullish">
                        +{{ item.value_score }}
                      </Badge>
                    </div>
                    <div class="text-xs font-semibold text-foreground">{{ item.headline }}</div>
                    <p class="text-xs text-muted-foreground leading-relaxed">{{ item.rationale }}</p>
                    <div class="text-xs text-emerald-300 font-medium bg-emerald-500/10 p-2 rounded border border-emerald-500/20">
                      💡 {{ item.investment_takeaway }}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            <!-- Bearish -->
            <Card class="border-rose-500/20 bg-rose-950/10">
              <CardHeader class="pb-3 border-b border-rose-500/20">
                <div class="flex items-center gap-2 text-xs font-mono font-bold text-rose-400">
                  <AlertTriangle class="h-4 w-4" />
                  RISK ALERTS (HEADWINDS & GOVERNANCE)
                </div>
              </CardHeader>
              <CardContent class="pt-4">
                <div v-if="!briefing.bearish_lookout || briefing.bearish_lookout.length === 0" class="text-xs text-muted-foreground font-mono">
                  No major corporate risk alerts today.
                </div>
                <div v-else class="space-y-3">
                  <div
                    v-for="item in briefing.bearish_lookout"
                    :key="item.ticker"
                    class="rounded-lg border border-rose-500/20 bg-card/80 p-3.5 space-y-2"
                  >
                    <div class="flex items-center justify-between">
                      <span class="font-mono text-sm font-bold text-rose-400">
                        ${{ item.ticker }}
                      </span>
                      <Badge variant="bearish">
                        {{ item.value_score }}
                      </Badge>
                    </div>
                    <div class="text-xs font-semibold text-foreground">{{ item.headline }}</div>
                    <p class="text-xs text-muted-foreground leading-relaxed">{{ item.rationale }}</p>
                    <div class="text-xs text-rose-300 font-medium bg-rose-500/10 p-2 rounded border border-rose-500/20">
                      ⚠️ {{ item.investment_takeaway }}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          <!-- Sector Highlights -->
          <div v-if="briefing.sector_highlights && briefing.sector_highlights.length > 0" class="space-y-3">
            <div class="flex items-center gap-2 text-xs font-mono font-bold text-foreground">
              <Factory class="h-4 w-4 text-primary" />
              SECTOR & INDUSTRY BREAKDOWN
            </div>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div
                v-for="sec in briefing.sector_highlights"
                :key="sec.sector"
                class="rounded-lg border border-border bg-card/60 p-3.5 space-y-2"
              >
                <div class="flex items-center justify-between gap-2">
                  <span class="font-mono text-xs font-semibold text-foreground">{{ sec.sector }}</span>
                  <Badge :variant="getSentimentVariant(sec.sentiment)">
                    {{ sec.sentiment }}
                  </Badge>
                </div>
                <p class="text-xs text-muted-foreground leading-relaxed">{{ sec.summary }}</p>
              </div>
            </div>
          </div>

          <!-- Action Plan -->
          <div v-if="briefing.action_plan" class="rounded-lg border border-primary/30 bg-primary/5 p-4 space-y-2">
            <div class="flex items-center gap-2 text-xs font-mono font-bold text-primary">
              <Target class="h-4 w-4" />
              DISCIPLINED VALUE INVESTOR ACTION PLAN
            </div>
            <p class="text-xs sm:text-sm text-foreground/90 leading-relaxed font-sans">
              {{ briefing.action_plan }}
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Radio, Copy, Check, Globe, TrendingUp, AlertTriangle, Factory, Target } from 'lucide-vue-next'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { Briefing } from '@/server/utils/types'

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

const getSentimentVariant = (s?: string) => {
  const lower = (s || '').toLowerCase()
  if (lower.includes('bull') || lower.includes('positive')) return 'bullish'
  if (lower.includes('bear') || lower.includes('negative')) return 'bearish'
  return 'neutral'
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
