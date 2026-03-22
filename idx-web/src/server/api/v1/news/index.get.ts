import { findAllNews, type NewsFilter } from '../../../utils/news-repo'
import { getAuthFromEvent } from '../../../utils/firebase-admin'
import { getUserWatchlist } from '../../../utils/user-service'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  
  const filter: NewsFilter = {}
  
  if (query.date_gte) {
    filter.date_gte = String(query.date_gte)
  }
  if (query.date_lte) {
    filter.date_lte = String(query.date_lte)
  }
  if (query.priority) {
    filter.priority = parseInt(String(query.priority), 10)
  }
  if (query.source) {
    filter.source = String(query.source)
  }
  
  const newsList = await findAllNews(filter)
  
  let isWatched = false
  const auth = getAuthFromEvent(event)
  
  if (auth) {
    const watchlist = await getUserWatchlist(auth.uid)
    isWatched = watchlist.length > 0
  }
  
  const response = newsList.map(n => ({
    ...n,
    is_watched: isWatched,
  }))
  
  return response
})
