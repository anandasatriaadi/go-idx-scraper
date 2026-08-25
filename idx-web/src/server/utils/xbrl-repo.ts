import { Db } from 'mongodb'
import { XBRLStatement } from './types'
import { getDb } from '../plugins/mongodb'

export interface ValueScreenerFilters {
  minRoic?: number
  minFScore?: number
  minMosPct?: number
  maxDebtEquity?: number
  sector?: string
  minTimingScore?: number
  timingStatus?: string
  peBand?: string
  peBandDiscount?: string
  limit?: number
  skip?: number
}

export class XBRLStatementRepository {
  private collection: any

  constructor(db?: Db) {
    const database = db || getDb()
    if (!database) throw new Error('Database not connected')
    this.collection = database.collection('xbrl_statements')
  }

  async findByTicker(ticker: string, limit: number = 24): Promise<XBRLStatement[]> {
    const docs = await this.collection
      .find({ ticker: ticker.toUpperCase().trim() })
      .sort({ year: -1, period_end_date: -1 })
      .limit(limit)
      .toArray()
    return docs.map((d: any) => ({ ...d, id: d._id.toString() }))
  }

  async findValueScreener(filters: ValueScreenerFilters): Promise<{ statements: XBRLStatement[]; total: number }> {
    const query: any = {}
    if (filters.minRoic !== undefined) query['computed_ratios.roic'] = { $gte: filters.minRoic }
    if (filters.minFScore !== undefined) query['computed_ratios.piotroski_f_score'] = { $gte: filters.minFScore }
    if (filters.minMosPct !== undefined) query['valuation.margin_of_safety_pct'] = { $gte: filters.minMosPct }
    if (filters.maxDebtEquity !== undefined) query['computed_ratios.debt_to_equity'] = { $lte: filters.maxDebtEquity }
    if (filters.sector) query['metadata.sector'] = filters.sector
    if (filters.minTimingScore !== undefined) query['timing_signal.score'] = { $gte: filters.minTimingScore }
    if (filters.timingStatus) query['timing_signal.status'] = { $regex: filters.timingStatus, $options: 'i' }

    const peBandFilter = filters.peBand || filters.peBandDiscount
    if (peBandFilter) {
      const norm = peBandFilter.toLowerCase().replace(/[^a-z0-9]/g, '')
      if (norm === 'minus2sd' || norm === '2sd' || norm === 'deepvalue') {
        query.$expr = {
          $and: [
            ...(query.$expr ? [query.$expr] : []),
            { $gt: ['$valuation_bands.minus_2sd_price_pe', 0] },
            { $lte: ['$valuation.current_price', '$valuation_bands.minus_2sd_price_pe'] }
          ]
        }
      } else if (norm === 'minus1sd' || norm === '1sd' || norm === 'discount') {
        query.$expr = {
          $and: [
            ...(query.$expr ? [query.$expr] : []),
            { $gt: ['$valuation_bands.minus_1sd_price_pe', 0] },
            { $lte: ['$valuation.current_price', '$valuation_bands.minus_1sd_price_pe'] }
          ]
        }
      } else if (norm === 'mean') {
        query.$expr = {
          $and: [
            ...(query.$expr ? [query.$expr] : []),
            { $gt: ['$valuation_bands.mean_price_pe', 0] },
            { $lte: ['$valuation.current_price', '$valuation_bands.mean_price_pe'] }
          ]
        }
      }
    }

    const limit = filters.limit || 50
    const skip = filters.skip || 0

    const [statements, total] = await Promise.all([
      this.collection.find(query).sort({ 'valuation.margin_of_safety_pct': -1 }).skip(skip).limit(limit).toArray(),
      this.collection.countDocuments(query)
    ])

    return {
      statements: statements.map((d: any) => ({ ...d, id: d._id.toString() })),
      total
    }
  }
}
