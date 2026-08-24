import { defineEventHandler, getQuery } from 'h3'
import { XBRLStatementRepository } from '../../../utils/xbrl-repo'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const repo = new XBRLStatementRepository()

  return await repo.findValueScreener({
    minRoic: query.min_roic ? parseFloat(String(query.min_roic)) : undefined,
    minFScore: query.min_f_score ? parseInt(String(query.min_f_score), 10) : undefined,
    minMosPct: query.min_mos ? parseFloat(String(query.min_mos)) : undefined,
    maxDebtEquity: query.max_de ? parseFloat(String(query.max_de)) : undefined,
    sector: query.sector ? String(query.sector) : undefined,
    minTimingScore: query.min_timing_score ? parseInt(String(query.min_timing_score), 10) : undefined,
    timingStatus: query.timing_status ? String(query.timing_status) : undefined,
    peBand: query.pe_band ? String(query.pe_band) : (query.pe_band_discount ? String(query.pe_band_discount) : undefined),
    peBandDiscount: query.pe_band_discount ? String(query.pe_band_discount) : undefined,
    limit: query.limit ? parseInt(String(query.limit), 10) : 50,
    skip: query.skip ? parseInt(String(query.skip), 10) : 0
  })
})
