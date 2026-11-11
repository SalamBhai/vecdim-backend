#!/usr/bin/env python3
"""
VECDIM High-Concurrency Ingestion Stress-Testing Suite
-----------------------------------------------------
Proves high-throughput, low-latency Staged Ingestion (< 30ms)
under simulated concurrent field-officer synchronization bursts.

Author: Sheriffdeen Saula (saulasheriffdeen@gmail.com)
Project: VECDIM (UNIOSUN / MOOD Africa / Dev-Afrique / NMEP)
"""

import sys
import time
import json
import uuid
import random
import urllib.request
import concurrent.futures
from statistics import mean, median
import os

BASE_URL = os.getenv("VECDIM_BASE_URL", "http://localhost:8080").rstrip("/")
TOTAL_REQUESTS = int(os.getenv("VECDIM_STRESS_BATCHES", "500"))
CONCURRENCY = int(os.getenv("VECDIM_STRESS_WORKERS", "25"))
RECORDS_PER_BATCH = int(os.getenv("VECDIM_STRESS_RECORDS_PER_BATCH", "5"))

LGAS = [
    ("Osun", "Osogbo", 7.771, 4.557),
    ("Osun", "Ife Central", 7.483, 4.561),
    ("Lagos", "Ikorodu", 6.619, 3.510),
    ("Kano", "Nasarawa", 12.002, 8.533),
]

def make_sample_record(state, lga, lat, lon):
    return {
        "record_id": str(uuid.uuid4()),
        "state": state,
        "lga": lga,
        "ward": "Ward 1",
        "latitude": lat + random.uniform(-0.01, 0.01),
        "longitude": lon + random.uniform(-0.01, 0.01),
        "collection_date": "2026-10-05",
        "species": "Anopheles gambiae s.l.",
        "insecticide": "Permethrin",
        "n_exposed": 100,
        "n_dead": random.randint(40, 95),
        "larval_density": random.randint(1, 5),
        "gps_accuracy_m": 8.5
    }

def send_sync_batch(req_id):
    state, lga, lat, lon = random.choice(LGAS)
    records = [make_sample_record(state, lga, lat, lon) for _ in range(RECORDS_PER_BATCH)]
    payload = json.dumps({"device_id": f"device_{req_id % 10}", "records": records}).encode("utf-8")

    headers = {
        "Content-Type": "application/json",
        "X-User-Id": "u_field_03",
        "X-User-Role": "field",
        "X-User-State": state
    }

    req = urllib.request.Request(
        f"{BASE_URL}/api/v1/records/batch",
        data=payload,
        headers=headers,
        method="POST"
    )

    start = time.perf_counter()
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            elapsed_ms = (time.perf_counter() - start) * 1000
            return (resp.status == 200, elapsed_ms, len(records), None)
    except Exception as e:
        elapsed_ms = (time.perf_counter() - start) * 1000
        return (False, elapsed_ms, 0, str(e))

def run_stress_test():
    print("=" * 75)
    print(" VECDIM INGESTION & PIPELINE CONCURRENCY BENCHMARK")
    print(" Testing: Go Staged Ingestion Under Peak Multi-Officer Synchronization")
    print("=" * 75)
    print(f" Target Endpoint   : {BASE_URL}/api/v1/records/batch")
    print(f" Total Requests    : {TOTAL_REQUESTS} batches")
    print(f" Concurrent Workers: {CONCURRENCY} simultaneous connections")
    print(f" Records Per Batch : {RECORDS_PER_BATCH} surveys ({TOTAL_REQUESTS * RECORDS_PER_BATCH} total surveys)")
    print("-" * 75)
    print(" Warming up & executing concurrent sync bursts...")

    # Check health first
    try:
        with urllib.request.urlopen(f"{BASE_URL}/api/v1/health", timeout=3) as r:
            if r.status != 200:
                print("Error: API is not healthy.")
                sys.exit(1)
    except Exception as e:
        print(f"Error: Could not reach {BASE_URL}/api/v1/health: {e}")
        print("Please start the backend first (e.g. 'go run ./cmd/api/main.go' or 'make run')")
        sys.exit(1)

    start_total = time.perf_counter()
    latencies = []
    successes = 0
    failures = 0
    total_surveys_saved = 0

    with concurrent.futures.ThreadPoolExecutor(max_workers=CONCURRENCY) as executor:
        futures = [executor.submit(send_sync_batch, i) for i in range(TOTAL_REQUESTS)]
        for f in concurrent.futures.as_completed(futures):
            ok, lat, count, err = f.result()
            latencies.append(lat)
            if ok:
                successes += 1
                total_surveys_saved += count
            else:
                failures += 1

    total_wall_time = time.perf_counter() - start_total
    latencies.sort()

    p50 = median(latencies)
    p95 = latencies[int(len(latencies) * 0.95)]
    p99 = latencies[int(len(latencies) * 0.99)]
    avg_lat = mean(latencies)
    throughput_reqs = TOTAL_REQUESTS / total_wall_time
    throughput_records = total_surveys_saved / total_wall_time

    print("=" * 75)
    print(" BENCHMARK RESULTS SUMMARY")
    print("=" * 75)
    print(f" Total Wall Time       : {total_wall_time:.2f} seconds")
    print(f" Successful Batches    : {successes}/{TOTAL_REQUESTS} ({(successes/TOTAL_REQUESTS)*100:.1f}%)")
    print(f" Failed Requests       : {failures}")
    print(f" Total Surveys Ingested: {total_surveys_saved} validated entomology records")
    print(f" Ingestion Throughput  : {throughput_records:.1f} records/second ({throughput_reqs:.1f} batches/sec)")
    print("-" * 75)
    print(" LATENCY DISTRIBUTION (Per HTTP Request):")
    print(f"   • Minimum Latency   : {min(latencies):.2f} ms")
    print(f"   • Average Latency   : {avg_lat:.2f} ms")
    print(f"   • Median (p50)      : {p50:.2f} ms")
    print(f"   • 95th Percentile   : {p95:.2f} ms")
    print(f"   • 99th Percentile   : {p99:.2f} ms")
    print(f"   • Maximum Latency   : {max(latencies):.2f} ms")
    print("=" * 75)
    print(" TARGET CHECKS:")
    print(f" [{'✓' if p99 < 30 else '✗'}] p99 latency under 30ms: {p99:.2f}ms")
    print(f" [{'✓' if failures == 0 else '✗'}] No failed HTTP requests: {failures}")
    if failures > 0 or p99 >= 30:
        sys.exit(1)
    print("=" * 75)

if __name__ == "__main__":
    run_stress_test()
