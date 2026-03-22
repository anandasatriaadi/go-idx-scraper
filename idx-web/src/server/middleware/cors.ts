export default defineEventHandler((event) => {
  const origin = getHeader(event, 'origin')
  const config = useRuntimeConfig()
  
  const allowedOrigins = config.cors?.allowedOrigins || ['*']
  
  if (origin) {
    const isAllowed = allowedOrigins.includes('*') || allowedOrigins.includes(origin)
    if (isAllowed) {
      setHeader(event, 'Access-Control-Allow-Origin', origin)
    }
    setHeader(event, 'Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS')
    setHeader(event, 'Access-Control-Allow-Headers', 'Authorization, Content-Type')
  }
  
  if (event.method === 'OPTIONS') {
    setHeader(event, 'Access-Control-Max-Age', '86400')
    event.node.res.statusCode = 204
    event.node.res.end()
    return
  }
})