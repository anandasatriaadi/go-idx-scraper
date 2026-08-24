import type { Db, Collection } from 'mongodb'
import type { PriceCandle } from './types'
import { getDb } from '../plugins/mongodb'

export class StockPriceRepository {
  private collection: Collection<PriceCandle>

  constructor(db?: Db) {
    const database = db || getDb()
    if (!database) throw new Error('Database not connected')
    this.collection = database.collection<PriceCandle>('stock_prices')
  }

  async getPricesByTicker(ticker: string, range?: string, limit?: number): Promise<PriceCandle[]> {
    let cleanTicker = ticker.toUpperCase().trim()
    if (cleanTicker.endsWith('.JK')) {
      cleanTicker = cleanTicker.slice(0, -3)
    }

    const query: any = { ticker: cleanTicker }

    if (range) {
      const r = range.toLowerCase().trim()
      const now = new Date()
      let startDate: Date | null = null

      if (r === '1d' || r === '1day') {
        startDate = new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000)
      } else if (r === '5d' || r === '5day') {
        startDate = new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000)
      } else if (r === '1mo' || r === '1m') {
        startDate = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
      } else if (r === '3mo' || r === '3m') {
        startDate = new Date(now.getTime() - 90 * 24 * 60 * 60 * 1000)
      } else if (r === '6mo' || r === '6m') {
        startDate = new Date(now.getTime() - 180 * 24 * 60 * 60 * 1000)
      } else if (r === '1y') {
        startDate = new Date(now.getTime() - 365 * 24 * 60 * 60 * 1000)
      } else if (r === '2y') {
        startDate = new Date(now.getTime() - 2 * 365 * 24 * 60 * 60 * 1000)
      } else if (r === '3y') {
        startDate = new Date(now.getTime() - 3 * 365 * 24 * 60 * 60 * 1000)
      } else if (r === '5y') {
        startDate = new Date(now.getTime() - 5 * 365 * 24 * 60 * 60 * 1000)
      } else if (r === '10y') {
        startDate = new Date(now.getTime() - 10 * 365 * 24 * 60 * 60 * 1000)
      } else if (r === 'ytd') {
        startDate = new Date(now.getFullYear(), 0, 1)
      }

      if (startDate) {
        query.date = { $gte: startDate }
      }
    }

    let cursor = this.collection.find(query).sort({ date: -1 })
    if (limit && limit > 0) {
      cursor = cursor.limit(limit)
    }

    const docs = await cursor.toArray()
    // Reverse array to return candles chronologically ascending (date: 1)
    return docs.reverse().map((d: any) => ({
      ...d,
      id: d._id ? d._id.toString() : d.id
    }))
  }
}
