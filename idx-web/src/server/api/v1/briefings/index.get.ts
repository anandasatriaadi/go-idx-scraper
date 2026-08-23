import { defineEventHandler, getQuery } from 'h3'
import { BriefingRepository } from '../../../utils/briefing-repo'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const limit = parseInt(String(query.limit || '20'), 10)
  const skip = parseInt(String(query.skip || '0'), 10)

  const repo = new BriefingRepository()
  return await repo.findAll(limit, skip)
})
