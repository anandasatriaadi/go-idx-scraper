<template>
  <div class="view-container">
    <div class="view-header">
      <div>
        <h1 class="title">📊 Financial Reports Archive</h1>
        <p class="sub">Quarterly and annual financial statements downloaded directly from IDX</p>
      </div>

      <input
        v-model="search"
        type="text"
        placeholder="Filter by Ticker (e.g. BBRI)..."
        class="search-box font-mono"
      />
    </div>

    <div v-if="loading" class="loading-box font-mono">Loading financial reports...</div>
    <div v-else-if="filteredList.length === 0" class="empty-box font-mono">No financial reports found.</div>
    <div v-else class="table-card">
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
          <tr v-for="item in filteredList" :key="item.id || item._id">
            <td class="font-mono issuer-col">${{ item.issuer_code }}</td>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  reports: any[]
  loading: boolean
}>()

const search = ref('')

const filteredList = computed(() => {
  if (!search.value) return props.reports
  const q = search.value.toLowerCase()
  return props.reports.filter(r => r.issuer_code?.toLowerCase().includes(q))
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
.search-box {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: #fff;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 0.85rem;
  width: 260px;
  outline: none;
}
.search-box:focus {
  border-color: #38bdf8;
}
.table-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow-x: auto;
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
}
.download-btn:hover {
  background: #059669;
  color: #fff;
}
.loading-box, .empty-box {
  text-align: center;
  padding: 40px;
  color: var(--text-muted);
}
</style>
