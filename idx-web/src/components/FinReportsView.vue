<template>
  <div class="view-container">
    <div class="view-header">
      <div>
        <h1 class="title">📊 Financial Reports Archive</h1>
        <p class="sub">Quarterly and annual financial statements downloaded directly from IDX</p>
      </div>

      <div class="search-wrap">
        <input
          v-model="search"
          type="text"
          placeholder="Filter by Ticker (e.g. BBRI)..."
          class="search-box font-mono"
          @input="onSearchInput"
          @keyup.enter="onSearchEnter"
        />
        <button v-if="search" class="clear-btn" @click="clearSearch">✕</button>
      </div>
    </div>

    <!-- Table or Loading -->
    <div v-if="loading && items.length === 0" class="loading-box font-mono">
      <div class="spinner"></div>
      <span>Loading financial reports...</span>
    </div>
    <div v-else-if="items.length === 0" class="empty-box font-mono">
      No financial reports found{{ search ? ` matching "${search}"` : '' }}.
    </div>
    <div v-else class="table-card">
      <div v-if="loading" class="table-overlay">
        <span class="overlay-text font-mono">Updating page...</span>
      </div>
      <table class="data-table">
        <thead>
          <tr class="font-mono">
            <th>Issuer Code</th>
            <th>Year</th>
            <th>Quarter / Period</th>
            <th>Report File</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id || item._id">
            <td class="font-mono issuer-col clickable" @click="$emit('open-ticker-financials', item.issuer_code)" title="Inspect 360° Financials">
              ${{ item.issuer_code }} 🔍
            </td>
            <td class="font-mono">{{ item.year || '-' }}</td>
            <td class="font-mono">{{ item.quarter ? 'Q' + item.quarter : (item.period_string || 'Annual') }}</td>
            <td>
              <a
                v-if="item.report_url"
                :href="item.report_url"
                target="_blank"
                rel="noopener"
                class="download-btn font-mono"
              >
                📥 Download Excel
              </a>
              <span v-else class="text-muted">-</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination Bar -->
    <div class="pagination-bar font-mono">
      <button
        class="pagination-btn"
        :disabled="page <= 1 || loading"
        @click="changePage(page - 1)"
      >
        ◀ Previous
      </button>

      <div class="pagination-info">
        Page <span class="page-current">{{ page }}</span> of <span class="page-total">{{ totalPages || 1 }}</span>
        <span class="items-total">({{ total }} Total)</span>
      </div>

      <button
        class="pagination-btn"
        :disabled="page >= totalPages || loading"
        @click="changePage(page + 1)"
      >
        Next ▶
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { FinancialReport } from '../server/utils/types'

const props = defineProps<{
  reports?: FinancialReport[]
  loading?: boolean
}>()

const items = ref<FinancialReport[]>(props.reports || [])
const loading = ref(props.loading || false)
const page = ref(1)
const limit = ref(20)
const total = ref(0)
const totalPages = ref(1)
defineEmits<{
  (e: 'open-ticker-financials', ticker: string): void
}>()

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

const changePage = (newPage: number) => {
  if (newPage < 1 || newPage > totalPages.value || newPage === page.value) return
  page.value = newPage
  fetchReports(newPage)
}

onMounted(() => {
  fetchReports(1)
})
</script>

<style scoped>
.view-container {
  max-width: 1440px;
  margin: 0 auto;
  padding: 24px 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}
.title {
  font-size: 1.5rem;
  font-weight: 700;
}
.sub {
  color: var(--text-secondary);
  font-size: 0.9rem;
}
.search-wrap {
  position: relative;
  display: flex;
  align-items: center;
}
.search-box {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: #fff;
  padding: 8px 32px 8px 12px;
  border-radius: 6px;
  font-size: 0.85rem;
  width: 280px;
  outline: none;
  transition: border-color 0.15s ease;
}
.search-box:focus {
  border-color: #38bdf8;
}
.clear-btn {
  position: absolute;
  right: 8px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 0.85rem;
  padding: 2px 6px;
}
.clear-btn:hover {
  color: #fff;
}
.table-card {
  position: relative;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow-x: auto;
  min-height: 200px;
}
.table-overlay {
  position: absolute;
  inset: 0;
  background: rgba(8, 12, 20, 0.6);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}
.overlay-text {
  color: #38bdf8;
  font-size: 0.85rem;
  background: var(--bg-card);
  padding: 6px 14px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
}
.data-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}
.data-table th {
  background: var(--bg-app);
  padding: 12px 16px;
  color: var(--text-muted);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid var(--border-color);
}
.data-table td {
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-color);
  font-size: 0.9rem;
}
.data-table tr:hover td {
  background: var(--bg-card-hover);
}
.issuer-col {
  color: #38bdf8;
  font-weight: 700;
}
.download-btn {
  background: #1e293b;
  color: #34d399;
  padding: 4px 10px;
  border-radius: 4px;
  text-decoration: none;
  font-size: 0.8rem;
  font-weight: 600;
  display: inline-block;
  transition: all 0.15s ease;
}
.download-btn:hover {
  background: #059669;
  color: #fff;
}
.pagination-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  gap: 12px;
}
.pagination-btn {
  background: #1e293b;
  border: 1px solid var(--border-subtle);
  color: #38bdf8;
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}
.pagination-btn:hover:not(:disabled) {
  background: #2563eb;
  color: #fff;
  border-color: #2563eb;
}
.pagination-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.pagination-info {
  color: var(--text-secondary);
  font-size: 0.85rem;
  display: flex;
  align-items: center;
  gap: 6px;
}
.page-current {
  color: #38bdf8;
  font-weight: 700;
}
.page-total {
  color: var(--text-primary);
  font-weight: 600;
}
.items-total {
  color: var(--text-muted);
  font-size: 0.8rem;
}
.loading-box, .empty-box {
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
</style>
