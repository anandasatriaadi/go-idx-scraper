import { findNewsById } from '../../../utils/news-repo'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  
  if (!id) {
    throw createError({
      statusCode: 400,
      statusMessage: 'ID is required',
    })
  }
  
  const news = await findNewsById(id)
  
  if (!news) {
    throw createError({
      statusCode: 404,
      statusMessage: 'News not found',
    })
  }
  
  return news
})
