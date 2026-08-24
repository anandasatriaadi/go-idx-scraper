<template>
  <div class="price-chart-card card-section font-mono">
    <!-- Header & Controls Bar -->
    <div class="chart-header-row">
      <div class="header-left">
        <div class="title-with-badge">
          <h2 class="chart-title">📈 PRICE & HISTORICAL VALUATION BANDS</h2>
          <span v-if="ticker" class="ticker-pill">${{ ticker }}</span>
        </div>
        <div class="chart-summary-stats">
          <span v-if="latestPrice" class="stat-item">
            <span class="stat-lbl">CLOSE:</span>
            <span class="stat-val text-cyan">Rp {{ formatNum(latestPrice.close) }}</span>
          </span>
          <span v-if="periodChangePct !== null" :class="['stat-item', periodChangePct >= 0 ? 'text-green' : 'text-red']">
            <span class="stat-lbl">CHG:</span>
            <span class="stat-val">{{ (periodChangePct >= 0 ? '+' : '') + periodChangePct.toFixed(2) }}%</span>
          </span>
          <span v-if="grahamNumber && grahamNumber > 0" class="stat-item">
            <span class="stat-lbl">GRAHAM FV:</span>
            <span class="stat-val text-green">Rp {{ formatNum(grahamNumber) }}</span>
          </span>
          <span v-if="currentMosPct !== null" :class="['stat-item', currentMosPct >= 30 ? 'text-green' : currentMosPct > 0 ? 'text-amber' : 'text-red']">
            <span class="stat-lbl">MOS:</span>
            <span class="stat-val">{{ (currentMosPct >= 0 ? '+' : '') + currentMosPct.toFixed(1) }}%</span>
          </span>
          <span v-if="timingSignal?.score !== undefined" :class="['stat-item timing-stat-item', getTimingScoreClass(timingSignal.score)]">
            <span class="stat-lbl">TIMING SCORE:</span>
            <span class="stat-val font-bold">{{ timingSignal.score }}/100</span>
          </span>
        </div>
      </div>

      <!-- Controls Right: Range Switcher & Indicator Toggles -->
      <div class="header-controls">
        <!-- Indicator Toggles -->
        <div class="indicator-toggles">
          <button
            v-if="hasPeBands"
            :class="['toggle-btn pe-toggle', { active: showPeBands }]"
            title="Toggle Historical P/E Standard Deviation Bands (±1σ, ±2σ)"
            @click="showPeBands = !showPeBands"
          >
            <span class="dot pe-dot"></span> P/E Bands
          </button>
          <button
            v-if="hasPbBands"
            :class="['toggle-btn pb-toggle', { active: showPbBands }]"
            title="Toggle Historical P/B Standard Deviation Bands (±1σ, ±2σ)"
            @click="showPbBands = !showPbBands"
          >
            <span class="dot pb-dot"></span> P/B Bands
          </button>
          <button
            :class="['toggle-btn sma50-toggle', { active: showSma50 }]"
            title="Toggle 50-day Simple Moving Average"
            @click="showSma50 = !showSma50"
          >
            <span class="dot sma50-dot"></span> SMA 50
          </button>
          <button
            :class="['toggle-btn sma200-toggle', { active: showSma200 }]"
            title="Toggle 200-day Simple Moving Average"
            @click="showSma200 = !showSma200"
          >
            <span class="dot sma200-dot"></span> SMA 200
          </button>
        </div>

        <!-- Range Buttons -->
        <div class="range-selector">
          <button
            v-for="r in rangeOptions"
            :key="r.id"
            :class="['range-btn', { active: selectedRange === r.id }]"
            @click="setRange(r.id)"
          >
            {{ r.label }}
          </button>
        </div>
      </div>
    </div>

    <!-- Smart Timing Score Banner -->
    <div v-if="timingSignal" :class="['smart-timing-banner font-mono', getTimingBannerClass(timingSignal.score)]">
      <div class="timing-banner-main">
        <div class="timing-score-badge">
          <span class="timing-indicator-dot">{{ timingSignal.score >= 70 ? '🟢' : timingSignal.score >= 50 ? '🟡' : '⚪' }}</span>
          <span class="score-number">{{ timingSignal.score }}/100</span>
          <span class="timing-status-label">{{ timingSignal.status || 'Timing Signal' }}</span>
        </div>
        <div class="timing-catalysts-list">
          <span v-if="timingSignal.rsi !== undefined" class="catalyst-chip" :class="{ 'chip-bullish': timingSignal.rsi < 35 || timingSignal.rsi_bullish_divergence }">
            RSI(14): {{ timingSignal.rsi.toFixed(1) }}
            <strong v-if="timingSignal.rsi_bullish_divergence" class="chip-alert">⚡ Bullish Div</strong>
          </span>
          <span v-if="timingSignal.stopping_volume" class="catalyst-chip chip-bullish">
            🛡️ VSA Stopping Volume
          </span>
          <span v-if="timingSignal.volume_dry_up" class="catalyst-chip chip-amber">
            💧 Volume Dry-Up (VDU: {{ timingSignal.vdu?.toFixed(2) || '0.00' }}x)
          </span>
          <span v-if="timingSignal.rvol !== undefined && timingSignal.rvol > 0" class="catalyst-chip">
            RVOL: {{ timingSignal.rvol.toFixed(2) }}x
          </span>
          <span v-if="timingSignal.clv !== undefined" class="catalyst-chip">
            CLV: {{ (timingSignal.clv > 0 ? '+' : '') + timingSignal.clv.toFixed(2) }}
          </span>
          <span v-if="timingSignal.valuation_discount_zone" class="catalyst-chip chip-emerald">
            Zone: {{ timingSignal.valuation_discount_zone }}
          </span>
          <span v-for="(sig, sIdx) in (timingSignal.signals || [])" :key="sIdx" class="catalyst-chip chip-extra">
            {{ sig }}
          </span>
        </div>
      </div>
    </div>

    <!-- Chart Legend & Guide -->
    <div class="chart-legend">
      <div class="legend-item">
        <span class="legend-line price-line"></span> Stock Price
      </div>
      <div v-if="grahamNumber && grahamNumber > 0" class="legend-item">
        <span class="legend-line graham-line"></span> Graham Fair Value (Rp {{ formatNum(grahamNumber) }})
      </div>
      <div v-if="grahamNumber && grahamNumber > 0" class="legend-item">
        <span class="legend-box mos-box"></span> 30% MoS Zone (≤ Rp {{ formatNum(grahamNumber * 0.7) }})
      </div>

      <!-- P/E Valuation Bands Legend -->
      <template v-if="showPeBands && valuationBands">
        <div class="legend-item">
          <span class="legend-line pe-mean-line"></span> P/E Mean (Rp {{ formatNum(valuationBands.mean_price_pe) }})
        </div>
        <div class="legend-item">
          <span class="legend-line pe-plus-line"></span> P/E +1σ / +2σ
        </div>
        <div class="legend-item">
          <span class="legend-box pe-accum-box"></span> P/E Accumulation Zone (-1σ to -2σ)
        </div>
      </template>

      <!-- P/B Valuation Bands Legend -->
      <template v-if="showPbBands && valuationBands">
        <div class="legend-item">
          <span class="legend-line pb-mean-line"></span> P/B Mean (Rp {{ formatNum(valuationBands.mean_price_pb) }})
        </div>
        <div class="legend-item">
          <span class="legend-box pb-accum-box"></span> P/B Accumulation Zone (-1σ to -2σ)
        </div>
      </template>

      <div v-if="showSma50" class="legend-item">
        <span class="legend-line sma50-line"></span> SMA 50
      </div>
      <div v-if="showSma200" class="legend-item">
        <span class="legend-line sma200-line"></span> SMA 200
      </div>
    </div>

    <!-- Chart Main Stage -->
    <div ref="chartContainerRef" class="chart-stage" @mouseleave="handleMouseLeave">
      <!-- Loading Overlay -->
      <div v-if="loading" class="chart-overlay-state">
        <div class="spinner"></div>
        <span>Loading daily price history...</span>
      </div>

      <!-- Empty State -->
      <div v-else-if="!prices || prices.length === 0" class="chart-overlay-state">
        <span>No historical price candles found for ${{ ticker }}. Run seed_ticker or price_updater.</span>
      </div>

      <!-- SVG Chart Canvas -->
      <svg
        v-else
        ref="svgRef"
        class="chart-svg"
        viewBox="0 0 800 340"
        preserveAspectRatio="none"
        @mousemove="handleMouseMove"
        @touchmove="handleTouchMove"
        @touchstart="handleTouchMove"
      >
        <defs>
          <!-- Price Area Gradient -->
          <linearGradient id="priceGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#38bdf8" stop-opacity="0.3" />
            <stop offset="60%" stop-color="#38bdf8" stop-opacity="0.08" />
            <stop offset="100%" stop-color="#38bdf8" stop-opacity="0.0" />
          </linearGradient>

          <!-- Margin of Safety Zone Gradient -->
          <linearGradient id="mosGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#10b981" stop-opacity="0.18" />
            <stop offset="100%" stop-color="#10b981" stop-opacity="0.06" />
          </linearGradient>

          <!-- P/E Accumulation Band Gradient (-1SD to -2SD) -->
          <linearGradient id="peAccumGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#34d399" stop-opacity="0.22" />
            <stop offset="100%" stop-color="#10b981" stop-opacity="0.10" />
          </linearGradient>

          <!-- P/B Accumulation Band Gradient (-1SD to -2SD) -->
          <linearGradient id="pbAccumGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#38bdf8" stop-opacity="0.20" />
            <stop offset="100%" stop-color="#2563eb" stop-opacity="0.08" />
          </linearGradient>

          <!-- Glow Filter for Active Dot -->
          <filter id="glow" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur in="SourceGraphic" stdDeviation="3" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>

        <!-- 1. Background Grid Horizontal Lines & Y-Axis Labels -->
        <g class="grid-lines">
          <g v-for="(grid, idx) in yGridLines" :key="idx">
            <line
              :x1="padding.left"
              :y1="grid.y"
              :x2="padding.left + plotWidth"
              :y2="grid.y"
              stroke="#1e293b"
              stroke-dasharray="3 3"
              stroke-width="1"
            />
            <text
              :x="padding.left + plotWidth + 6"
              :y="grid.y + 4"
              fill="#64748b"
              font-size="10"
              font-family="var(--font-mono)"
              text-anchor="start"
            >
              {{ formatCompact(grid.val) }}
            </text>
          </g>
        </g>

        <!-- 2. Background Vertical Grid Lines & X-Axis Date Ticks -->
        <g class="x-grid-lines">
          <g v-for="(tick, idx) in xGridTicks" :key="idx">
            <line
              :x1="tick.x"
              :y1="padding.top"
              :x2="tick.x"
              :y2="padding.top + plotHeight"
              stroke="#1e293b"
              stroke-dasharray="2 4"
              stroke-width="1"
            />
            <text
              :x="tick.x"
              :y="padding.top + plotHeight + 18"
              fill="#64748b"
              font-size="10"
              font-family="var(--font-mono)"
              text-anchor="middle"
            >
              {{ tick.label }}
            </text>
          </g>
        </g>

        <!-- 3. Margin of Safety (MoS) 30% Shaded Zone -->
        <g v-if="grahamNumber && grahamNumber > 0 && grahamY !== null && mosY !== null" class="mos-band">
          <rect
            :x="padding.left"
            :y="Math.min(grahamY, mosY)"
            :width="plotWidth"
            :height="Math.abs(mosY - grahamY)"
            fill="url(#mosGradient)"
          />
          <!-- 30% Discount Lower Boundary -->
          <line
            :x1="padding.left"
            :y1="mosY"
            :x2="padding.left + plotWidth"
            :y2="mosY"
            stroke="#10b981"
            stroke-dasharray="3 3"
            stroke-width="1.2"
            stroke-opacity="0.8"
          />
          <text
            :x="padding.left + 8"
            :y="mosY - 5"
            fill="#10b981"
            font-size="9"
            font-weight="600"
            font-family="var(--font-mono)"
            fill-opacity="0.85"
          >
            -30% MoS Entry (Rp {{ formatNum(grahamNumber * 0.7) }})
          </text>

          <!-- Graham Fair Value Top Line -->
          <line
            :x1="padding.left"
            :y1="grahamY"
            :x2="padding.left + plotWidth"
            :y2="grahamY"
            stroke="#10b981"
            stroke-dasharray="6 4"
            stroke-width="1.75"
          />
          <text
            :x="padding.left + 8"
            :y="grahamY - 5"
            fill="#34d399"
            font-size="9.5"
            font-weight="700"
            font-family="var(--font-mono)"
          >
            Graham Fair Value (Rp {{ formatNum(grahamNumber) }})
          </text>
        </g>

        <!-- 4. P/E Standard Deviation Bands Overlay -->
        <g v-if="showPeBands && valuationBands && peBandCoords" class="pe-bands-group">
          <!-- P/E Accumulation Shaded Band between -1SD and -2SD -->
          <rect
            v-if="peBandCoords.minus1sdY !== null && peBandCoords.minus2sdY !== null"
            :x="padding.left"
            :y="Math.min(peBandCoords.minus1sdY, peBandCoords.minus2sdY)"
            :width="plotWidth"
            :height="Math.abs(peBandCoords.minus2sdY - peBandCoords.minus1sdY)"
            fill="url(#peAccumGradient)"
          />

          <!-- +2 SD P/E Line -->
          <g v-if="peBandCoords.plus2sdY !== null">
            <line
              :x1="padding.left"
              :y1="peBandCoords.plus2sdY"
              :x2="padding.left + plotWidth"
              :y2="peBandCoords.plus2sdY"
              stroke="#f87171"
              stroke-dasharray="4 4"
              stroke-width="1.2"
              stroke-opacity="0.8"
            />
            <text
              :x="padding.left + plotWidth - 8"
              :y="peBandCoords.plus2sdY - 4"
              fill="#f87171"
              font-size="8.5"
              font-weight="600"
              font-family="var(--font-mono)"
              text-anchor="end"
            >
              +2σ P/E (Rp {{ formatNum(valuationBands.plus_2sd_price_pe) }} | {{ valuationBands.plus_2sd_pe.toFixed(1) }}x)
            </text>
          </g>

          <!-- +1 SD P/E Line -->
          <g v-if="peBandCoords.plus1sdY !== null">
            <line
              :x1="padding.left"
              :y1="peBandCoords.plus1sdY"
              :x2="padding.left + plotWidth"
              :y2="peBandCoords.plus1sdY"
              stroke="#fbbf24"
              stroke-dasharray="4 4"
              stroke-width="1.2"
              stroke-opacity="0.8"
            />
            <text
              :x="padding.left + plotWidth - 8"
              :y="peBandCoords.plus1sdY - 4"
              fill="#fbbf24"
              font-size="8.5"
              font-weight="600"
              font-family="var(--font-mono)"
              text-anchor="end"
            >
              +1σ P/E (Rp {{ formatNum(valuationBands.plus_1sd_price_pe) }} | {{ valuationBands.plus_1sd_pe.toFixed(1) }}x)
            </text>
          </g>

          <!-- Mean P/E Line -->
          <g v-if="peBandCoords.meanY !== null">
            <line
              :x1="padding.left"
              :y1="peBandCoords.meanY"
              :x2="padding.left + plotWidth"
              :y2="peBandCoords.meanY"
              stroke="#38bdf8"
              stroke-dasharray="6 3"
              stroke-width="1.5"
              stroke-opacity="0.9"
            />
            <text
              :x="padding.left + plotWidth - 8"
              :y="peBandCoords.meanY - 4"
              fill="#38bdf8"
              font-size="8.5"
              font-weight="700"
              font-family="var(--font-mono)"
              text-anchor="end"
            >
              Mean P/E (Rp {{ formatNum(valuationBands.mean_price_pe) }} | {{ valuationBands.mean_pe.toFixed(1) }}x)
            </text>
          </g>

          <!-- -1 SD P/E Line -->
          <g v-if="peBandCoords.minus1sdY !== null">
            <line
              :x1="padding.left"
              :y1="peBandCoords.minus1sdY"
              :x2="padding.left + plotWidth"
              :y2="peBandCoords.minus1sdY"
              stroke="#34d399"
              stroke-dasharray="3 3"
              stroke-width="1.4"
              stroke-opacity="0.85"
            />
            <text
              :x="padding.left + plotWidth - 8"
              :y="peBandCoords.minus1sdY - 4"
              fill="#34d399"
              font-size="8.5"
              font-weight="600"
              font-family="var(--font-mono)"
              text-anchor="end"
            >
              -1σ P/E (Rp {{ formatNum(valuationBands.minus_1sd_price_pe) }} | {{ valuationBands.minus_1sd_pe.toFixed(1) }}x)
            </text>
          </g>

          <!-- -2 SD P/E Line -->
          <g v-if="peBandCoords.minus2sdY !== null">
            <line
              :x1="padding.left"
              :y1="peBandCoords.minus2sdY"
              :x2="padding.left + plotWidth"
              :y2="peBandCoords.minus2sdY"
              stroke="#10b981"
              stroke-width="1.8"
            />
            <text
              :x="padding.left + plotWidth - 8"
              :y="peBandCoords.minus2sdY - 4"
              fill="#10b981"
              font-size="8.5"
              font-weight="700"
              font-family="var(--font-mono)"
              text-anchor="end"
            >
              -2σ P/E Accumulation (Rp {{ formatNum(valuationBands.minus_2sd_price_pe) }} | {{ valuationBands.minus_2sd_pe.toFixed(1) }}x)
            </text>
          </g>
        </g>

        <!-- 5. P/B Standard Deviation Bands Overlay -->
        <g v-if="showPbBands && valuationBands && pbBandCoords" class="pb-bands-group">
          <!-- P/B Accumulation Shaded Band between -1SD and -2SD -->
          <rect
            v-if="pbBandCoords.minus1sdY !== null && pbBandCoords.minus2sdY !== null"
            :x="padding.left"
            :y="Math.min(pbBandCoords.minus1sdY, pbBandCoords.minus2sdY)"
            :width="plotWidth"
            :height="Math.abs(pbBandCoords.minus2sdY - pbBandCoords.minus1sdY)"
            fill="url(#pbAccumGradient)"
          />

          <!-- Mean P/B Line -->
          <g v-if="pbBandCoords.meanY !== null">
            <line
              :x1="padding.left"
              :y1="pbBandCoords.meanY"
              :x2="padding.left + plotWidth"
              :y2="pbBandCoords.meanY"
              stroke="#60a5fa"
              stroke-dasharray="5 3"
              stroke-width="1.3"
              stroke-opacity="0.8"
            />
            <text
              :x="padding.left + 8"
              :y="pbBandCoords.meanY - 4"
              fill="#60a5fa"
              font-size="8.5"
              font-weight="600"
              font-family="var(--font-mono)"
            >
              Mean P/B (Rp {{ formatNum(valuationBands.mean_price_pb) }} | {{ valuationBands.mean_pb.toFixed(2) }}x)
            </text>
          </g>

          <!-- -1 SD P/B Line -->
          <g v-if="pbBandCoords.minus1sdY !== null">
            <line
              :x1="padding.left"
              :y1="pbBandCoords.minus1sdY"
              :x2="padding.left + plotWidth"
              :y2="pbBandCoords.minus1sdY"
              stroke="#38bdf8"
              stroke-dasharray="3 3"
              stroke-width="1.2"
            />
            <text
              :x="padding.left + 8"
              :y="pbBandCoords.minus1sdY - 4"
              fill="#38bdf8"
              font-size="8.5"
              font-family="var(--font-mono)"
            >
              -1σ P/B (Rp {{ formatNum(valuationBands.minus_1sd_price_pb) }})
            </text>
          </g>

          <!-- -2 SD P/B Line -->
          <g v-if="pbBandCoords.minus2sdY !== null">
            <line
              :x1="padding.left"
              :y1="pbBandCoords.minus2sdY"
              :x2="padding.left + plotWidth"
              :y2="pbBandCoords.minus2sdY"
              stroke="#0284c7"
              stroke-width="1.6"
            />
            <text
              :x="padding.left + 8"
              :y="pbBandCoords.minus2sdY - 4"
              fill="#0284c7"
              font-size="8.5"
              font-weight="700"
              font-family="var(--font-mono)"
            >
              -2σ P/B (Rp {{ formatNum(valuationBands.minus_2sd_price_pb) }})
            </text>
          </g>
        </g>

        <!-- 6. SMA 200 Curve -->
        <path
          v-if="showSma200 && sma200Path"
          :d="sma200Path"
          fill="none"
          stroke="#c084fc"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
          opacity="0.85"
        />

        <!-- 7. SMA 50 Curve -->
        <path
          v-if="showSma50 && sma50Path"
          :d="sma50Path"
          fill="none"
          stroke="#fbbf24"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
          opacity="0.9"
        />

        <!-- 8. Stock Price Area & Curve -->
        <path
          v-if="priceAreaPath"
          :d="priceAreaPath"
          fill="url(#priceGradient)"
        />
        <path
          v-if="priceLinePath"
          :d="priceLinePath"
          fill="none"
          stroke="#38bdf8"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />

        <!-- 9. Interactive Crosshair & Highlight Dot -->
        <g v-if="hoveredIndex !== null && hoveredPoint" class="crosshair-group">
          <!-- Vertical Crosshair Line -->
          <line
            :x1="hoverX"
            :y1="padding.top"
            :x2="hoverX"
            :y2="padding.top + plotHeight"
            stroke="#94a3b8"
            stroke-dasharray="3 3"
            stroke-width="1"
            stroke-opacity="0.7"
          />
          <!-- Horizontal Crosshair Line -->
          <line
            :x1="padding.left"
            :y1="hoverY"
            :x2="padding.left + plotWidth"
            :y2="hoverY"
            stroke="#94a3b8"
            stroke-dasharray="3 3"
            stroke-width="1"
            stroke-opacity="0.5"
          />
          <!-- Highlight Glow & Circle on Price Line -->
          <circle
            :cx="hoverX"
            :cy="hoverY"
            r="5"
            fill="#38bdf8"
            stroke="#080c14"
            stroke-width="2"
            filter="url(#glow)"
          />
          <!-- Glowing Outer Pulse Ring -->
          <circle
            :cx="hoverX"
            :cy="hoverY"
            r="8"
            fill="none"
            stroke="#38bdf8"
            stroke-width="1"
            stroke-opacity="0.6"
          />
        </g>
      </svg>

      <!-- Floating Terminal Hover Tooltip -->
      <div
        v-if="hoveredIndex !== null && hoveredPoint"
        class="chart-tooltip font-mono"
        :style="tooltipStyle"
      >
        <div class="tooltip-header">
          <span class="tooltip-date">📅 {{ formatFullDate(hoveredPoint.date) }}</span>
          <span v-if="hoverChangePct !== null" :class="['tooltip-badge', hoverChangePct >= 0 ? 'bullish' : 'bearish']">
            {{ (hoverChangePct >= 0 ? '+' : '') + hoverChangePct.toFixed(2) }}%
          </span>
        </div>

        <div class="tooltip-grid">
          <div class="tooltip-row highlight">
            <span class="lbl">Close:</span>
            <span class="val text-cyan">Rp {{ formatNum(hoveredPoint.close) }}</span>
          </div>
          <div class="tooltip-row">
            <span class="lbl">Open / High / Low:</span>
            <span class="val">
              {{ formatNum(hoveredPoint.open) }} / {{ formatNum(hoveredPoint.high) }} / {{ formatNum(hoveredPoint.low) }}
            </span>
          </div>
          <div class="tooltip-row">
            <span class="lbl">Volume:</span>
            <span class="val">{{ formatVolume(hoveredPoint.volume) }}</span>
          </div>

          <div v-if="grahamNumber && grahamNumber > 0" class="tooltip-row val-row">
            <span class="lbl">Graham Fair Value:</span>
            <span class="val text-green">Rp {{ formatNum(grahamNumber) }}</span>
          </div>
          <div v-if="grahamNumber && grahamNumber > 0 && hoverMosPct !== null" class="tooltip-row val-row">
            <span class="lbl">Margin of Safety:</span>
            <span :class="['val font-bold', hoverMosPct >= 30 ? 'text-green' : hoverMosPct > 0 ? 'text-amber' : 'text-red']">
              {{ (hoverMosPct >= 0 ? '+' : '') + hoverMosPct.toFixed(1) }}%
              <span class="sub-lbl">{{ hoverMosPct >= 30 ? '(Deep MoS)' : hoverMosPct > 0 ? '(Fair/MoS)' : '(Overvalued)' }}</span>
            </span>
          </div>

          <!-- P/E Band Position in Tooltip -->
          <div v-if="valuationBands && valuationBands.minus_2sd_price_pe > 0" class="tooltip-row val-row pe-band-tooltip-row">
            <span class="lbl">P/E Band Zone:</span>
            <span :class="['val font-bold', getPeZoneClass(hoveredPoint.close)]">
              {{ getPeZoneLabel(hoveredPoint.close) }}
            </span>
          </div>

          <div v-if="showSma50 && hoveredSma50 !== null" class="tooltip-row sma-row">
            <span class="lbl sma50-text">SMA 50:</span>
            <span class="val">Rp {{ formatNum(hoveredSma50) }}</span>
          </div>
          <div v-if="showSma200 && hoveredSma200 !== null" class="tooltip-row sma-row">
            <span class="lbl sma200-text">SMA 200:</span>
            <span class="val">Rp {{ formatNum(hoveredSma200) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { PriceCandle, StockPriceResponse, ValuationBands, TimingSignal } from '../server/utils/types'

const props = defineProps<{
  ticker: string | null
  grahamNumber?: number | null
  currentPrice?: number | null
  valuationBands?: ValuationBands | null
  timingSignal?: TimingSignal | null
}>()

const selectedRange = ref<string>('1y')
const loading = ref<boolean>(false)
const prices = ref<PriceCandle[]>([])

const showSma50 = ref<boolean>(true)
const showSma200 = ref<boolean>(true)
const showPeBands = ref<boolean>(true)
const showPbBands = ref<boolean>(false)

const hoveredIndex = ref<number | null>(null)
const chartContainerRef = ref<HTMLElement | null>(null)
const svgRef = ref<SVGSVGElement | null>(null)

const rangeOptions = [
  { id: '1m', label: '1M' },
  { id: '6m', label: '6M' },
  { id: '1y', label: '1Y' },
  { id: '3y', label: '3Y' },
  { id: '5y', label: '5Y' },
  { id: 'max', label: 'MAX' }
]

// Availability flags
const hasPeBands = computed(() => {
  return !!(props.valuationBands && props.valuationBands.mean_price_pe > 0)
})

const hasPbBands = computed(() => {
  return !!(props.valuationBands && props.valuationBands.mean_price_pb > 0)
})

// SVG Layout Dimensions
const svgWidth = 800
const svgHeight = 340
const padding = { top: 25, right: 65, bottom: 35, left: 15 }
const plotWidth = svgWidth - padding.left - padding.right // 720
const plotHeight = svgHeight - padding.top - padding.bottom // 280

const setRange = (r: string) => {
  if (selectedRange.value === r) return
  selectedRange.value = r
  if (props.ticker) {
    fetchPriceHistory(props.ticker, r)
  }
}

const fetchPriceHistory = async (t: string, range: string) => {
  if (!t) return
  loading.value = true
  hoveredIndex.value = null
  try {
    const res = await $fetch<StockPriceResponse>(`/api/v1/stocks/${t}/prices?range=${range}`)
    if (res && Array.isArray(res.prices)) {
      // Ensure sorted chronologically ascending
      prices.value = res.prices.slice().sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())
    } else {
      prices.value = []
    }
  } catch (err) {
    console.error(`Failed to fetch prices for ${t}:`, err)
    prices.value = []
  } finally {
    loading.value = false
  }
}

watch(
  () => props.ticker,
  (newTicker) => {
    if (newTicker) {
      fetchPriceHistory(newTicker, selectedRange.value)
    } else {
      prices.value = []
    }
  },
  { immediate: true }
)

// Computed Moving Averages
const sma50Values = computed(() => computeSMA(prices.value, 50))
const sma200Values = computed(() => computeSMA(prices.value, 200))

function computeSMA(data: PriceCandle[], period: number): (number | null)[] {
  const result: (number | null)[] = []
  let sum = 0
  for (let i = 0; i < data.length; i++) {
    sum += data[i].close
    if (i >= period) {
      sum -= data[i - period].close
    }
    if (i >= period - 1) {
      result.push(sum / period)
    } else {
      result.push(null)
    }
  }
  return result
}

// Coordinate Scaling Calculations
const yBounds = computed(() => {
  if (prices.value.length === 0) {
    return { min: 0, max: 100 }
  }

  let min = Infinity
  let max = -Infinity

  for (const p of prices.value) {
    if (p.close < min) min = p.close
    if (p.close > max) max = p.close
  }

  // Factor in Graham Number and MoS line if present and realistic (0.1x to 10x range)
  if (props.grahamNumber && props.grahamNumber > 0) {
    const gVal = props.grahamNumber
    const mosVal = gVal * 0.70
    if (gVal < max * 4 && gVal > min * 0.1) {
      if (gVal > max) max = gVal
      if (mosVal < min) min = mosVal
    }
  }

  // Factor in P/E Standard Deviation Bands if active
  if (showPeBands.value && props.valuationBands) {
    const vb = props.valuationBands
    const pePrices = [
      vb.plus_2sd_price_pe,
      vb.plus_1sd_price_pe,
      vb.mean_price_pe,
      vb.minus_1sd_price_pe,
      vb.minus_2sd_price_pe
    ].filter(p => p && p > 0 && p < max * 5 && p > min * 0.1)

    for (const p of pePrices) {
      if (p < min) min = p
      if (p > max) max = p
    }
  }

  // Factor in P/B Standard Deviation Bands if active
  if (showPbBands.value && props.valuationBands) {
    const vb = props.valuationBands
    const pbPrices = [
      vb.plus_2sd_price_pb,
      vb.plus_1sd_price_pb,
      vb.mean_price_pb,
      vb.minus_1sd_price_pb,
      vb.minus_2sd_price_pb
    ].filter(p => p && p > 0 && p < max * 5 && p > min * 0.1)

    for (const p of pbPrices) {
      if (p < min) min = p
      if (p > max) max = p
    }
  }

  // Factor in visible SMAs
  if (showSma50.value) {
    for (const val of sma50Values.value) {
      if (val !== null) {
        if (val < min) min = val
        if (val > max) max = val
      }
    }
  }
  if (showSma200.value) {
    for (const val of sma200Values.value) {
      if (val !== null) {
        if (val < min) min = val
        if (val > max) max = val
      }
    }
  }

  if (min === Infinity || max === -Infinity || min === max) {
    min = Math.max(0, min - 100)
    max = max + 100
  }

  const span = max - min
  const paddedMin = Math.max(0, min - span * 0.06)
  const paddedMax = max + span * 0.06

  return { min: paddedMin, max: paddedMax }
})

const getX = (index: number): number => {
  if (prices.value.length <= 1) return padding.left + plotWidth / 2
  return padding.left + (index / (prices.value.length - 1)) * plotWidth
}

const getY = (val: number): number => {
  const { min, max } = yBounds.value
  if (max === min) return padding.top + plotHeight / 2
  const ratio = (val - min) / (max - min)
  return padding.top + plotHeight - ratio * plotHeight
}

// Paths
const priceLinePath = computed(() => {
  if (prices.value.length === 0) return ''
  return prices.value
    .map((p, i) => `${i === 0 ? 'M' : 'L'} ${getX(i).toFixed(1)},${getY(p.close).toFixed(1)}`)
    .join(' ')
})

const priceAreaPath = computed(() => {
  if (prices.value.length === 0) return ''
  const firstX = getX(0).toFixed(1)
  const lastX = getX(prices.value.length - 1).toFixed(1)
  const bottomY = (padding.top + plotHeight).toFixed(1)
  const linePart = prices.value
    .map((p, i) => `${i === 0 ? 'M' : 'L'} ${getX(i).toFixed(1)},${getY(p.close).toFixed(1)}`)
    .join(' ')
  return `${linePart} L ${lastX},${bottomY} L ${firstX},${bottomY} Z`
})

const sma50Path = computed(() => {
  return buildLinePath(sma50Values.value)
})

const sma200Path = computed(() => {
  return buildLinePath(sma200Values.value)
})

function buildLinePath(data: (number | null)[]): string {
  let path = ''
  let inSegment = false
  for (let i = 0; i < data.length; i++) {
    const val = data[i]
    if (val !== null && !isNaN(val)) {
      const x = getX(i)
      const y = getY(val)
      if (!inSegment) {
        path += `M ${x.toFixed(1)},${y.toFixed(1)}`
        inSegment = true
      } else {
        path += ` L ${x.toFixed(1)},${y.toFixed(1)}`
      }
    } else {
      inSegment = false
    }
  }
  return path
}

// Graham Fair Value Coordinates
const grahamY = computed(() => {
  if (!props.grahamNumber || props.grahamNumber <= 0) return null
  return getY(props.grahamNumber)
})

const mosY = computed(() => {
  if (!props.grahamNumber || props.grahamNumber <= 0) return null
  return getY(props.grahamNumber * 0.70)
})

// P/E Valuation Band Y Coordinates
const peBandCoords = computed(() => {
  if (!props.valuationBands) return null
  const vb = props.valuationBands
  return {
    plus2sdY: vb.plus_2sd_price_pe > 0 ? getY(vb.plus_2sd_price_pe) : null,
    plus1sdY: vb.plus_1sd_price_pe > 0 ? getY(vb.plus_1sd_price_pe) : null,
    meanY: vb.mean_price_pe > 0 ? getY(vb.mean_price_pe) : null,
    minus1sdY: vb.minus_1sd_price_pe > 0 ? getY(vb.minus_1sd_price_pe) : null,
    minus2sdY: vb.minus_2sd_price_pe > 0 ? getY(vb.minus_2sd_price_pe) : null
  }
})

// P/B Valuation Band Y Coordinates
const pbBandCoords = computed(() => {
  if (!props.valuationBands) return null
  const vb = props.valuationBands
  return {
    plus2sdY: vb.plus_2sd_price_pb > 0 ? getY(vb.plus_2sd_price_pb) : null,
    plus1sdY: vb.plus_1sd_price_pb > 0 ? getY(vb.plus_1sd_price_pb) : null,
    meanY: vb.mean_price_pb > 0 ? getY(vb.mean_price_pb) : null,
    minus1sdY: vb.minus_1sd_price_pb > 0 ? getY(vb.minus_1sd_price_pb) : null,
    minus2sdY: vb.minus_2sd_price_pb > 0 ? getY(vb.minus_2sd_price_pb) : null
  }
})

// Grid Lines & Ticks
const yGridLines = computed(() => {
  const { min, max } = yBounds.value
  const steps = 4
  const lines = []
  for (let i = 0; i <= steps; i++) {
    const val = min + (i / steps) * (max - min)
    lines.push({
      val,
      y: getY(val)
    })
  }
  return lines
})

const xGridTicks = computed(() => {
  if (prices.value.length === 0) return []
  const count = Math.min(5, prices.value.length)
  if (count <= 1) {
    return [{ x: getX(0), label: formatTickDate(prices.value[0].date) }]
  }

  const ticks = []
  const step = (prices.value.length - 1) / (count - 1)
  for (let i = 0; i < count; i++) {
    const idx = Math.round(i * step)
    const p = prices.value[idx]
    if (p) {
      ticks.push({
        x: getX(idx),
        label: formatTickDate(p.date)
      })
    }
  }
  return ticks
})

// Summary Stats
const latestPrice = computed(() => {
  if (prices.value.length === 0) return null
  return prices.value[prices.value.length - 1]
})

const periodChangePct = computed(() => {
  if (prices.value.length < 2) return null
  const first = prices.value[0].close
  const last = prices.value[prices.value.length - 1].close
  if (first === 0) return null
  return ((last - first) / first) * 100
})

const currentMosPct = computed(() => {
  if (!props.grahamNumber || props.grahamNumber <= 0 || !latestPrice.value) return null
  const close = latestPrice.value.close
  return ((props.grahamNumber - close) / props.grahamNumber) * 100
})

// Timing Helpers
const getTimingScoreClass = (score: number) => {
  if (score >= 70) return 'text-green'
  if (score >= 50) return 'text-amber'
  return 'text-muted'
}

const getTimingBannerClass = (score: number) => {
  if (score >= 70) return 'banner-actionable'
  if (score >= 50) return 'banner-accumulation'
  return 'banner-neutral'
}

const getPeZoneClass = (price: number) => {
  if (!props.valuationBands) return ''
  const vb = props.valuationBands
  if (vb.minus_2sd_price_pe > 0 && price <= vb.minus_2sd_price_pe) return 'text-green'
  if (vb.minus_1sd_price_pe > 0 && price <= vb.minus_1sd_price_pe) return 'text-green'
  if (vb.mean_price_pe > 0 && price <= vb.mean_price_pe) return 'text-cyan'
  if (vb.plus_1sd_price_pe > 0 && price <= vb.plus_1sd_price_pe) return 'text-amber'
  return 'text-red'
}

const getPeZoneLabel = (price: number) => {
  if (!props.valuationBands) return '-'
  const vb = props.valuationBands
  if (vb.minus_2sd_price_pe > 0 && price <= vb.minus_2sd_price_pe) return '≤ -2σ (Deep Value)'
  if (vb.minus_1sd_price_pe > 0 && price <= vb.minus_1sd_price_pe) return '-1σ to -2σ (Accumulation)'
  if (vb.mean_price_pe > 0 && price <= vb.mean_price_pe) return 'Mean to -1σ (Discount)'
  if (vb.plus_1sd_price_pe > 0 && price <= vb.plus_1sd_price_pe) return 'Mean to +1σ (Fair/High)'
  if (vb.plus_2sd_price_pe > 0 && price <= vb.plus_2sd_price_pe) return '+1σ to +2σ (Premium)'
  return '> +2σ (Overextended)'
}

// Hover / Crosshair Interactions
const hoveredPoint = computed(() => {
  if (hoveredIndex.value === null || !prices.value[hoveredIndex.value]) return null
  return prices.value[hoveredIndex.value]
})

const hoverX = computed(() => {
  if (hoveredIndex.value === null) return 0
  return getX(hoveredIndex.value)
})

const hoverY = computed(() => {
  if (!hoveredPoint.value) return 0
  return getY(hoveredPoint.value.close)
})

const hoverChangePct = computed(() => {
  if (hoveredIndex.value === null || hoveredIndex.value === 0) return null
  const curr = prices.value[hoveredIndex.value]?.close
  const prev = prices.value[hoveredIndex.value - 1]?.close
  if (!curr || !prev) return null
  return ((curr - prev) / prev) * 100
})

const hoverMosPct = computed(() => {
  if (!props.grahamNumber || props.grahamNumber <= 0 || !hoveredPoint.value) return null
  const close = hoveredPoint.value.close
  return ((props.grahamNumber - close) / props.grahamNumber) * 100
})

const hoveredSma50 = computed(() => {
  if (hoveredIndex.value === null) return null
  return sma50Values.value[hoveredIndex.value] || null
})

const hoveredSma200 = computed(() => {
  if (hoveredIndex.value === null) return null
  return sma200Values.value[hoveredIndex.value] || null
})

const tooltipStyle = computed(() => {
  if (hoveredIndex.value === null || !chartContainerRef.value) return { display: 'none' }
  const total = prices.value.length
  if (total === 0) return { display: 'none' }

  // Flip tooltip side if cursor is in right half
  const isRightHalf = hoveredIndex.value > total / 2
  const xRatio = (hoveredIndex.value / (total - 1)) * 100

  return {
    left: isRightHalf ? `${Math.max(10, xRatio - 36)}%` : `${Math.min(65, xRatio + 4)}%`,
    top: '20px'
  }
})

const handleMouseMove = (event: MouseEvent) => {
  if (!svgRef.value || prices.value.length === 0) return
  const rect = svgRef.value.getBoundingClientRect()
  const mouseX = event.clientX - rect.left
  const relativeX = (mouseX / rect.width) * svgWidth

  // Clamp inside plot area
  const clampedX = Math.max(padding.left, Math.min(padding.left + plotWidth, relativeX))
  const normRatio = (clampedX - padding.left) / plotWidth
  const idx = Math.round(normRatio * (prices.value.length - 1))
  hoveredIndex.value = Math.max(0, Math.min(prices.value.length - 1, idx))
}

const handleTouchMove = (event: TouchEvent) => {
  if (!svgRef.value || prices.value.length === 0 || event.touches.length === 0) return
  const touch = event.touches[0]
  const rect = svgRef.value.getBoundingClientRect()
  const mouseX = touch.clientX - rect.left
  const relativeX = (mouseX / rect.width) * svgWidth

  const clampedX = Math.max(padding.left, Math.min(padding.left + plotWidth, relativeX))
  const normRatio = (clampedX - padding.left) / plotWidth
  const idx = Math.round(normRatio * (prices.value.length - 1))
  hoveredIndex.value = Math.max(0, Math.min(prices.value.length - 1, idx))
}

const handleMouseLeave = () => {
  hoveredIndex.value = null
}

// Formatters
const formatNum = (val?: number) => {
  if (val === undefined || val === null || isNaN(val)) return '-'
  return new Intl.NumberFormat('id-ID', { maximumFractionDigits: 0 }).format(val)
}

const formatCompact = (val?: number) => {
  if (val === undefined || val === null || isNaN(val)) return '-'
  if (Math.abs(val) >= 1e6) return (val / 1e6).toFixed(1) + 'M'
  if (Math.abs(val) >= 1e3) return (val / 1e3).toFixed(1) + 'k'
  return val.toFixed(0)
}

const formatVolume = (vol?: number) => {
  if (vol === undefined || vol === null || isNaN(vol)) return '-'
  if (vol >= 1e9) return (vol / 1e9).toFixed(2) + ' B shares'
  if (vol >= 1e6) return (vol / 1e6).toFixed(2) + ' M shares'
  if (vol >= 1e3) return (vol / 1e3).toFixed(1) + ' K shares'
  return vol.toLocaleString('id-ID') + ' shares'
}

const formatTickDate = (d?: string | Date) => {
  if (!d) return ''
  const dt = new Date(d)
  if (selectedRange.value === '1m' || selectedRange.value === '6m') {
    return dt.toLocaleDateString('en-GB', { day: '2-digit', month: 'short' })
  }
  return dt.toLocaleDateString('en-GB', { month: 'short', year: '2-digit' })
}

const formatFullDate = (d?: string | Date) => {
  if (!d) return ''
  const dt = new Date(d)
  return dt.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}
</script>

<style scoped>
.price-chart-card {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.chart-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.title-with-badge {
  display: flex;
  align-items: center;
  gap: 10px;
}

.chart-title {
  font-size: 0.85rem;
  font-weight: 700;
  color: #38bdf8;
  letter-spacing: 0.05em;
}

.ticker-pill {
  background: #2563eb;
  color: #fff;
  font-size: 0.75rem;
  font-weight: 800;
  padding: 1px 6px;
  border-radius: 4px;
}

.chart-summary-stats {
  display: flex;
  gap: 14px;
  flex-wrap: wrap;
  font-size: 0.78rem;
  align-items: center;
}

.stat-item {
  display: flex;
  gap: 5px;
  align-items: center;
}

.stat-lbl {
  color: var(--text-muted);
  font-size: 0.72rem;
}

.stat-val {
  font-weight: 700;
}

.timing-stat-item {
  background: rgba(255, 255, 255, 0.04);
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px solid var(--border-color);
}

.header-controls {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.indicator-toggles {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.toggle-btn {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  color: var(--text-secondary);
  font-size: 0.72rem;
  font-weight: 600;
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 5px;
  transition: all 0.15s ease;
}

.toggle-btn:hover {
  background: var(--bg-card-hover);
  color: #fff;
}

.toggle-btn.active {
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}

.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  opacity: 0.5;
}

.toggle-btn.active .dot {
  opacity: 1;
}

.pe-dot { background: #34d399; }
.pb-dot { background: #38bdf8; }
.sma50-dot { background: #fbbf24; }
.sma200-dot { background: #c084fc; }

.range-selector {
  display: flex;
  gap: 2px;
  background: var(--bg-card);
  padding: 2px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

.range-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 0.72rem;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.range-btn:hover {
  color: #fff;
}

.range-btn.active {
  background: #2563eb;
  color: #fff;
}

/* Smart Timing Banner */
.smart-timing-banner {
  border-radius: 8px;
  padding: 10px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: all 0.2s ease;
}

.banner-actionable {
  background: rgba(16, 185, 129, 0.08);
  border: 1px solid rgba(16, 185, 129, 0.4);
  box-shadow: 0 0 12px rgba(16, 185, 129, 0.1);
}

.banner-accumulation {
  background: rgba(245, 158, 11, 0.08);
  border: 1px solid rgba(245, 158, 11, 0.4);
  box-shadow: 0 0 12px rgba(245, 158, 11, 0.1);
}

.banner-neutral {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
}

.timing-banner-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 10px;
}

.timing-score-badge {
  display: flex;
  align-items: center;
  gap: 8px;
}

.timing-indicator-dot {
  font-size: 0.85rem;
}

.score-number {
  font-weight: 800;
  font-size: 0.95rem;
  color: #f8fafc;
}

.timing-status-label {
  font-weight: 700;
  font-size: 0.82rem;
  color: #38bdf8;
  letter-spacing: 0.02em;
}

.timing-catalysts-list {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.catalyst-chip {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-secondary);
  font-size: 0.72rem;
  padding: 2px 7px;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.chip-bullish {
  background: rgba(16, 185, 129, 0.15);
  border-color: rgba(16, 185, 129, 0.4);
  color: #34d399;
}

.chip-amber {
  background: rgba(245, 158, 11, 0.12);
  border-color: rgba(245, 158, 11, 0.35);
  color: #fbbf24;
}

.chip-emerald {
  background: rgba(52, 211, 153, 0.12);
  border-color: rgba(52, 211, 153, 0.35);
  color: #6ee7b7;
}

.chip-extra {
  background: rgba(56, 189, 248, 0.1);
  border-color: rgba(56, 189, 248, 0.3);
  color: #7dd3fc;
}

.chip-alert {
  color: #fef08a;
  font-weight: 800;
}

/* Legend */
.chart-legend {
  display: flex;
  gap: 14px;
  flex-wrap: wrap;
  font-size: 0.72rem;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 8px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.legend-line {
  display: inline-block;
  width: 14px;
  height: 2px;
  border-radius: 1px;
}

.price-line { background: #38bdf8; }
.graham-line {
  background: #10b981;
  border-top: 1px dashed #10b981;
}
.pe-mean-line {
  background: #38bdf8;
  border-top: 1px dashed #38bdf8;
}
.pe-plus-line {
  background: #fbbf24;
  border-top: 1px dashed #f87171;
}
.pb-mean-line {
  background: #60a5fa;
  border-top: 1px dashed #60a5fa;
}
.sma50-line { background: #fbbf24; }
.sma200-line { background: #c084fc; }

.legend-box.mos-box {
  display: inline-block;
  width: 12px;
  height: 10px;
  background: rgba(16, 185, 129, 0.25);
  border: 1px solid #10b981;
  border-radius: 2px;
}

.legend-box.pe-accum-box {
  display: inline-block;
  width: 12px;
  height: 10px;
  background: rgba(52, 211, 153, 0.25);
  border: 1px solid #34d399;
  border-radius: 2px;
}

.legend-box.pb-accum-box {
  display: inline-block;
  width: 12px;
  height: 10px;
  background: rgba(56, 189, 248, 0.22);
  border: 1px solid #38bdf8;
  border-radius: 2px;
}

.chart-stage {
  position: relative;
  width: 100%;
  height: 320px;
}

.chart-svg {
  width: 100%;
  height: 100%;
  display: block;
  user-select: none;
}

.chart-overlay-state {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-muted);
  font-size: 0.85rem;
  background: rgba(8, 12, 20, 0.85);
  z-index: 10;
}

.spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--border-color);
  border-top-color: #38bdf8;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.chart-tooltip {
  position: absolute;
  background: rgba(15, 23, 42, 0.95);
  backdrop-filter: blur(8px);
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  padding: 10px 14px;
  width: 270px;
  pointer-events: none;
  z-index: 20;
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.6);
  display: flex;
  flex-direction: column;
  gap: 6px;
  transition: left 0.08s ease-out, top 0.08s ease-out;
}

.tooltip-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 4px;
  font-size: 0.72rem;
}

.tooltip-date {
  color: #f8fafc;
  font-weight: 700;
}

.tooltip-badge {
  font-size: 0.7rem;
  font-weight: 800;
  padding: 1px 4px;
  border-radius: 3px;
}

.tooltip-badge.bullish {
  background: var(--bullish-bg);
  color: var(--bullish-text);
}

.tooltip-badge.bearish {
  background: var(--bearish-bg);
  color: var(--bearish-text);
}

.tooltip-grid {
  display: flex;
  flex-direction: column;
  gap: 3px;
  font-size: 0.73rem;
}

.tooltip-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.tooltip-row .lbl {
  color: var(--text-secondary);
}

.tooltip-row .val {
  color: #f8fafc;
  font-weight: 600;
}

.tooltip-row.highlight {
  font-size: 0.82rem;
  border-bottom: 1px dashed var(--border-color);
  padding-bottom: 3px;
  margin-bottom: 2px;
}

.tooltip-row.val-row {
  border-top: 1px solid var(--border-color);
  padding-top: 3px;
  margin-top: 2px;
}

.pe-band-tooltip-row {
  background: rgba(56, 189, 248, 0.04);
  padding: 2px 4px;
  border-radius: 3px;
}

.sub-lbl {
  font-size: 0.65rem;
  opacity: 0.85;
}

.sma50-text { color: #fbbf24; }
.sma200-text { color: #c084fc; }

.text-cyan { color: #38bdf8; }
.text-green { color: #34d399; }
.text-amber { color: #fbbf24; }
.text-red { color: #f87171; }
.font-bold { font-weight: 700; }

@media (max-width: 640px) {
  .chart-header-row {
    flex-direction: column;
    align-items: flex-start;
  }
  .header-controls {
    width: 100%;
    justify-content: space-between;
  }
  .timing-banner-main {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
