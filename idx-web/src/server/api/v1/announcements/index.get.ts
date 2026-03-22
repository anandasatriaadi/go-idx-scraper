import { findAllAnnouncements } from '../../../utils/announcement-repo'
import { getAuthFromEvent } from '../../../utils/firebase-admin'
import { getUserWatchlist } from '../../../utils/user-service'

export default defineEventHandler(async (event) => {
  const announcements = await findAllAnnouncements()
  
  let watchlist: string[] = []
  const auth = getAuthFromEvent(event)
  
  if (auth) {
    watchlist = await getUserWatchlist(auth.uid)
  }
  
  const watchlistSet = new Set(watchlist)
  
  const response = announcements.map(ann => ({
    ...ann,
    is_watched: ann.kode_emiten ? watchlistSet.has(ann.kode_emiten) : false,
  }))
  
  return response
})
