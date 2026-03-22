import { getDb } from '../plugins/mongodb'
import type { Collection } from 'mongodb'
import type { User } from './types'

export function getUsersCollection(): Collection<User> {
  const db = getDb()
  if (!db) throw new Error('Database not connected')
  return db.collection<User>('users')
}

export async function findByFirebaseUID(uid: string): Promise<User | null> {
  const collection = getUsersCollection()
  return collection.findOne({ firebase_uid: uid })
}

export async function updateWatchlist(uid: string, watchlist: string[]): Promise<void> {
  const collection = getUsersCollection()
  await collection.updateOne(
    { firebase_uid: uid },
    { 
      $set: { 
        watchlist, 
        updated_at: new Date().toISOString() 
      } 
    }
  )
}

export async function upsertUser(uid: string, email: string): Promise<User> {
  const collection = getUsersCollection()
  const now = new Date().toISOString()
  
  const existing = await findByFirebaseUID(uid)
  if (existing) return existing
  
  const newUser: User = {
    firebase_uid: uid,
    email,
    watchlist: [],
    created_at: now,
    updated_at: now,
  }
  
  await collection.insertOne(newUser as any)
  return newUser
}
