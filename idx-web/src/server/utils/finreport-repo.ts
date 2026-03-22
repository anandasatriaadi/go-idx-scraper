import { getDb } from '../plugins/mongodb'
import type { Collection } from 'mongodb'
import type { FinancialReport } from './types'

export function getFinancialReportsCollection(): Collection<FinancialReport> {
  const db = getDb()
  if (!db) throw new Error('Database not connected')
  return db.collection<FinancialReport>('financial_reports')
}

export async function findAllFinancialReports(): Promise<FinancialReport[]> {
  const collection = getFinancialReportsCollection()
  return collection.find({}).toArray()
}
