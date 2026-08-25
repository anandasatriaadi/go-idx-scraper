<template>
  <header class="sticky top-0 z-50 w-full border-b border-border bg-card/90 backdrop-blur-md">
    <div class="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6">
      <!-- Brand Logo -->
      <div
        class="flex cursor-pointer items-center gap-2.5 transition-opacity hover:opacity-80"
        @click="$emit('select-tab', 'overview')"
      >
        <Badge variant="default" class="rounded px-1.5 py-0.5 font-mono text-[11px] font-bold tracking-wider">
          IDX
        </Badge>
        <span class="font-mono text-sm font-bold tracking-tight text-foreground">
          TERMINAL
        </span>
        <span class="relative flex h-2 w-2" title="Market Ingestion Engine Live">
          <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"></span>
          <span class="relative inline-flex h-2 w-2 rounded-full bg-emerald-500"></span>
        </span>
      </div>

      <!-- Navigation Tabs -->
      <nav class="hidden items-center gap-1 md:flex">
        <Button
          v-for="item in navItems"
          :key="item.id"
          :variant="activeTab === item.id ? 'secondary' : 'ghost'"
          size="sm"
          class="font-mono text-xs transition-all"
          :class="activeTab === item.id ? 'bg-secondary text-primary font-semibold shadow-xs border border-border/80' : 'text-muted-foreground hover:text-foreground'"
          @click="$emit('select-tab', item.id)"
        >
          <component :is="item.icon" class="mr-1.5 h-3.5 w-3.5" />
          {{ item.label }}
        </Button>
      </nav>

      <!-- Right Controls -->
      <div class="flex items-center gap-3">
        <!-- Quick Search -->
        <div class="relative w-48 sm:w-64">
          <Search class="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            v-model="searchQuery"
            type="text"
            placeholder="Search ticker (e.g. BBRI)..."
            class="h-8 pl-8 pr-8 font-mono text-xs bg-background/60 border-border focus-visible:ring-1 focus-visible:ring-primary uppercase placeholder:normal-case"
            @keyup.enter="handleSearch"
          />
          <kbd class="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 rounded border border-border bg-muted px-1.5 py-0.5 text-[10px] font-mono text-muted-foreground">
            ↵
          </kbd>
        </div>

        <!-- User Auth -->
        <Button
          v-if="!user"
          variant="outline"
          size="sm"
          class="font-mono text-xs border-primary/40 text-primary hover:bg-primary/10"
          @click="$emit('open-auth')"
        >
          <User class="mr-1.5 h-3.5 w-3.5" />
          Sign In
        </Button>
        <div v-else class="flex items-center gap-1.5 rounded-md border border-border bg-muted/40 p-1 pl-2.5">
          <span class="font-mono text-xs font-medium text-foreground">
            {{ userEmail }}
          </span>
          <Button
            variant="ghost"
            size="iconSm"
            class="h-6 w-6 rounded text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
            title="Sign Out"
            @click="logout"
          >
            <LogOut class="h-3 w-3" />
          </Button>
        </div>
      </div>
    </div>

    <!-- Mobile Nav Bar -->
    <div class="flex overflow-x-auto border-t border-border/50 px-2 py-1.5 md:hidden scrollbar-none gap-1 bg-card/50">
      <Button
        v-for="item in navItems"
        :key="item.id"
        :variant="activeTab === item.id ? 'secondary' : 'ghost'"
        size="xs"
        class="whitespace-nowrap font-mono text-[11px]"
        @click="$emit('select-tab', item.id)"
      >
        {{ item.label }}
      </Button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  LayoutDashboard,
  Radio,
  SlidersHorizontal,
  Newspaper,
  BellRing,
  FileText,
  Search,
  User,
  LogOut,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { useAuth } from '@/composables/useAuth'

defineProps<{
  activeTab: string
}>()

const emit = defineEmits<{
  (e: 'select-tab', tab: string): void
  (e: 'open-auth'): void
  (e: 'search-ticker', ticker: string): void
}>()

const { user, logout } = useAuth()
const searchQuery = ref('')

const navItems = [
  { id: 'overview', label: 'Overview', icon: LayoutDashboard },
  { id: 'briefing', label: 'Daily Briefing', icon: Radio },
  { id: 'screener', label: 'Value Screener', icon: SlidersHorizontal },
  { id: 'news', label: 'News Terminal', icon: Newspaper },
  { id: 'announcements', label: 'Disclosures', icon: BellRing },
  { id: 'reports', label: 'Filings', icon: FileText },
]

const userEmail = computed(() => {
  if (!user.value?.email) return 'User'
  return user.value.email.split('@')[0]
})

const handleSearch = () => {
  if (searchQuery.value.trim()) {
    emit('search-ticker', searchQuery.value.trim().toUpperCase())
  }
}
</script>
