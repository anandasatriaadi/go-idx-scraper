import { ref } from 'vue'
import { useAuth } from './useAuth'

export const useWatchlist = () => {
  const { token, user } = useAuth()
  const watchlist = ref<string[]>([])
  const loading = ref(false)

  const fetchWatchlist = async () => {
    if (!token.value) {
      watchlist.value = []
      return
    }
    loading.value = true
    try {
      const data = await $fetch<{ watchlist: string[] }>('/api/v1/user/watchlist', {
        headers: {
          Authorization: `Bearer ${token.value}`
        }
      })
      if (data && Array.isArray(data.watchlist)) {
        watchlist.value = data.watchlist
      }
    } catch (e) {
      console.warn('Could not fetch watchlist', e)
    } finally {
      loading.value = false
    }
  }

  const toggleWatchlist = async (ticker: string) => {
    if (!token.value) {
      alert('Please sign in to manage your watchlist')
      return
    }
    const upper = ticker.toUpperCase().trim()
    const index = watchlist.value.indexOf(upper)
    const newWatchlist = [...watchlist.value]
    if (index >= 0) {
      newWatchlist.splice(index, 1)
    } else {
      newWatchlist.push(upper)
    }

    watchlist.value = newWatchlist

    try {
      await $fetch('/api/v1/user/watchlist', {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${token.value}`
        },
        body: {
          watchlist: newWatchlist
        }
      })
    } catch (e) {
      console.error('Failed to sync watchlist', e)
      // Rollback on failure
      fetchWatchlist()
    }
  }

  const isWatched = (ticker: string) => {
    if (!ticker) return false
    return watchlist.value.includes(ticker.toUpperCase().trim())
  }

  return {
    watchlist,
    loading,
    fetchWatchlist,
    toggleWatchlist,
    isWatched
  }
}
