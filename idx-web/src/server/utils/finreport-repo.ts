import { getDb } from '../plugins/mongodb'
import type { Collection } from 'mongodb'
import type { FinancialReport } from './types'

export function getFinancialReportsCollection(): Collection<FinancialReport> {
  const db = getDb()
  if (!db) throw new Error('Database not connected')
  return db.collection<FinancialReport>('financial_reports')
}

export interface FinReportFilter {
  issuer_code?: string;
  year?: number;
  quarter?: number;
  search?: string;
  limit?: number;
  skip?: number;
}

export async function findAllFinancialReportsPaginated(filter: FinReportFilter = {}) {
  const db = getDb()
  if (!db) throw new Error('Database not connected')
  const collection = getFinancialReportsCollection()
  const xbrlCollection = db.collection('xbrl_statements')

  const query: any = {}
  if (filter.issuer_code) query.issuer_code = filter.issuer_code.toUpperCase().trim()
  if (filter.year) query.year = filter.year
  if (filter.quarter) query.quarter = filter.quarter
  if (filter.search) {
    query.$or = [
      { issuer_code: { $regex: filter.search.trim(), $options: 'i' } },
      { ticker: { $regex: filter.search.trim(), $options: 'i' } },
      { company_name: { $regex: filter.search.trim(), $options: 'i' } }
    ]
  }

  const limit = Math.min(Math.max(filter.limit || 20, 1), 100)
  const skip = Math.max(filter.skip || 0, 0)

  // Check primary collection first
  let [data, total] = await Promise.all([
    collection
      .find(query)
      .sort({ year: -1, downloaded_at: -1, _id: -1 })
      .skip(skip)
      .limit(limit)
      .toArray(),
    collection.countDocuments(query)
  ])

  // If financial_reports collection is empty, seamlessly query xbrl_statements
  if (total === 0) {
    const xbrlQuery: any = {}
    if (filter.issuer_code) xbrlQuery.ticker = filter.issuer_code.toUpperCase().trim()
    if (filter.year) xbrlQuery.year = filter.year
    if (filter.search) {
      xbrlQuery.$or = [
        { ticker: { $regex: filter.search.trim(), $options: 'i' } },
        { company_name: { $regex: filter.search.trim(), $options: 'i' } }
      ]
    }

    const [xbrlDocs, xbrlTotal] = await Promise.all([
      xbrlCollection
        .find(xbrlQuery)
        .sort({ year: -1, period: -1, _id: -1 })
        .skip(skip)
        .limit(limit)
        .toArray(),
      xbrlCollection.countDocuments(xbrlQuery)
    ])

    data = xbrlDocs.map((d: any) => ({
      _id: d._id,
      id: d._id.toString(),
      issuer_code: d.ticker,
      year: d.year,
      period_string: d.period,
      quarter: d.period === 'Q1' ? 1 : d.period === 'Q2' ? 2 : d.period === 'Q3' ? 3 : 4,
      report_url: d.metadata?.source_file ? `/saham/${d.metadata.source_file}` : undefined,
      downloaded_at: d.created_at ? new Date(d.created_at).getTime() : Date.now(),
      created_at: d.created_at,
      updated_at: d.updated_at
    }))
    total = xbrlTotal
  }

  return {
    data: data.map((d: any) => ({ ...d, id: d._id ? d._id.toString() : d.id })),
    total,
    page: Math.floor(skip / limit) + 1,
    total_pages: Math.ceil(total / limit)
  }
}
