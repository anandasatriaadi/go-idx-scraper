import { defineEventHandler, createError, getQuery } from 'h3'
import { XBRLStatementRepository } from '../../../../utils/xbrl-repo'

export default defineEventHandler(async (event) => {
  const ticker = event.context.params?.ticker
  if (!ticker) {
    throw createError({ statusCode: 400, statusMessage: 'Ticker is required' })
  }
  const query = getQuery(event)
  const limit = query.limit ? parseInt(query.limit as string) : 24
  const repo = new XBRLStatementRepository()
  return await repo.findByTicker(ticker, limit)
})
