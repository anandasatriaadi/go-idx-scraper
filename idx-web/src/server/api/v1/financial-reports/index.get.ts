import { findAllFinancialReports } from '../../../utils/finreport-repo'

export default defineEventHandler(async () => {
  const reports = await findAllFinancialReports()
  return reports
})
