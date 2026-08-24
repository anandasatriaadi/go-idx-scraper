import { defineEventHandler, getQuery } from 'h3'
import { findAllAnnouncementsPaginated } from '../../../utils/announcement-repo'
import { getAuthFromEvent } from '../../../utils/firebase-admin'
import { getUserWatchlist } from '../../../utils/user-service'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const limit = parseInt(String(query.limit || '20'), 10)
  const page = parseInt(String(query.page || '1'), 10)
  const skip = query.skip ? parseInt(String(query.skip), 10) : (page - 1) * limit
  const ticker = query.ticker ? String(query.ticker) : undefined
  const search = query.search ? String(query.search) : undefined

  const result = await findAllAnnouncementsPaginated({ limit, skip, ticker, search })

  let watchlist: string[] = []
  const auth = getAuthFromEvent(event)
  if (auth) {
    watchlist = await getUserWatchlist(auth.uid)
  }
  const watchlistSet = new Set(watchlist)

  return {
    ...result,
    data: result.data.map(ann => ({
      ...ann,
      is_watched: ann.kode_emiten ? watchlistSet.has(ann.kode_emiten) : false,
    }))
  }
})
