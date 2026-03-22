import { getAuthFromEvent } from '../../../utils/firebase-admin'
import { setUserWatchlist } from '../../../utils/user-service'

export default defineEventHandler(async (event) => {
  const auth = getAuthFromEvent(event)
  
  if (!auth) {
    throw createError({
      statusCode: 401,
      statusMessage: 'User not authenticated',
    })
  }

  const body = await readBody<{ watchlist: string[] }>(event)
  
  if (!body || !Array.isArray(body.watchlist)) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Invalid request body',
    })
  }

  await setUserWatchlist(auth.uid, body.watchlist)
  
  return {
    watchlist: body.watchlist,
  }
})
