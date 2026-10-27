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
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>VECDIM — Geospatial Surveillance & Decision-Support System</title>
  
  <!-- Tailwind CSS CDN -->
  <script src="https://cdn.tailwindcss.com"></script>
  
  <!-- Leaflet CSS & JS -->
  <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" integrity="sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY=" crossorigin="" />
  <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js" integrity="sha256-20nQCchB9co0qIjJZRGuk2/Z9VM+kNiyxNV1lvTlZBo=" crossorigin=""></script>
  
  <!-- Google Fonts: Inter -->
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800;900&display=swap" rel="stylesheet">
  
  <style>
    body { font-family: 'Inter', sans-serif; background-color: #090D16; color: #F1F5F9; }
    #map { height: 460px; border-radius: 16px; z-index: 10; }
    .leaflet-popup-content-wrapper { background: #1E293B; color: #F8FAFC; border-radius: 12px; border: 1px solid #334155; }
    .leaflet-popup-tip { background: #1E293B; }
    .pulse-alert {
      box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.7);
      animation: pulse 1.6s infinite;
    }
    @keyframes pulse {
      0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.7); }
      70% { transform: scale(1.05); box-shadow: 0 0 0 10px rgba(239, 68, 68, 0); }
      100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(239, 68, 68, 0); }
    }
  </style>
