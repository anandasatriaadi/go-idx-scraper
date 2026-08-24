import { getDb } from '../plugins/mongodb'
import type { Collection } from 'mongodb'
import type { Announcement } from './types'

export function getAnnouncementsCollection(): Collection<Announcement> {
  const db = getDb()
  if (!db) throw new Error('Database not connected')
  return db.collection<Announcement>('announcements')
}

export interface AnnouncementFilter {
  ticker?: string;
  search?: string;
  limit?: number;
  skip?: number;
}

export async function findAllAnnouncementsPaginated(filter: AnnouncementFilter = {}) {
  const collection = getAnnouncementsCollection()
  const query: any = {}

  if (filter.ticker) {
    query.kode_emiten = filter.ticker.toUpperCase().trim()
  }
  if (filter.search) {
    const q = filter.search.trim()
    query.$or = [
      { judul_pengumuman: { $regex: q, $options: 'i' } },
      { no_pengumuman: { $regex: q, $options: 'i' } },
      { kode_emiten: { $regex: q, $options: 'i' } },
      { title: { $regex: q, $options: 'i' } }
    ]
  }

  const limit = Math.min(Math.max(filter.limit || 20, 1), 100)
  const skip = Math.max(filter.skip || 0, 0)

  const [data, total] = await Promise.all([
    collection
      .find(query)
      .sort({ tgl_pengumuman: -1, created_at: -1, _id: -1 })
      .skip(skip)
      .limit(limit)
      .toArray(),
    collection.countDocuments(query)
  ])

  return {
    data: data.map(d => ({ ...d, id: d._id ? d._id.toString() : d.id })),
    total,
    page: Math.floor(skip / limit) + 1,
    total_pages: Math.ceil(total / limit)
  }
}

export async function findAnnouncementById(id: string): Promise<Announcement | null> {
  const collection = getAnnouncementsCollection()
  return collection.findOne({ _id: id } as any)
}
