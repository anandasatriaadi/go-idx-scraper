import { upsertUser, findByFirebaseUID, updateWatchlist } from './user-repo'

export async function getOrCreateUser(uid: string, email: string) {
  return upsertUser(uid, email)
}

export async function getUserWatchlist(uid: string): Promise<string[]> {
  const user = await findByFirebaseUID(uid)
  if (!user) return []
  return user.watchlist || []
}

export async function setUserWatchlist(uid: string, watchlist: string[]): Promise<void> {
  await updateWatchlist(uid, watchlist)
}
