# Technical Architecture & Future Roadmap: Institutional-Grade Value Investing Platform for IDX

**Version:** 1.0.0  
**Author:** AI Engineering & Financial Analysis Team  
**Date:** 2026-08-24  
**Target Repository:** `go-idx-scraper` & `idx-web`  

---

## Table of Contents

1. [Executive Summary & Current System Context](#1-executive-summary--current-system-context)
2. [Domain Foundation: Value Investing on the Indonesia Stock Exchange (IDX)](#2-domain-foundation-value-investing-on-the-indonesia-stock-exchange-idx)
3. [Pillar 1: Financial Statement Parser & Automated Valuation Engine](#3-pillar-1-financial-statement-parser--automated-valuation-engine)
   - [3.1 Excel Data Ingestion Architecture](#31-excel-data-ingestion-architecture)
   - [3.2 Mathematical Formulas & Line-Item Extractions](#32-mathematical-formulas--line-item-extractions)
   - [3.3 Bankruptcy & Earnings Manipulation Forensics (Altman Z & Piotroski F)](#33-bankruptcy--earnings-manipulation-forensics-altman-z--piotroski-f)
   - [3.4 Intrinsic Valuation Models (DCF & Graham Number)](#34-intrinsic-valuation-models-dcf--graham-number)
   - [3.5 Database Schema & API Contract](#35-database-schema--api-contract)
4. [Pillar 2: On-Demand 1-Page AI Investment Dossier & Due Diligence Memo](#4-pillar-2-on-demand-1-page-ai-investment-dossier--due-diligence-memo)
   - [4.1 Multi-Modal Context Synthesis Pipeline](#41-multi-modal-context-synthesis-pipeline)
   - [4.2 Prompt Engineering & Structured Output Schemas](#42-prompt-engineering--structured-output-schemas)
   - [4.3 UI Presentation & Ticker 360° Dossier View](#43-ui-presentation--ticker-360-dossier-view)
5. [Pillar 3: Insider Tracking & Corporate Integrity Radar](#5-pillar-3-insider-tracking--corporate-integrity-radar)
   - [5.1 Insider Trade Ingestion & Signal Classification](#51-insider-trade-ingestion--signal-classification)
   - [5.2 Share Buyback Retirement vs. Dilution Radar](#52-share-buyback-retirement-vs-dilution-radar)
   - [5.3 Dividend Compounding & Sustainability Forecaster](#53-dividend-compounding--sustainability-forecaster)
6. [Pillar 4: Portfolio Health & Fundamental Moat Monitor](#6-pillar-4-portfolio-health--fundamental-moat-monitor)
   - [6.1 User Portfolio Tracking Engine](#61-user-portfolio-tracking-engine)
   - [6.2 Automated Moat Degradation & Risk Alerts](#62-automated-moat-degradation--risk-alerts)
7. [System Data Flow & Unified Schema Reference](#7-system-data-flow--unified-schema-reference)
8. [Phased Implementation Guide](#8-phased-implementation-guide)

---

## 1. Executive Summary & Current System Context

This document is an exhaustive engineering and financial specification designed to onboard any engineer, data analyst, or financial professional to the **go-idx-scraper** and **idx-web** ecosystem.

### Current System Capabilities (Production Status)
The current repository is divided into two decoupled layers:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ 1. DATA COLLECTION & INTELLIGENCE ENGINE (Go 1.24+ / DDD Hexagonal) [ACTIVE]│
│    - Downloader CLI: Downloads XBRL instance.zip and Excel financial reports│
│    - XBRL Parser CLI (`cmd/xbrl_parser`): Native Go XML streaming parser    │
│    - Valuation Engine: ROIC, Piotroski F-Score (0-9), Altman Z'', Graham No │
│    - Scraper CLI: Multi-channel Kontan scraper (investasi + keuangan)       │
│    - Daily Briefing Engine: Runs at 7 AM GMT+8 using Gemini 3.7 Flash        │
│    - Announcer CLI: Scrapes official IDX disclosure attachments             │
│    - Makefile: Unified runner (`make web`, `make briefing`, `make parse-xbrl`)
│    - Database: MongoDB persistence (`xbrl_statements`, `daily_briefings`, etc)
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
                                       ▼ MongoDB URI
┌─────────────────────────────────────────────────────────────────────────────┐
│ 2. API & TERMINAL DASHBOARD (Nuxt 4 / Nitro / Vue 3 / Firebase Auth) [ACTIVE]│
│    - Dark Bloomberg-inspired Terminal (http://localhost:3000)               │
│    - Views: Overview, Daily Briefing, Value Screener, News, Disclosures     │
│    - Modals: Ticker 360° Financials, Article Reader, Firebase Auth          │
│    - API Endpoints:                                                         │
│      • /api/v1/briefings/latest, /api/v1/briefings                          │
│      • /api/v1/news (supports ?ticker=..., ?industry=...)                   │
│      • /api/v1/stocks/:ticker/financials                                    │
│      • /api/v1/screener/value (filters: min_mos, min_roic, min_f_score, etc) │
│      • /api/v1/announcements, /api/v1/financial-reports, /api/v1/user        │
│    - Firebase Client & Admin SDK Token Verification + Watchlist Sync        │
└─────────────────────────────────────────────────────────────────────────────┘
```

### The Strategic Vision
While the current platform aggregates and summarizes raw news and disclosures, **this roadmap transforms the platform into an automated institutional-grade Value Investing Intelligence Operating System**. It moves from descriptive analysis (what happened) to normative and predictive decision support (what an asset is worth, whether management is aligned, and whether to buy).

---

## 2. Domain Foundation: Value Investing on the Indonesia Stock Exchange (IDX)

The investment philosophy underpinning this entire architecture is rooted in the fundamental teachings of **Benjamin Graham, Warren Buffett, Charlie Munger, and Aswath Damodaran**, adapted for emerging market dynamics in Indonesia.

### Core Value Investing Tenets Enforced by System Logic:
1. **A Stock is an Ownership Stake in a Real Business:** Market price fluctuations are noise; true valuation depends on the long-term cash flows generated by the company's productive assets.
2. **Economic Moat (Competitive Advantage):** A business must possess durable barriers to entry (e.g., pricing power, brand monopoly, regulatory licenses, lowest-cost distribution network) that protect high Returns on Invested Capital (ROIC).
3. **Owner Earnings & Free Cash Flow:** Accounting net income can be distorted by non-cash items, depreciation schedules, or aggressive working capital assumptions. The platform prioritizes **Owner Earnings** (Cash Flow from Operations minus Maintenance Capital Expenditures).
4. **Margin of Safety (MOS):** An asset must never be purchased at fair value; it must be purchased at a measurable discount (typically $\ge 30\%$) to conservative intrinsic value to protect against unexpected downside risks.
5. **Management Integrity & Capital Allocation:** Capital allocation is management's most critical job. The platform tracks whether retained earnings produce more than one dollar of market value and whether executives buy or dump shares with their personal funds.
6. **Emerging Market Governance & Debt Traps:** Indonesian companies often operate in cyclical commodity or interest-rate sensitive sectors. The platform rigorously tests for liquidity distress, debt covenants, and auditor qualifications.

---

## 3. Pillar 1: Financial Statement Parser & Automated Valuation Engine

### 3.1 IDX XBRL Taxonomy, Instance XBRL & Data Ingestion Architecture
The platform is engineered directly around the **Official IDX XBRL Taxonomy (2020-01-01)** (`01-idx-taxonomy-2020-01-01.zip`) and verified against production IDX filings (`instance.zip` and `inlineXBRL.zip`).

#### A. The Three Financial Data Formats on IDX

```
                                  IDX Financial Reporting
                                             │
      ┌──────────────────────────────────────┼──────────────────────────────────────┐
      ▼                                      ▼                                      ▼
1. Raw Instance XBRL                   2. Inline XBRL (iXBRL)                 3. Excel Spreadsheet
   (`instance.xbrl`)                      (`1000000.html`, `1210000.html`)       (`FinancialStatement-*.xlsx`)
   ────────────────                    ─────────────────────────              ─────────────────────────────
   • Single 3.6MB XML document         • HTML files with `<ix:nonFraction>`   • Standard Excel workbook
   • Standard `<xbrli:xbrl>` root       • Human-readable + Machine-tagged     • Direct worksheet projection
   • 3,000+ distinct contexts          • Sheet-by-sheet organization          • Uses same numerical codes
   • Direct, pure XML facts            • Full styling & footnotes             • Derived from the XBRL data
```

---

### 3.2 Native Go XBRL Parser Architecture (`instance.xbrl`)

Parsing native `instance.xbrl` directly in Go via streaming XML (`encoding/xml`) offers massive advantages over brittle spreadsheet cell extraction:
- **100% Semantic Fidelity:** No cell coordinate drifts or multilingual translation errors.
- **Microsecond Parsing Speed:** Direct SAX-style token streaming without heavy spreadsheet rendering overhead.
- **Context-Aware Multi-Year Extraction:** A single file contains `CurrentYearInstant`, `PriorEndYearInstant`, `CurrentYearDuration`, and `PriorYearDuration` contexts.

#### Go XBRL Struct & Parser Blueprint

```go
package xbrl

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"
)

type InstanceFact struct {
	XMLName    xml.Name
	ContextRef string `xml:"contextRef,attr"`
	UnitRef    string `xml:"unitRef,attr"`
	Decimals   string `xml:"decimals,attr"`
	Scale      string `xml:"scale,attr"`
	IsNil      bool   `xml:"nil,attr"`
	Value      string `xml:",chardata"`
}

type ParsedStatement struct {
	Ticker             string
	EntityName         string
	Sector             string
	PeriodEndDate      string
	Currency           string
	TotalAssets        float64
	CashAndEquivalents float64
	TotalLiabilities   float64
	TotalEquity        float64
	Revenue            float64
	GrossProfit        float64
	NetIncome          float64
	OperatingCashFlow  float64
	CapEx              float64
}
```

---

### 3.3 Complete Line-Item Mapping Catalog (Verified from Real IDX Filings)

#### 1. Document and Entity Information (`idx-dei:`)
| DEI Concept | XML Tag | Description / Example (AADI Filing) |
| :--- | :--- | :--- |
| **Entity Ticker** | `idx-dei:EntityCode` | `AADI`, `BBRI`, `TLKM` |
| **Entity Legal Name** | `idx-dei:EntityName` | `PT Adaro Andalan Indonesia Tbk` |
| **Sector** | `idx-dei:Sector` | `A. Energy` |
| **Subsector** | `idx-dei:Subsector` | `A1. Oil, Gas & Coal` |
| **Industry** | `idx-dei:Industry` | `A12. Coal` |
| **Subindustry** | `idx-dei:Subindustry` | `A121. Coal Production` |
| **Filing Period** | `idx-dei:PeriodOfFinancialStatementsSubmissions` | `Kuartal I / First Quarter`, `Tahunan / Annual` |
| **Period End Date** | `idx-dei:CurrentPeriodEndDate` | `2026-03-31` |
| **Presentation Currency** | `idx-dei:DescriptionOfPresentationCurrency` | `Dollar Amerika / USD`, `Rupiah / IDR` |
| **Level of Rounding** | `idx-dei:LevelOfRoundingUsedInFinancialStatements` | `Ribuan / In Thousand`, `Satuan Penuh` |
| **Audit Status** | `idx-dei:TypeOfReportOnFinancialStatements` | `Tidak Diaudit / Unaudit`, `Diaudit / Audited` |
| **Auditor Opinion** | `idx-dei:TypeOfAuditorsOpinion` | `Wajar Tanpa Pengecualian / Unqualified` |

---

#### 2. Core Financial Concepts (`idx-cor:`)
| Financial Concept | XML Tag (`idx-cor:`) | Context Required | Sample Metric Value (AADI Q1 2026) |
| :--- | :--- | :--- | :--- |
| **Total Assets** | `Assets` | `CurrentYearInstant` | `$5,780,540,000` |
| **Cash & Cash Equivalents** | `CashAndCashEquivalents` | `CurrentYearInstant` | `$914,431,000` |
| **Total Liabilities** | `Liabilities` | `CurrentYearInstant` | `$1,999,310,000` |
| **Total Equity** | `Equity` | `CurrentYearInstant` | `$3,781,230,000` |
| **Sales & Revenue** | `SalesAndRevenue` | `CurrentYearDuration` | `$1,044,192,000` |
| **Gross Profit** | `GrossProfit` | `CurrentYearDuration` | `$257,553,000` |
| **Net Income / Profit** | `ProfitLoss` | `CurrentYearDuration` | `$153,768,000` |
| **Operating Cash Flow** | `NetCashFlowsFromUsedInOperatingActivities` | `CurrentYearDuration` | Extracted from `1510000` / `CashFlow` |
| **Capital Expenditures** | `PaymentsForPropertyPlantEquipment` | `CurrentYearDuration` | Extracted from `1510000` / `CashFlow` |

---

### 3.4 Mathematical Formulas & Line-Item Extractions

The engine extracts historical line items across trailing twelve months (TTM) and multi-year annuals:

#### 1. Free Cash Flow (FCF) & Owner Earnings
$$\text{Free Cash Flow} = \text{Cash Flow from Operations (CFO)} - \text{Capital Expenditures (CapEx)}$$
$$\text{Owner Earnings} = \text{Net Income} + \text{Depreciation \& Amortization} - \text{Maintenance CapEx} \pm \Delta\text{Working Capital}$$

*Where $\text{CapEx}$ is extracted from `CashFlow` sheet under "Perolehan aset tetap / Payments for property, plant and equipment".*

#### 2. Return on Invested Capital (ROIC)
$$\text{ROIC} = \frac{\text{NOPAT}}{\text{Invested Capital}} = \frac{\text{Operating Income (EBIT)} \times (1 - \text{Effective Tax Rate})}{(\text{Total Equity} + \text{Total Debt}) - \text{Cash \& Short-term Investments}}$$
- **Significance:** Companies consistently generating $\text{ROIC} > 15\%$ across 5+ years possess proven economic moats.

#### 3. Net Debt & Solvency
$$\text{Net Debt} = (\text{Short-term Borrowings} + \text{Long-term Debt}) - \text{Cash \& Cash Equivalents}$$
$$\text{Interest Coverage Ratio} = \frac{\text{EBIT}}{\text{Finance Costs / Interest Expense}}$$

---

### 3.3 Bankruptcy & Earnings Manipulation Forensics

#### A. Piotroski F-Score (0 to 9 Points)
A discrete score between 0 and 9 assessing financial trend strength across three buckets:

| Category | Signal | Condition for +1 Point |
| :--- | :--- | :--- |
| **Profitability** | 1. Net Income | $\text{ROA} > 0$ |
| | 2. Cash Flow | $\text{CFO} > 0$ |
| | 3. $\Delta\text{ROA}$ | $\text{ROA}_{\text{current}} > \text{ROA}_{\text{prior}}$ |
| | 4. Quality of Earnings | $\text{CFO} > \text{Net Income}$ (Accrual check) |
| **Leverage / Liquidity** | 5. $\Delta\text{Long-term Debt}$ | $\text{LTDebt}_{\text{current}} < \text{LTDebt}_{\text{prior}}$ |
| | 6. $\Delta\text{Current Ratio}$ | $\text{CR}_{\text{current}} > \text{CR}_{\text{prior}}$ |
| | 7. Dilution Check | $\text{Shares Outstanding}_{\text{current}} \le \text{Shares Outstanding}_{\text{prior}}$ |
| **Operating Efficiency** | 8. $\Delta\text{Gross Margin}$ | $\text{Gross Margin}_{\text{current}} > \text{Gross Margin}_{\text{prior}}$ |
| | 9. $\Delta\text{Asset Turnover}$ | $\text{Asset Turnover}_{\text{current}} > \text{Asset Turnover}_{\text{prior}}$ |

- **Interpretation:** $8–9$: Exceptionally strong compounder; $0–3$: Weak financial state.

#### B. Altman Z-Score for Emerging Markets ($Z''$-Score)
Adapted for emerging markets and non-manufacturing/service issuers to predict bankruptcy risk:
$$Z'' = 6.56X_1 + 3.26X_2 + 6.72X_3 + 1.05X_4$$
- $X_1 = \frac{\text{Working Capital}}{\text{Total Assets}}$
- $X_2 = \frac{\text{Retained Earnings}}{\text{Total Assets}}$
- $X_3 = \frac{\text{EBIT}}{\text{Total Assets}}$
- $X_4 = \frac{\text{Book Value of Equity}}{\text{Total Liabilities}}$
- **Thresholds:** $Z'' > 2.60$ (Safe Zone), $1.10 \le Z'' \le 2.60$ (Grey Zone), $Z'' < 1.10$ (Distress Zone / High Default Risk).

---

### 3.4 Intrinsic Valuation Models

#### A. Multi-Stage Discounted Cash Flow (DCF) Model
$$\text{Intrinsic Value} = \sum_{t=1}^{N} \frac{\text{FCF}_0 \times (1 + g_1)^t}{(1 + r)^t} + \frac{\text{FCF}_N \times (1 + g_2)}{(r - g_2) \times (1 + r)^N}$$
- $g_1$: Stage 1 conservative growth rate (e.g. 5–10% capped at sector growth).
- $g_2$: Terminal growth rate (capped at Indonesia GDP long-term trend, e.g. 4.0%).
- $r$: Discount rate / Hurdle rate (e.g. Risk-free Sukuk Rate $6.85\% + \text{Equity Risk Premium } 5.5\% = 12.35\%$).

#### B. Benjamin Graham Fair Value Formula
$$\text{Graham Number} = \sqrt{22.5 \times \text{EPS} \times \text{Book Value Per Share (BVPS)}}$$
$$\text{Margin of Safety (MOS \%)} = \frac{\text{Intrinsic Value} - \text{Current Market Price}}{\text{Intrinsic Value}} \times 100\%$$

---

### 3.5 Database Schema & API Contract

#### MongoDB Collection: `financial_metrics`
```json
{
  "_id": "6a9000000000000000000001",
  "issuer_code": "BBRI",
  "year": 2025,
  "period": "III",
  "statement_date": "2025-09-30T00:00:00.000Z",
  "metrics": {
    "revenue": 198500000000000,
    "net_income": 45300000000000,
    "operating_cash_flow": 52100000000000,
    "capex": 8400000000000,
    "free_cash_flow": 43700000000000,
    "roic": 0.185,
    "roe": 0.198,
    "debt_to_equity": 0.65,
    "current_ratio": 1.42,
    "piotroski_f_score": 8,
    "altman_z_score": 3.12
  },
  "valuation": {
    "eps": 365,
    "bvps": 2150,
    "graham_number": 4200,
    "dcf_fair_value": 5400,
    "current_price": 3800,
    "margin_of_safety_pct": 29.6
  },
  "created_at": "2026-08-24T00:00:00.000Z",
  "updated_at": "2026-08-24T00:00:00.000Z"
}
```

#### API Endpoints:
- `GET /api/v1/stocks/:ticker/financials`: Returns multi-year historical line items and ratios.
- `GET /api/v1/screener/value`: Filters stocks by `min_roic`, `min_f_score`, `min_mos_pct`, and `max_debt_equity`.

---

## 4. Pillar 2: On-Demand 1-Page AI Investment Dossier & Due Diligence Memo

### 4.1 Multi-Modal Context Synthesis Pipeline
When a user clicks any stock ticker on the terminal, the system aggregates:
1. **Parsed Financial Metrics (Pillar 1):** 5-year ROIC trend, FCF conversion, F-Score, Debt/Equity.
2. **Latest Scraped News (Kontan):** Value scores and headlines from the last 90 days.
3. **Official Disclosures (IDX):** Shareholder structure, insider filings, dividend schedules.

```
[Financial Metrics] + [Kontan News] + [IDX Disclosures]
                           │
                           ▼
             [OpenRouter Gemini 3.7 Flash]
       (Role: Senior Value Investment Partner)
                           │
                           ▼
          [Structured 1-Page Investment Memo]
```

---

### 4.2 Prompt Engineering & Structured Output Schemas

```typescript
export interface InvestmentDossier {
  ticker: string;
  company_name: string;
  verdict: 'Strong Buy' | 'Accumulate' | 'Hold' | 'Avoid / Divest';
  moat_rating: 'Wide Moat' | 'Narrow Moat' | 'No Moat';
  moat_analysis: string; // 2-3 sentences evaluating pricing power and switching costs
  capital_allocation_grade: 'A' | 'B' | 'C' | 'D' | 'F';
  capital_allocation_rationale: string;
  bull_case_thesis: string[]; // 3 core drivers of upside intrinsic value
  bear_case_risks: string[];  // 3 fatal risks that could destroy the thesis
  intrinsic_value_estimate: number;
  conservative_buy_target: number; // Fair value minus 30% Margin of Safety
  overvalued_sell_threshold: number;
  one_paragraph_summary: string;
}
```

#### Gemini 3.7 Flash System Prompt Template:
```
You are the Chief Investment Officer of a disciplined Graham-Buffett-Munger Value Fund.
Evaluate ticker {{TICKER}} with absolute objectivity, analytical rigor, and zero marketing hype.

Input Data Provided:
- 5-Year Financial Profile: {{FINANCIAL_DATA_JSON}}
- Recent Material News & Sentiment: {{NEWS_SUMMARY_LIST}}
- Corporate Actions & Insider Activity: {{CORPORATE_ACTIONS_JSON}}

Execute your evaluation following strict principles:
1. Moat: Does the business possess structural pricing power or low-cost advantages, or is it a commodity price-taker?
2. Capital Allocation: Does management reinvest capital at high ROIC, or do they destroy value with empire-building acquisitions?
3. Margin of Safety: Calculate an intrinsic value estimate and deduct a 30% margin of safety to derive the Conservative Buy Target.
```

---

## 5. Pillar 3: Insider Tracking & Corporate Integrity Radar

### 5.1 Insider Trade Ingestion & Signal Classification
In Indonesia, direct market purchases by Directors, Commissioners, or $\ge 5\%$ Shareholders must be disclosed to IDX under regulation `POJK 11/POJK.04/2017`.

The Announcer CLI scrapes disclosure attachments with classification `"Keterbukaan Informasi Pemegang Saham Tertentu"`.

#### Data Extracted:
- Insider Name & Title (e.g. *Direktur Utama / Komisaris Utama*).
- Transaction Type: Open Market Purchase (`Beli`) vs. Disposal (`Jual`).
- Number of Shares & Average Execution Price.
- Percentage Ownership Change.

#### Core Value Signal Engine:
- **High-Conviction Buy Signal:** Open-market purchases by C-level executives using personal capital (insiders only buy for one reason: they believe the stock is undervalued).
- **Red Flag Sell Signal:** Systematic liquidation by controlling shareholders or sudden resignations prior to earnings releases.

---

### 5.2 Share Buyback Retirement vs. Dilution Radar
The platform differentiates between shareholder-friendly capital allocators and predatory diluters:

```
                      Share Count Delta (YoY)
                                │
        ┌───────────────────────┴───────────────────────┐
        ▼                                               ▼
  Negative $\Delta$ Shares                        Positive $\Delta$ Shares
  (Net Share Reduction)                           (Shareholder Dilution)
        │                                               │
  ✔ Real Share Cancellation                       ❌ Rights Issues / Warrants
  ✔ Increases EPS & BVPS                          ❌ EPS & Dividend Dilution
  (Example: DSSA share retirement)                 (Example: Highly dilutive penny stocks)
```

---

### 5.3 Dividend Compounding & Sustainability Forecaster

#### Metric: Free Cash Flow Dividend Coverage
$$\text{FCF Dividend Payout Ratio} = \frac{\text{Total Cash Dividends Paid}}{\text{Free Cash Flow}}$$
- **Safety Rule:** Ratios $< 60\%$ are highly sustainable compounders. Ratios $> 100\%$ indicate dividends are funded by debt or asset sales.

---

## 6. Pillar 4: Portfolio Health & Fundamental Moat Monitor

### 6.1 User Portfolio Tracking Engine
Users log their investment holdings through the authenticated terminal UI:

```json
{
  "_id": "user_portfolio_001",
  "firebase_uid": "user_12345",
  "holdings": [
    {
      "ticker": "BBRI",
      "shares": 50000,
      "average_buy_price": 4150,
      "total_invested": 207500000
    },
    {
      "ticker": "DSSA",
      "shares": 1000,
      "average_buy_price": 38000,
      "total_invested": 38000000
    }
  ]
}
```

### 6.2 Automated Moat Degradation & Risk Alerts
The background engine constantly correlates portfolio holdings against new incoming news, disclosures, and financial reports:
1. **Auditor Warning Alert:** Instant flag if an auditor issues a qualified opinion (`Wajar Dengan Pengecualian`) or disclaimer of opinion.
2. **Moat Deterioration Alert:** Fired if gross margins contract for 3 consecutive quarters or ROIC falls below the cost of capital.
3. **Debt Distress Alert:** Fired if Altman $Z''$-Score drops below $1.10$ or interest coverage ratio falls below $2.0\times$.

---

## 7. System Data Flow & Unified Schema Reference

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          UNIFIED DATA LAKE (MongoDB)                        │
├───────────────────────┬─────────────────────────────┬───────────────────────┤
│ `news`                │ `financial_metrics`         │ `daily_briefings`     │
│ - Tickers, Industry   │ - ROIC, FCF, F-Score        │ - Macro pulse         │
│ - ValueScore, Takeaway│ - DCF Fair Value, Graham No │ - Buy / Risk Lookouts │
├───────────────────────┼─────────────────────────────┼───────────────────────┤
│ `announcements`       │ `insider_trades`            │ `investment_dossiers` │
│ - Disclosures, PDFs   │ - Executive Buys / Sells    │ - 1-Page AI Memos     │
└───────────────────────┴─────────────────────────────┴───────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                  NUXT 4 / NITRO API & DARK TERMINAL FRONTEND                │
│                                                                             │
│  [Top Navbar: Overview | Briefings | News | Screener | Portfolio | Auth]    │
│                                                                             │
│  ┌─────────────────────────┬──────────────────────────────────────────────┐ │
│  │ Financial Screener View │ Stock 360° Due Diligence Modal               │ │
│  │ - Min ROIC > 15%        │ - 1-Page Investment Thesis Memo              │ │
│  │ - Piotroski F >= 7      │ - Valuation Gauge (Price vs. Graham / DCF)   │ │
│  │ - Margin of Safety > 30%│ - Insider Transactions History                │ │
│  └─────────────────────────┴──────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 8. Phased Implementation Guide & Execution Status

| Phase | Milestone | Core Deliverables | Target Architecture | Status |
| :--- | :--- | :--- | :--- | :--- |
| **Phase 1** | **Native Go XBRL Streaming Parser & Valuation Engine (Pillar 1)** | • Native Go XML streaming parser (`internal/infra/xbrl/parser.go`).<br>• 700+ raw facts map persistence (`FactMap`).<br>• Metric calculations (ROIC, FCF, Piotroski F-Score 0–9, Altman Z''-Score).<br>• Benjamin Graham Number & Margin of Safety % calculation.<br>• MongoDB collection `xbrl_statements` & `cmd/xbrl_parser` CLI tool. | Go CLI + MongoDB | **✅ COMPLETED & PRODUCTION READY** |
| **Phase 2** | **Value Screener & Ticker 360° Financials UI** | • Nuxt API endpoints `/api/v1/stocks/:ticker/financials` and `/api/v1/screener/value`.<br>• Interactive **Value Screener** tab with strategy presets (*Deep Value*, *Buffett Moat*, *High F-Score*).<br>• **Ticker 360° Financials Modal** with valuation gauge and multi-year 3-statement trends. | Nuxt 4 Frontend + Nitro API | **✅ COMPLETED & PRODUCTION READY** |
| **Phase 3** | **On-Demand 1-Page AI Dossier & Due Diligence Memo (Pillar 2)** | • OpenRouter Gemini 3.7 Flash structured memo synthesis.<br>• Integration into Ticker 360° Drawer / Modal.<br>• Moat rating, Capital Allocation grade, 3 Bull drivers, 3 Bear risks, and Buy target price. | Gemini 3.7 + Vue 3 Modal | **⏳ PLANNED (NEXT MILESTONE)** |
| **Phase 4** | **Insider Trades & Corporate Integrity (Pillar 3)** | • Scraper for Director/Commissioner stock filings (`POJK 11/2017`).<br>• Buyback tracking vs. dilution alert engine ($-\Delta\text{Shares}$ vs. $+\Delta\text{Shares}$).<br>• Dividend safety & FCF payout coverage forecaster. | Go Scraper + MongoDB | **⏳ PLANNED** |
| **Phase 5** | **Portfolio Health & Moat Monitor (Pillar 4)** | • Firebase-authenticated user portfolio tracking (`user_portfolios`).<br>• Automated alerts for auditor red flags, margin contractions, & debt distress.<br>• Portfolio-weighted ROIC and cash flow yield summary. | Full Stack Nuxt 4 + Firebase | **⏳ PLANNED** |

---

*This document serves as the definitive reference specification for all subsequent implementation plans and sprint tasks on this repository.*
