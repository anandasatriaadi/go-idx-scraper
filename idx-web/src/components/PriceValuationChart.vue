<template>
  <div class="price-chart-card card-section font-mono">
    <!-- Header & Controls Bar -->
    <div class="chart-header-row">
      <div class="header-left">
        <div class="title-with-badge">
          <h2 class="chart-title">📈 PRICE & GRAHAM VALUATION BANDS</h2>
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
        </div>
      </div>

      <!-- Controls Right: Range Switcher & Indicator Toggles -->
      <div class="header-controls">
        <!-- Indicator Toggles -->
        <div class="indicator-toggles">
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

        <!-- 4. SMA 200 Curve -->
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

        <!-- 5. SMA 50 Curve -->
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

        <!-- 6. Stock Price Area & Curve -->
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

        <!-- 7. Interactive Crosshair & Highlight Dot -->
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
import { ref, computed, watch, onMounted } from 'vue'
import type { PriceCandle, StockPriceResponse } from '../server/utils/types'

const props = defineProps<{
  ticker: string | null
  grahamNumber?: number | null
  currentPrice?: number | null
}>()

const selectedRange = ref<string>('1y')
const loading = ref<boolean>(false)
const prices = ref<PriceCandle[]>([])

const showSma50 = ref<boolean>(true)
const showSma200 = ref<boolean>(true)

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

.header-controls {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.indicator-toggles {
  display: flex;
  gap: 6px;
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

.chart-legend {
  display: flex;
  gap: 16px;
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
  width: 260px;
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
}
</style>
