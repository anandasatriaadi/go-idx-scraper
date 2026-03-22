import { getDb } from '../plugins/mongodb'
import type { Collection, Filter } from 'mongodb'
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
}

export async function findAllNews(filter: NewsFilter = {}): Promise<News[]> {
  const collection = getNewsCollection()
  
  const query: Filter<News> = {}
  
  if (filter.date_gte || filter.date_lte) {
    query.date = {}
    if (filter.date_gte) {
      (query.date as any).$gte = new Date(filter.date_gte)
    }
    if (filter.date_lte) {
      (query.date as any).$lte = new Date(filter.date_lte)
    }
  }
  
  if (filter.priority !== undefined) {
    query.priority = filter.priority
  }
  
  if (filter.source) {
    query.link = filter.source
  }
  
  const results = await collection.find(query).toArray()
  return results
}

export async function findNewsById(id: string): Promise<News | null> {
  const collection = getNewsCollection()
  return collection.findOne({ _id: id } as any)
}
