import { getAuthFromEvent } from '../../../utils/firebase-admin'
import { getUserWatchlist } from '../../../utils/user-service'

export default defineEventHandler(async (event) => {
  const auth = getAuthFromEvent(event)
  
  if (!auth) {
    throw createError({
      statusCode: 401,
      statusMessage: 'User not authenticated',
    })
  }

  const watchlist = await getUserWatchlist(auth.uid)
  
  return {
    watchlist,
  }
})
