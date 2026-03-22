import { getDb } from '../plugins/mongodb'
import type { Collection } from 'mongodb'
import type { Announcement } from './types'

export function getAnnouncementsCollection(): Collection<Announcement> {
  const db = getDb()
  if (!db) throw new Error('Database not connected')
  return db.collection<Announcement>('announcements')
}

export async function findAllAnnouncements(filter: Record<string, any> = {}): Promise<Announcement[]> {
  const collection = getAnnouncementsCollection()
  const results = await collection.find(filter).toArray()
  return results
}

export async function findAnnouncementById(id: string): Promise<Announcement | null> {
  const collection = getAnnouncementsCollection()
  return collection.findOne({ _id: id })
}
