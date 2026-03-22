export default defineEventHandler(async (event) => {
  const authHeader = getHeader(event, 'authorization')
  
  if (!authHeader) {
    event.context.auth = null
    return
  }
  
  const bearerPrefix = 'Bearer '
  if (!authHeader.startsWith(bearerPrefix)) {
    event.context.auth = null
    return
  }
  
  const token = authHeader.slice(bearerPrefix.length)
  
  try {
    const auth = await verifyFirebaseToken(token)
    event.context.auth = auth
  } catch (err) {
    event.context.auth = null
  }
})