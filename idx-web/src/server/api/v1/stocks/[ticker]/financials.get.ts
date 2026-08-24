import { defineEventHandler, createError } from 'h3'
import { XBRLStatementRepository } from '../../../../utils/xbrl-repo'

export default defineEventHandler(async (event) => {
  const ticker = event.context.params?.ticker
  if (!ticker) {
    throw createError({ statusCode: 400, statusMessage: 'Ticker is required' })
  }
  const repo = new XBRLStatementRepository()
  return await repo.findByTicker(ticker)
})
