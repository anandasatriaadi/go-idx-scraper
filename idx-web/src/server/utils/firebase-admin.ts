import type { H3Event } from 'h3'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import firebaseAdmin from 'firebase-admin'

const { admin } = firebaseAdmin

let app: admin.app.App | null = null

function getFirebaseApp(): admin.app.App {
  if (app) return app
  
  const config = useRuntimeConfig()
  const credentialsPath = config.firebaseCredentialsPath
  
  if (!credentialsPath) {
    throw new Error('FIREBASE_CREDENTIALS_PATH is not configured')
  }
  
  try {
    const resolvedPath = resolve(credentialsPath)
    const credentials = JSON.parse(readFileSync(resolvedPath, 'utf-8'))
    app = admin.initializeApp({
      credential: admin.credential.cert(credentials),
    })
    return app
  } catch (err) {
    throw new Error(`Failed to initialize Firebase app: ${err}`)
  }
}

export interface FirebaseUser {
  uid: string
  email?: string
}

export async function verifyFirebaseToken(token: string): Promise<FirebaseUser> {
  const firebaseApp = getFirebaseApp()
  const decodedToken = await firebaseApp.auth().verifyIdToken(token)
  
  return {
    uid: decodedToken.uid,
    email: decodedToken.email as string | undefined,
  }
}

export function getAuthFromEvent(event: H3Event): FirebaseUser | null {
  return event.context.auth as FirebaseUser | null
}

export function requireAuth(event: H3Event): FirebaseUser {
  const auth = getAuthFromEvent(event)
  if (!auth) {
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized - Authentication required',
    })
  }
  return auth
}