import { defineEventHandler, getQuery } from 'h3'
import { findAllNewsPaginated } from '../../../utils/news-repo'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const limit = parseInt(String(query.limit || '20'), 10)
  const page = parseInt(String(query.page || '1'), 10)
  const skip = query.skip ? parseInt(String(query.skip), 10) : (page - 1) * limit

  return await findAllNewsPaginated({
    limit,
    skip,
    ticker: query.ticker ? String(query.ticker) : undefined,
    sector: query.sector ? String(query.sector) : undefined,
    subsector: query.subsector ? String(query.subsector) : undefined,
    industry: query.industry ? String(query.industry) : undefined,
    search: query.search ? String(query.search) : undefined,
    priority: query.priority ? parseInt(String(query.priority), 10) : undefined,
    date_gte: query.date_gte ? String(query.date_gte) : undefined,
    date_lte: query.date_lte ? String(query.date_lte) : undefined,
    source: query.source ? String(query.source) : undefined,
  })
})
