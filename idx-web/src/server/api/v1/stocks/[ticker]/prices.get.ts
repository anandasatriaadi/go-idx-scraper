import { defineEventHandler, createError, getQuery } from 'h3'
import { StockPriceRepository } from '../../../../utils/price-repo'
import type { StockPriceResponse } from '../../../../utils/types'

export default defineEventHandler(async (event): Promise<StockPriceResponse> => {
  const ticker = event.context.params?.ticker
  if (!ticker || typeof ticker !== 'string' || !ticker.trim()) {
    throw createError({ statusCode: 400, statusMessage: 'Ticker is required' })
  }

  const query = getQuery(event)
  const range = query.range ? String(query.range).toLowerCase().trim() : '5y'
  const limit = query.limit ? parseInt(String(query.limit), 10) : (range === 'max' ? 5000 : 1250)

  if (isNaN(limit) || limit <= 0) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid limit parameter' })
  }

  const cleanTicker = ticker.toUpperCase().trim().replace(/\.JK$/, '')
  const repo = new StockPriceRepository()
  const prices = await repo.getPricesByTicker(cleanTicker, range, limit)

  return {
    ticker: cleanTicker,
    range,
    count: prices.length,
    prices
  }
})
