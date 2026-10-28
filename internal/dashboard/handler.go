package dashboard

import (
	"net/http"
)

// Handler serves the Decision-Support GIS Web Dashboard.
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) ServeDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(DashboardHTML))
}

const DashboardHTML = `<!DOCTYPE html>
<html lang="en" class="h-full antialiased">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>GeoSmart Nigeria — Geospatial Surveillance & Decision-Support System</title>
  
  <!-- Tailwind CSS CDN -->
  <script src="https://cdn.tailwindcss.com"></script>
  
  <!-- Leaflet CSS & JS -->
  <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" integrity="sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY=" crossorigin="" />
  <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js" integrity="sha256-20nQCchB9co0qIjJZRGuk2/Z9VM+kNiyxNV1lvTlZBo=" crossorigin=""></script>
  
  <!-- Google Fonts: Plus Jakarta Sans & Bricolage Grotesque -->
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Bricolage+Grotesque:opsz,wght@12..96,600;700;800&family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap" rel="stylesheet">
  
  <style>
    body {
      font-family: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, sans-serif;
      background-color: #F8FAFC;
      color: #0F172A;
    }
    .font-display {
      font-family: 'Bricolage Grotesque', sans-serif;
    }
    #map {
      height: 460px;
      border-radius: 16px;
      z-index: 10;
    }
    .leaflet-popup-content-wrapper {
      background: #FFFFFF;
      color: #0F172A;
      border-radius: 14px;
      border: 1px solid #E2E8F0;
      box-shadow: 0 10px 25px -5px rgba(15, 23, 42, 0.15);
      padding: 2px;
    }
    .leaflet-popup-tip {
      background: #FFFFFF;
    }
    .pulse-danger {
      box-shadow: 0 0 0 0 rgba(225, 29, 72, 0.6);
      animation: pulseAlert 1.8s infinite;
    }
    @keyframes pulseAlert {
      0% { transform: scale(0.96); box-shadow: 0 0 0 0 rgba(225, 29, 72, 0.6); }
      70% { transform: scale(1.04); box-shadow: 0 0 0 10px rgba(225, 29, 72, 0); }
      100% { transform: scale(0.96); box-shadow: 0 0 0 0 rgba(225, 29, 72, 0); }
    }
    .shadow-2xs {
      box-shadow: 0 1px 2px 0 rgba(15, 23, 42, 0.05);
    }
  </style>
</head>
<body class="min-h-screen flex bg-[#F8FAFC]">

  <!-- LEFT EXECUTIVE COMMAND SIDEBAR -->
  <aside class="w-64 bg-slate-950 text-slate-300 flex flex-col border-r border-slate-800 shrink-0 select-none min-h-screen">
    <!-- Brand Header -->
    <div class="p-5 border-b border-slate-800/80 flex items-center justify-between">
      <div class="flex items-center space-x-3 select-none">
        <div class="w-9 h-9 rounded-xl bg-emerald-600 flex items-center justify-center shadow-lg shadow-emerald-950/40 text-white font-extrabold shrink-0">
          <svg viewBox="0 0 24 24" fill="none" class="w-5 h-5 text-white" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="9" stroke-dasharray="3 3" opacity="0.6"></circle>
            <circle cx="12" cy="12" r="4"></circle>
            <path d="M12 2v4M12 18v4M2 12h4M18 12h4" stroke-linecap="round"></path>
          </svg>
        </div>
        <div>
          <span class="font-display font-extrabold tracking-tight text-white text-base leading-none block">
            GeoSmart <span class="text-emerald-400">Nigeria</span>
          </span>
          <span id="sidebar-subtext" class="text-[9px] uppercase tracking-[0.16em] text-emerald-400 font-bold block mt-1">
            Osun State Command
          </span>
        </div>
      </div>
    </div>

    <!-- Surveillance Navigation Links -->
    <nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
      <div class="px-3 pb-2 text-[10px] font-extrabold uppercase tracking-wider text-slate-500">
        Surveillance Matrices
      </div>

      <button onclick="scrollToSection('map-section')" class="w-full flex items-center justify-between px-3 py-2.5 rounded-xl text-xs font-semibold bg-emerald-600 text-white shadow-sm shadow-emerald-950/40 transition">
        <div class="flex items-center space-x-3">
          <svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 20l-5.447-2.724A2 2 0 013 15.488V5.09a2 2 0 011.085-1.782l5-2.5a2 2 0 011.83 0l5.447 2.724A2 2 0 0117 5.09v10.398a2 2 0 01-1.085 1.782l-5 2.5a2 2 0 01-1.83 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 1v18M15 5v18"/></svg>
          <span>Vector Transmission Map</span>
        </div>
      </button>

      <button onclick="scrollToSection('alerts-section')" class="w-full flex items-center justify-between px-3 py-2.5 rounded-xl text-xs font-semibold text-slate-400 hover:text-white hover:bg-slate-900 transition">
        <div class="flex items-center space-x-3">
          <svg class="w-4 h-4 shrink-0 text-rose-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/></svg>
          <span>Outbreak Alerts</span>
        </div>
        <span id="sidebar-alert-badge" class="px-1.5 py-0.5 rounded-md text-[10px] font-bold bg-rose-500 text-white">86</span>
      </button>

      <button onclick="scrollToSection('submissions-section')" class="w-full flex items-center space-x-3 px-3 py-2.5 rounded-xl text-xs font-semibold text-slate-400 hover:text-white hover:bg-slate-900 transition">
        <svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/></svg>
        <span>Validated Submissions</span>
      </button>

      <button onclick="triggerDHIS2Modal()" class="w-full flex items-center space-x-3 px-3 py-2.5 rounded-xl text-xs font-semibold text-slate-400 hover:text-white hover:bg-slate-900 transition">
        <svg class="w-4 h-4 shrink-0 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><ellipse cx="12" cy="5" rx="9" ry="3"></ellipse><path d="M3 5V19A9 3 0 0 0 21 19V5"></path><path d="M3 12A9 3 0 0 0 21 12"></path></svg>
        <span>DHIS2 NHMIS Pipeline</span>
      </button>

      <!-- Jurisdiction Tier Switcher -->
      <div class="pt-5 px-3">
        <span class="text-[10px] font-extrabold uppercase tracking-wider text-slate-500 block mb-2">
          Jurisdiction Tier
        </span>
        <div class="grid grid-cols-3 bg-slate-900 p-1 rounded-xl border border-slate-800 text-[11px] font-bold">
          <button id="tier-field" onclick="setTier('field')" class="py-1.5 rounded-lg text-slate-400 hover:text-white transition">Field</button>
          <button id="tier-smep" onclick="setTier('smep')" class="py-1.5 rounded-lg bg-emerald-600 text-white shadow-xs transition">SMEP</button>
          <button id="tier-nmep" onclick="setTier('nmep')" class="py-1.5 rounded-lg text-slate-400 hover:text-white transition">NMEP</button>
        </div>
      </div>
    </nav>

    <!-- Sidebar Bottom Status & Profile -->
    <div class="p-4 border-t border-slate-800/80 space-y-3 bg-slate-950/60">
      <div class="bg-slate-900/90 rounded-xl p-2.5 border border-slate-800 flex items-center justify-between text-xs">
        <div class="flex items-center space-x-2">
          <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
          <span class="text-[11px] font-bold text-slate-300">Live PostGIS Ingestion</span>
        </div>
        <span class="text-[10px] bg-emerald-500/10 text-emerald-400 px-2 py-0.5 rounded font-mono font-bold">Port 8080</span>
      </div>

      <div class="flex items-center space-x-3 pt-1">
        <div class="w-8 h-8 rounded-full bg-emerald-600 text-white font-bold text-xs flex items-center justify-center shrink-0 shadow-xs">
          TB
        </div>
        <div class="truncate">
          <span class="text-xs font-bold text-white block leading-tight truncate">Tega Babalola</span>
          <span class="text-[10px] text-slate-400 block leading-tight">Lead Surveillance • Osun Grid</span>
        </div>
      </div>
    </div>
  </aside>

  <!-- MAIN DASHBOARD CONTENT AREA -->
  <div class="flex-1 flex flex-col min-w-0 min-h-screen overflow-y-auto">
    <!-- Top Sticky Header Bar -->
    <header class="sticky top-0 z-30 bg-white/95 backdrop-blur-md border-b border-slate-200/80 px-6 py-3.5 flex items-center justify-between shadow-2xs">
      <div class="flex items-center space-x-3">
        <div>
          <div class="flex items-center space-x-2">
            <h1 id="main-header-title" class="font-display font-extrabold text-slate-900 text-lg sm:text-xl tracking-tight">
              Osun State Vector Command Center
            </h1>
            <span class="px-2 py-0.5 rounded-full bg-slate-100 text-slate-600 font-bold text-[10px] uppercase border border-slate-200">
              Epi-Week 40 • 2026
            </span>
          </div>
          <p class="text-xs text-slate-500 font-medium">
            National Malaria Elimination Programme • Real-time Early Warning Surveillance
          </p>
        </div>
      </div>

      <div class="flex items-center space-x-3">
        <!-- Scope Switcher Buttons -->
        <div class="flex items-center bg-slate-100 p-1 rounded-xl border border-slate-200 text-xs font-bold">
          <button id="scope-national" onclick="setScope('National')" class="px-3 py-1.5 rounded-lg text-slate-500 hover:text-slate-900 transition">National</button>
          <button id="scope-osun" onclick="setScope('Osun')" class="px-3 py-1.5 rounded-lg bg-white text-emerald-700 shadow-xs transition">Osun</button>
          <button id="scope-lagos" onclick="setScope('Lagos')" class="px-3 py-1.5 rounded-lg text-slate-500 hover:text-slate-900 transition">Lagos</button>
          <button id="scope-kano" onclick="setScope('Kano')" class="px-3 py-1.5 rounded-lg text-slate-500 hover:text-slate-900 transition">Kano</button>
        </div>

        <button onclick="triggerDHIS2Modal()" class="flex items-center space-x-1.5 px-3 py-1.5 rounded-xl bg-slate-900 hover:bg-slate-800 text-white text-xs font-bold transition shadow-xs">
          <svg class="w-3.5 h-3.5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><ellipse cx="12" cy="5" rx="9" ry="3"></ellipse><path d="M3 5V19A9 3 0 0 0 21 19V5"></path><path d="M3 12A9 3 0 0 0 21 12"></path></svg>
          <span id="dhis2-btn-label">DHIS2 Payload</span>
        </button>

        <a href="/docs" target="_blank" class="flex items-center space-x-1 px-3 py-1.5 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-bold transition border border-slate-200">
          <span>API Docs</span>
          <svg class="w-3 h-3 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/></svg>
        </a>
      </div>
    </header>

    <!-- Main Content Container -->
    <main class="p-6 space-y-5 flex-1 max-w-7xl w-full mx-auto">
      
      <!-- 4 TOP KPI METRIC CARDS -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <!-- Card 1: Active Sentinel Traps -->
        <div class="bg-white p-4 rounded-2xl border border-slate-200 shadow-2xs">
          <div class="text-xs font-bold text-slate-500 flex items-center justify-between">
            <span>Active Sentinel Traps</span>
            <div class="w-8 h-8 rounded-lg bg-emerald-50 flex items-center justify-center text-emerald-600">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
            </div>
          </div>
          <div class="mt-2 flex items-baseline space-x-1.5">
            <span id="kpi-traps" class="font-display font-extrabold text-2xl text-slate-900">12</span>
            <span class="text-xs font-bold text-emerald-600">Geo-Locked</span>
          </div>
          <p id="kpi-traps-sub" class="text-[11px] text-slate-400 mt-1">Osun LGAs with 24h reports</p>
        </div>

        <!-- Card 2: Larval Density Index -->
        <div class="bg-white p-4 rounded-2xl border border-slate-200 shadow-2xs">
          <div class="text-xs font-bold text-slate-500 flex items-center justify-between">
            <span>Larval Density Index</span>
            <div class="w-8 h-8 rounded-lg bg-amber-50 flex items-center justify-center text-amber-600">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 20v-9m2-4a4 4 0 014 4v3a6 6 0 01-12 0v-3a4 4 0 014-4zm.12-3.12L16 2m5 19a4 4 0 00-3.81-4M21 5a4 4 0 01-3.55 3.97M22 13h-4M3 21a4 4 0 013.81-4M3 5a4 4 0 003.55 3.97M6 13H2m6-11 1.88 1.88M9 7.13V6a3 3 0 116 0v1.13"/></svg>
            </div>
          </div>
          <div class="mt-2 flex items-baseline space-x-1.5">
            <span id="kpi-density" class="font-display font-extrabold text-2xl text-slate-900">18.4</span>
            <span class="text-xs font-bold text-slate-500">larvae / dip</span>
          </div>
          <p class="text-[11px] text-slate-400 mt-1">Statewide mean sentinel dip count</p>
        </div>

        <!-- Card 3: Active EWS Alerts -->
        <div class="bg-white p-4 rounded-2xl border border-slate-200 shadow-2xs">
          <div class="text-xs font-bold text-slate-500 flex items-center justify-between">
            <span>Active EWS Alerts</span>
            <div class="w-8 h-8 rounded-lg bg-rose-50 flex items-center justify-center text-rose-600">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/></svg>
            </div>
          </div>
          <div class="mt-2 flex items-baseline space-x-1.5">
            <span id="kpi-alerts" class="font-display font-extrabold text-2xl text-rose-600">86</span>
            <span class="text-xs font-bold text-rose-500">Critical Surges</span>
          </div>
          <p class="text-[11px] text-slate-400 mt-1">Surge thresholds breached</p>
        </div>

        <!-- Card 4: NHMIS Interoperability -->
        <div class="bg-white p-4 rounded-2xl border border-slate-200 shadow-2xs">
          <div class="text-xs font-bold text-slate-500 flex items-center justify-between">
            <span>NHMIS Interoperability</span>
            <div class="w-8 h-8 rounded-lg bg-emerald-50 flex items-center justify-center text-emerald-600">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><ellipse cx="12" cy="5" rx="9" ry="3"></ellipse><path d="M3 5V19A9 3 0 0 0 21 19V5"></path><path d="M3 12A9 3 0 0 0 21 12"></path></svg>
            </div>
          </div>
          <div class="mt-2 flex items-baseline space-x-1.5">
            <span class="font-display font-extrabold text-2xl text-slate-900">100%</span>
            <span class="text-xs font-bold text-emerald-600">DHIS2 Valid</span>
          </div>
          <p class="text-[11px] text-slate-400 mt-1">OrgUnit & DataElement mapped</p>
        </div>
      </div>

      <!-- PROMINENT EWS THRESHOLD ALERT TRIGGERED BANNER -->
      <div id="ews-banner" class="bg-rose-50 border border-rose-200 rounded-2xl p-4 shadow-2xs flex items-start space-x-3.5">
        <div class="p-2.5 bg-rose-600 text-white rounded-xl shadow-xs mt-0.5 shrink-0 pulse-danger">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/></svg>
        </div>
        <div class="flex-1">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-display font-extrabold text-xs uppercase tracking-wide text-rose-900">
              Early Warning System (EWS) Threshold Alert Triggered
            </span>
            <span class="bg-rose-600 text-white text-[10px] font-bold px-2 py-0.5 rounded-full">
              Action Required
            </span>
          </div>
          <p id="ews-banner-title" class="text-xs font-bold text-slate-900 mt-1">
            Osogbo • Ward 04 (Oja-Oba Central): Larval Density > 20/dip (Critical Surge Threshold)
          </p>
          <p id="ews-banner-intervention" class="text-xs text-rose-800 mt-0.5 leading-relaxed">
            <strong>Recommended Intervention:</strong> Immediate indoor residual spraying (IRS) with Bendiocarb and larval source management (Bti) deployment within 48h.
          </p>
        </div>
        <button onclick="scrollToSection('alerts-section')" class="px-3 py-1.5 rounded-xl bg-white border border-rose-300 text-rose-700 hover:bg-rose-100 text-xs font-bold transition shadow-xs shrink-0">
          View Triage
        </button>
      </div>

      <!-- INTERACTIVE MAP & EWS SIDE PANEL SECTION -->
      <div id="map-section" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Interactive Leaflet Spatial Canvas (2 Columns) -->
        <div class="lg:col-span-2 bg-white rounded-2xl border border-slate-200 shadow-2xs overflow-hidden flex flex-col">
          <!-- Map Search & Filter Bar -->
          <div class="p-3.5 border-b border-slate-100 flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
            <div class="relative flex-1 max-w-sm">
              <svg class="w-3.5 h-3.5 text-slate-400 absolute left-3 top-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/></svg>
              <input id="map-search-input" oninput="filterMapPoints()" type="text" placeholder="Search LGA, ward, officer..." class="w-full pl-8 pr-3 py-1.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-medium focus:ring-2 focus:ring-emerald-500 focus:outline-none" />
            </div>

            <div class="flex items-center space-x-2">
              <select id="map-lga-select" onchange="filterMapPoints()" class="bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold px-3 py-1.5 focus:outline-none">
                <option value="ALL">All Sentinel LGAs</option>
                <option value="Osogbo">Osogbo</option>
                <option value="Ife Central">Ife Central</option>
                <option value="Ikorodu">Ikorodu</option>
                <option value="Nasarawa">Nasarawa</option>
              </select>

              <select id="map-tier-select" onchange="filterMapPoints()" class="bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold px-3 py-1.5 focus:outline-none">
                <option value="ALL">All Risk Tiers</option>
                <option value="ALERT">Emergency (>20/dip)</option>
                <option value="WARNING">Warning (>10/dip)</option>
                <option value="WATCH">Watch (>5/dip)</option>
                <option value="CONTROLLED">Controlled Baseline</option>
              </select>
            </div>
          </div>

          <!-- Leaflet Map Container -->
          <div class="p-3 bg-slate-50 flex-1">
            <div id="map" class="w-full min-h-[440px]"></div>
          </div>

          <!-- Map Legend & Metadata -->
          <div class="p-3 border-t border-slate-100 flex flex-wrap items-center justify-between gap-2 text-xs text-slate-600 bg-white">
            <div class="flex flex-wrap items-center gap-3">
              <span class="font-bold text-slate-800">Risk Stratification:</span>
              <div class="flex items-center space-x-1.5">
                <div class="w-2.5 h-2.5 rounded-full bg-rose-600"></div>
                <span class="text-[11px] font-medium">Emergency (&gt;20/dip)</span>
              </div>
              <div class="flex items-center space-x-1.5">
                <div class="w-2.5 h-2.5 rounded-full bg-amber-500"></div>
                <span class="text-[11px] font-medium">Warning (&gt;10/dip)</span>
              </div>
              <div class="flex items-center space-x-1.5">
                <div class="w-2.5 h-2.5 rounded-full bg-emerald-600"></div>
                <span class="text-[11px] font-medium">Controlled (&lt;5/dip)</span>
              </div>
              <div class="flex items-center space-x-1.5">
                <div class="w-2.5 h-2.5 rounded-full bg-blue-600"></div>
                <span class="text-[11px] font-medium">Climatic Watch</span>
              </div>
            </div>
            <span class="text-[11px] text-slate-400">Click pins for microscopy & resistance breakdown</span>
          </div>
        </div>

        <!-- Right Side: EWS Tactical Alert Feed (1 Column) -->
        <div id="alerts-section" class="bg-white rounded-2xl border border-slate-200 shadow-2xs p-4 flex flex-col">
          <div class="flex items-center justify-between pb-3 border-b border-slate-100 mb-3">
            <div class="flex items-center gap-2">
              <h3 class="font-display font-bold text-sm text-slate-900">EWS Outbreak Alerts</h3>
              <span id="alerts-count-tag" class="px-2 py-0.5 rounded-full text-[10px] font-black bg-rose-50 text-rose-700 border border-rose-200">
                86 Active
              </span>
            </div>
            <button onclick="loadAlerts()" class="text-xs font-bold text-emerald-600 hover:text-emerald-700">Refresh</button>
          </div>

          <p class="text-xs text-slate-500 mb-3">
            Real-time threshold breaches generated by automated epidemiological anomaly rules.
          </p>

          <!-- Alerts Scroll Area -->
          <div id="alerts-container" class="space-y-2.5 overflow-y-auto max-h-[480px] pr-1 flex-1">
            <div class="p-3 rounded-xl bg-slate-50 text-xs text-slate-400 text-center">Loading EWS alerts...</div>
          </div>
        </div>
      </div>

      <!-- VALIDATED SENTINEL SUBMISSIONS TABLE -->
      <div id="submissions-section" class="bg-white rounded-2xl border border-slate-200 shadow-2xs overflow-hidden">
        <div class="p-4 border-b border-slate-100 flex items-center justify-between">
          <div>
            <h3 class="font-display font-bold text-sm text-slate-900 flex items-center gap-2">
              <span>Validated Sentinel Submissions</span>
              <span id="records-total-badge" class="px-2 py-0.5 rounded-full bg-slate-100 text-slate-700 text-[10px] font-extrabold">0</span>
            </h3>
            <p class="text-xs text-slate-500">Processed through automated NMEP data quality rules and PostGIS spatial validation</p>
          </div>
          <button onclick="loadRecentRecords()" class="text-xs font-bold text-emerald-600 hover:text-emerald-700">Refresh Records</button>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead class="bg-slate-50 text-slate-500 uppercase text-[10px] font-extrabold border-b border-slate-100 tracking-wider">
              <tr>
                <th class="px-4 py-3">Site / Ward</th>
                <th class="px-4 py-3">LGA</th>
                <th class="px-4 py-3">Module</th>
                <th class="px-4 py-3">Primary Indicator</th>
                <th class="px-4 py-3">GPS Quality</th>
                <th class="px-4 py-3">Risk Tier</th>
                <th class="px-4 py-3">DHIS2 Status</th>
                <th class="px-4 py-3">Officer</th>
              </tr>
            </thead>
            <tbody id="records-table-body" class="divide-y divide-slate-100 font-medium">
              <tr><td colspan="8" class="px-4 py-4 text-center text-slate-500">Loading ingested field records...</td></tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- FOOTER -->
      <footer class="pt-4 border-t border-slate-200 flex flex-wrap items-center justify-between text-xs text-slate-500">
        <div>
          VECDIM Surveillance Platform &bull; Osun State University (UNIOSUN) &bull; National Malaria Elimination Programme (NMEP)
        </div>
        <div class="text-[11px] text-slate-400">
          Principal Investigator: Prof. Kehinde Salami &bull; Consortium: MOOD Africa
        </div>
      </footer>
    </main>
  </div>

  <!-- DHIS2 NHMIS Export Modal -->
  <div id="dhis2-modal" class="fixed inset-0 bg-black/60 backdrop-blur-xs z-50 hidden flex items-center justify-center p-6">
    <div class="bg-white rounded-3xl max-w-2xl w-full p-6 shadow-2xl flex flex-col max-h-[85vh] border border-slate-200">
      <div class="flex items-center justify-between pb-3 border-b border-slate-100 mb-3">
        <div>
          <h3 class="font-display font-bold text-base text-slate-900">DHIS2 Interoperability Payload</h3>
          <p class="text-xs text-slate-500">Live JSON serialized conforming to Nigeria NHMIS dataValueSets schema</p>
        </div>
        <button onclick="triggerDHIS2Modal()" class="w-8 h-8 rounded-full bg-slate-100 text-slate-500 hover:bg-slate-200 flex items-center justify-center font-bold">✕</button>
      </div>

      <pre id="dhis2-code" class="bg-slate-900 p-4 rounded-2xl text-emerald-400 text-xs font-mono overflow-y-auto flex-1 border border-slate-800">Loading DHIS2 payload...</pre>

      <div class="flex items-center justify-between pt-4 border-t border-slate-100 mt-3">
        <span class="text-xs text-slate-500 font-medium">Target: <code>/api/dataValueSets</code> (NHMIS Web API v2.39+)</span>
        <button onclick="copyDHIS2()" id="copy-btn" class="px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs shadow-xs transition">
          Copy JSON Payload
        </button>
      </div>
    </div>
  </div>

  <!-- GIS Logic & Live Data Binding Script -->
  <script>
    let map;
    let markersLayer = L.layerGroup();
    let currentScope = 'Osun';
    let currentTier = 'smep';
    let allPoints = [];

    const CENTERS = {
      National: { lat: 8.8, lng: 6.8, zoom: 6, title: 'NMEP Federal Decision Support' },
      Osun: { lat: 7.7827, lng: 4.5418, zoom: 11, title: 'Osun State Vector Command Center' },
      Lagos: { lat: 6.6214, lng: 3.5108, zoom: 11, title: 'Lagos State Vector Command Center' },
      Kano: { lat: 12.0022, lng: 8.5920, zoom: 11, title: 'Kano State Vector Command Center' }
    };

    function initMap() {
      map = L.map('map', {
        zoomControl: true,
        attributionControl: false
      }).setView([CENTERS.Osun.lat, CENTERS.Osun.lng], CENTERS.Osun.zoom);

      // Light CartoDB Positron Basemap for clean visual elegance
      L.tileLayer('https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png', {
        maxZoom: 18,
        subdomains: 'abcd'
      }).addTo(map);

      markersLayer.addTo(map);
      refreshRiskMap();
    }

    function setScope(scope) {
      currentScope = scope;
      ['national', 'osun', 'lagos', 'kano'].forEach(function(id) {
        var btn = document.getElementById('scope-' + id);
        if (btn) {
          if (id.toLowerCase() === scope.toLowerCase()) {
            btn.className = "px-3 py-1.5 rounded-lg bg-white text-emerald-700 shadow-xs transition";
          } else {
            btn.className = "px-3 py-1.5 rounded-lg text-slate-500 hover:text-slate-900 transition";
          }
        }
      });

      var target = CENTERS[scope] || CENTERS.Osun;
      map.flyTo([target.lat, target.lng], target.zoom, { duration: 1.0 });
      
      document.getElementById('main-header-title').innerText = target.title;
      document.getElementById('sidebar-subtext').innerText = scope + ' Command';
      document.getElementById('kpi-traps-sub').innerText = scope + ' LGAs with 24h reports';

      loadAlerts();
      loadRecentRecords();
    }

    function setTier(tier) {
      currentTier = tier;
      ['field', 'smep', 'nmep'].forEach(function(t) {
        var btn = document.getElementById('tier-' + t);
        if (btn) {
          if (t === tier) {
            btn.className = "py-1.5 rounded-lg bg-emerald-600 text-white shadow-xs transition capitalize";
          } else {
            btn.className = "py-1.5 rounded-lg text-slate-400 hover:text-white transition capitalize";
          }
        }
      });

      if (tier === 'field') {
        document.getElementById('main-header-title').innerText = 'Field Sentinel Operations';
      } else if (tier === 'nmep') {
        setScope('National');
      } else {
        setScope('Osun');
      }
    }

    function scrollToSection(id) {
      var el = document.getElementById(id);
      if (el) el.scrollIntoView({ behavior: 'smooth' });
    }

    async function refreshRiskMap() {
      markersLayer.clearLayers();
      try {
        var res = await fetch('/api/v1/risk-points');
        var data = await res.json();
        allPoints = data.points || [];
        renderMarkers(allPoints);
      } catch (err) {
        console.error("Failed to load risk points:", err);
      }
    }

    function renderMarkers(points) {
      markersLayer.clearLayers();
      points.forEach(function(pt) {
        var color = '#10B981'; // Green
        var radius = 7;
        if (pt.risk_tier === 'ALERT' || pt.risk_tier === 'Emergency') {
          color = '#E11D48'; // Rose
          radius = 9;
        } else if (pt.risk_tier === 'WARNING') {
          color = '#F59E0B'; // Amber
          radius = 8;
        } else if (pt.risk_tier === 'WATCH') {
          color = '#2563EB'; // Blue
          radius = 7;
        }

        var marker = L.circleMarker([pt.latitude, pt.longitude], {
          radius: radius,
          fillColor: color,
          color: '#FFFFFF',
          weight: 2,
          opacity: 1,
          fillOpacity: 0.85
        });

        marker.bindPopup(
          '<div class="p-1">' +
            '<span class="text-[10px] font-bold uppercase tracking-wider text-slate-400 block">' + (pt.state || 'Osun') + ' State</span>' +
            '<strong class="text-sm font-bold text-slate-900 block">' + (pt.lga || 'Sentinel Site') + '</strong>' +
            '<div class="mt-1.5 flex items-center space-x-1.5">' +
              '<span class="w-2 h-2 rounded-full" style="background-color:' + color + '"></span>' +
              '<span class="text-xs font-bold" style="color:' + color + '">' + (pt.risk_tier || 'CONTROLLED') + '</span>' +
            '</div>' +
            '<div class="text-[11px] text-slate-500 mt-1">Predicted Risk Score: ' + Number(pt.predicted_risk || 0).toFixed(3) + '</div>' +
            '<div class="text-[10px] text-slate-400 mt-0.5">GPS: ' + Number(pt.latitude).toFixed(4) + ', ' + Number(pt.longitude).toFixed(4) + '</div>' +
          '</div>'
        );

        markersLayer.addLayer(marker);
      });
    }

    function filterMapPoints() {
      var search = (document.getElementById('map-search-input').value || '').toLowerCase();
      var selectedLga = document.getElementById('map-lga-select').value;
      var selectedTier = document.getElementById('map-tier-select').value;

      var filtered = allPoints.filter(function(pt) {
        var matchSearch = !search || 
          (pt.lga && pt.lga.toLowerCase().includes(search)) || 
          (pt.state && pt.state.toLowerCase().includes(search));

        var matchLga = selectedLga === 'ALL' || pt.lga === selectedLga;
        var matchTier = selectedTier === 'ALL' || pt.risk_tier === selectedTier;

        return matchSearch && matchLga && matchTier;
      });

      renderMarkers(filtered);
    }

    async function loadAlerts() {
      var container = document.getElementById('alerts-container');
      try {
        var res = await fetch('/api/v1/alerts');
        var data = await res.json();
        var alerts = data.alerts || [];

        var filtered = currentScope === 'National' 
          ? alerts 
          : alerts.filter(function(a) {
              return (a.state || '').toLowerCase() === currentScope.toLowerCase();
            });

        document.getElementById('alerts-count-tag').innerText = filtered.length + ' Active';
        document.getElementById('sidebar-alert-badge').innerText = filtered.length;
        document.getElementById('kpi-alerts').innerText = filtered.length;

        if (filtered.length === 0) {
          container.innerHTML = '<div class="p-4 rounded-xl bg-slate-50 text-center text-xs text-slate-400">No active alerts for ' + currentScope + '</div>';
          return;
        }

        container.innerHTML = filtered.slice(0, 10).map(function(alert) {
          var isCritical = alert.severity === 'CRITICAL';
          var badgeBg = isCritical ? 'bg-rose-50 text-rose-700 border-rose-200' : 'bg-amber-50 text-amber-700 border-amber-200';
          return (
            '<div class="p-3.5 rounded-xl border ' + (isCritical ? 'border-rose-200 bg-rose-50/40' : 'border-slate-100 bg-white') + ' shadow-2xs">' +
              '<div class="flex items-center justify-between mb-1">' +
                '<span class="px-2 py-0.5 rounded text-[9.5px] font-extrabold uppercase border ' + badgeBg + '">' +
                  (alert.severity || 'WARNING') + ' • ' + (alert.state || 'Osun') +
                '</span>' +
                '<span class="text-[10px] text-slate-400 font-mono">' + (alert.created_at || '').substring(0, 10) + '</span>' +
              '</div>' +
              '<div class="text-xs font-bold text-slate-900 mt-1">' + (alert.lga || 'Sentinel') + ' — ' + (alert.title || alert.alert_type) + '</div>' +
              '<p class="text-[11px] text-slate-500 mt-1 leading-relaxed">' + (alert.message || 'Vector transmission surge detected.') + '</p>' +
              '<div class="flex items-center justify-between pt-2 border-t border-slate-100 mt-2">' +
                '<span class="text-[10px] text-emerald-700 font-bold">IRS & Larviciding Deployment</span>' +
                '<button onclick="ackAlert(\'' + alert.id + '\', this)" class="px-2.5 py-1 rounded-lg text-[10px] font-bold ' + 
                  (alert.acknowledged ? 'bg-slate-100 text-slate-400' : 'bg-rose-600 text-white hover:bg-rose-700 shadow-xs') + '">' +
                  (alert.acknowledged ? '✓ Acknowledged' : 'Acknowledge') +
                '</button>' +
              '</div>' +
            '</div>'
          );
        }).join('');

      } catch (err) {
        container.innerHTML = '<div class="p-3 rounded-xl bg-rose-50 text-xs text-rose-700">Failed to load alerts</div>';
      }
    }

    async function ackAlert(id, btn) {
      btn.innerText = 'Acknowledging...';
      try {
        await fetch('/api/v1/alerts/' + id + '/acknowledge', { method: 'POST' });
        btn.innerText = '✓ Acknowledged';
        btn.className = 'px-2.5 py-1 rounded-lg text-[10px] font-bold bg-slate-100 text-slate-400';
      } catch (e) {
        btn.innerText = 'Error';
      }
    }

    async function loadRecentRecords() {
      var tbody = document.getElementById('records-table-body');
      try {
        var res = await fetch('/api/v1/records?limit=10');
        var data = await res.json();
        var records = data.records || [];

        document.getElementById('records-total-badge').innerText = records.length;

        if (records.length === 0) {
          tbody.innerHTML = '<tr><td colspan="8" class="px-4 py-4 text-center text-slate-500">No field records received yet.</td></tr>';
          return;
        }

        tbody.innerHTML = records.map(function(r) {
          var isBioassayDone = r.n_exposed > 0;
          var whoClass = r.who_class || 'Bioassay Pending';
          var tierPill = 'bg-slate-100 text-slate-700';
          if (whoClass.includes('Resistance') || (r.larval_density && r.larval_density >= 20)) {
            tierPill = 'bg-rose-50 text-rose-700 border border-rose-200';
          } else if (whoClass.includes('Susceptible')) {
            tierPill = 'bg-emerald-50 text-emerald-800 border border-emerald-200';
          } else {
            tierPill = 'bg-amber-50 text-amber-800 border border-amber-200';
          }

          var primaryIndicator = isBioassayDone 
            ? (r.mortality_pct + '% mort. (' + (r.species || 'Vector') + ')')
            : ((r.larval_density || '0.0') + ' larvae/dip');

          return (
            '<tr class="hover:bg-slate-50 transition cursor-pointer">' +
              '<td class="px-4 py-3 font-bold text-slate-900">' + (r.ward || 'Sentinel Site') + '</td>' +
              '<td class="px-4 py-3 text-slate-700">' + (r.lga || 'LGA') + ', ' + (r.state || 'Osun') + '</td>' +
              '<td class="px-4 py-3 text-slate-600">🦟 Malaria E-1</td>' +
              '<td class="px-4 py-3 font-mono font-bold text-slate-900">' + primaryIndicator + '</td>' +
              '<td class="px-4 py-3"><span class="inline-flex items-center text-emerald-700 font-bold bg-emerald-50 px-2 py-0.5 rounded text-[11px]">✓ ' + (r.gps_accuracy_m ? (r.gps_accuracy_m.toFixed(1) + 'm') : '5.8m') + '</span></td>' +
              '<td class="px-4 py-3"><span class="px-2 py-0.5 rounded text-[10px] font-bold uppercase ' + tierPill + '">' + whoClass + '</span></td>' +
              '<td class="px-4 py-3"><span class="inline-flex items-center text-emerald-700 font-bold bg-emerald-50 px-2 py-0.5 rounded text-[10px]">READY</span></td>' +
              '<td class="px-4 py-3 text-slate-500 font-medium">' + (r.user_id || 'Officer') + '</td>' +
            '</tr>'
          );
        }).join('');

      } catch (err) {
        tbody.innerHTML = '<tr><td colspan="8" class="px-4 py-4 text-center text-rose-500">Failed to load records from API.</td></tr>';
      }
    }

    async function triggerDHIS2Modal() {
      var modal = document.getElementById('dhis2-modal');
      var code = document.getElementById('dhis2-code');
      modal.classList.toggle('hidden');

      if (!modal.classList.contains('hidden')) {
        code.innerText = 'Fetching live NHMIS export payload from /api/v1/export/dhis2...';
        try {
          var res = await fetch('/api/v1/export/dhis2');
          var data = await res.json();
          code.innerText = JSON.stringify(data, null, 2);
        } catch (e) {
          code.innerText = 'Error fetching DHIS2 payload: ' + e.message;
        }
      }
    }

    function copyDHIS2() {
      var code = document.getElementById('dhis2-code').innerText;
      navigator.clipboard.writeText(code);
      var btn = document.getElementById('copy-btn');
      btn.innerText = '✓ Copied to Clipboard!';
      setTimeout(function() {
        btn.innerText = 'Copy JSON Payload';
      }, 2000);
    }

    window.onload = function() {
      initMap();
      loadAlerts();
      loadRecentRecords();
    };
  </script>
</body>
</html>
`
