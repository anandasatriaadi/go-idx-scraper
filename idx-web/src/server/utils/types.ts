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
