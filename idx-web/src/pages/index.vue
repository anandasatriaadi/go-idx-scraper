<template>
  <div style="font-family: system-ui, -apple-system, sans-serif; max-width: 800px; margin: 40px auto; padding: 20px; line-height: 1.6;">
    <h1 style="color: #0f172a;">IDX Intelligence & News API</h1>
    <p style="color: #475569;">
      Nuxt 4 / Nitro backend server for Indonesia Stock Exchange data collection and value investing intelligence.
    </p>

    <div style="margin-top: 24px; padding: 16px; background: #f8fafc; border-radius: 8px; border: 1px solid #e2e8f0;">
      <h2 style="font-size: 1.1rem; color: #1e293b; margin-top: 0;">Available API Endpoints:</h2>
      <ul style="color: #334155;">
        <li><code>GET /api/v1/briefings/latest</code> - Latest Daily Market Briefing</li>
        <li><code>GET /api/v1/briefings</code> - Historical Market Briefings</li>
        <li><code>GET /api/v1/news</code> - Financial & Investment News (supports <code>?ticker=...</code>, <code>?industry=...</code>)</li>
        <li><code>GET /api/v1/announcements</code> - IDX Official Disclosures</li>
        <li><code>GET /api/v1/financial-reports</code> - Financial Reports</li>
        <li><code>GET /api/v1/user/watchlist</code> - User Watchlist (Auth required)</li>
      </ul>
    </div>
  </div>
</template>
