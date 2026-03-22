import { MongoClient } from 'mongodb'

let client: MongoClient | null = null
let db: any = null

export default defineNitroPlugin(async () => {
  const config = useRuntimeConfig()
  
  try {
    client = new MongoClient(config.mongoUri)
    await client.connect()
    db = client.db(config.mongoDbName)
    console.log('[MongoDB] Connected successfully')
  } catch (err) {
    console.error('[MongoDB] Connection failed:', err)
  }
})

export function getDb() {
  return db
}

export function getClient() {
  return client
}
