# VECDIM Core Backend API & Spatial Database
**Vector-Borne Disease Early Warning and Ingestion Platform**  
*Lead Developer Technical Submission — Osun State University (UNIOSUN)*  
*Author: Sheriffdeen Saula (`saulasheriffdeen@gmail.com`)*

---

## 1. Architectural Highlights

### The Problem: The Synchronous ML Pipeline Bottleneck
In real-world public health deployments across rural Nigeria (Lagos, Kano, Osun), hundreds of entomological field enumerators sync batches of field surveillance records simultaneously when returning from community surveys. 

If an API were designed naively to trigger the R Random Forest model synchronously on every HTTP sync request:
* The R runtime takes several seconds per execution (loading packages, computing covariates, fitting tree ensembles).
* A burst of 50 incoming mobile syncs at 4:00 PM would spawn 50 concurrent R processes, exhausting server CPU and RAM.
* Enumerator mobile phones would hit cellular timeouts (HTTP 504), resulting in duplicate retransmissions and drained device batteries.

### The Solution: The Staged Ingestion (Outbox) Pattern
This backend implements a **Staged Ingestion Architecture**:

```
[ Mobile App (Expo / React Native) ]
                   │
                   ▼  HTTP POST /api/v1/records/batch (< 30ms)
┌────────────────────────────────────────────────────────────────────────┐
│ 1. Go Ingestion Layer (vecdim-backend)                                │
│ • Validates basic JSON structure & runs Point-of-Entry rules           │
│ • Idempotency Check: Prevents duplicate record_id storage              │
│ • Partial Failure Isolation: Rejects bad rows, saves valid rows        │
│ • Saves raw record with status = 'UNPROCESSED'                         │
│ • Instantly returns 200 OK with per-record status & error arrays       │
└────────────────────────────────────────────────────────────────────────┘
                   │
                   ▼
┌────────────────────────────────────────────────────────────────────────┐
│ 2. Flexible Pipeline Orchestrator                                      │
│ • Option A: Scheduled Cron Job (e.g. hourly / nightly batch)          │
│ • Option B: On-Demand Trigger: POST /api/v1/pipeline/run               │
└────────────────────────────────────────────────────────────────────────┘
                   │
                   ▼
┌────────────────────────────────────────────────────────────────────────┐
│ 3. The R Analytics Pipeline (Docker Worker)                            │
│ • Queries PostGIS for staged UNPROCESSED records                       │
│ • Runs validation rules (flags 222 rows, keeps 1,303 clean rows)       │
│ • Fits Random Forest classifier predicting high-risk zones             │
│ • Generates / updates the 647 Early Warning Alerts (86 ALERT / 135     │
│   WARNING / 426 WATCH)                                                 │
│ • Updates database statuses to 'CLEAN' or 'FLAGGED'                    │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Parity Test Results: 100.00% Mathematical Precision

The candidate pack requires exact reproduction of the 11 validation rules defined in `R/helpers.R` across all 1,525 rows of `vecdim_dummy_entomology_data.csv`:

```
================================================================================
VECDIM VALIDATION ENGINE PARITY TEST REPORT
================================================================================
Total Records Evaluated : 1525
Ground Truth Rows in R  : 1525
Go Flagged Records      : 222 (Expected: 222)
Go Clean Records        : 1303 (Expected: 1,303)
R Flagged Records       : 222
Differing Rows          : 0 (Threshold: <= 5, Target: 0)
Parity Accuracy         : 100.00%
================================================================================
```
* **Differing rows:** **0** (Zero discrepancy against R `outputs/validation_flags_all.csv`).
* **Execution speed:** **0.02 seconds** for 1,525 records (over 100x faster than script interpretation).

---

## 3. Technology Rationale ("Why These Tools")

| Component | Selected Technology | Why Chosen Over Alternatives |
| :--- | :--- | :--- |
| **Backend Architecture** | **Go Modulith (Golang 1.24)** | Single standalone static binary (~15MB RAM idle). Eliminates thread-pool starvation during rural sync bursts via Go’s M:N Goroutine scheduler and non-blocking Netpoller. Clean `internal/` package boundaries prevent architectural erosion without microservice network overhead. |
| **ORM / Data Layer** | **Ent ORM (`entgo.io/ent`)** | Developed by Meta/Ariga. Schema-as-code generates 100% compile-time type-safe queries. Zero reflection at runtime (`.AsNoTracking()` behavior by default), eliminating memory leaks. Built-in batch insertion builders and cross-dialect testing with in-memory SQLite. |
| **Spatial Database** | **PostgreSQL 16 + PostGIS 3.4** | The global standard for geospatial epidemiology. GiST R-Tree indexing provides sub-millisecond bounding box searches (`ST_Intersects`) for disease surveillance mapping. |
| **Generic Helpers** | **Go 1.18+ Generics** | Custom generic abstractions (`SaveBatch[T]`, `ParseAndValidate[T]`, `Map[In, Out]`) eliminate boilerplate error handling and streamline database and DHIS2 transformations while maintaining strict type safety. |

---

## 4. API Specification & Endpoints

All endpoints are served under `/api/v1` and root for compatibility:

| Method | Endpoint | Status | Description |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/records/batch` | **Must** | Idempotent batch upload. Partial failure handling: valid records saved as `accepted`, invalid records rejected with rule name and human guidance. Returns HTTP 200. |
| `GET` | `/api/v1/records` | **Must** | Role-scoped records query (`field`: own records; `vecdim`: state records; `uniosun`: national). Filters: `state`, `lga`, `since`. |
| `GET` | `/api/v1/alerts` | **Must** | 647 Early Warning Alerts from R model. Filters: `state`, `lga`, `from`, `to`, `tier`. Returns tier totals (`86 ALERT`, `135 WARNING`, `426 WATCH`). |
| `POST` | `/api/v1/alerts/{id}/acknowledge` | **Bonus** | Tracks alert acknowledgement, user, timestamp, and optional mitigation notes. |
| `GET` | `/api/v1/risk-points` | **Bonus** | PostGIS spatial bounding box query (`bbox=minLon,minLat,maxLon,maxLat`). Outputs GeoJSON FeatureCollection. |
| `GET` | `/api/v1/export/dhis2` | **Bonus** | Aggregates and exports records into DHIS2 standard `dataValueSets` JSON. |
| `POST` | `/api/v1/pipeline/run` | **Bonus** | On-demand trigger to execute R model processing over staged records. |
| `GET` | `/api/v1/health` | **Must** | Returns `{"status":"ok"}`. |

