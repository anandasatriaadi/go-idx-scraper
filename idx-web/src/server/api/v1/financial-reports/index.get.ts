import { defineEventHandler, getQuery } from 'h3'
import { findAllFinancialReportsPaginated } from '../../../utils/finreport-repo'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const limit = parseInt(String(query.limit || '20'), 10)
  const page = parseInt(String(query.page || '1'), 10)
  const skip = query.skip ? parseInt(String(query.skip), 10) : (page - 1) * limit

  return await findAllFinancialReportsPaginated({
    limit,
    skip,
    issuer_code: query.issuer_code ? String(query.issuer_code) : (query.ticker ? String(query.ticker) : undefined),
    year: query.year ? parseInt(String(query.year), 10) : undefined,
    quarter: query.quarter ? parseInt(String(query.quarter), 10) : undefined,
    search: query.search ? String(query.search) : undefined,
  })
})
