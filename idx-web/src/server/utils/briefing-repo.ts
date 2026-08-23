import { Db } from 'mongodb'
import { Briefing } from './types'
import { getDb } from '../plugins/mongodb'

export class BriefingRepository {
  private collection: any

  constructor(db?: Db) {
    const database = db || getDb()
    if (!database) throw new Error('Database not connected')
    this.collection = database.collection('daily_briefings')
  }

  async findLatest(): Promise<Briefing | null> {
    const doc = await this.collection.find().sort({ date: -1 }).limit(1).toArray()
    if (!doc || doc.length === 0) return null
    return { ...doc[0], id: doc[0]._id.toString() }
  }

  async findAll(limit: number = 20, skip: number = 0): Promise<{ briefings: Briefing[]; total: number }> {
    const [briefings, total] = await Promise.all([
      this.collection.find().sort({ date: -1 }).skip(skip).limit(limit).toArray(),
      this.collection.countDocuments()
    ])
    return {
      briefings: briefings.map((b: any) => ({ ...b, id: b._id.toString() })),
      total
    }
  }
}