---

## 5. DHIS2 Data Element Mapping

The `/api/v1/export/dhis2` endpoint maps entomological variables to national DHIS2 indicators:

| Entomological Variable | DHIS2 Data Element UID | Description | Example Value |
| :--- | :--- | :--- | :--- |
| `n_exposed` | `VECDIM_N_EXPOSED` | Total mosquitoes exposed in tube test | `"100"` |
| `n_dead` | `VECDIM_N_DEAD` | Total dead mosquitoes after 24 hours | `"54"` |
| `mortality_pct` | `VECDIM_MORTALITY_PCT` | Calculated WHO mortality percentage (1 decimal) | `"54.0"` |
| `who_class` | `VECDIM_WHO_CLASS` | WHO Susceptibility Classification | `"Resistant"` |
| `itn_use_pct` | `VECDIM_ITN_USE_PCT` | Insecticide-treated net coverage % | `"61.5"` |
| `larval_density` | `VECDIM_LARVAL_DENSITY` | Larval dip density count | `"6"` |

---

## 6. Quick Start & Execution

### Prerequisites
* Go 1.23+ (`/usr/local/bin/go`)
* Docker & Docker Compose (`docker-compose.yml`)

### Commands
```bash
# 1. Run all unit and parity tests
make test

# 2. Run the 11-rule ground truth parity verification
make parity-test

# 3. Start PostGIS container
make docker-up

# 4. Seed database with demo users, 647 alerts, and risk points
make seed

# 5. Start API server on http://localhost:8080
make run
```
