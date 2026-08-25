<template>
  <div class="mx-auto max-w-7xl space-y-6 px-4 py-6 sm:px-6">
    <div class="flex flex-col gap-4 border-b border-border/80 pb-6 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <div class="flex items-center gap-2">
          <FileText class="h-5 w-5 text-primary" />
          <h1 class="text-xl font-bold tracking-tight text-foreground font-mono">
            FINANCIAL REPORTS ARCHIVE
          </h1>
        </div>
        <p class="text-xs text-muted-foreground mt-1">
          Quarterly and annual financial statements downloaded directly from IDX
        </p>
      </div>

      <div class="relative w-full sm:w-72">
        <Search class="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
        <Input
          v-model="search"
          type="text"
          placeholder="Filter by Ticker (e.g. BBRI)..."
          class="h-8 pl-8 pr-8 font-mono text-xs bg-background/80 uppercase placeholder:normal-case"
          @input="onSearchInput"
          @keyup.enter="onSearchEnter"
        />
        <button
          v-if="search"
          class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-muted-foreground hover:text-foreground"
          @click="clearSearch"
        >
          ✕
        </button>
      </div>
    </div>

    <!-- Table -->
    <div v-if="loading && items.length === 0" class="flex h-64 items-center justify-center rounded-lg border border-dashed border-border bg-card/40 font-mono text-xs text-muted-foreground">
      Loading financial reports...
    </div>
    <div v-else-if="items.length === 0" class="flex h-64 items-center justify-center rounded-lg border border-dashed border-border bg-card/40 font-mono text-xs text-muted-foreground">
      No financial reports found{{ search ? ` matching "${search}"` : '' }}.
    </div>
    <Card v-else class="overflow-hidden border-border bg-card">
      <div class="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow class="border-border hover:bg-transparent font-mono text-[11px]">
              <TableHead class="w-32 font-bold text-foreground">Issuer Code</TableHead>
              <TableHead class="w-32 font-bold text-foreground">Year</TableHead>
              <TableHead class="w-48 font-bold text-foreground">Quarter / Period</TableHead>
              <TableHead class="font-bold text-foreground">Report Filing File</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="item in items"
              :key="item.id || item._id"
              class="border-border hover:bg-muted/40 transition-colors font-mono text-xs"
            >
              <TableCell>
                <Badge
                  variant="outline"
                  class="cursor-pointer font-mono text-xs font-bold text-primary border-primary/30 hover:bg-primary/15 transition-colors"
                  title="Inspect 360° Financials"
                  @click="$emit('open-ticker-financials', item.issuer_code)"
                >
                  ${{ item.issuer_code }} ↗
                </Badge>
              </TableCell>
              <TableCell class="text-muted-foreground">
                {{ item.year || '-' }}
              </TableCell>
              <TableCell class="text-muted-foreground">
                {{ item.quarter ? 'Q' + item.quarter : (item.period_string || 'Annual') }}
              </TableCell>
              <TableCell>
                <a
                  v-if="item.report_url"
                  :href="item.report_url"
                  target="_blank"
                  rel="noopener"
                  class="inline-flex items-center gap-1 rounded bg-secondary/80 px-2 py-0.5 font-mono text-[11px] text-primary hover:bg-primary/20 transition-colors"
                >
                  <Download class="h-3 w-3" />
                  Download Excel
                </a>
                <span v-else class="text-muted-foreground">-</span>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </Card>

    <!-- Pagination -->
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
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { FileText, Search, Download } from 'lucide-vue-next'
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
import type { FinancialReport } from '@/server/utils/types'

const props = defineProps<{
  reports?: FinancialReport[]
  loading?: boolean
}>()

defineEmits<{
  (e: 'open-ticker-financials', ticker: string): void
}>()

const items = ref<FinancialReport[]>(props.reports || [])
const loading = ref(props.loading || false)
const page = ref(1)
const limit = ref(20)
const total = ref(0)
const totalPages = ref(1)
const search = ref('')

let searchDebounce: ReturnType<typeof setTimeout> | null = null

const fetchReports = async (p = page.value) => {
  loading.value = true
  try {
    const params = new URLSearchParams()
    params.append('page', p.toString())
    params.append('limit', limit.value.toString())
    if (search.value.trim()) {
      params.append('search', search.value.trim())
    }

    const res = await $fetch<{
      data: FinancialReport[]
      total: number
      page: number
      total_pages: number
    }>(`/api/v1/financial-reports?${params.toString()}`)

    items.value = res.data || []
    total.value = res.total || 0
    totalPages.value = res.total_pages || 1
    page.value = res.page || p
  } catch (err) {
    console.error('Failed to fetch financial reports', err)
  } finally {
    loading.value = false
  }
}

const onSearchInput = () => {
  if (searchDebounce) clearTimeout(searchDebounce)
  searchDebounce = setTimeout(() => {
    page.value = 1
    fetchReports(1)
  }, 350)
}

const onSearchEnter = () => {
  if (searchDebounce) clearTimeout(searchDebounce)
  page.value = 1
  fetchReports(1)
}

const clearSearch = () => {
  search.value = ''
  page.value = 1
  fetchReports(1)
}

const changePage = (p: number) => {
  page.value = p
  fetchReports(p)
}

onMounted(() => {
  fetchReports(1)
})
</script>
