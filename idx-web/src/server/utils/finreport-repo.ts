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
  const collection = getFinancialReportsCollection()
  const query: any = {}

  if (filter.issuer_code) query.issuer_code = filter.issuer_code.toUpperCase().trim()
  if (filter.year) query.year = filter.year
  if (filter.quarter) query.quarter = filter.quarter
  if (filter.search) {
    query.issuer_code = { $regex: filter.search.trim(), $options: 'i' }
  }

  const limit = Math.min(Math.max(filter.limit || 20, 1), 100)
  const skip = Math.max(filter.skip || 0, 0)

  const [data, total] = await Promise.all([
    collection
      .find(query)
      .sort({ year: -1, downloaded_at: -1, _id: -1 })
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
