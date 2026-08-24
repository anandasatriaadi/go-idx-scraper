# Design Specification: IDX Intelligence Terminal UI (Dark Bloomberg-Inspired Dashboard)

**Date**: 2026-08-24  
**Status**: Approved  

---

## 1. Overview & Goals

This specification defines the frontend user interface for `idx-web`, transforming the Nuxt 4 application into a sleek, high-contrast, dark-themed financial intelligence terminal.

Key deliverables:
1. **Dark Terminal Design System**: Deep slate background (`#080c14`), card background (`#0f172a`), border styling (`#1e293b`), emerald green bullish accents (`#10b981`), crimson red bearish accents (`#ef4444`), and cyber sky-blue accents (`#38bdf8`).
2. **Top Navigation Bar**: Brand identity, active tab switcher, global quick ticker search, and Firebase Auth profile / sign-in modal trigger.
3. **Unified Executive Overview**:
   - Top Hero Widget showcasing the latest 7 AM Daily Briefing Macro Pulse, Bullish Opportunities, Bearish Risk Alerts, and Today's Value Investor Action Plan.
   - Real-time News Stream with quick sector chips and instant filtering.
4. **Dedicated Dedicated Views**:
   - **Daily Briefings View**: Full interactive reader with historical date selector and Markdown view.
   - **News Terminal View**: Dense data table / cards filterable by ticker symbol, industry sector, value score ($-10$ to $+10$), and search query.
   - **Announcements View**: Real-time official IDX disclosure feed with watchlist highlight badges (`is_watched`).
   - **Financial Reports View**: Searchable quarterly statement archive with Excel download triggers.
5. **Interactive Modals & Composables**:
   - **Article Reader Modal**: Renders clean Markdown content, Reporter/Editor metadata, Value Investing takeaway callout, and external source link.
   - **Firebase Auth Composable & Modal**: Supports Google Sign-In and Email login; syncs user watchlist with `/api/v1/user/watchlist`.

---

## 2. Component Architecture

```
idx-web/src/
├── app.vue                   # Root shell with global layout, Navbar, and Modals
├── assets/
│   └── main.css              # Dark terminal CSS variables, custom scrollbars, animations
├── composables/
│   ├── useAuth.ts            # Firebase client auth state, ID token management, login/logout
│   └── useWatchlist.ts       # Watchlist sync with /api/v1/user/watchlist
├── components/
│   ├── Navbar.vue            # Top navigation bar, ticker search, auth button
│   ├── OverviewView.vue      # Default dashboard: Briefing Hero + Live News Grid
│   ├── BriefingView.vue      # Historical briefing reader & Markdown viewer
│   ├── NewsTerminalView.vue  # Filterable multi-channel news terminal
│   ├── AnnouncementsView.vue # Official IDX disclosures with watchlist tags
│   ├── FinReportsView.vue    # Financial reports search & download table
│   ├── ArticleModal.vue      # Full article markdown reader modal
│   └── AuthModal.vue         # Firebase login/signup dialog
└── pages/
    └── index.vue             # Single-page terminal controller switching views smoothly
```

---

## 3. Visual Design System & Palette

```css
:root {
  --bg-app: #080c14;
  --bg-card: #0f172a;
  --bg-card-hover: #1e293b;
  --border-color: #1e293b;
  --border-subtle: #334155;
  
  --text-primary: #f8fafc;
  --text-secondary: #94a3b8;
  --text-muted: #64748b;
  
  --bullish-bg: rgba(16, 185, 129, 0.12);
  --bullish-border: #10b981;
  --bullish-text: #34d399;
  
  --bearish-bg: rgba(239, 68, 68, 0.12);
  --bearish-border: #ef4444;
  --bearish-text: #f87171;
  
  --neutral-bg: rgba(56, 189, 248, 0.12);
  --neutral-border: #38bdf8;
  --neutral-text: #7dd3fc;
  
  --accent-amber: #f59e0b;
}
```

---

## 4. Feature View Specifications

### 4.1 Overview View
- **Hero Briefing Card**:
  - Displays `title`, `date`, and `macro_pulse`.
  - **Bullish Lookout Column**: Cards with ticker pills, headline, rationale, value score badge (`+1` to `+10`), and investment takeaway.
  - **Bearish Lookout Column**: Cards with ticker pills, headline, rationale, value score badge (`-1` to `-10`), and risk warnings.
  - **Action Plan Callout**: Bold, highlighted box with Graham-Buffett-Munger actionable guidance for today's market.
- **News Stream Grid**:
  - Quick filter chips (`All`, `🟢 Bullish`, `🔴 Bearish`, `Banking`, `Poultry`, `Mining`, `Energy`, `Consumer`).
  - Search input with debounce.
  - Interactive cards triggering `ArticleModal`.

### 4.2 News Terminal View
- Left/Top filters: Ticker search, Industry dropdown, Direction filter, Date range.
- Card & Table toggle for dense terminal inspection.
- Ticker pills clickable to instantly filter news by that company.

### 4.3 Daily Briefings View
- Date dropdown to switch between historical briefing dates.
- Interactive sections for Macro Pulse, Stock Lookouts, and Sector Highlights.
- Raw Markdown tab with Copy-to-Clipboard functionality.

### 4.4 Announcements & Financial Reports Views
- Fast client-side search across Issuer Code and Title.
- Direct external download links for IDX attachment files.
- Star button on each ticker to add/remove from Watchlist (authenticated via Firebase).

---

## 5. Composables & State Management

### 5.1 `useAuth.ts`
- Tracks current Firebase user, email, and ID token.
- Methods: `loginWithGoogle()`, `loginWithEmail(email, pass)`, `signupWithEmail(email, pass)`, `logout()`.
- Automatically attaches `Authorization: Bearer <ID_TOKEN>` to `$fetch` requests.

### 5.2 `useWatchlist.ts`
- Manages user's ticker array (e.g. `['BBRI', 'TLKM']`).
- Methods: `fetchWatchlist()`, `toggleWatchlist(ticker: string)`.
- Updates `/api/v1/user/watchlist` with optimistic UI updates.

---

## 6. Testing & Build Strategy

1. **Component Rendering & Build Verification**:
   - `npm run build` in `idx-web` to verify full SSR/client compilation.
2. **Endpoint Integration Testing**:
   - Verify `/api/v1/briefings/latest`, `/api/v1/news`, `/api/v1/announcements`, `/api/v1/financial-reports` populate frontend views cleanly.
3. **Responsive Testing**:
   - Ensure clean layout on Desktop (1440px+), Laptop (1024px), Tablet (768px), and Mobile (375px).
