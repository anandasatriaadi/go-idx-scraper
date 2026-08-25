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
    const stmts = docs.map((d: any) => ({ ...d, id: d._id.toString() }))
    return adjustStatementsForStockSplits(stmts)
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

export function adjustStatementsForStockSplits(statements: XBRLStatement[]): XBRLStatement[] {
  if (!statements || statements.length < 2) return statements

  // Latest statement is at index 0 (sorted year desc, period_end_date desc)
  const latestShares = statements[0]?.core?.shares_outstanding || 0
  if (latestShares <= 1000) return statements

  for (let i = 1; i < statements.length; i++) {
    const s = statements[i]
    const curShares = s.core?.shares_outstanding || 0
    if (curShares <= 1) continue

    const ratio = latestShares / curShares
    if (ratio >= 1.8 || ratio <= 0.55) {
      const fxRate = (s.metadata?.currency === 'USD')
        ? (s.metadata?.conversion_rate && s.metadata.conversion_rate >= 1000 ? s.metadata.conversion_rate : (s.metadata?.conversion_rate ? s.metadata.conversion_rate * 1000 : 16000))
        : 1.0

      const netIncome = s.core?.net_income_parent || s.core?.net_income || 0
      const totalEquity = s.core?.total_equity || 0
      const revenue = s.core?.revenue || 0
      const cash = s.core?.cash_and_equivalents || 0
      const fcf = s.core?.free_cash_flow || 0

      const adjEps = (netIncome * fxRate) / latestShares
      const adjBvps = (totalEquity * fxRate) / latestShares
      const adjRevPerShare = (revenue * fxRate) / latestShares
      const adjCashPerShare = (cash * fxRate) / latestShares
      const adjFcfPerShare = (fcf * fxRate) / latestShares

      let adjGraham = 0
      if (adjEps > 0 && adjBvps > 0) {
        adjGraham = Math.sqrt(22.5 * adjEps * adjBvps)
      }

      if (s.valuation) {
        s.valuation.normalized_eps = adjEps
        s.valuation.normalized_bvps = adjBvps
        s.valuation.revenue_per_share = adjRevPerShare
        s.valuation.cash_per_share = adjCashPerShare
        s.valuation.free_cash_flow_per_share = adjFcfPerShare
        s.valuation.graham_number = adjGraham
      }
    }
  }

  return statements
}
