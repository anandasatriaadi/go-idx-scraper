export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: true },
  
  modules: [],
  
  srcDir: 'src/',
  
  runtimeConfig: {
    firebaseCredentialsPath: process.env.FIREBASE_CREDENTIALS_PATH || '',
    cors: {
      allowedOrigins: process.env.CORS_ALLOWED_ORIGINS?.split(',') || ['*'],
    },
    mongoUri: process.env.MONGODB_URI || 'mongodb://localhost:27017',
    mongoDbName: process.env.MONGODB_DB_NAME || 'idx_scraper',
    openrouterApiKey: process.env.OPENROUTER_API_KEY || '',
    public: {
      firebaseApiKey: process.env.FIREBASE_API_KEY || '',
      firebaseAuthDomain: process.env.FIREBASE_AUTH_DOMAIN || '',
      firebaseProjectId: process.env.FIREBASE_PROJECT_ID || '',
    }
  },

  nitro: {
    experimental: {
      openAPI: true
    },
    routeRules: {
      '/api/**': { cors: false }
    }
  }
})