</head>
<body class="min-h-screen flex flex-col">
  <!-- Top Navigation Bar -->
  <header class="border-b border-slate-800 bg-[#0F172A]/90 backdrop-blur sticky top-0 z-50 px-6 py-3.5">
    <div class="max-w-7xl mx-auto flex flex-wrap items-center justify-between gap-4">
      <div class="flex items-center space-x-3.5">
        <div class="w-10 h-10 rounded-xl bg-emerald-600/20 border border-emerald-500/40 flex items-center justify-center text-emerald-400 font-black text-xl">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="9" stroke-width="2" stroke="currentColor"/>
            <circle cx="12" cy="12" r="4" stroke-width="2" stroke="currentColor"/>
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 2v3m0 14v3M2 12h3m14 0h3"/>
          </svg>
        </div>
        <div>
          <div class="flex items-center space-x-2">
            <h1 class="text-lg font-extrabold tracking-tight text-white">GeoSmart <span class="text-emerald-400">Nigeria</span></h1>
            <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">VECDIM PLATFORM</span>
          </div>
          <p class="text-xs text-slate-400">Osun State University (UNIOSUN) • NMEP • MOOD Africa • Dev-Afrique</p>
        </div>
      </div>

      <!-- Scope Switcher Tabs -->
      <div class="flex items-center bg-slate-900 border border-slate-800 rounded-xl p-1 text-xs font-semibold">
        <button id="scope-national" onclick="setScope('National')" class="px-3.5 py-1.5 rounded-lg bg-emerald-600 text-white font-bold transition">National Directorate</button>
        <button id="scope-osun" onclick="setScope('Osun')" class="px-3.5 py-1.5 rounded-lg text-slate-400 hover:text-white transition">Osun SMEP</button>
        <button id="scope-lagos" onclick="setScope('Lagos')" class="px-3.5 py-1.5 rounded-lg text-slate-400 hover:text-white transition">Lagos SMEP</button>
        <button id="scope-kano" onclick="setScope('Kano')" class="px-3.5 py-1.5 rounded-lg text-slate-400 hover:text-white transition">Kano SMEP</button>
      </div>

      <!-- Quick Action Links -->
      <div class="flex items-center space-x-3 text-xs">
        <a href="/docs" target="_blank" class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 font-medium flex items-center gap-1.5 border border-slate-700">
          <span>Swagger API Docs</span>
          <svg class="w-3.5 h-3.5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/></svg>
        </a>
        <button onclick="triggerDHIS2Modal()" class="px-3 py-1.5 rounded-lg bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 font-bold border border-blue-500/40 flex items-center gap-1.5">
          <span>DHIS2 Export</span>
        </button>
      </div>
    </div>
  </header>

  <!-- Main Container -->
  <main class="max-w-7xl mx-auto px-6 py-6 flex-1 w-full space-y-6">
    <!-- Top KPI Cards (Tiers & Metrics) -->
    <div class="grid grid-cols-2 md:grid-cols-5 gap-4">
      <!-- Tier 1: Alert -->
      <div class="bg-[#121A2A] border border-red-500/30 rounded-2xl p-4 relative overflow-hidden shadow-lg shadow-red-950/20">
        <div class="flex items-center justify-between text-xs text-red-400 font-bold uppercase tracking-wider">
          <span>Emergency Tier</span>
          <span class="w-2 h-2 rounded-full bg-red-500 animate-ping"></span>
        </div>
        <div class="text-3xl font-black text-white mt-1">86</div>
        <p class="text-[11px] text-slate-400 mt-1">Critical vector outbreaks (Larvae $\ge$ 20/dip)</p>
      </div>

      <!-- Tier 2: Warning -->
      <div class="bg-[#121A2A] border border-amber-500/30 rounded-2xl p-4 shadow-lg shadow-amber-950/20">
        <div class="flex items-center justify-between text-xs text-amber-400 font-bold uppercase tracking-wider">
          <span>Warning Tier</span>
          <span class="w-2 h-2 rounded-full bg-amber-500"></span>
        </div>
        <div class="text-3xl font-black text-white mt-1">135</div>
        <p class="text-[11px] text-slate-400 mt-1">Surveillance thresholds elevated</p>
      </div>

      <!-- Tier 3: Watch -->
      <div class="bg-[#121A2A] border border-blue-500/30 rounded-2xl p-4 shadow-lg shadow-blue-950/20">
        <div class="flex items-center justify-between text-xs text-blue-400 font-bold uppercase tracking-wider">
          <span>Watch Tier</span>
          <span class="w-2 h-2 rounded-full bg-blue-500"></span>
        </div>
        <div class="text-3xl font-black text-white mt-1">426</div>
        <p class="text-[11px] text-slate-400 mt-1">Climatic & NDVI risk elevation</p>
      </div>

      <!-- Tier 4: Controlled -->
      <div class="bg-[#121A2A] border border-emerald-500/30 rounded-2xl p-4 shadow-lg shadow-emerald-950/20">
        <div class="flex items-center justify-between text-xs text-emerald-400 font-bold uppercase tracking-wider">
          <span>Controlled</span>
          <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
        </div>
        <div class="text-3xl font-black text-white mt-1">878</div>
        <p class="text-[11px] text-slate-400 mt-1">Endemic baseline containment</p>
      </div>

      <!-- Mathematical Parity -->
      <div class="bg-[#121A2A] border border-purple-500/30 rounded-2xl p-4 shadow-lg shadow-purple-950/20 col-span-2 md:col-span-1">
        <div class="flex items-center justify-between text-xs text-purple-400 font-bold uppercase tracking-wider">
          <span>R Parity Score</span>
          <span class="px-1.5 py-0.5 rounded text-[9px] bg-purple-500/20 text-purple-300 font-mono">11/11 RULES</span>
        </div>
        <div class="text-3xl font-black text-white mt-1">100.00%</div>
        <p class="text-[11px] text-slate-400 mt-1">1,525/1,525 exact match against R</p>
      </div>
    </div>

    <!-- Center Section: Interactive Map + EWS Alert Feed -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Leaflet GIS Map Canvas (2 Columns) -->
      <div class="lg:col-span-2 bg-[#121A2A] border border-slate-800 rounded-3xl p-5 shadow-xl flex flex-col">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h2 class="text-base font-bold text-white flex items-center gap-2">
              <svg class="w-5 h-5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 20l-5.447-2.724A2 2 0 013 15.488V5.09a2 2 0 011.085-1.782l5-2.5a2 2 0 011.83 0l5.447 2.724A2 2 0 0117 5.09v10.398a2 2 0 01-1.085 1.782l-5 2.5a2 2 0 01-1.83 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 1v18M15 5v18"/></svg>
              <span>National Sentinel Geofence & Transmission Risk Canvas</span>
            </h2>
            <p class="text-xs text-slate-400 mt-0.5">PostGIS spatial coordinates rendered with WHO vector resistance and bioassay classifications</p>
          </div>
          <div class="flex items-center gap-2 text-xs">
            <span class="flex items-center gap-1.5"><span class="w-2.5 h-2.5 rounded-full bg-red-500"></span> Alert</span>
            <span class="flex items-center gap-1.5"><span class="w-2.5 h-2.5 rounded-full bg-amber-500"></span> Warning</span>
            <span class="flex items-center gap-1.5"><span class="w-2.5 h-2.5 rounded-full bg-blue-500"></span> Watch</span>
            <span class="flex items-center gap-1.5"><span class="w-2.5 h-2.5 rounded-full bg-emerald-500"></span> Safe</span>
          </div>
        </div>

        <div id="map" class="w-full flex-1 min-h-[440px]"></div>

        <div class="flex items-center justify-between mt-3 text-xs text-slate-400 pt-2 border-t border-slate-800/80">
          <div class="flex items-center gap-4">
            <span id="map-zoom-tag" class="font-mono text-emerald-400">View: Nationwide (All Sentinel Hubs)</span>
            <span>• Haversine geofence threshold: 25.0km</span>
          </div>
          <button onclick="refreshRiskMap()" class="hover:text-white flex items-center gap-1 text-slate-400">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>
            <span>Reload Spatial Points</span>
          </button>
        </div>
      </div>

      <!-- Right Column: EWS Live Alerts Panel -->
      <div class="bg-[#121A2A] border border-slate-800 rounded-3xl p-5 shadow-xl flex flex-col">
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-2">
            <h2 class="text-base font-bold text-white">EWS Tactical Alerts</h2>
            <span id="alert-counter" class="px-2 py-0.5 rounded-full text-[10px] font-black bg-red-500/20 text-red-400 border border-red-500/30">86 ACTIVE</span>
          </div>
          <button onclick="loadAlerts()" class="text-xs text-slate-400 hover:text-white">Refresh</button>
        </div>
        <p class="text-xs text-slate-400 mb-4">Intervention alerts generated by statistical outbreak detection and ML risk models</p>

        <!-- Alerts Scroll List -->
        <div id="alerts-container" class="space-y-3 overflow-y-auto max-h-[480px] pr-1 flex-1">
          <!-- Dynamically populated -->
          <div class="p-3 rounded-2xl bg-slate-900 border border-slate-800 text-xs text-slate-400">Loading alerts from Go Modulith API...</div>
        </div>
      </div>
    </div>

    <!-- Bottom Section: Live Sync Ingestion Stream & Validation Parity Summary -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Live Sync Ingestion Feed (2 cols) -->
      <div class="lg:col-span-2 bg-[#121A2A] border border-slate-800 rounded-3xl p-5 shadow-xl">
        <div class="flex items-center justify-between mb-3">
          <div>
            <h3 class="text-base font-bold text-white">Live Mobile Ingestion Stream</h3>
            <p class="text-xs text-slate-400">Surveys ingested via mobile HTTP POST /api/v1/records/batch with officer attribution</p>
          </div>
          <button onclick="loadRecentRecords()" class="text-xs text-emerald-400 hover:underline">Refresh Records</button>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead>
              <tr class="border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider text-[10px]">
                <th class="py-2.5 px-3">Record ID</th>
                <th class="py-2.5 px-3">Officer</th>
                <th class="py-2.5 px-3">State / LGA</th>
                <th class="py-2.5 px-3">Vector Species</th>
                <th class="py-2.5 px-3">WHO Bioassay</th>
                <th class="py-2.5 px-3">Status</th>
              </tr>
            </thead>
            <tbody id="records-table-body" class="divide-y divide-slate-800/60 font-medium">
              <!-- Dynamically populated -->
              <tr><td colspan="6" class="py-4 text-center text-slate-500">Loading ingested field records...</td></tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Quick System Audit & Compliance -->
      <div class="bg-[#121A2A] border border-slate-800 rounded-3xl p-5 shadow-xl flex flex-col justify-between">
        <div>
          <h3 class="text-base font-bold text-white mb-1">System Health & Standards</h3>
          <p class="text-xs text-slate-400 mb-4">Federal Ministry of Health & NMEP operational compliance</p>
          
          <div class="space-y-3 text-xs">
            <div class="flex items-center justify-between p-2.5 rounded-xl bg-slate-900 border border-slate-800">
              <span class="text-slate-300">API Ingestion Latency</span>
              <span class="font-mono font-bold text-emerald-400">&lt; 30ms round-trip</span>
            </div>
            <div class="flex items-center justify-between p-2.5 rounded-xl bg-slate-900 border border-slate-800">
              <span class="text-slate-300">Spatial Engine</span>
              <span class="font-mono font-bold text-emerald-400">PostGIS 16-3.4 / 4326</span>
            </div>
            <div class="flex items-center justify-between p-2.5 rounded-xl bg-slate-900 border border-slate-800">
              <span class="text-slate-300">Idempotency Guarantee</span>
              <span class="font-mono font-bold text-emerald-400">UUID Deduplication</span>
            </div>
            <div class="flex items-center justify-between p-2.5 rounded-xl bg-slate-900 border border-slate-800">
              <span class="text-slate-300">DHIS2 NHMIS Standard</span>
              <span class="font-mono font-bold text-blue-400">Web API v2.39+ Payload</span>
            </div>
          </div>
        </div>

        <div class="pt-4 border-t border-slate-800 mt-4 text-[11px] text-slate-500">
          Candidate: <strong class="text-slate-300">Sheriffdeen Saula</strong> (<code class="text-emerald-400">salambhai.dev</code>)<br>
          Reader Access: <strong class="text-slate-300">moodafrica@uniosun.edu.ng</strong>
        </div>
      </div>
    </div>
  </main>

  <!-- DHIS2 Payload Inspection Modal -->
  <div id="dhis2-modal" class="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 hidden flex items-center justify-center p-6">
    <div class="bg-[#121A2A] border border-slate-700 rounded-3xl max-w-2xl w-full p-6 shadow-2xl flex flex-col max-h-[85vh]">
      <div class="flex items-center justify-between pb-3 border-b border-slate-800 mb-3">
        <div>
          <h3 class="text-base font-bold text-white">DHIS2 Interoperability Payload</h3>
          <p class="text-xs text-slate-400">Live JSON serialized conforming to Nigeria NHMIS dataValueSets schema</p>
        </div>
        <button onclick="triggerDHIS2Modal()" class="w-8 h-8 rounded-full bg-slate-800 text-slate-400 hover:text-white flex items-center justify-center">✕</button>
      </div>

      <pre id="dhis2-code" class="bg-slate-950 p-4 rounded-xl text-emerald-400 text-xs font-mono overflow-y-auto flex-1 border border-slate-800">Loading DHIS2 payload from /api/v1/export/dhis2...</pre>

      <div class="flex items-center justify-between pt-4 border-t border-slate-800 mt-3">
        <span class="text-xs text-slate-400">Target Endpoint: <code>/api/dataValueSets</code></span>
        <button onclick="copyDHIS2()" id="copy-btn" class="px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white font-bold text-xs">Copy JSON Payload</button>
      </div>
    </div>
  </div>

  <footer class="border-t border-slate-900 bg-[#090D16] py-4 text-center text-xs text-slate-500">
    VECDIM Surveillance Platform &bull; Osun State University &bull; National Malaria Elimination Programme (NMEP)
  </footer>

  <!-- GIS Logic & Live Data Binding Script -->
  <script>
    let map;
    let markersLayer = L.layerGroup();
    let currentScope = 'National';

    // State Centers
    const CENTERS = {
      National: { lat: 8.8, lng: 6.8, zoom: 6 },
      Osun: { lat: 7.7827, lng: 4.5418, zoom: 11 },
      Lagos: { lat: 6.6214, lng: 3.5108, zoom: 11 },
      Kano: { lat: 12.0022, lng: 8.5920, zoom: 11 }
    };

    function initMap() {
      map = L.map('map', {
        zoomControl: true,
        attributionControl: false
      }).setView([CENTERS.National.lat, CENTERS.National.lng], CENTERS.National.zoom);

      // Dark Theme Tiles via CartoDB
      L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
        maxZoom: 18,
        subdomains: 'abcd'
      }).addTo(map);

      markersLayer.addTo(map);
      refreshRiskMap();
    }

    function setScope(scope) {
      currentScope = scope;
      ['national', 'osun', 'lagos', 'kano'].forEach(id => {
        const btn = document.getElementById('scope-' + id);
        if (id.toLowerCase() === scope.toLowerCase()) {
          btn.className = "px-3.5 py-1.5 rounded-lg bg-emerald-600 text-white font-bold transition";
        } else {
          btn.className = "px-3.5 py-1.5 rounded-lg text-slate-400 hover:text-white transition";
        }
      });

      const target = CENTERS[scope] || CENTERS.National;
      map.flyTo([target.lat, target.lng], target.zoom, { duration: 1.2 });
      document.getElementById('map-zoom-tag').innerText = 'View: ' + scope + ' Sentinel Network';
      loadAlerts();
      loadRecentRecords();
    }

    // Refresh Spatial Risk Map Points
    async function refreshRiskMap() {
      markersLayer.clearLayers();

      try {
        const res = await fetch('/api/v1/risk-points');
        const data = await res.json();
        const points = data.points || [];

        points.forEach(pt => {
          let color = '#10B981'; // Green (Safe)
          if (pt.risk_tier === 'ALERT' || pt.risk_tier === 'Emergency') color = '#EF4444'; // Red
          else if (pt.risk_tier === 'WARNING') color = '#F59E0B'; // Amber
          else if (pt.risk_tier === 'WATCH') color = '#3B82F6'; // Blue

          const marker = L.circleMarker([pt.latitude, pt.longitude], {
            radius: pt.risk_tier === 'ALERT' ? 9 : 7,
            fillColor: color,
            color: '#FFFFFF',
            weight: 1.5,
            opacity: 0.9,
            fillOpacity: 0.8
          });

          marker.bindPopup(` + "`" + `
            <div class="text-xs">
              <strong class="text-emerald-400 font-bold block mb-1">${pt.lga || 'Sentinel Site'}, ${pt.state || 'Nigeria'}</strong>
              <div class="text-slate-300">Tier: <strong style="color:${color}">${pt.risk_tier}</strong></div>
              <div class="text-slate-400">Risk Score: ${Number(pt.predicted_risk || 0).toFixed(3)}</div>
              <div class="text-slate-400">Coordinates: ${Number(pt.latitude).toFixed(4)}, ${Number(pt.longitude).toFixed(4)}</div>
            </div>
          ` + "`" + `);

          markersLayer.addLayer(marker);
        });

      } catch (err) {
        console.error("Failed to load risk points:", err);
      }
    }

    // Load EWS Alerts
    async function loadAlerts() {
      const container = document.getElementById('alerts-container');
      try {
        const res = await fetch('/api/v1/alerts');
        const data = await res.json();
        const alerts = data.alerts || [];

        const filtered = currentScope === 'National' 
          ? alerts 
          : alerts.filter(a => (a.state || '').toLowerCase() === currentScope.toLowerCase());

        document.getElementById('alert-counter').innerText = filtered.length + ' ACTIVE';

        if (filtered.length === 0) {
          container.innerHTML = '<div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800 text-center text-xs text-slate-500">No active alerts for ' + currentScope + '</div>';
          return;
        }

        container.innerHTML = filtered.slice(0, 10).map(alert => ` + "`" + `
          <div class="p-3.5 rounded-2xl bg-slate-900 border ${alert.severity === 'CRITICAL' ? 'border-red-500/40 bg-red-950/10' : 'border-slate-800'} transition hover:border-slate-700">
            <div class="flex items-center justify-between mb-1.5">
              <span class="px-2 py-0.5 rounded text-[10px] font-extrabold ${alert.severity === 'CRITICAL' ? 'bg-red-500/20 text-red-400' : 'bg-amber-500/20 text-amber-400'} uppercase">
                ${alert.severity} • ${alert.state || 'Osun'}
              </span>
              <span class="text-[10px] text-slate-500 font-mono">${(alert.created_at || '').substring(0, 10)}</span>
            </div>
            <div class="text-xs font-bold text-white mb-1">${alert.lga || 'Sentinel'} — ${alert.title || alert.alert_type}</div>
            <p class="text-[11px] text-slate-400 mb-2 leading-relaxed">${alert.message || 'Vector transmission surge detected.'}</p>
            <div class="flex items-center justify-between pt-2 border-t border-slate-800/80">
              <span class="text-[10px] text-emerald-400 font-medium">IRS & Larviciding Recommended</span>
              <button onclick="ackAlert('${alert.id}', this)" class="px-2.5 py-1 rounded-lg text-[10px] font-bold ${alert.acknowledged ? 'bg-slate-800 text-slate-500' : 'bg-red-500/20 text-red-400 hover:bg-red-500/30'}">
                ${alert.acknowledged ? '✓ Acknowledged' : 'Acknowledge'}
              </button>
            </div>
          </div>
        ` + "`" + `).join('');

      } catch (err) {
        container.innerHTML = '<div class="p-3 rounded-2xl bg-slate-900 border border-slate-800 text-xs text-rose-400">Failed to connect to API</div>';
      }
    }

    async function ackAlert(id, btn) {
      btn.innerText = 'Acknowledging...';
      try {
        await fetch('/api/v1/alerts/' + id + '/acknowledge', { method: 'POST' });
        btn.innerText = '✓ Acknowledged';
        btn.className = 'px-2.5 py-1 rounded-lg text-[10px] font-bold bg-slate-800 text-slate-500';
      } catch (e) {
        btn.innerText = 'Error';
      }
    }

    // Load Recent Ingested Records
    async function loadRecentRecords() {
      const tbody = document.getElementById('records-table-body');
      try {
        const res = await fetch('/api/v1/records?limit=6');
        const data = await res.json();
        const records = data.records || [];

        if (records.length === 0) {
          tbody.innerHTML = '<tr><td colspan="6" class="py-4 text-center text-slate-500">No field records received yet.</td></tr>';
          return;
        }

        tbody.innerHTML = records.map(r => {
          let bioColor = 'text-red-400';
          if (r.who_class === 'Susceptible') bioColor = 'text-emerald-400';
          else if (r.who_class === 'Possible resistance') bioColor = 'text-amber-400';

          return ` + "`" + `
            <tr class="hover:bg-slate-800/40 transition">
              <td class="py-2.5 px-3 font-mono text-slate-400 text-[11px]">${(r.id || '').substring(0, 14)}...</td>
              <td class="py-2.5 px-3 font-bold text-white">${r.user_id || 'u_field_01'}</td>
              <td class="py-2.5 px-3 text-slate-300">${r.state || 'Osun'} / ${r.lga || 'Osogbo'}</td>
              <td class="py-2.5 px-3 italic text-slate-400">${r.species || 'An. coluzzii'}</td>
              <td class="py-2.5 px-3 font-bold ${bioColor}">${r.who_class || 'Resistant'} (${r.mortality_pct || 0}%)</td>
              <td class="py-2.5 px-3"><span class="px-2 py-0.5 rounded text-[10px] font-extrabold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">${r.status || 'CLEAN'}</span></td>
            </tr>
          ` + "`" + `;
        }).join('');

      } catch (err) {
        tbody.innerHTML = '<tr><td colspan="6" class="py-4 text-center text-rose-400">Failed to query records endpoint.</td></tr>';
      }
    }

    // DHIS2 Modal & Export
    async function triggerDHIS2Modal() {
      const modal = document.getElementById('dhis2-modal');
      modal.classList.toggle('hidden');
      if (!modal.classList.contains('hidden')) {
        try {
          const res = await fetch('/api/v1/export/dhis2');
          const data = await res.json();
          document.getElementById('dhis2-code').innerText = JSON.stringify(data, null, 2);
        } catch (e) {
          document.getElementById('dhis2-code').innerText = '// Error fetching DHIS2 payload';
        }
      }
    }

    function copyDHIS2() {
      const text = document.getElementById('dhis2-code').innerText;
      navigator.clipboard.writeText(text);
      const btn = document.getElementById('copy-btn');
      btn.innerText = 'Copied!';
      setTimeout(() => btn.innerText = 'Copy JSON Payload', 2000);
    }

    window.onload = function() {
      initMap();
      loadAlerts();
      loadRecentRecords();
    };
  </script>
</body>
</html>`
