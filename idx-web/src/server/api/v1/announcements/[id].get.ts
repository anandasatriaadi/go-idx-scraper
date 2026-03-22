import { findAnnouncementById } from '../../../utils/announcement-repo'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  
  if (!id) {
    throw createError({
      statusCode: 400,
      statusMessage: 'ID is required',
    })
  }
  
  const announcement = await findAnnouncementById(id)
  
  if (!announcement) {
    throw createError({
      statusCode: 404,
      statusMessage: 'Announcement not found',
    })
  }
  
  return announcement
})
