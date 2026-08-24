import { getDb } from '../plugins/mongodb'
import type { Collection } from 'mongodb'
import type { News } from './types'

export function getNewsCollection(): Collection<News> {
  const db = getDb()
  if (!db) throw new Error('Database not connected')
  return db.collection<News>('news')
}

export interface NewsFilter {
  date_gte?: string;
  date_lte?: string;
  priority?: number;
  source?: string;
  ticker?: string;
  sector?: string;
  subsector?: string;
  industry?: string;
  search?: string;
  limit?: number;
  skip?: number;
}

export async function findAllNewsPaginated(filter: NewsFilter = {}) {
  const collection = getNewsCollection()
  const query: any = {}

  if (filter.date_gte || filter.date_lte) {
    query.date = {}
    if (filter.date_gte) query.date.$gte = new Date(filter.date_gte)
    if (filter.date_lte) query.date.$lte = new Date(filter.date_lte)
  }

  if (filter.priority !== undefined) query.priority = filter.priority
  if (filter.source) query.link = filter.source
  if (filter.ticker) query.tickers = filter.ticker.toUpperCase().trim()
  if (filter.sector) query.sector = filter.sector
  if (filter.subsector) query.subsector = filter.subsector
  if (filter.industry && !filter.subsector) {
    query.$or = [{ subsector: filter.industry }, { industry: filter.industry }, { sector: filter.industry }]
  }
  if (filter.search) {
    const q = filter.search.trim()
    query.$or = [
      { title: { $regex: q, $options: 'i' } },
      { summary: { $regex: q, $options: 'i' } },
      { tickers: filter.search.toUpperCase().trim() }
    ]
  }

  const limit = Math.min(Math.max(filter.limit || 20, 1), 100)
  const skip = Math.max(filter.skip || 0, 0)

  const [data, total] = await Promise.all([
    collection
      .find(query)
      .sort({ date: -1, created_at: -1, _id: -1 })
      .skip(skip)
      .limit(limit)
      .toArray(),
    collection.countDocuments(query)
  ])

  return {
    data: data.map((d: any) => ({ ...d, id: d._id ? d._id.toString() : d.id })),
    total,
    page: Math.floor(skip / limit) + 1,
    total_pages: Math.ceil(total / limit)
  }
}

export async function findNewsById(id: string): Promise<News | null> {
  const collection = getNewsCollection()
  return collection.findOne({ _id: id } as any)
}
