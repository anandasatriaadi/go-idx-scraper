import { Db } from 'mongodb'
import { XBRLStatement } from './types'
import { getDb } from '../plugins/mongodb'

export class XBRLStatementRepository {
  private collection: any

  constructor(db?: Db) {
    const database = db || getDb()
    if (!database) throw new Error('Database not connected')
    this.collection = database.collection('xbrl_statements')
  }

  async findByTicker(ticker: string, limit: number = 8): Promise<XBRLStatement[]> {
    const docs = await this.collection
      .find({ ticker: ticker.toUpperCase().trim() })
      .sort({ year: -1, period: -1 })
      .limit(limit)
      .toArray()
    return docs.map((d: any) => ({ ...d, id: d._id.toString() }))
  }

  async findValueScreener(filters: {
    minRoic?: number
    minFScore?: number
    minMosPct?: number
    maxDebtEquity?: number
    sector?: string
    limit?: number
    skip?: number
  }): Promise<{ statements: XBRLStatement[]; total: number }> {
    const query: any = {}
    if (filters.minRoic !== undefined) query['computed_ratios.roic'] = { $gte: filters.minRoic }
    if (filters.minFScore !== undefined) query['computed_ratios.piotroski_f_score'] = { $gte: filters.minFScore }
    if (filters.minMosPct !== undefined) query['valuation.margin_of_safety_pct'] = { $gte: filters.minMosPct }
    if (filters.maxDebtEquity !== undefined) query['computed_ratios.debt_to_equity'] = { $lte: filters.maxDebtEquity }
    if (filters.sector) query['metadata.sector'] = filters.sector

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
