<template>
  <div v-if="ticker" class="modal-backdrop" @click.self="$emit('close')">
    <div class="modal-card">
      <!-- Modal Header -->
      <div class="modal-header">
        <div class="header-left">
          <div class="title-row">
            <span class="ticker-badge font-mono">${{ ticker }}</span>
            <h1 class="company-name">{{ latestStatement?.company_name || ticker }}</h1>
          </div>
          <div class="meta-tags font-mono">
            <span v-if="latestStatement?.metadata?.sector" class="tag sector-tag">{{ latestStatement.metadata.sector }}</span>
            <span v-if="latestStatement?.metadata?.industry" class="tag">{{ latestStatement.metadata.industry }}</span>
            <span v-if="latestStatement?.metadata?.currency" class="tag">Currency: {{ latestStatement.metadata.currency }}</span>
            <span v-if="latestStatement?.metadata?.conversion_rate" class="tag">Rate: Rp {{ latestStatement.metadata.conversion_rate }}</span>
            <span v-if="latestStatement?.year" class="tag">Latest: {{ latestStatement.year }} {{ latestStatement.period }}</span>
          </div>
        </div>
        <button class="btn-close" @click="$emit('close')">✕</button>
      </div>

      <!-- Top Navigation Tabs -->
      <div class="modal-tab-bar font-mono">
        <button
          :class="['modal-tab-btn', { active: activeModalTab === 'terminal' }]"
          @click="activeModalTab = 'terminal'"
        >
          📊 Institutional Matrix
        </button>
        <button
          :class="['modal-tab-btn', { active: activeModalTab === 'moat' }]"
          @click="activeModalTab = 'moat'"
        >
          🛡️ 5-Year Moat & Quality Radar
          <span v-if="fiveYearAvgRoic !== null" :class="['tab-badge', fiveYearAvgRoic >= 15 ? 'badge-green' : 'badge-amber']">
            ROIC: {{ fiveYearAvgRoic.toFixed(1) }}%
          </span>
        </button>
        <button
          :class="['modal-tab-btn', { active: activeModalTab === 'chart' }]"
          @click="activeModalTab = 'chart'"
        >
          📈 Price & Graham Valuation Chart
        </button>
        <button
          :class="['modal-tab-btn', { active: activeModalTab === 'news' }]"
          @click="activeModalTab = 'news'"
        >
          📰 Related Sector News
          <span v-if="latestStatement?.metadata?.sector" class="tab-badge">
            {{ latestStatement.metadata.sector }}
          </span>
        </button>
      </div>

      <!-- Modal Body Scroll -->
      <div class="modal-body">
        <div v-if="loading" class="loading-state font-mono">
          <div class="spinner"></div>
          <span>Loading financial intelligence for ${{ ticker }}...</span>
        </div>
        <div v-else-if="statements.length === 0" class="empty-state font-mono">
          No XBRL financial statements found for ticker ${{ ticker }}.
        </div>

        <!-- 1. INSTITUTIONAL 3-COLUMN MATRIX (Stockbit Style) -->
        <div v-else-if="activeModalTab === 'terminal'" class="matrix-layout">
          <!-- LEFT COLUMN: Valuation & Solvency Multiples -->
          <div class="matrix-col">
            <!-- Smart Timing & VSA Signals Card -->
            <div v-if="latestTimingSignal" class="metric-card timing-card">
              <div class="timing-card-header">
                <h3 class="card-heading timing-heading">⚡ Smart Timing & VSA</h3>
                <span :class="['timing-score-pill font-mono', getTimingScoreClass(latestTimingSignal.score)]">
                  {{ latestTimingSignal.score }}/100
                </span>
              </div>
              <div :class="['timing-status-banner font-mono', getTimingBannerClass(latestTimingSignal.score)]">
                <span class="dot-icon">{{ latestTimingSignal.score >= 70 ? '🟢' : latestTimingSignal.score >= 50 ? '🟡' : '⚪' }}</span>
                <span class="status-txt">{{ latestTimingSignal.status || 'Timing Watch' }}</span>
              </div>
              <div class="data-rows font-mono">
                <div class="data-row">
                  <span class="label">RSI (14-Day)</span>
                  <span class="val">
                    {{ latestTimingSignal.rsi ? latestTimingSignal.rsi.toFixed(1) : '-' }}
                    <span v-if="latestTimingSignal.rsi_bullish_divergence" class="badge-bullish-div">BULLISH DIV ⚡</span>
                  </span>
                </div>
                <div class="data-row">
                  <span class="label">VSA Stopping Volume</span>
                  <span :class="['val', latestTimingSignal.stopping_volume ? 'text-green font-bold' : '']">
                    {{ latestTimingSignal.stopping_volume ? 'DETECTED 🛡️' : 'None' }}
                  </span>
                </div>
                <div class="data-row">
                  <span class="label">Volume Dry-Up (VDU)</span>
                  <span :class="['val', latestTimingSignal.volume_dry_up ? 'text-amber font-bold' : '']">
                    {{ latestTimingSignal.vdu ? latestTimingSignal.vdu.toFixed(2) + 'x' : '-' }}
                    <span v-if="latestTimingSignal.volume_dry_up" class="badge-vdu">DRY-UP 💧</span>
                  </span>
                </div>
                <div class="data-row">
                  <span class="label">Relative Volume (RVOL)</span>
                  <span class="val">{{ latestTimingSignal.rvol ? latestTimingSignal.rvol.toFixed(2) + 'x' : '-' }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Close Location Value (CLV)</span>
                  <span class="val">{{ latestTimingSignal.clv !== undefined ? (latestTimingSignal.clv > 0 ? '+' : '') + latestTimingSignal.clv.toFixed(2) : '-' }}</span>
                </div>
                <div v-if="latestTimingSignal.valuation_discount_zone || latestValuationBands" class="data-row">
                  <span class="label">P/E Discount Zone</span>
                  <span class="val text-cyan">{{ latestTimingSignal.valuation_discount_zone || getPeZoneDesc }}</span>
                </div>
              </div>

              <!-- Catalyst Signals Pills -->
              <div v-if="latestTimingSignal.signals && latestTimingSignal.signals.length > 0" class="catalyst-chips-row font-mono">
                <span v-for="(sig, sIdx) in latestTimingSignal.signals" :key="sIdx" class="catalyst-pill">
                  ✓ {{ sig }}
                </span>
              </div>
            </div>

            <!-- Current Valuation Card -->
            <div class="metric-card">
              <h3 class="card-heading">Current Valuation</h3>
              <div class="data-rows">
                <div class="data-row">
                  <span class="label">Current PE Ratio (TTM)</span>
                  <span class="val font-mono">{{ formatMultiple(latestStatement?.valuation?.pe_ratio) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Price to Book Value</span>
                  <span class="val font-mono">{{ formatMultiple(latestStatement?.valuation?.pb_ratio) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Price to Sales (TTM)</span>
                  <span class="val font-mono">{{ formatMultiple(latestStatement?.valuation?.ps_ratio) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Price to Free Cashflow (TTM)</span>
                  <span class="val font-mono">{{ formatMultiple(latestStatement?.valuation?.p_fcf_ratio) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Earnings Yield (TTM)</span>
                  <span class="val font-mono text-green">{{ formatPct(latestStatement?.valuation?.earnings_yield_pct) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">EV to EBIT (TTM)</span>
                  <span class="val font-mono">{{ formatMultiple(latestStatement?.valuation?.ev_to_ebit) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">EV to EBITDA (TTM)</span>
                  <span class="val font-mono">{{ formatMultiple(latestStatement?.valuation?.ev_to_ebitda) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Graham Number</span>
                  <span class="val font-mono text-green">{{ formatIDRPrice(latestStatement?.valuation?.graham_number) }}</span>
                </div>
                <div class="data-row highlight-row">
                  <span class="label">Margin of Safety</span>
                  <span :class="['val font-mono', (latestStatement?.valuation?.margin_of_safety_pct || 0) > 0 ? 'text-green' : 'text-red']">
                    {{ formatSignedPct(latestStatement?.valuation?.margin_of_safety_pct) }}
                  </span>
                </div>
                <div v-if="latestValuationBands?.mean_price_pe" class="data-row">
                  <span class="label">P/E Mean (Price)</span>
                  <span class="val font-mono text-cyan">{{ formatIDRPrice(latestValuationBands.mean_price_pe) }} ({{ latestValuationBands.mean_pe.toFixed(1) }}x)</span>
                </div>
                <div v-if="latestValuationBands?.minus_1sd_price_pe" class="data-row">
                  <span class="label">P/E -1σ Entry</span>
                  <span class="val font-mono text-green">{{ formatIDRPrice(latestValuationBands.minus_1sd_price_pe) }}</span>
                </div>
                <div v-if="latestValuationBands?.minus_2sd_price_pe" class="data-row">
                  <span class="label">P/E -2σ Deep Value</span>
                  <span class="val font-mono text-green font-bold">{{ formatIDRPrice(latestValuationBands.minus_2sd_price_pe) }}</span>
                </div>
              </div>
            </div>

            <!-- Per Share Metrics Card -->
            <div class="metric-card">
              <h3 class="card-heading">Per Share</h3>
              <div class="data-rows">
                <div class="data-row">
                  <span class="label">Current EPS (IDR)</span>
                  <span class="val font-mono">{{ formatIDRPrice(latestStatement?.valuation?.normalized_eps) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Book Value Per Share</span>
                  <span class="val font-mono">{{ formatIDRPrice(latestStatement?.valuation?.normalized_bvps) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Revenue Per Share</span>
                  <span class="val font-mono">{{ formatIDRPrice(latestStatement?.valuation?.revenue_per_share) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Cash Per Share</span>
                  <span class="val font-mono">{{ formatIDRPrice(latestStatement?.valuation?.cash_per_share) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Free Cashflow Per Share</span>
                  <span class="val font-mono">{{ formatIDRPrice(latestStatement?.valuation?.free_cash_flow_per_share) }}</span>
                </div>
              </div>
            </div>

            <!-- Solvency & Health Card -->
            <div class="metric-card">
              <h3 class="card-heading">Solvency & Health</h3>
              <div class="data-rows">
                <div class="data-row">
                  <span class="label">Piotroski F-Score</span>
                  <span class="val font-mono text-cyan">{{ latestStatement?.computed_ratios?.piotroski_f_score || 0 }}/9</span>
                </div>
                <div class="data-row">
                  <span class="label">Altman Z''-Score</span>
                  <span :class="['val font-mono', (latestStatement?.computed_ratios?.altman_z_score || 0) > 2.6 ? 'text-green' : 'text-amber']">
                    {{ (latestStatement?.computed_ratios?.altman_z_score || 0).toFixed(2) }}
                  </span>
                </div>
                <div class="data-row">
                  <span class="label">Current Ratio</span>
                  <span class="val font-mono">{{ (latestStatement?.computed_ratios?.current_ratio || 0).toFixed(2) }}x</span>
                </div>
                <div class="data-row">
                  <span class="label">Debt to Equity</span>
                  <span class="val font-mono">{{ (latestStatement?.computed_ratios?.debt_to_equity || 0).toFixed(2) }}x</span>
                </div>
                <div class="data-row">
                  <span class="label">Interest Coverage</span>
                  <span class="val font-mono">{{ (latestStatement?.computed_ratios?.interest_coverage_ratio || 0).toFixed(2) }}x</span>
                </div>
              </div>
            </div>
          </div>

          <!-- MIDDLE COLUMN: Multi-Year Historical Performance Matrix -->
          <div class="matrix-col col-span-2">
            <!-- Multi-Year Historical Matrix -->
            <div class="metric-card matrix-table-card">
              <div class="table-header-bar">
                <div class="matrix-tabs font-mono">
                  <button
                    :class="['m-tab', { active: matrixMetric === 'net_income' }]"
                    @click="matrixMetric = 'net_income'"
                  >
                    Net Income
                  </button>
                  <button
                    :class="['m-tab', { active: matrixMetric === 'revenue' }]"
                    @click="matrixMetric = 'revenue'"
                  >
                    Revenue
                  </button>
                  <button
                    :class="['m-tab', { active: matrixMetric === 'fcf' }]"
                    @click="matrixMetric = 'fcf'"
                  >
                    Free Cash Flow
                  </button>
                  <button
                    :class="['m-tab', { active: matrixMetric === 'roic' }]"
                    @click="matrixMetric = 'roic'"
                  >
                    ROIC (%)
                  </button>
                  <button
                    :class="['m-tab', { active: matrixMetric === 'roe' }]"
                    @click="matrixMetric = 'roe'"
                  >
                    ROE (%)
                  </button>
                  <button
                    :class="['m-tab', { active: matrixMetric === 'op_margin' }]"
                    @click="matrixMetric = 'op_margin'"
                  >
                    Op Margin (%)
                  </button>
                  <button
                    :class="['m-tab', { active: matrixMetric === 'gross_margin' }]"
                    @click="matrixMetric = 'gross_margin'"
                  >
                    Gross Margin (%)
                  </button>
                  <button
                    :class="['m-tab', { active: matrixMetric === 'invested_capital' }]"
                    @click="matrixMetric = 'invested_capital'"
                  >
                    Invested Cap
                  </button>
                </div>
                <span class="matrix-currency font-mono">Unit: {{ latestStatement?.metadata?.currency || 'IDR' }}</span>
              </div>

              <!-- Historical Table Grid -->
              <div class="table-responsive">
                <table class="matrix-table font-mono">
                  <thead>
                    <tr>
                      <th>Period</th>
                      <th v-for="y in uniqueYears" :key="y">{{ y }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="p in ['Q1', 'Q2', 'Q3', 'FY']" :key="p">
                      <td class="period-cell">{{ p }}</td>
                      <td v-for="y in uniqueYears" :key="y + p">
                        {{ getMatrixValue(y, p, matrixMetric) }}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <!-- Key Company Stats Bar -->
              <div class="key-stats-grid font-mono">
                <div class="stat-item">
                  <span class="stat-label">Market Cap</span>
                  <span class="stat-val">{{ formatCompact(latestStatement?.valuation?.market_cap) }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">Enterprise Value</span>
                  <span class="stat-val">{{ formatCompact(latestStatement?.valuation?.enterprise_value) }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">Shares Outstanding</span>
                  <span class="stat-val">{{ formatCompact(latestStatement?.core?.shares_outstanding) }}</span>
                </div>
              </div>
            </div>

            <!-- Profitability & Growth Cards -->
            <div class="dual-cards-grid">
              <div class="metric-card">
                <h3 class="card-heading">Profitability</h3>
                <div class="data-rows">
                  <div class="data-row">
                    <span class="label">Return on Invested Capital (ROIC)</span>
                    <span class="val font-mono text-green">{{ formatPct((latestStatement?.computed_ratios?.roic || 0) * 100) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Return on Equity (ROE)</span>
                    <span class="val font-mono text-green">{{ formatPct((latestStatement?.computed_ratios?.roe || 0) * 100) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Gross Profit Margin</span>
                    <span class="val font-mono">{{ formatPct(latestStatement?.computed_ratios?.gross_margin_pct) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Operating Profit Margin</span>
                    <span class="val font-mono">{{ formatPct(latestStatement?.computed_ratios?.operating_margin_pct) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Net Profit Margin</span>
                    <span class="val font-mono">{{ formatPct(latestStatement?.computed_ratios?.net_margin_pct) }}</span>
                  </div>
                </div>
              </div>

              <div class="metric-card">
                <h3 class="card-heading">Cash Flow & Capital Allocation</h3>
                <div class="data-rows">
                  <div class="data-row">
                    <span class="label">Cash from Operations (CFO)</span>
                    <span class="val font-mono text-green">{{ formatCompact(latestStatement?.core?.operating_cash_flow) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Capital Expenditure (CapEx)</span>
                    <span class="val font-mono text-amber">{{ formatCompact(latestStatement?.core?.capex) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Free Cash Flow (FCF)</span>
                    <span class="val font-mono text-green">{{ formatCompact(latestStatement?.core?.free_cash_flow) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Cash from Investing (CFI)</span>
                    <span class="val font-mono">{{ formatCompact(latestStatement?.core?.investing_cash_flow) }}</span>
                  </div>
                  <div class="data-row">
                    <span class="label">Cash from Financing (CFF)</span>
                    <span class="val font-mono">{{ formatCompact(latestStatement?.core?.financing_cash_flow) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- RIGHT COLUMN: 3-Statement Snapshot & Balance Sheet -->
          <div class="matrix-col">
            <!-- Balance Sheet Card -->
            <div class="metric-card">
              <h3 class="card-heading">Balance Sheet</h3>
              <div class="data-rows">
                <div class="data-row">
                  <span class="label">Cash & Equivalents</span>
                  <span class="val font-mono text-cyan">{{ formatCompact(latestStatement?.core?.cash_and_equivalents) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Current Assets</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.current_assets) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Total Assets</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.total_assets) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Current Liabilities</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.current_liabilities) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Total Liabilities</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.total_liabilities) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Working Capital</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.working_capital) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Total Debt</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.total_debt) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Net Debt (Cash)</span>
                  <span :class="['val font-mono', (latestStatement?.computed_ratios?.net_debt || 0) <= 0 ? 'text-green' : 'text-red']">
                    {{ (latestStatement?.computed_ratios?.net_debt || 0) <= 0 ? 'Net Cash ' + formatCompact(Math.abs(latestStatement?.computed_ratios?.net_debt || 0)) : formatCompact(latestStatement?.computed_ratios?.net_debt) }}
                  </span>
                </div>
                <div class="data-row">
                  <span class="label">Total Equity</span>
                  <span class="val font-mono text-cyan">{{ formatCompact(latestStatement?.core?.total_equity) }}</span>
                </div>
              </div>
            </div>

            <!-- Income Statement Card -->
            <div class="metric-card">
              <h3 class="card-heading">Income Statement</h3>
              <div class="data-rows">
                <div class="data-row">
                  <span class="label">Revenue / Sales</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.revenue) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Cost of Revenue</span>
                  <span class="val font-mono">{{ formatCompact(latestStatement?.core?.cost_of_revenue) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Gross Profit</span>
                  <span class="val font-mono text-green">{{ formatCompact(latestStatement?.core?.gross_profit) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Operating Profit (EBIT)</span>
                  <span class="val font-mono text-green">{{ formatCompact(latestStatement?.core?.operating_income) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Finance Costs</span>
                  <span class="val font-mono text-amber">{{ formatCompact(latestStatement?.core?.finance_costs) }}</span>
                </div>
                <div class="data-row">
                  <span class="label">Net Income</span>
                  <span class="val font-mono text-green">{{ formatCompact(latestStatement?.core?.net_income) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 2. 5-YEAR MOAT & QUALITY RADAR VIEW TAB -->
        <div v-else-if="activeModalTab === 'moat'" class="moat-radar-tab font-mono">
          <!-- Top Executive Summary KPI Cards -->
          <div class="moat-kpi-grid">
            <!-- 1. 5-Year Average ROIC -->
            <div class="moat-kpi-card">
              <div class="kpi-header">
                <span class="kpi-title">5-Year Avg ROIC</span>
                <span :class="['kpi-pill', (fiveYearAvgRoic || 0) >= 15 ? 'pill-green' : (fiveYearAvgRoic || 0) >= 10 ? 'pill-amber' : 'pill-red']">
                  {{ (fiveYearAvgRoic || 0) >= 15 ? 'Moat Proven' : 'Sub-Hurdle' }}
                </span>
              </div>
              <div class="kpi-big-num" :class="(fiveYearAvgRoic || 0) >= 15 ? 'text-green' : (fiveYearAvgRoic || 0) >= 10 ? 'text-amber' : 'text-red'">
                {{ fiveYearAvgRoic !== null ? fiveYearAvgRoic.toFixed(1) + '%' : '-' }}
              </div>
              <div class="kpi-subtext">
                <span>Indonesian WACC: <strong>11.5%</strong></span>
                <span class="spread-tag" :class="(fiveYearAvgRoic || 0) >= 11.5 ? 'text-green' : 'text-red'">
                  {{ fiveYearAvgRoic !== null ? ((fiveYearAvgRoic - 11.5 >= 0 ? '+' : '') + (fiveYearAvgRoic - 11.5).toFixed(1) + '% EVA') : '-' }}
                </span>
              </div>
            </div>

            <!-- 2. Moat Verdict -->
            <div class="moat-kpi-card">
              <div class="kpi-header">
                <span class="kpi-title">Economic Moat Rating</span>
                <span class="kpi-pill pill-cyan">5Y Durability</span>
              </div>
              <div class="kpi-big-label" :class="moatVerdict.class">
                {{ moatVerdict.badge }}
              </div>
              <div class="kpi-desc">
                {{ moatVerdict.desc }}
              </div>
            </div>

            <!-- 3. ROE vs ROIC Spread (Leverage Check) -->
            <div class="moat-kpi-card">
              <div class="kpi-header">
                <span class="kpi-title">ROE vs ROIC Spread</span>
                <span :class="['kpi-pill', Math.abs(fiveYearAvgSpread || 0) <= 10 ? 'pill-green' : 'pill-amber']">
                  {{ Math.abs(fiveYearAvgSpread || 0) <= 10 ? 'Organic' : 'Leverage' }}
                </span>
              </div>
              <div class="kpi-big-num text-cyan">
                {{ fiveYearAvgSpread !== null ? (fiveYearAvgSpread > 0 ? '+' : '') + fiveYearAvgSpread.toFixed(1) + '%' : '-' }}
              </div>
              <div class="kpi-subtext">
                <span>5Y Avg ROE: <strong>{{ fiveYearAvgRoe !== null ? fiveYearAvgRoe.toFixed(1) + '%' : '-' }}</strong></span>
                <span class="text-secondary">{{ leverageVerdict.title }}</span>
              </div>
            </div>

            <!-- 4. FCF Conversion & Cash Quality -->
            <div class="moat-kpi-card">
              <div class="kpi-header">
                <span class="kpi-title">5Y FCF Conversion</span>
                <span :class="['kpi-pill', (fiveYearAvgFcfConversion || 0) >= 80 ? 'pill-green' : (fiveYearAvgFcfConversion || 0) >= 50 ? 'pill-amber' : 'pill-red']">
                  Owner Earnings
                </span>
              </div>
              <div class="kpi-big-num" :class="(fiveYearAvgFcfConversion || 0) >= 80 ? 'text-green' : (fiveYearAvgFcfConversion || 0) >= 50 ? 'text-amber' : 'text-red'">
                {{ fiveYearAvgFcfConversion !== null ? fiveYearAvgFcfConversion.toFixed(1) + '%' : '-' }}
              </div>
              <div class="kpi-desc">
                {{ cashFlowVerdict.title }}
              </div>
            </div>
          </div>

          <!-- 2x2 Grid of 5-Year Visual Charts -->
          <div class="moat-charts-grid">
            <!-- Chart 1: 5-Year ROIC vs 15% Moat Hurdle Line -->
            <div class="moat-chart-card">
              <div class="chart-card-header">
                <div>
                  <h4 class="chart-card-title">1. 5-Year ROIC Trend vs. 15% Economic Moat Hurdle</h4>
                  <p class="chart-card-sub">Bars indicate annual ROIC (%). Dashed lines represent 15% Moat Hurdle & 20% Elite Compounder thresholds.</p>
                </div>
                <div class="chart-legend-row">
                  <span class="legend-dot dot-green"></span> ROIC ≥ 15%
                  <span class="legend-dot dot-amber"></span> 10% - 15%
                  <span class="legend-dot dot-red"></span> &lt; 10%
                </div>
              </div>

              <!-- SVG ROIC Bar Chart -->
              <div class="svg-chart-container">
                <svg viewBox="0 0 520 200" class="moat-svg">
                  <!-- Threshold lines -->
                  <!-- 20% Elite Line -->
                  <line x1="45" y1="50" x2="500" y2="50" stroke="#10b981" stroke-dasharray="4,4" stroke-width="1.2" opacity="0.6" />
                  <text x="502" y="53" fill="#10b981" font-size="9" text-anchor="start">20% Elite</text>

                  <!-- 15% Hurdle Line -->
                  <line x1="45" y1="80" x2="500" y2="80" stroke="#38bdf8" stroke-dasharray="4,4" stroke-width="1.2" opacity="0.8" />
                  <text x="502" y="83" fill="#38bdf8" font-size="9" text-anchor="start">15% WACC</text>

                  <!-- 0% Baseline -->
                  <line x1="45" y1="160" x2="500" y2="160" stroke="#334155" stroke-width="1" />
                  <text x="35" y="163" fill="#64748b" font-size="9" text-anchor="end">0%</text>
                  <text x="35" y="83" fill="#64748b" font-size="9" text-anchor="end">15%</text>
                  <text x="35" y="53" fill="#64748b" font-size="9" text-anchor="end">20%</text>

                  <!-- Bars for each year -->
                  <g v-for="(item, idx) in fiveYearAnnuals" :key="item.year">
                    <!-- Bar positioning: X = 70 + idx * 90 -->
                    <rect
                      :x="70 + idx * 85"
                      :y="getRoicBarY(item.roic)"
                      width="42"
                      :height="getRoicBarHeight(item.roic)"
                      :fill="getRoicColor(item.roic)"
                      rx="3"
                      class="svg-bar"
                    />
                    <!-- Value Label Top of Bar -->
                    <text
                      :x="70 + idx * 85 + 21"
                      :y="getRoicBarY(item.roic) - 6"
                      :fill="getRoicColor(item.roic)"
                      font-size="11"
                      font-weight="700"
                      text-anchor="middle"
                    >
                      {{ item.roic.toFixed(1) }}%
                    </text>
                    <!-- Year Label Bottom -->
                    <text
                      :x="70 + idx * 85 + 21"
                      y="180"
                      fill="#94a3b8"
                      font-size="11"
                      font-weight="600"
                      text-anchor="middle"
                    >
                      {{ item.year }}
                    </text>
                  </g>
                </svg>
              </div>
            </div>

            <!-- Chart 2: ROE vs ROIC Spread & Leverage Multiplier -->
            <div class="moat-chart-card">
              <div class="chart-card-header">
                <div>
                  <h4 class="chart-card-title">2. ROE vs. ROIC Spread (Financial Leverage Check)</h4>
                  <p class="chart-card-sub">Compares Return on Equity against ROIC. A large gap indicates returns boosted by balance sheet debt.</p>
                </div>
                <div class="chart-legend-row">
                  <span class="legend-dot dot-cyan"></span> ROE
                  <span class="legend-dot dot-green"></span> ROIC
                </div>
              </div>

              <!-- SVG ROE vs ROIC Grouped Bar Chart -->
              <div class="svg-chart-container">
                <svg viewBox="0 0 520 200" class="moat-svg">
                  <!-- 0% Baseline -->
                  <line x1="45" y1="160" x2="500" y2="160" stroke="#334155" stroke-width="1" />
                  <text x="35" y="163" fill="#64748b" font-size="9" text-anchor="end">0%</text>

                  <!-- 25% Grid line -->
                  <line x1="45" y1="80" x2="500" y2="80" stroke="rgba(255,255,255,0.05)" stroke-width="1" />
                  <text x="35" y="83" fill="#64748b" font-size="9" text-anchor="end">25%</text>

                  <!-- Grouped Bars for each year -->
                  <g v-for="(item, idx) in fiveYearAnnuals" :key="'roe-' + item.year">
                    <!-- ROE Bar (Cyan) -->
                    <rect
                      :x="65 + idx * 85"
                      :y="getRateBarY(item.roe)"
                      width="22"
                      :height="getRateBarHeight(item.roe)"
                      fill="#38bdf8"
                      rx="2"
                      class="svg-bar"
                    />
                    <!-- ROIC Bar (Green) -->
                    <rect
                      :x="90 + idx * 85"
                      :y="getRateBarY(item.roic)"
                      width="22"
                      :height="getRateBarHeight(item.roic)"
                      fill="#10b981"
                      rx="2"
                      class="svg-bar"
                    />
                    <!-- Spread Label -->
                    <text
                      :x="65 + idx * 85 + 23"
                      :y="Math.min(getRateBarY(item.roe), getRateBarY(item.roic)) - 6"
                      fill="#fde68a"
                      font-size="9"
                      font-weight="600"
                      text-anchor="middle"
                    >
                      Δ{{ (item.spread >= 0 ? '+' : '') + item.spread.toFixed(0) }}%
                    </text>
                    <!-- Year Label Bottom -->
                    <text
                      :x="65 + idx * 85 + 23"
                      y="180"
                      fill="#94a3b8"
                      font-size="11"
                      font-weight="600"
                      text-anchor="middle"
                    >
                      {{ item.year }}
                    </text>
                  </g>
                </svg>
              </div>
            </div>

            <!-- Chart 3: Owner Earnings: Net Income vs CFO vs Free Cash Flow -->
            <div class="moat-chart-card">
              <div class="chart-card-header">
                <div>
                  <h4 class="chart-card-title">3. Owner Earnings: Net Income vs. CFO vs. Free Cash Flow</h4>
                  <p class="chart-card-sub">Verifies whether accounting Net Income translates into real Free Cash Flow (accruals check).</p>
                </div>
                <div class="chart-legend-row">
                  <span class="legend-dot dot-blue"></span> Net Income
                  <span class="legend-dot dot-cyan"></span> CFO
                  <span class="legend-dot dot-emerald"></span> FCF
                </div>
              </div>

              <!-- SVG Cash Flow Clustered Bars -->
              <div class="svg-chart-container">
                <svg viewBox="0 0 520 200" class="moat-svg">
                  <!-- Baseline -->
                  <line x1="45" y1="160" x2="500" y2="160" stroke="#334155" stroke-width="1" />
                  <text x="35" y="163" fill="#64748b" font-size="9" text-anchor="end">0</text>

                  <g v-for="(item, idx) in fiveYearAnnuals" :key="'fcf-' + item.year">
                    <!-- Net Income Bar -->
                    <rect
                      :x="60 + idx * 85"
                      :y="getCashBarY(item.netIncome)"
                      width="16"
                      :height="getCashBarHeight(item.netIncome)"
                      fill="#3b82f6"
                      rx="2"
                      class="svg-bar"
                    />
                    <!-- CFO Bar -->
                    <rect
                      :x="78 + idx * 85"
                      :y="getCashBarY(item.cfo)"
                      width="16"
                      :height="getCashBarHeight(item.cfo)"
                      fill="#06b6d4"
                      rx="2"
                      class="svg-bar"
                    />
                    <!-- FCF Bar -->
                    <rect
                      :x="96 + idx * 85"
                      :y="getCashBarY(item.fcf)"
                      width="16"
                      :height="getCashBarHeight(item.fcf)"
                      fill="#10b981"
                      rx="2"
                      class="svg-bar"
                    />
                    <!-- FCF % Label -->
                    <text
                      :x="60 + idx * 85 + 26"
                      :y="Math.min(getCashBarY(item.netIncome), getCashBarY(item.cfo), getCashBarY(item.fcf)) - 6"
                      fill="#34d399"
                      font-size="9"
                      font-weight="700"
                      text-anchor="middle"
                    >
                      FCF {{ item.fcfConversion.toFixed(0) }}%
                    </text>
                    <!-- Year Label Bottom -->
                    <text
                      :x="60 + idx * 85 + 26"
                      y="180"
                      fill="#94a3b8"
                      font-size="11"
                      font-weight="600"
                      text-anchor="middle"
                    >
                      {{ item.year }}
                    </text>
                  </g>
                </svg>
              </div>
            </div>

            <!-- Chart 4: Pricing Power: Gross Margin & Operating Margin Trajectory -->
            <div class="moat-chart-card">
              <div class="chart-card-header">
                <div>
                  <h4 class="chart-card-title">4. Pricing Power: Gross & Operating Margins</h4>
                  <p class="chart-card-sub">Durable moats maintain or expand gross margins during inflationary and commodity price swings.</p>
                </div>
                <div class="chart-legend-row">
                  <span class="legend-dot dot-emerald"></span> Gross Margin %
                  <span class="legend-dot dot-amber"></span> Operating Margin %
                </div>
              </div>

              <!-- SVG Dual Margin Line / Bar Chart -->
              <div class="svg-chart-container">
                <svg viewBox="0 0 520 200" class="moat-svg">
                  <!-- 0% Baseline -->
                  <line x1="45" y1="160" x2="500" y2="160" stroke="#334155" stroke-width="1" />
                  <text x="35" y="163" fill="#64748b" font-size="9" text-anchor="end">0%</text>

                  <!-- 50% Line -->
                  <line x1="45" y1="80" x2="500" y2="80" stroke="rgba(255,255,255,0.05)" stroke-width="1" />
                  <text x="35" y="83" fill="#64748b" font-size="9" text-anchor="end">50%</text>

                  <g v-for="(item, idx) in fiveYearAnnuals" :key="'margin-' + item.year">
                    <!-- Gross Margin Bar -->
                    <rect
                      :x="65 + idx * 85"
                      :y="getMarginBarY(item.grossMargin)"
                      width="22"
                      :height="getMarginBarHeight(item.grossMargin)"
                      fill="#10b981"
                      rx="2"
                      class="svg-bar"
                    />
                    <!-- Operating Margin Bar -->
                    <rect
                      :x="90 + idx * 85"
                      :y="getMarginBarY(item.operatingMargin)"
                      width="22"
                      :height="getMarginBarHeight(item.operatingMargin)"
                      fill="#f59e0b"
                      rx="2"
                      class="svg-bar"
                    />
                    <!-- Gross % Label Top -->
                    <text
                      :x="65 + idx * 85 + 11"
                      :y="getMarginBarY(item.grossMargin) - 5"
                      fill="#34d399"
                      font-size="9"
                      font-weight="700"
                      text-anchor="middle"
                    >
                      {{ item.grossMargin.toFixed(0) }}%
                    </text>
                    <!-- Op % Label Top -->
                    <text
                      :x="90 + idx * 85 + 11"
                      :y="getMarginBarY(item.operatingMargin) - 5"
                      fill="#fbbf24"
                      font-size="9"
                      font-weight="700"
                      text-anchor="middle"
                    >
                      {{ item.operatingMargin.toFixed(0) }}%
                    </text>
                    <!-- Year Label Bottom -->
                    <text
                      :x="65 + idx * 85 + 23"
                      y="180"
                      fill="#94a3b8"
                      font-size="11"
                      font-weight="600"
                      text-anchor="middle"
                    >
                      {{ item.year }}
                    </text>
                  </g>
                </svg>
              </div>
            </div>
          </div>

          <!-- Comprehensive 5-Year Quality & Moat Table -->
          <div class="metric-card moat-table-card">
            <div class="moat-table-header">
              <h4 class="card-heading">📊 5-Year Multi-Year Fundamental Health & Moat Ledger</h4>
              <span class="text-secondary text-xs">Currency: {{ latestStatement?.metadata?.currency || 'IDR' }} (Audited Annuals)</span>
            </div>
            <div class="table-responsive">
              <table class="matrix-table font-mono">
                <thead>
                  <tr>
                    <th>Year (Period)</th>
                    <th>ROIC (%)</th>
                    <th>ROE (%)</th>
                    <th>ROE-ROIC Spread</th>
                    <th>Gross Margin</th>
                    <th>Op Margin</th>
                    <th>FCF Conversion</th>
                    <th>Invested Capital</th>
                    <th>CapEx / CFO</th>
                    <th>Debt/Equity</th>
                    <th>Piotroski</th>
                    <th>Altman Z''</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in fiveYearAnnuals" :key="'row-' + item.year">
                    <td class="period-cell">{{ item.year }} ({{ item.period }})</td>
                    <td :class="item.roic >= 15 ? 'text-green font-bold' : item.roic >= 10 ? 'text-amber' : 'text-red'">
                      {{ item.roic.toFixed(1) }}%
                    </td>
                    <td class="text-cyan">{{ item.roe.toFixed(1) }}%</td>
                    <td :class="Math.abs(item.spread) <= 10 ? 'text-secondary' : 'text-amber'">
                      {{ (item.spread >= 0 ? '+' : '') + item.spread.toFixed(1) }}%
                    </td>
                    <td>{{ item.grossMargin.toFixed(1) }}%</td>
                    <td>{{ item.operatingMargin.toFixed(1) }}%</td>
                    <td :class="item.fcfConversion >= 80 ? 'text-green' : item.fcfConversion >= 50 ? 'text-amber' : 'text-red'">
                      {{ item.fcfConversion.toFixed(1) }}%
                    </td>
                    <td>{{ formatCompact(item.investedCapital) }}</td>
                    <td>{{ item.capexToCfo.toFixed(1) }}%</td>
                    <td>{{ item.debtToEquity.toFixed(2) }}x</td>
                    <td class="text-cyan">{{ item.piotroski }}/9</td>
                    <td :class="item.altmanZ >= 2.6 ? 'text-green' : item.altmanZ >= 1.1 ? 'text-amber' : 'text-red'">
                      {{ item.altmanZ.toFixed(2) }}
                    </td>
                  </tr>
                  <!-- 5-Year Average Footer Row -->
                  <tr class="avg-footer-row font-bold">
                    <td class="text-cyan">5-Year Avg</td>
                    <td :class="(fiveYearAvgRoic || 0) >= 15 ? 'text-green' : 'text-amber'">
                      {{ fiveYearAvgRoic !== null ? fiveYearAvgRoic.toFixed(1) + '%' : '-' }}
                    </td>
                    <td class="text-cyan">{{ fiveYearAvgRoe !== null ? fiveYearAvgRoe.toFixed(1) + '%' : '-' }}</td>
                    <td class="text-secondary">{{ fiveYearAvgSpread !== null ? (fiveYearAvgSpread >= 0 ? '+' : '') + fiveYearAvgSpread.toFixed(1) + '%' : '-' }}</td>
                    <td>{{ fiveYearAvgGrossMargin !== null ? fiveYearAvgGrossMargin.toFixed(1) + '%' : '-' }}</td>
                    <td>{{ fiveYearAvgOpMargin !== null ? fiveYearAvgOpMargin.toFixed(1) + '%' : '-' }}</td>
                    <td :class="(fiveYearAvgFcfConversion || 0) >= 80 ? 'text-green' : 'text-amber'">
                      {{ fiveYearAvgFcfConversion !== null ? fiveYearAvgFcfConversion.toFixed(1) + '%' : '-' }}
                    </td>
                    <td class="text-secondary">-</td>
                    <td class="text-secondary">-</td>
                    <td class="text-secondary">-</td>
                    <td class="text-cyan">{{ fiveYearAvgPiotroski !== null ? fiveYearAvgPiotroski.toFixed(1) + '/9' : '-' }}</td>
                    <td :class="(fiveYearAvgAltmanZ || 0) >= 2.6 ? 'text-green' : 'text-amber'">
                      {{ fiveYearAvgAltmanZ !== null ? fiveYearAvgAltmanZ.toFixed(2) : '-' }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- 6-Pillar Fundamental Quality Checklist -->
          <div class="metric-card quality-checklist-card">
            <h4 class="card-heading">🛡️ 6-Pillar Forensic Moat & Quality Audit</h4>
            <div class="checklist-grid">
              <!-- Pillar 1 -->
              <div class="check-item">
                <div class="check-header">
                  <span class="check-icon">{{ (fiveYearAvgRoic || 0) >= 15 ? '✅' : '❌' }}</span>
                  <span class="check-title">1. Sustainable ROIC (≥ 15%)</span>
                </div>
                <p class="check-desc">
                  5Y average ROIC of <strong>{{ fiveYearAvgRoic !== null ? fiveYearAvgRoic.toFixed(1) + '%' : '-' }}</strong> beats the 11.5% Indonesian WACC cost of capital, generating genuine positive economic value.
                </p>
              </div>

              <!-- Pillar 2 -->
              <div class="check-item">
                <div class="check-header">
                  <span class="check-icon">{{ Math.abs(fiveYearAvgSpread || 0) <= 12 ? '✅' : '⚠️' }}</span>
                  <span class="check-title">2. Organic Returns (ROE vs ROIC)</span>
                </div>
                <p class="check-desc">
                  Spread of <strong>{{ fiveYearAvgSpread !== null ? fiveYearAvgSpread.toFixed(1) + '%' : '-' }}</strong> indicates returns are generated by operational strength rather than risky debt leverage.
                </p>
              </div>

              <!-- Pillar 3 -->
              <div class="check-item">
                <div class="check-header">
                  <span class="check-icon">{{ (fiveYearAvgFcfConversion || 0) >= 70 ? '✅' : '⚠️' }}</span>
                  <span class="check-title">3. Owner Earnings Backing (FCF ≥ 70%)</span>
                </div>
                <p class="check-desc">
                  Free Cash Flow conversion averages <strong>{{ fiveYearAvgFcfConversion !== null ? fiveYearAvgFcfConversion.toFixed(1) + '%' : '-' }}</strong>, confirming that profits are collected in cash rather than trapped in working capital.
                </p>
              </div>

              <!-- Pillar 4 -->
              <div class="check-item">
                <div class="check-header">
                  <span class="check-icon">{{ pricingPowerVerdict.class.includes('green') ? '✅' : '⚠️' }}</span>
                  <span class="check-title">4. Pricing Power & Gross Margins</span>
                </div>
                <p class="check-desc">
                  Gross margins average <strong>{{ fiveYearAvgGrossMargin !== null ? fiveYearAvgGrossMargin.toFixed(1) + '%' : '-' }}</strong> over 5 years. {{ pricingPowerVerdict.desc }}
                </p>
              </div>

              <!-- Pillar 5 -->
              <div class="check-item">
                <div class="check-header">
                  <span class="check-icon">{{ (fiveYearRevenueCagr || 0) >= 5 ? '✅' : '⚪' }}</span>
                  <span class="check-title">5. 5-Year Compounding Runway</span>
                </div>
                <p class="check-desc">
                  5-year Revenue CAGR: <strong>{{ fiveYearRevenueCagr !== null ? (fiveYearRevenueCagr >= 0 ? '+' : '') + fiveYearRevenueCagr.toFixed(1) + '%' : '-' }}</strong> | Net Income CAGR: <strong>{{ fiveYearNetIncomeCagr !== null ? (fiveYearNetIncomeCagr >= 0 ? '+' : '') + fiveYearNetIncomeCagr.toFixed(1) + '%' : '-' }}</strong>.
                </p>
              </div>

              <!-- Pillar 6 -->
              <div class="check-item">
                <div class="check-header">
                  <span class="check-icon">{{ (fiveYearAvgAltmanZ || 0) >= 2.6 ? '✅' : '⚠️' }}</span>
                  <span class="check-title">6. Balance Sheet Fortification</span>
                </div>
                <p class="check-desc">
                  Average Altman Z''-Score of <strong>{{ fiveYearAvgAltmanZ !== null ? fiveYearAvgAltmanZ.toFixed(2) : '-' }}</strong> places the firm in the {{ (fiveYearAvgAltmanZ || 0) >= 2.6 ? 'Safe Zone' : 'Grey/Caution Zone' }} with average Piotroski score of <strong>{{ fiveYearAvgPiotroski !== null ? fiveYearAvgPiotroski.toFixed(1) + '/9' : '-' }}</strong>.
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- 3. CHART VIEW TAB -->
        <div v-else-if="activeModalTab === 'chart'" class="chart-tab-content">
          <PriceValuationChart
            :ticker="ticker"
            :graham-number="latestStatement?.valuation?.graham_number || 0"
            :current-price="latestStatement?.valuation?.current_price || 0"
            :valuation-bands="latestValuationBands"
            :timing-signal="latestTimingSignal"
          />
        </div>

        <!-- 4. RELATED SECTOR NEWS TAB -->
        <div v-else-if="activeModalTab === 'news'" class="sector-news-tab">
          <div class="news-tab-header">
            <div>
              <h3 class="news-tab-title font-mono">Related Industry & Sector Intelligence</h3>
              <p class="news-tab-sub">Stream of news developments affecting {{ latestStatement?.metadata?.sector || 'the sector' }}</p>
            </div>
            <button class="btn-refresh font-mono" @click="fetchSectorNews">🔄 Refresh Stream</button>
          </div>

          <div v-if="loadingNews" class="loading-state font-mono">Loading related news...</div>
          <div v-else-if="sectorNews.length === 0" class="empty-state font-mono">No news articles found for this sector.</div>
          <div v-else class="news-cards-list">
            <article v-for="item in sectorNews" :key="item.id || item._id" class="sector-news-card">
              <div class="card-meta font-mono">
                <span :class="['sentiment-tag', item.impact_direction?.toLowerCase()]">{{ item.impact_direction || 'Neutral' }}</span>
                <span class="score-tag font-mono">Score: {{ (item.value_score && item.value_score > 0 ? '+' : '') + (item.value_score || 0) }}</span>
                <span class="date">{{ formatDate(item.date || item.created_at) }}</span>
              </div>
              <h4 class="card-headline">{{ item.title }}</h4>
              <p class="card-summary">{{ item.summary }}</p>
              <div v-if="item.investment_takeaway" class="card-takeaway">
                💡 {{ item.investment_takeaway }}
              </div>
            </article>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import PriceValuationChart from './PriceValuationChart.vue'
import type { XBRLStatement, News } from '../server/utils/types'

const props = defineProps<{
  ticker: string
}>()

defineEmits<{
  (e: 'close'): void
}>()

const activeModalTab = ref<'terminal' | 'moat' | 'chart' | 'news'>('terminal')
const matrixMetric = ref<'net_income' | 'revenue' | 'fcf' | 'roic' | 'roe' | 'op_margin' | 'gross_margin' | 'invested_capital'>('net_income')

const statements = ref<XBRLStatement[]>([])
const sectorNews = ref<News[]>([])
const loading = ref(true)
const loadingNews = ref(false)

const latestStatement = computed(() => {
  if (statements.value.length === 0) return null
  return statements.value[0]
})

const latestTimingSignal = computed(() => {
  return latestStatement.value?.timing_signal || latestStatement.value?.valuation?.timing_signal || null
})

const latestValuationBands = computed(() => {
  return latestStatement.value?.valuation_bands || latestStatement.value?.valuation?.valuation_bands || null
})

const getTimingScoreClass = (score?: number) => {
  if (!score) return 'timing-low'
  if (score >= 70) return 'timing-high'
  if (score >= 50) return 'timing-mid'
  return 'timing-low'
}

const getTimingBannerClass = (score?: number) => {
  if (!score) return 'banner-neutral'
  if (score >= 70) return 'banner-bullish'
  if (score >= 50) return 'banner-amber'
  return 'banner-neutral'
}

const getPeZoneDesc = computed(() => {
  const vb = latestValuationBands.value
  const p = latestStatement.value?.valuation?.current_price
  if (!vb || !p || vb.mean_price_pe <= 0) return '-'
  if (vb.minus_2sd_price_pe > 0 && p <= vb.minus_2sd_price_pe) return '≤ -2σ (Deep Value)'
  if (vb.minus_1sd_price_pe > 0 && p <= vb.minus_1sd_price_pe) return '-1σ to -2σ (Accumulation)'
  if (vb.mean_price_pe > 0 && p <= vb.mean_price_pe) return 'Mean to -1σ (Discount)'
  return 'Fair / Premium'
})

const uniqueYears = computed(() => {
  const years = Array.from(new Set(statements.value.map(s => s.year))).sort((a, b) => b - a)
  return years.slice(0, 5)
})

const fiveYearAnnuals = computed(() => {
  if (statements.value.length === 0) return []
  const years = Array.from(new Set(statements.value.map(s => s.year))).sort((a, b) => a - b).slice(-5)

  return years.map(y => {
    const yearStmts = statements.value.filter(s => s.year === y)
    let stmt = yearStmts.find(s => {
      const p = s.period.toUpperCase()
      return p === 'FY' || p === 'TAHUNAN' || p === 'AUDIT' || p === 'Q4' || p === 'IV'
    })
    if (!stmt && yearStmts.length > 0) {
      stmt = yearStmts[0]
    }
    if (!stmt) return null

    const roic = (stmt.computed_ratios?.roic || 0) * 100
    const roe = (stmt.computed_ratios?.roe || 0) * 100
    const spread = roe - roic
    const grossMargin = stmt.computed_ratios?.gross_margin_pct || 0
    const operatingMargin = stmt.computed_ratios?.operating_margin_pct || 0
    const netMargin = stmt.computed_ratios?.net_margin_pct || 0
    const revenue = stmt.core?.revenue || 0
    const netIncome = stmt.core?.net_income_parent || stmt.core?.net_income || 0
    const cfo = stmt.core?.operating_cash_flow || 0
    const capex = stmt.core?.capex || 0
    const fcf = stmt.core?.free_cash_flow || 0
    const fcfConversion = netIncome > 0 ? (fcf / netIncome) * 100 : 0
    const capexToCfo = cfo > 0 ? (capex / cfo) * 100 : 0
    const investedCapital = (stmt.core?.total_equity || 0) + (stmt.core?.total_debt || 0) - (stmt.core?.cash_and_equivalents || 0)
    const debtToEquity = stmt.computed_ratios?.debt_to_equity || 0
    const totalAssets = stmt.core?.total_assets || 0
    const totalEquity = stmt.core?.total_equity || 1
    const leverageMultiplier = totalEquity > 0 ? totalAssets / totalEquity : 1
    const piotroski = stmt.computed_ratios?.piotroski_f_score || 0
    const altmanZ = stmt.computed_ratios?.altman_z_score || 0

    return {
      year: y,
      period: stmt.period,
      roic,
      roe,
      spread,
      grossMargin,
      operatingMargin,
      netMargin,
      revenue,
      netIncome,
      cfo,
      capex,
      fcf,
      fcfConversion,
      capexToCfo,
      investedCapital,
      debtToEquity,
      leverageMultiplier,
      piotroski,
      altmanZ
    }
  }).filter(Boolean) as Array<{
    year: number
    period: string
    roic: number
    roe: number
    spread: number
    grossMargin: number
    operatingMargin: number
    netMargin: number
    revenue: number
    netIncome: number
    cfo: number
    capex: number
    fcf: number
    fcfConversion: number
    capexToCfo: number
    investedCapital: number
    debtToEquity: number
    leverageMultiplier: number
    piotroski: number
    altmanZ: number
  }>
})

const fiveYearAvgRoic = computed(() => {
  const items = fiveYearAnnuals.value
  if (items.length === 0) return null
  const sum = items.reduce((acc, it) => acc + it.roic, 0)
  return sum / items.length
})

const fiveYearAvgRoe = computed(() => {
  const items = fiveYearAnnuals.value
  if (items.length === 0) return null
  const sum = items.reduce((acc, it) => acc + it.roe, 0)
  return sum / items.length
})

const fiveYearAvgSpread = computed(() => {
  if (fiveYearAvgRoe.value === null || fiveYearAvgRoic.value === null) return null
  return fiveYearAvgRoe.value - fiveYearAvgRoic.value
})

const fiveYearAvgFcfConversion = computed(() => {
  const items = fiveYearAnnuals.value
  if (items.length === 0) return null
  const valid = items.filter(it => it.netIncome > 0)
  if (valid.length === 0) return 0
  const sum = valid.reduce((acc, it) => acc + it.fcfConversion, 0)
  return sum / valid.length
})

const fiveYearAvgGrossMargin = computed(() => {
  const items = fiveYearAnnuals.value
  if (items.length === 0) return null
  const sum = items.reduce((acc, it) => acc + it.grossMargin, 0)
  return sum / items.length
})

const fiveYearAvgOpMargin = computed(() => {
  const items = fiveYearAnnuals.value
  if (items.length === 0) return null
  const sum = items.reduce((acc, it) => acc + it.operatingMargin, 0)
  return sum / items.length
})

const fiveYearAvgPiotroski = computed(() => {
  const items = fiveYearAnnuals.value
  if (items.length === 0) return null
  const sum = items.reduce((acc, it) => acc + it.piotroski, 0)
  return sum / items.length
})

const fiveYearAvgAltmanZ = computed(() => {
  const items = fiveYearAnnuals.value
  if (items.length === 0) return null
  const sum = items.reduce((acc, it) => acc + it.altmanZ, 0)
  return sum / items.length
})

const fiveYearRevenueCagr = computed(() => {
  const items = fiveYearAnnuals.value
  if (items.length < 2) return null
  const first = items[0].revenue
  const last = items[items.length - 1].revenue
  if (first <= 0 || last <= 0) return null
  const n = items.length - 1
  return (Math.pow(last / first, 1 / n) - 1) * 100
})

const fiveYearNetIncomeCagr = computed(() => {
  const items = fiveYearAnnuals.value
  if (items.length < 2) return null
  const first = items[0].netIncome
  const last = items[items.length - 1].netIncome
  if (first <= 0 || last <= 0) return null
  const n = items.length - 1
  return (Math.pow(last / first, 1 / n) - 1) * 100
})

const moatVerdict = computed(() => {
  const roic = fiveYearAvgRoic.value || 0
  if (roic >= 20) {
    return {
      badge: '💎 Wide Moat Compounder',
      class: 'text-green font-bold',
      desc: 'Sustained elite capital returns beating 15% hurdle rate. Durable pricing power & barriers to entry.'
    }
  }
  if (roic >= 15) {
    return {
      badge: '🛡️ Solid Economic Moat',
      class: 'text-green font-bold',
      desc: 'Reliably exceeds cost of capital (11.5% Indonesian WACC). Solid competitive moat & shareholder value creator.'
    }
  }
  if (roic >= 10) {
    return {
      badge: '🟡 Narrow / Average Moat',
      class: 'text-amber font-bold',
      desc: 'Covers cost of capital with modest spread. Vulnerable to cyclical downturns or margin pressure.'
    }
  }
  return {
    badge: '🔴 Value Destructive / No Moat',
    class: 'text-red font-bold',
    desc: 'Fails to beat 11.5% WACC hurdle rate. Reinvestment does not compound economic value.'
  }
})

const leverageVerdict = computed(() => {
  const spread = fiveYearAvgSpread.value || 0
  const roic = fiveYearAvgRoic.value || 0
  if (spread > 15 && roic < 12) {
    return { title: '⚠️ Financial Leverage Masking Low ROIC', class: 'text-amber' }
  }
  if (roic >= 15) {
    return { title: '🟢 Pure Organic Compounding', class: 'text-green' }
  }
  return { title: '⚪ Balanced Capital Structure', class: 'text-secondary' }
})

const cashFlowVerdict = computed(() => {
  const fcfConv = fiveYearAvgFcfConversion.value || 0
  if (fcfConv >= 80) {
    return { title: '🟢 Pristine Owner Earnings (FCF ≥ 80%)', class: 'text-green' }
  }
  if (fcfConv >= 50) {
    return { title: '🟡 Moderate Cash Flow Conversion', class: 'text-amber' }
  }
  return { title: '🔴 Working Capital Drag / Heavy CapEx', class: 'text-red' }
})

const pricingPowerVerdict = computed(() => {
  const items = fiveYearAnnuals.value
  if (items.length < 2) return { title: 'Stable Margins', class: 'text-secondary', desc: 'Sufficient margin stability.' }
  const first = items[0].grossMargin
  const last = items[items.length - 1].grossMargin
  if (last >= first) {
    return { title: '🟢 Expanding Pricing Power', class: 'text-green', desc: 'Gross margins expanded over 5 years, proving strong pricing power.' }
  }
  if (first - last < 3) {
    return { title: '🟢 Resilient Margins', class: 'text-green', desc: 'Gross margins held resilient against inflationary pressures.' }
  }
  return { title: '⚠️ Margin Compression', class: 'text-amber', desc: 'Gross margins contracted over 5 years, indicating pricing competition.' }
})

const roicChartMax = computed(() => {
  const maxVal = Math.max(...fiveYearAnnuals.value.map(it => it.roic), 25)
  return Math.ceil(maxVal / 5) * 5
})

const getRoicBarY = (val: number) => {
  const max = roicChartMax.value
  const clamped = Math.max(0, Math.min(val, max))
  const h = (clamped / max) * 120
  return 160 - h
}

const getRoicBarHeight = (val: number) => {
  const max = roicChartMax.value
  const clamped = Math.max(0, Math.min(val, max))
  return Math.max(3, (clamped / max) * 120)
}

const getRoicColor = (val: number) => {
  if (val >= 20) return '#10b981'
  if (val >= 15) return '#34d399'
  if (val >= 10) return '#fbbf24'
  return '#f87171'
}

const rateChartMax = computed(() => {
  const maxVal = Math.max(
    ...fiveYearAnnuals.value.map(it => Math.max(it.roe, it.roic)),
    30
  )
  return Math.ceil(maxVal / 10) * 10
})

const getRateBarY = (val: number) => {
  const max = rateChartMax.value
  const clamped = Math.max(0, Math.min(val, max))
  const h = (clamped / max) * 120
  return 160 - h
}

const getRateBarHeight = (val: number) => {
  const max = rateChartMax.value
  const clamped = Math.max(0, Math.min(val, max))
  return Math.max(3, (clamped / max) * 120)
}

const cashChartMax = computed(() => {
  const maxVal = Math.max(
    ...fiveYearAnnuals.value.map(it => Math.max(it.netIncome, it.cfo, it.fcf)),
    1000
  )
  return maxVal > 0 ? maxVal : 1000
})

const getCashBarY = (val: number) => {
  const max = cashChartMax.value
  const clamped = Math.max(0, Math.min(val, max))
  const h = (clamped / max) * 120
  return 160 - h
}

const getCashBarHeight = (val: number) => {
  const max = cashChartMax.value
  const clamped = Math.max(0, Math.min(val, max))
  return Math.max(3, (clamped / max) * 120)
}

const marginChartMax = computed(() => {
  const maxVal = Math.max(
    ...fiveYearAnnuals.value.map(it => Math.max(it.grossMargin, it.operatingMargin)),
    40
  )
  return Math.ceil(maxVal / 10) * 10
})

const getMarginBarY = (val: number) => {
  const max = marginChartMax.value
  const clamped = Math.max(0, Math.min(val, max))
  const h = (clamped / max) * 120
  return 160 - h
}

const getMarginBarHeight = (val: number) => {
  const max = marginChartMax.value
  const clamped = Math.max(0, Math.min(val, max))
  return Math.max(3, (clamped / max) * 120)
}

const fetchFinancials = async () => {
  if (!props.ticker) return
  loading.value = true
  try {
    const data = await $fetch<XBRLStatement[]>(`/api/v1/stocks/${props.ticker}/financials`)
    statements.value = data || []
  } catch (e) {
    console.error('Failed to load financials', e)
    statements.value = []
  } finally {
    loading.value = false
  }
}

const fetchSectorNews = async () => {
  const sector = latestStatement.value?.metadata?.sector
  if (!sector) return
  loadingNews.value = true
  try {
    const res = await $fetch<any>(`/api/v1/news?sector=${encodeURIComponent(sector)}&limit=20`)
    sectorNews.value = Array.isArray(res) ? res : (res.data || [])
  } catch (e) {
    console.error('Failed to fetch sector news', e)
  } finally {
    loadingNews.value = false
  }
}

const getMatrixValue = (year: number, period: string, metric: 'net_income' | 'revenue' | 'fcf' | 'roic' | 'roe' | 'op_margin' | 'gross_margin' | 'invested_capital') => {
  const stmt = statements.value.find(s => s.year === year && s.period.toUpperCase() === period.toUpperCase())
  if (!stmt) return '-'
  if (metric === 'net_income') return formatCompact(stmt.core?.net_income_parent || stmt.core?.net_income || 0)
  if (metric === 'revenue') return formatCompact(stmt.core?.revenue || 0)
  if (metric === 'fcf') return formatCompact(stmt.core?.free_cash_flow || 0)
  if (metric === 'roic') return formatPct((stmt.computed_ratios?.roic || 0) * 100)
  if (metric === 'roe') return formatPct((stmt.computed_ratios?.roe || 0) * 100)
  if (metric === 'op_margin') return formatPct(stmt.computed_ratios?.operating_margin_pct)
  if (metric === 'gross_margin') return formatPct(stmt.computed_ratios?.gross_margin_pct)
  if (metric === 'invested_capital') {
    const ic = (stmt.core?.total_equity || 0) + (stmt.core?.total_debt || 0) - (stmt.core?.cash_and_equivalents || 0)
    return formatCompact(ic)
  }
  return '-'
}

const formatMultiple = (val?: number) => {
  if (!val || val <= 0) return '-'
  return val.toFixed(2) + 'x'
}

const formatPct = (val?: number) => {
  if (!val || isNaN(val)) return '-'
  return val.toFixed(2) + '%'
}

const formatSignedPct = (val?: number) => {
  if (val === undefined || isNaN(val)) return '-'
  return (val > 0 ? '+' : '') + val.toFixed(1) + '%'
}

const formatIDRPrice = (val?: number) => {
  if (!val || isNaN(val) || val <= 0) return '-'
  return 'Rp ' + Math.round(val).toLocaleString('en-US')
}

const formatCompact = (val?: number) => {
  if (val === undefined || val === null || isNaN(val) || val === 0) return '-'
  const abs = Math.abs(val)
  if (abs >= 1e12) return (val / 1e12).toFixed(2) + ' T'
  if (abs >= 1e9) return (val / 1e9).toFixed(2) + ' B'
  if (abs >= 1e6) return (val / 1e6).toFixed(2) + ' M'
  return Math.round(val).toLocaleString('en-US')
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

watch(() => props.ticker, () => {
  fetchFinancials()
}, { immediate: true })

watch(activeModalTab, (tab) => {
  if (tab === 'news' && sectorNews.value.length === 0) {
    fetchSectorNews()
  }
})
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.85);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 20px;
}
.modal-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  width: 100%;
  max-width: 1400px;
  max-height: 94vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.7);
}
.modal-header {
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.title-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
.ticker-badge {
  background: #2563eb;
  color: #fff;
  font-weight: 800;
  font-size: 0.95rem;
  padding: 3px 8px;
  border-radius: 4px;
}
.company-name {
  font-size: 1.35rem;
  font-weight: 700;
}
.meta-tags {
  display: flex;
  gap: 8px;
  margin-top: 6px;
  flex-wrap: wrap;
}
.tag {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  font-size: 0.75rem;
  padding: 2px 8px;
  border-radius: 4px;
}
.sector-tag {
  color: #38bdf8;
  border-color: rgba(56, 189, 248, 0.3);
}
.btn-close {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 1.25rem;
  cursor: pointer;
}
.btn-close:hover {
  color: #fff;
}
.modal-tab-bar {
  display: flex;
  background: var(--bg-app);
  padding: 6px 24px;
  gap: 8px;
  border-bottom: 1px solid var(--border-color);
}
.modal-tab-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 6px 14px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
}
.modal-tab-btn:hover {
  color: #fff;
  background: var(--bg-card-hover);
}
.modal-tab-btn.active {
  background: var(--bg-card);
  color: #38bdf8;
  border: 1px solid var(--border-color);
}
.tab-badge {
  background: #1e293b;
  color: #38bdf8;
  font-size: 0.7rem;
  padding: 1px 6px;
  border-radius: 4px;
}
.tab-badge.badge-green {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.4);
}
.tab-badge.badge-amber {
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.4);
}
.modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}
.matrix-layout {
  display: grid;
  grid-template-columns: 320px 1fr 340px;
  gap: 20px;
  align-items: start;
}
.matrix-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.metric-card {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
}
.card-heading {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  padding-bottom: 6px;
}
.data-rows {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.data-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
}
.data-row .label {
  color: var(--text-secondary);
}
.data-row .val {
  font-weight: 600;
  color: #f8fafc;
}
.highlight-row {
  background: rgba(255, 255, 255, 0.03);
  padding: 4px 6px;
  border-radius: 4px;
}
.text-green {
  color: #34d399 !important;
}
.text-red {
  color: #f87171 !important;
}
.text-amber {
  color: #fbbf24 !important;
}
.text-cyan {
  color: #38bdf8 !important;
}
.matrix-table-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.table-header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.matrix-tabs {
  display: flex;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  padding: 2px;
  border-radius: 6px;
  gap: 2px;
}
.m-tab {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 4px 10px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  border-radius: 4px;
}
.m-tab.active {
  background: #2563eb;
  color: #fff;
}
.matrix-currency {
  font-size: 0.75rem;
  color: var(--text-muted);
}
.matrix-table {
  width: 100%;
  border-collapse: collapse;
  text-align: right;
  font-size: 0.85rem;
}
.matrix-table th {
  background: rgba(255, 255, 255, 0.02);
  padding: 8px 12px;
  color: var(--text-muted);
  border-bottom: 1px solid var(--border-color);
}
.matrix-table th:first-child, .matrix-table td:first-child {
  text-align: left;
}
.matrix-table td {
  padding: 10px 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}
.period-cell {
  color: #38bdf8;
  font-weight: 700;
}
.key-stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  background: var(--bg-card);
  padding: 10px 14px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
}
.stat-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.stat-label {
  font-size: 0.7rem;
  color: var(--text-muted);
}
.stat-val {
  font-size: 0.9rem;
  font-weight: 700;
  color: #f8fafc;
}
.dual-cards-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.chart-tab-content {
  background: var(--bg-app);
  border-radius: 8px;
  padding: 20px;
}
.sector-news-tab {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.news-tab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.news-tab-title {
  font-size: 1.1rem;
  font-weight: 700;
}
.news-tab-sub {
  color: var(--text-secondary);
  font-size: 0.85rem;
}
.btn-refresh {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: #38bdf8;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 0.8rem;
  cursor: pointer;
}
.news-cards-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.sector-news-card {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.card-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.75rem;
}
.sentiment-tag {
  font-size: 0.7rem;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 4px;
  text-transform: uppercase;
}
.sentiment-tag.bullish { background: var(--bullish-bg); color: var(--bullish-text); }
.sentiment-tag.bearish { background: var(--bearish-bg); color: var(--bearish-text); }
.sentiment-tag.neutral { background: var(--neutral-bg); color: var(--neutral-text); }
.score-tag {
  color: var(--text-secondary);
}
.card-headline {
  font-size: 1rem;
  font-weight: 600;
}
.card-summary {
  font-size: 0.85rem;
  color: var(--text-secondary);
  line-height: 1.4;
}
.card-takeaway {
  font-size: 0.8rem;
  color: #fde68a;
  background: rgba(245, 158, 11, 0.06);
  padding: 6px 8px;
  border-radius: 4px;
}
.loading-state, .empty-state {
  text-align: center;
  padding: 60px;
  color: var(--text-muted);
}

/* Smart Timing Card & Badges */
.timing-card {
  border: 1px solid rgba(56, 189, 248, 0.25);
  background: rgba(15, 23, 42, 0.7);
}
.timing-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  padding-bottom: 6px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}
.timing-heading {
  margin-bottom: 0;
  border-bottom: none;
  padding-bottom: 0;
  color: #38bdf8;
}
.timing-score-pill {
  font-size: 0.8rem;
  font-weight: 800;
  padding: 2px 8px;
  border-radius: 4px;
}
.timing-high {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
  border: 1px solid #10b981;
}
.timing-mid {
  background: rgba(245, 158, 11, 0.18);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.4);
}
.timing-low {
  background: rgba(100, 116, 139, 0.2);
  color: var(--text-muted);
  border: 1px solid var(--border-color);
}
.timing-status-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 6px;
  margin-bottom: 12px;
  font-size: 0.8rem;
}
.timing-status-banner.banner-bullish {
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.35);
  color: #34d399;
}
.timing-status-banner.banner-amber {
  background: rgba(245, 158, 11, 0.12);
  border: 1px solid rgba(245, 158, 11, 0.35);
  color: #fbbf24;
}
.timing-status-banner.banner-neutral {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
}
.badge-bullish-div {
  background: rgba(56, 189, 248, 0.2);
  color: #38bdf8;
  font-size: 0.68rem;
  font-weight: 800;
  padding: 1px 5px;
  border-radius: 3px;
  margin-left: 6px;
}
.badge-vdu {
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
  font-size: 0.68rem;
  font-weight: 800;
  padding: 1px 5px;
  border-radius: 3px;
  margin-left: 6px;
}
.catalyst-chips-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
  padding-top: 8px;
  border-top: 1px dashed rgba(255, 255, 255, 0.06);
}
.catalyst-pill {
  background: rgba(56, 189, 248, 0.1);
  border: 1px solid rgba(56, 189, 248, 0.25);
  color: #7dd3fc;
  font-size: 0.7rem;
  padding: 2px 6px;
  border-radius: 4px;
}

/* 5-Year Moat Radar & Quality Dashboard Styles */
.moat-radar-tab {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.moat-kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.moat-kpi-card {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.kpi-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.kpi-title {
  font-size: 0.75rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  font-weight: 700;
}
.kpi-pill {
  font-size: 0.65rem;
  font-weight: 800;
  padding: 1px 6px;
  border-radius: 4px;
}
.pill-green {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.4);
}
.pill-amber {
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.4);
}
.pill-red {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.4);
}
.pill-cyan {
  background: rgba(56, 189, 248, 0.2);
  color: #38bdf8;
  border: 1px solid rgba(56, 189, 248, 0.4);
}
.kpi-big-num {
  font-size: 1.8rem;
  font-weight: 800;
  line-height: 1.1;
}
.kpi-big-label {
  font-size: 1.1rem;
  font-weight: 800;
  line-height: 1.2;
}
.kpi-subtext {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.75rem;
  color: var(--text-secondary);
}
.kpi-desc {
  font-size: 0.75rem;
  color: var(--text-secondary);
  line-height: 1.35;
}
.spread-tag {
  font-weight: 700;
}
.moat-charts-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.moat-chart-card {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.chart-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}
.chart-card-title {
  font-size: 0.88rem;
  font-weight: 700;
  color: #f8fafc;
  margin-bottom: 2px;
}
.chart-card-sub {
  font-size: 0.72rem;
  color: var(--text-secondary);
  line-height: 1.3;
}
.chart-legend-row {
  display: flex;
  gap: 8px;
  align-items: center;
  font-size: 0.7rem;
  color: var(--text-muted);
  flex-shrink: 0;
}
.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}
.dot-green { background: #10b981; }
.dot-amber { background: #f59e0b; }
.dot-red { background: #ef4444; }
.dot-cyan { background: #38bdf8; }
.dot-blue { background: #3b82f6; }
.dot-emerald { background: #10b981; }
.svg-chart-container {
  width: 100%;
  background: rgba(0, 0, 0, 0.25);
  border: 1px solid rgba(255, 255, 255, 0.04);
  border-radius: 6px;
  padding: 8px 4px;
}
.moat-svg {
  width: 100%;
  height: auto;
  overflow: visible;
}
.svg-bar {
  transition: opacity 0.2s, transform 0.2s;
  cursor: pointer;
}
.svg-bar:hover {
  opacity: 0.85;
  filter: brightness(1.2);
}
.moat-table-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.moat-table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.avg-footer-row {
  background: rgba(56, 189, 248, 0.06);
  border-top: 2px solid rgba(56, 189, 248, 0.3);
}
.quality-checklist-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.checklist-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
}
.check-item {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.check-header {
  display: flex;
  align-items: center;
  gap: 8px;
}
.check-icon {
  font-size: 1rem;
}
.check-title {
  font-size: 0.8rem;
  font-weight: 700;
  color: #f8fafc;
}
.check-desc {
  font-size: 0.72rem;
  color: var(--text-secondary);
  line-height: 1.35;
}
@media (max-width: 1200px) {
  .moat-kpi-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .moat-charts-grid {
    grid-template-columns: 1fr;
  }
  .checklist-grid {
    grid-template-columns: 1fr 1fr;
  }
}
@media (max-width: 768px) {
  .moat-kpi-grid {
    grid-template-columns: 1fr;
  }
  .checklist-grid {
    grid-template-columns: 1fr;
  }
}
@media (max-width: 1200px) {
  .matrix-layout {
    grid-template-columns: 1fr 1fr;
  }
}
@media (max-width: 768px) {
  .matrix-layout {
    grid-template-columns: 1fr;
  }
  .dual-cards-grid {
    grid-template-columns: 1fr;
  }
}
</style>
