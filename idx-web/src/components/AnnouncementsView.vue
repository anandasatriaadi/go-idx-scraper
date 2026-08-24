<template>
  <div class="view-container">
    <div class="view-header">
      <div>
        <h1 class="title">📢 IDX Official Disclosures</h1>
        <p class="sub">Corporate action announcements, material disclosures, and shareholder reports</p>
      </div>

      <input
        v-model="search"
        type="text"
        placeholder="Filter announcements (e.g. BBRI)..."
        class="search-box font-mono"
      />
    </div>

    <div v-if="loading" class="loading-box font-mono">Loading announcements...</div>
    <div v-else-if="filteredList.length === 0" class="empty-box font-mono">No announcements found.</div>
    <div v-else class="table-card">
      <table class="data-table">
        <thead>
          <tr class="font-mono">
            <th>Date</th>
            <th>No. Pengumuman</th>
            <th>Issuer / Title</th>
            <th>Attachment</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in filteredList" :key="item.id || item._id">
            <td class="font-mono date-col">{{ formatDate(item.tgl_pengumuman || item.created_at) }}</td>
            <td class="font-mono no-col">{{ item.no_pengumuman || '-' }}</td>
            <td>
              <div class="title-cell">
                <span class="announcement-title">{{ item.judul_pengumuman || item.title || 'Official Announcement' }}</span>
              </div>
            </td>
            <td>
              <div v-if="item.attachments && item.attachments.length > 0" class="attachments-row">
                <a
                  v-for="(att, i) in item.attachments"
                  :key="i"
                  :href="att.url || '#'"
                  target="_blank"
                  rel="noopener"
                  class="att-link font-mono"
                >
                  📥 {{ att.file_name || 'PDF' }}
                </a>
              </div>
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
import type { Announcement } from '../server/utils/types'

const props = defineProps<{
  announcements: Announcement[]
  loading: boolean
}>()

const search = ref('')

const filteredList = computed(() => {
  if (!search.value) return props.announcements
  const q = search.value.toLowerCase()
  return props.announcements.filter(a => {
    return (
      a.judul_pengumuman?.toLowerCase().includes(q) ||
      a.no_pengumuman?.toLowerCase().includes(q) ||
      a.title?.toLowerCase().includes(q)
    )
  })
})

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}
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
.date-col {
  color: var(--text-muted);
  font-size: 0.8rem;
  white-space: nowrap;
}
.no-col {
  color: #38bdf8;
  font-size: 0.8rem;
  white-space: nowrap;
}
.announcement-title {
  font-weight: 500;
  color: var(--text-primary);
}
.attachments-row {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.att-link {
  background: #1e293b;
  color: #38bdf8;
  padding: 2px 6px;
  border-radius: 4px;
  text-decoration: none;
  font-size: 0.75rem;
}
.att-link:hover {
  background: #2563eb;
  color: #fff;
}
.loading-box, .empty-box {
  text-align: center;
  padding: 40px;
  color: var(--text-muted);
}
</style>
