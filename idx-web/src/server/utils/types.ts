export interface Attachment {
  pdf_filename?: string;
  full_save_path?: string;
  original_filename?: string;
}

export interface Announcement {
  _id?: string;
  id: string;
  created_at?: string;
  updated_at?: string;
  efek_emiten_dire?: boolean;
  efek_emiten_dinfra?: boolean;
  final_id?: number;
  old_final_id?: number;
  no_pengumuman?: string;
  tgl_pengumuman?: string;
  judul_pengumuman?: string;
  jenis_pengumuman?: string;
  kode_emiten?: string;
  created_date?: string;
  form_id?: string;
  perkara_pengumuman?: string;
  jmsx_group_id?: string;
  attachments?: Attachment[];
}

export interface AnnouncementResponse extends Announcement {
  is_watched: boolean;
}

export interface News {
  _id?: string;
  id: string;
  created_at?: string;
  updated_at?: string;
  date?: string;
  title?: string;
  summary?: string;
  content?: string;
  link?: string;
  priority?: number;
  value_score?: number;
  impact_direction?: 'Bullish' | 'Bearish' | 'Neutral';
  investment_takeaway?: string;
  tickers?: string[];
  sector?: string;
  subsector?: string;
  industry?: string;
  is_industry_wide?: boolean;
}

export interface BriefingItem {
  ticker?: string;
  issuer_name?: string;
  headline: string;
  rationale: string;
  value_score: number;
  investment_takeaway: string;
}

export interface SectorHighlight {
  sector: string;
  summary: string;
  sentiment: 'Bullish' | 'Bearish' | 'Neutral';
}

export interface Briefing {
  _id?: string;
  id: string;
  date: string;
  title: string;
  macro_pulse: string;
  bullish_lookout: BriefingItem[];
  bearish_lookout: BriefingItem[];
  sector_highlights: SectorHighlight[];
  action_plan: string;
  raw_markdown?: string;
  created_at?: string;
  updated_at?: string;
}

export interface NewsResponse extends News {
  is_watched: boolean;
}

export interface User {
  _id?: string;
  id?: string;
  firebase_uid: string;
  email: string;
  watchlist: string[];
  created_at?: string;
  updated_at?: string;
}

export interface FinancialReport {
  _id?: string;
  id?: string;
  created_at?: string;
  updated_at?: string;
  issuer_code?: string;
  report_url?: string;
  year?: number;
  quarter?: number;
  downloaded_at?: number;
}

export interface FactValue {
  value: number;
  unit?: string;
  decimals?: number;
}

export type FactMap = Record<string, Record<string, FactValue>>;

export interface StatementMetadata {
  sector: string;
  subsector?: string;
  industry: string;
  subindustry?: string;
  accounting_standard: string;
  currency: string;
  rounding_level: string;
  rounding_multiplier: number;
  audit_status: string;
  auditor_opinion?: string;
  conversion_rate?: number;
}

export interface CoreFinancials {
  shares_outstanding: number;
  total_assets: number;
  cash_and_equivalents: number;
  current_assets: number;
  total_liabilities: number;
  current_liabilities: number;
  short_term_debt: number;
  long_term_debt: number;
  working_capital?: number;
  total_debt: number;
  total_equity: number;
  retained_earnings: number;
  revenue: number;
  cost_of_revenue: number;
  gross_profit: number;
  operating_income: number;
  ebitda?: number;
  finance_costs: number;
  net_income: number;
  net_income_parent: number;
  operating_cash_flow: number;
  investing_cash_flow?: number;
  financing_cash_flow?: number;
  capex: number;
  free_cash_flow: number;
  dividends_paid: number;
}

export interface ComputedRatios {
  roic: number;
  roe: number;
  roa: number;
  gross_margin_pct: number;
  operating_margin_pct: number;
  net_margin_pct: number;
  debt_to_equity: number;
  net_debt: number;
  interest_coverage_ratio: number;
  current_ratio: number;
  fcf_conversion_pct: number;
  piotroski_f_score: number;
  altman_z_score: number;
}

export interface ValuationMetrics {
  normalized_eps: number;
  normalized_bvps: number;
  revenue_per_share?: number;
  cash_per_share?: number;
  free_cash_flow_per_share?: number;
  market_cap?: number;
  enterprise_value?: number;
  graham_number: number;
  dcf_fair_value: number;
  current_price: number;
  margin_of_safety_pct: number;
  pe_ratio: number;
  pb_ratio: number;
  ps_ratio?: number;
  p_fcf_ratio?: number;
  ev_to_ebit?: number;
  ev_to_ebitda?: number;
  earnings_yield_pct?: number;
  quick_ratio?: number;
}

export interface XBRLStatement {
  _id?: string;
  id: string;
  ticker: string;
  company_name: string;
  year: number;
  period: string;
  period_type: string;
  period_end_date: string;
  statement_type: string;
  metadata: StatementMetadata;
  core: CoreFinancials;
  facts?: FactMap;
  computed_ratios: ComputedRatios;
  valuation: ValuationMetrics;
  created_at?: string;
  updated_at?: string;
}

export interface PriceCandle {
  _id?: string;
  id?: string;
  ticker: string;
  date: string | Date;
  open: number;
  high: number;
  low: number;
  close: number;
  adj_close: number;
  volume: number;
  created_at?: string | Date;
  updated_at?: string | Date;
}

export interface StockPriceResponse {
  ticker: string;
  range: string;
  count: number;
  prices: PriceCandle[];
}

