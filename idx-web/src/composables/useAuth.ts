import { ref } from 'vue'
import {
  signInWithPopup,
  GoogleAuthProvider,
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  signOut,
  onAuthStateChanged,
  type User
} from 'firebase/auth'

export const useAuth = () => {
  const nuxtApp = useNuxtApp()
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const loading = ref(true)
  const error = ref<string | null>(null)

  const auth = nuxtApp.$firebaseAuth as any

  if (import.meta.client && auth) {
    onAuthStateChanged(auth, async (u) => {
      user.value = u
      if (u) {
        token.value = await u.getIdToken()
      } else {
        token.value = null
      }
      loading.value = false
    })
  } else {
    loading.value = false
  }

  const loginWithGoogle = async () => {
    error.value = null
    try {
      if (!auth) throw new Error('Firebase Auth not available')
      const provider = new GoogleAuthProvider()
      const result = await signInWithPopup(auth, provider)
      user.value = result.user
      token.value = await result.user.getIdToken()
      return result.user
    } catch (e: any) {
      error.value = e.message || 'Failed to sign in with Google'
      throw e
    }
  }

  const loginWithEmail = async (email: string, pass: string) => {
    error.value = null
    try {
      if (!auth) throw new Error('Firebase Auth not available')
      const result = await signInWithEmailAndPassword(auth, email, pass)
      user.value = result.user
      token.value = await result.user.getIdToken()
      return result.user
    } catch (e: any) {
      error.value = e.message || 'Failed to sign in with email'
      throw e
    }
  }

  const signupWithEmail = async (email: string, pass: string) => {
    error.value = null
    try {
      if (!auth) throw new Error('Firebase Auth not available')
      const result = await createUserWithEmailAndPassword(auth, email, pass)
      user.value = result.user
      token.value = await result.user.getIdToken()
      return result.user
    } catch (e: any) {
      error.value = e.message || 'Failed to create account'
      throw e
    }
  }

  const logout = async () => {
    try {
      if (auth) await signOut(auth)
      user.value = null
      token.value = null
    } catch (e: any) {
      error.value = e.message || 'Failed to sign out'
    }
  }

  return {
    user,
    token,
    loading,
    error,
    loginWithGoogle,
    loginWithEmail,
    signupWithEmail,
    logout
  }
}
