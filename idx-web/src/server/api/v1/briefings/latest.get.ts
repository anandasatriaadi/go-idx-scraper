import { defineEventHandler, createError } from 'h3'
import { BriefingRepository } from '../../../utils/briefing-repo'

export default defineEventHandler(async () => {
  const repo = new BriefingRepository()
  const latest = await repo.findLatest()
  if (!latest) {
    throw createError({ statusCode: 404, statusMessage: 'No briefing found' })
  }
  return latest
})
