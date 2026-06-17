#!/usr/bin/env python3
"""KARVON API Full E2E Audit"""

import requests
import json
import io
import struct
import zlib
import time
from datetime import datetime

BASE = "https://fan.sarbon.me/api/v1"
HEALTH_URL = "https://fan.sarbon.me/health"
OTP = "136092"
PHONE1 = "+998901234567"
PHONE2 = "+998901234568"
ADMIN_LOGIN = "karvonadmin"
ADMIN_PASS = "karvon321321"

results = []
access_token = None
refresh_token = None
admin_token = None
user_id = None
company_id = None
cargo_id = None
warehouse_id = None
payment_id = None
route_id = None
favorite_id = None
notification_id = None
moderator_id = None
category_id = None
pricing_key = None

def log(section, method, path, status, ok, note="", req_body=None, resp_body=None):
    results.append({
        "section": section,
        "method": method,
        "path": path,
        "status": status,
        "ok": ok,
        "note": note,
        "req_body": req_body,
        "resp_body": resp_body,
    })
    icon = "OK" if ok else "FAIL"
    print(f"[{icon}] [{status}] {method} {path} - {note}")

def req(method, url, **kwargs):
    try:
        r = getattr(requests, method)(url, timeout=20, **kwargs)
        try:
            body = r.json()
        except:
            body = r.text[:500]
        return r, body
    except Exception as e:
        return None, str(e)

def auth_headers(token=None):
    t = token or access_token
    return {"Authorization": f"Bearer {t}"} if t else {}

def admin_headers():
    return {"Authorization": f"Bearer {admin_token}"} if admin_token else {}

# ─────────────────────────────────────────────────────────────────────────────
# 1. PUBLIC / HEALTH
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("1. PUBLIC / HEALTH")
print("="*60)

r, b = req("get", HEALTH_URL)
if r:
    ok = r.status_code < 400
    log("Public", "GET", "/health", r.status_code, ok, json.dumps(b)[:100] if isinstance(b, dict) else str(b)[:100], resp_body=b)
else:
    log("Public", "GET", "/health", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/config")
if r:
    ok = r.status_code == 200
    log("Public", "GET", "/config", r.status_code, ok, json.dumps(b)[:120] if isinstance(b, dict) else str(b)[:120], resp_body=b)
else:
    log("Public", "GET", "/config", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/geo/countries", params={"q": "Узб"})
if r:
    ok = r.status_code == 200
    log("Public", "GET", "/geo/countries?q=Узб", r.status_code, ok,
        f"count={len(b.get('data', b)) if isinstance(b, dict) else '?'}", resp_body=b)
else:
    log("Public", "GET", "/geo/countries", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/geo/cities", params={"q": "Таш"})
if r:
    ok = r.status_code == 200
    log("Public", "GET", "/geo/cities?q=Таш", r.status_code, ok,
        f"count={len(b.get('data', b)) if isinstance(b, dict) else '?'}", resp_body=b)
else:
    log("Public", "GET", "/geo/cities", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/payments/packages")
if r:
    ok = r.status_code == 200
    log("Public", "GET", "/payments/packages", r.status_code, ok,
        f"packages={len(b.get('data', [])) if isinstance(b, dict) else '?'}", resp_body=b)
    if ok and isinstance(b, dict):
        packages = b.get("data", [])
        if isinstance(packages, list):
            print(f"  Packages preview: {json.dumps(packages[:2])[:200]}")
        else:
            print(f"  Packages preview (dict): {json.dumps(packages)[:200]}")
else:
    log("Public", "GET", "/payments/packages", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/listings/cargo")
if r:
    ok = r.status_code == 200
    cargo_list = b.get("data", []) if isinstance(b, dict) else []
    if cargo_list and len(cargo_list) > 0:
        first_cargo = cargo_list[0]
        cargo_id = first_cargo.get("id")
        print(f"  Found cargo id for later tests: {cargo_id}")
    log("Public", "GET", "/listings/cargo", r.status_code, ok,
        f"count={len(cargo_list)}", resp_body=b)
else:
    log("Public", "GET", "/listings/cargo", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/warehouses")
if r:
    ok = r.status_code == 200
    wh_list = b.get("data", []) if isinstance(b, dict) else []
    if wh_list and len(wh_list) > 0:
        warehouse_id = wh_list[0].get("id")
        print(f"  Found warehouse id for later tests: {warehouse_id}")
    log("Public", "GET", "/warehouses", r.status_code, ok,
        f"count={len(wh_list)}", resp_body=b)
else:
    log("Public", "GET", "/warehouses", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/search", params={"type": "cargo"})
if r:
    ok = r.status_code == 200
    log("Public", "GET", "/search?type=cargo", r.status_code, ok, str(b)[:100], resp_body=b)
else:
    log("Public", "GET", "/search?type=cargo", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/search", params={"type": "warehouse"})
if r:
    ok = r.status_code == 200
    log("Public", "GET", "/search?type=warehouse", r.status_code, ok, str(b)[:100], resp_body=b)
else:
    log("Public", "GET", "/search?type=warehouse", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/search/cities", params={"q": "Та"})
if r:
    ok = r.status_code == 200
    log("Public", "GET", "/search/cities?q=Та", r.status_code, ok, str(b)[:100], resp_body=b)
else:
    log("Public", "GET", "/search/cities", 0, False, f"EXCEPTION: {b}")

# ─────────────────────────────────────────────────────────────────────────────
# 2. AUTH FLOW
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("2. AUTH FLOW")
print("="*60)

body = {"phone": PHONE1, "channel": "whatsapp"}
r, b = req("post", f"{BASE}/auth/send-otp", json=body)
if r:
    ok = r.status_code in (200, 201)
    log("Auth", "POST", "/auth/send-otp", r.status_code, ok, str(b)[:150], req_body=body, resp_body=b)
    print(f"  send-otp response: {json.dumps(b)[:200]}")
else:
    log("Auth", "POST", "/auth/send-otp", 0, False, f"EXCEPTION: {b}")

body = {"phone": PHONE1, "code": OTP}
r, b = req("post", f"{BASE}/auth/verify-otp", json=body)
if r:
    ok = r.status_code in (200, 201)
    log("Auth", "POST", "/auth/verify-otp", r.status_code, ok,
        f"requires_reg={b.get('data',{}).get('requires_registration','?') if isinstance(b,dict) else '?'}",
        req_body=body, resp_body=b)
    print(f"  verify-otp response: {json.dumps(b)[:300]}")
    if ok and isinstance(b, dict):
        data = b.get("data", {})
        access_token = data.get("access_token")
        refresh_token = data.get("refresh_token")
        requires_reg = data.get("requires_registration", False)
        print(f"  access_token={'SET' if access_token else 'MISSING'}")
        print(f"  requires_registration={requires_reg}")
else:
    log("Auth", "POST", "/auth/verify-otp", 0, False, f"EXCEPTION: {b}")

# Complete registration if needed
if access_token:
    body = {"name": "Test User", "role": "shipper", "city": "Ташкент"}
    r, b = req("post", f"{BASE}/auth/complete-registration", json=body, headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Auth", "POST", "/auth/complete-registration", r.status_code, ok, str(b)[:150], req_body=body, resp_body=b)
        print(f"  complete-registration: {json.dumps(b)[:200]}")
    else:
        log("Auth", "POST", "/auth/complete-registration", 0, False, f"EXCEPTION: {b}")

# GET /users/me
if access_token:
    r, b = req("get", f"{BASE}/users/me", headers=auth_headers())
    if r:
        ok = r.status_code == 200
        log("Auth", "GET", "/users/me", r.status_code, ok, str(b)[:150], resp_body=b)
        if ok and isinstance(b, dict):
            user_data = b.get("data", b)
            user_id = user_data.get("id")
            print(f"  user_id={user_id}")
            print(f"  user profile: {json.dumps(user_data)[:300]}")
    else:
        log("Auth", "GET", "/users/me", 0, False, f"EXCEPTION: {b}")

# Refresh token
if refresh_token:
    body = {"refresh_token": refresh_token}
    r, b = req("post", f"{BASE}/auth/refresh", json=body)
    if r:
        ok = r.status_code in (200, 201)
        log("Auth", "POST", "/auth/refresh", r.status_code, ok, str(b)[:150], req_body=body, resp_body=b)
        if ok and isinstance(b, dict):
            new_data = b.get("data", {})
            if new_data.get("access_token"):
                access_token = new_data["access_token"]
                refresh_token = new_data.get("refresh_token", refresh_token)
                print(f"  Tokens refreshed successfully")
    else:
        log("Auth", "POST", "/auth/refresh", 0, False, f"EXCEPTION: {b}")

# Second user auth
print("\n  [Auth for second user]")
body2 = {"phone": PHONE2, "channel": "whatsapp"}
r, b = req("post", f"{BASE}/auth/send-otp", json=body2)
body2v = {"phone": PHONE2, "code": OTP}
r2, b2 = req("post", f"{BASE}/auth/verify-otp", json=body2v)
user2_token = None
if r2 and r2.status_code in (200, 201) and isinstance(b2, dict):
    d2 = b2.get("data", {})
    user2_token = d2.get("access_token")
    req("post", f"{BASE}/auth/complete-registration",
        json={"name": "Test User 2", "role": "carrier", "city": "Самарканд"},
        headers={"Authorization": f"Bearer {user2_token}"} if user2_token else {})
    print(f"  user2_token={'SET' if user2_token else 'MISSING'}")

# Logout
if access_token:
    r, b = req("post", f"{BASE}/auth/logout", headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201, 204)
        log("Auth", "POST", "/auth/logout", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Auth", "POST", "/auth/logout", 0, False, f"EXCEPTION: {b}")

# Re-login to continue testing
body = {"phone": PHONE1, "code": OTP}
r, b = req("post", f"{BASE}/auth/verify-otp", json=body)
if r and r.status_code in (200, 201) and isinstance(b, dict):
    data = b.get("data", {})
    access_token = data.get("access_token")
    refresh_token = data.get("refresh_token")
    print(f"  Re-login: access_token={'SET' if access_token else 'MISSING'}")

# ─────────────────────────────────────────────────────────────────────────────
# 3. USER PROFILE
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("3. USER PROFILE")
print("="*60)

body = {"name": "Updated Name", "email": "test@test.com"}
r, b = req("put", f"{BASE}/users/me", json=body, headers=auth_headers())
if r:
    ok = r.status_code in (200, 201)
    log("Profile", "PUT", "/users/me", r.status_code, ok, str(b)[:150], req_body=body, resp_body=b)
    print(f"  PUT /users/me: {json.dumps(b)[:200]}")
else:
    log("Profile", "PUT", "/users/me", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/users/me/stats", headers=auth_headers())
if r:
    ok = r.status_code == 200
    log("Profile", "GET", "/users/me/stats", r.status_code, ok, str(b)[:150], resp_body=b)
else:
    log("Profile", "GET", "/users/me/stats", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/users/me/events", headers=auth_headers())
if r:
    ok = r.status_code == 200
    log("Profile", "GET", "/users/me/events", r.status_code, ok, str(b)[:150], resp_body=b)
else:
    log("Profile", "GET", "/users/me/events", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/users/me/tokens", headers=auth_headers())
if r:
    ok = r.status_code == 200
    log("Profile", "GET", "/users/me/tokens", r.status_code, ok, str(b)[:150], resp_body=b)
    if ok and isinstance(b, dict):
        print(f"  token balance: {b.get('data', b)}")
else:
    log("Profile", "GET", "/users/me/tokens", 0, False, f"EXCEPTION: {b}")

# ─────────────────────────────────────────────────────────────────────────────
# 4. COMPANIES
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("4. COMPANIES")
print("="*60)

company_body = {
    "name": "Test QA Company",
    "type": "llc",
    "inn": "123456789",
    "phone": "+998901111111",
    "address": "Ташкент, ул. Тест 1"
}
r, b = req("post", f"{BASE}/companies", json=company_body, headers=auth_headers())
if r:
    ok = r.status_code in (200, 201)
    log("Companies", "POST", "/companies", r.status_code, ok, str(b)[:150], req_body=company_body, resp_body=b)
    print(f"  POST /companies: {json.dumps(b)[:300]}")
    if ok and isinstance(b, dict):
        company_id = b.get("data", {}).get("id")
        print(f"  company_id={company_id}")
else:
    log("Companies", "POST", "/companies", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/companies", headers=auth_headers())
if r:
    ok = r.status_code == 200
    log("Companies", "GET", "/companies", r.status_code, ok, str(b)[:150], resp_body=b)
    if not company_id and ok and isinstance(b, dict):
        companies = b.get("data", [])
        if companies:
            company_id = companies[0].get("id")
            print(f"  Found company_id from list: {company_id}")
else:
    log("Companies", "GET", "/companies", 0, False, f"EXCEPTION: {b}")

if company_id:
    r, b = req("get", f"{BASE}/companies/{company_id}", headers=auth_headers())
    if r:
        ok = r.status_code == 200
        log("Companies", "GET", f"/companies/{company_id}", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Companies", "GET", f"/companies/:id", 0, False, f"EXCEPTION: {b}")

    update_body = {"name": "Updated QA Company", "phone": "+998902222222"}
    r, b = req("put", f"{BASE}/companies/{company_id}", json=update_body, headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Companies", "PUT", f"/companies/{company_id}", r.status_code, ok, str(b)[:150], req_body=update_body, resp_body=b)
    else:
        log("Companies", "PUT", f"/companies/:id", 0, False, f"EXCEPTION: {b}")

# Test: create without required fields
r, b = req("post", f"{BASE}/companies", json={}, headers=auth_headers())
if r:
    ok = r.status_code in (400, 422)
    log("Companies", "POST", "/companies (no fields)", r.status_code, ok,
        f"Validation error: {r.status_code==400 or r.status_code==422}", req_body={}, resp_body=b)
    print(f"  Empty body → {r.status_code}: {json.dumps(b)[:200]}")
else:
    log("Companies", "POST", "/companies (no fields)", 0, False, f"EXCEPTION: {b}")

# ─────────────────────────────────────────────────────────────────────────────
# 5. CARGO
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("5. CARGO")
print("="*60)

cargo_body = {
    "from_city": "Ташкент",
    "to_city": "Самарканд",
    "weight": 1000,
    "volume": 10,
    "price": 500000,
    "currency": "UZS",
    "category": "general",
    "description": "QA Test cargo listing"
}
r, b = req("post", f"{BASE}/listings/cargo", json=cargo_body, headers=auth_headers())
if r:
    ok = r.status_code in (200, 201)
    log("Cargo", "POST", "/listings/cargo", r.status_code, ok, str(b)[:200], req_body=cargo_body, resp_body=b)
    print(f"  POST /listings/cargo: {json.dumps(b)[:300]}")
    if ok and isinstance(b, dict):
        new_cargo_id = b.get("data", {}).get("id")
        if new_cargo_id:
            cargo_id = new_cargo_id
            print(f"  New cargo_id={cargo_id}")
else:
    log("Cargo", "POST", "/listings/cargo", 0, False, f"EXCEPTION: {b}")

# GET with filters
r, b = req("get", f"{BASE}/listings/cargo",
           params={"from_city": "Ташкент", "to_city": "Самарканд", "page": 1, "per_page": 5})
if r:
    ok = r.status_code == 200
    log("Cargo", "GET", "/listings/cargo (filtered)", r.status_code, ok,
        f"count={len(b.get('data',[]) if isinstance(b,dict) else [])}", resp_body=b)
else:
    log("Cargo", "GET", "/listings/cargo (filtered)", 0, False, f"EXCEPTION: {b}")

if cargo_id:
    r, b = req("get", f"{BASE}/listings/cargo/{cargo_id}")
    if r:
        ok = r.status_code == 200
        log("Cargo", "GET", f"/listings/cargo/{cargo_id}", r.status_code, ok, str(b)[:150], resp_body=b)
        print(f"  Cargo detail: {json.dumps(b.get('data',b) if isinstance(b,dict) else b)[:300]}")
    else:
        log("Cargo", "GET", "/listings/cargo/:id", 0, False, f"EXCEPTION: {b}")

    update_cargo = {"price": 600000, "description": "Updated QA cargo"}
    r, b = req("put", f"{BASE}/listings/cargo/{cargo_id}", json=update_cargo, headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Cargo", "PUT", f"/listings/cargo/{cargo_id}", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Cargo", "PUT", "/listings/cargo/:id", 0, False, f"EXCEPTION: {b}")

    status_body = {"status": "archived"}
    r, b = req("patch", f"{BASE}/listings/cargo/{cargo_id}/status", json=status_body, headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Cargo", "PATCH", f"/listings/cargo/{cargo_id}/status", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Cargo", "PATCH", "/listings/cargo/:id/status", 0, False, f"EXCEPTION: {b}")

    r, b = req("get", f"{BASE}/listings/cargo/{cargo_id}/stats", headers=auth_headers())
    if r:
        ok = r.status_code == 200
        log("Cargo", "GET", f"/listings/cargo/{cargo_id}/stats", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Cargo", "GET", "/listings/cargo/:id/stats", 0, False, f"EXCEPTION: {b}")

    r, b = req("post", f"{BASE}/listings/cargo/{cargo_id}/duplicate", headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Cargo", "POST", f"/listings/cargo/{cargo_id}/duplicate", r.status_code, ok, str(b)[:150], resp_body=b)
        print(f"  duplicate: {json.dumps(b)[:200]}")
    else:
        log("Cargo", "POST", "/listings/cargo/:id/duplicate", 0, False, f"EXCEPTION: {b}")

    tmpl_body = {"name": "My QA Template"}
    r, b = req("post", f"{BASE}/listings/cargo/{cargo_id}/template", json=tmpl_body, headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Cargo", "POST", f"/listings/cargo/{cargo_id}/template", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Cargo", "POST", "/listings/cargo/:id/template", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/listings/cargo/templates", headers=auth_headers())
if r:
    ok = r.status_code == 200
    log("Cargo", "GET", "/listings/cargo/templates", r.status_code, ok, str(b)[:150], resp_body=b)
else:
    log("Cargo", "GET", "/listings/cargo/templates", 0, False, f"EXCEPTION: {b}")

# Cross-user access test
if user2_token and cargo_id:
    r, b = req("put", f"{BASE}/listings/cargo/{cargo_id}", json={"price": 999},
               headers={"Authorization": f"Bearer {user2_token}"})
    if r:
        ok = r.status_code == 403
        log("Cargo", "PUT", "/listings/cargo/:id (other user)", r.status_code, ok,
            f"403 expected, got {r.status_code}", resp_body=b)
        print(f"  Cross-user access → {r.status_code}: {json.dumps(b)[:200]}")
    else:
        log("Cargo", "PUT", "/listings/cargo/:id (other user)", 0, False, f"EXCEPTION: {b}")

if cargo_id:
    r, b = req("delete", f"{BASE}/listings/cargo/{cargo_id}", headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201, 204)
        log("Cargo", "DELETE", f"/listings/cargo/{cargo_id}", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Cargo", "DELETE", "/listings/cargo/:id", 0, False, f"EXCEPTION: {b}")

# ─────────────────────────────────────────────────────────────────────────────
# 6. WAREHOUSES
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("6. WAREHOUSES")
print("="*60)

wh_body = {
    "city": "Ташкент",
    "address": "ул. Тест, 1",
    "area": 500,
    "price_per_day": 10000,
    "currency": "UZS",
    "description": "QA Test warehouse"
}
r, b = req("post", f"{BASE}/warehouses", json=wh_body, headers=auth_headers())
if r:
    ok = r.status_code in (200, 201)
    log("Warehouses", "POST", "/warehouses", r.status_code, ok, str(b)[:200], req_body=wh_body, resp_body=b)
    print(f"  POST /warehouses: {json.dumps(b)[:300]}")
    if ok and isinstance(b, dict):
        new_wh_id = b.get("data", {}).get("id")
        if new_wh_id:
            warehouse_id = new_wh_id
            print(f"  New warehouse_id={warehouse_id}")
else:
    log("Warehouses", "POST", "/warehouses", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/warehouses")
if r:
    ok = r.status_code == 200
    log("Warehouses", "GET", "/warehouses", r.status_code, ok,
        f"count={len(b.get('data',[]) if isinstance(b,dict) else [])}", resp_body=b)
else:
    log("Warehouses", "GET", "/warehouses", 0, False, f"EXCEPTION: {b}")

if warehouse_id:
    r, b = req("get", f"{BASE}/warehouses/{warehouse_id}")
    if r:
        ok = r.status_code == 200
        log("Warehouses", "GET", f"/warehouses/{warehouse_id}", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Warehouses", "GET", "/warehouses/:id", 0, False, f"EXCEPTION: {b}")

    update_wh = {"price_per_day": 12000, "description": "Updated QA warehouse"}
    r, b = req("put", f"{BASE}/warehouses/{warehouse_id}", json=update_wh, headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Warehouses", "PUT", f"/warehouses/{warehouse_id}", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Warehouses", "PUT", "/warehouses/:id", 0, False, f"EXCEPTION: {b}")

    r, b = req("patch", f"{BASE}/warehouses/{warehouse_id}/status",
               json={"status": "archived"}, headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Warehouses", "PATCH", f"/warehouses/{warehouse_id}/status", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Warehouses", "PATCH", "/warehouses/:id/status", 0, False, f"EXCEPTION: {b}")

    r, b = req("get", f"{BASE}/warehouses/{warehouse_id}/stats", headers=auth_headers())
    if r:
        ok = r.status_code == 200
        log("Warehouses", "GET", f"/warehouses/{warehouse_id}/stats", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Warehouses", "GET", "/warehouses/:id/stats", 0, False, f"EXCEPTION: {b}")

    r, b = req("delete", f"{BASE}/warehouses/{warehouse_id}", headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201, 204)
        log("Warehouses", "DELETE", f"/warehouses/{warehouse_id}", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Warehouses", "DELETE", "/warehouses/:id", 0, False, f"EXCEPTION: {b}")

# ─────────────────────────────────────────────────────────────────────────────
# 7. CONTACTS & TOKENS
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("7. CONTACTS & TOKENS")
print("="*60)

# Get a public cargo id to test contact view
r, b = req("get", f"{BASE}/listings/cargo", params={"per_page": 5})
pub_cargo_id = None
if r and r.status_code == 200 and isinstance(b, dict):
    items = b.get("data", [])
    if items:
        pub_cargo_id = items[0].get("id")
        print(f"  Using public cargo id: {pub_cargo_id}")

if pub_cargo_id:
    contact_body = {"listing_type": "cargo", "listing_id": pub_cargo_id}
    r, b = req("post", f"{BASE}/contacts/view", json=contact_body, headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201, 402, 403)
        note = str(b)[:150]
        if r.status_code == 402:
            note = "Insufficient tokens (expected)"
        log("Contacts", "POST", "/contacts/view", r.status_code, ok, note, req_body=contact_body, resp_body=b)
        print(f"  contacts/view → {r.status_code}: {json.dumps(b)[:300]}")
    else:
        log("Contacts", "POST", "/contacts/view", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/contacts/history", headers=auth_headers())
if r:
    ok = r.status_code == 200
    log("Contacts", "GET", "/contacts/history", r.status_code, ok, str(b)[:150], resp_body=b)
else:
    log("Contacts", "GET", "/contacts/history", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/users/me/tokens", headers=auth_headers())
if r:
    ok = r.status_code == 200
    log("Contacts", "GET", "/users/me/tokens", r.status_code, ok, str(b)[:150], resp_body=b)
else:
    log("Contacts", "GET", "/users/me/tokens", 0, False, f"EXCEPTION: {b}")

# ─────────────────────────────────────────────────────────────────────────────
# 8. PAYMENTS
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("8. PAYMENTS")
print("="*60)

r, b = req("get", f"{BASE}/payments/packages")
if r:
    ok = r.status_code == 200
    log("Payments", "GET", "/payments/packages", r.status_code, ok, str(b)[:150], resp_body=b)
    if ok and isinstance(b, dict):
        pkgs = b.get("data", {})
        print(f"  packages: {json.dumps(pkgs)[:300]}")
        # data may be {"packages": [...]} or a list
        pkg_list = pkgs.get("packages", []) if isinstance(pkgs, dict) else pkgs
        if pkg_list and isinstance(pkg_list, list):
            first_pkg = pkg_list[0] if isinstance(pkg_list[0], dict) else {}
            pricing_key = first_pkg.get("key") or first_pkg.get("pricing_key")
            print(f"  Using pricing_key={pricing_key}")
else:
    log("Payments", "GET", "/payments/packages", 0, False, f"EXCEPTION: {b}")

pay_body = {"payment_type": "tokens", "pricing_key": pricing_key or "tokens_basic", "currency": "UZS"}
r, b = req("post", f"{BASE}/payments/create", json=pay_body, headers=auth_headers())
if r:
    ok = r.status_code in (200, 201)
    log("Payments", "POST", "/payments/create", r.status_code, ok, str(b)[:200], req_body=pay_body, resp_body=b)
    print(f"  POST /payments/create: {json.dumps(b)[:400]}")
    if ok and isinstance(b, dict):
        payment_id = b.get("data", {}).get("id")
        print(f"  payment_id={payment_id}")
else:
    log("Payments", "POST", "/payments/create", 0, False, f"EXCEPTION: {b}")

if payment_id:
    r, b = req("get", f"{BASE}/payments/{payment_id}", headers=auth_headers())
    if r:
        ok = r.status_code == 200
        log("Payments", "GET", f"/payments/{payment_id}", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Payments", "GET", "/payments/:id", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/payments/history", headers=auth_headers())
if r:
    ok = r.status_code == 200
    log("Payments", "GET", "/payments/history", r.status_code, ok, str(b)[:150], resp_body=b)
else:
    log("Payments", "GET", "/payments/history", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/subscriptions/active", headers=auth_headers())
if r:
    ok = r.status_code == 200
    log("Payments", "GET", "/subscriptions/active", r.status_code, ok, str(b)[:150], resp_body=b)
else:
    log("Payments", "GET", "/subscriptions/active", 0, False, f"EXCEPTION: {b}")

# ─────────────────────────────────────────────────────────────────────────────
# 9. FAVORITES
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("9. FAVORITES")
print("="*60)

# Get a fresh cargo id
r, b = req("get", f"{BASE}/listings/cargo", params={"per_page": 3})
fav_cargo_id = None
if r and r.status_code == 200 and isinstance(b, dict):
    items = b.get("data", [])
    if items:
        fav_cargo_id = items[0].get("id")
        print(f"  Using cargo id for favorite: {fav_cargo_id}")

if fav_cargo_id:
    fav_body = {"listing_type": "cargo", "listing_id": fav_cargo_id}
    r, b = req("post", f"{BASE}/favorites", json=fav_body, headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Favorites", "POST", "/favorites", r.status_code, ok, str(b)[:150], req_body=fav_body, resp_body=b)
        print(f"  POST /favorites: {json.dumps(b)[:200]}")
    else:
        log("Favorites", "POST", "/favorites", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/favorites", headers=auth_headers())
if r:
    ok = r.status_code == 200
    log("Favorites", "GET", "/favorites", r.status_code, ok,
        f"count={len(b.get('data',[]) if isinstance(b,dict) else [])}", resp_body=b)
else:
    log("Favorites", "GET", "/favorites", 0, False, f"EXCEPTION: {b}")

if fav_cargo_id:
    fav_body = {"listing_type": "cargo", "listing_id": fav_cargo_id}
    r, b = req("delete", f"{BASE}/favorites", json=fav_body, headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201, 204)
        log("Favorites", "DELETE", "/favorites", r.status_code, ok, str(b)[:150], req_body=fav_body, resp_body=b)
    else:
        log("Favorites", "DELETE", "/favorites", 0, False, f"EXCEPTION: {b}")

# ─────────────────────────────────────────────────────────────────────────────
# 10. ROUTES
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("10. ROUTES")
print("="*60)

route_body = {"from_city": "Ташкент", "to_city": "Самарканд"}
r, b = req("post", f"{BASE}/routes", json=route_body, headers=auth_headers())
if r:
    ok = r.status_code in (200, 201)
    log("Routes", "POST", "/routes", r.status_code, ok, str(b)[:150], req_body=route_body, resp_body=b)
    print(f"  POST /routes: {json.dumps(b)[:250]}")
    if ok and isinstance(b, dict):
        route_id = b.get("data", {}).get("id")
        print(f"  route_id={route_id}")
else:
    log("Routes", "POST", "/routes", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/routes", headers=auth_headers())
if r:
    ok = r.status_code == 200
    log("Routes", "GET", "/routes", r.status_code, ok,
        f"count={len(b.get('data',[]) if isinstance(b,dict) else [])}", resp_body=b)
    if not route_id and ok and isinstance(b, dict):
        routes = b.get("data", [])
        if routes:
            route_id = routes[0].get("id")
else:
    log("Routes", "GET", "/routes", 0, False, f"EXCEPTION: {b}")

if route_id:
    notif_body = {"enabled": False}
    r, b = req("patch", f"{BASE}/routes/{route_id}/notifications", json=notif_body, headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Routes", "PATCH", f"/routes/{route_id}/notifications", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Routes", "PATCH", "/routes/:id/notifications", 0, False, f"EXCEPTION: {b}")

    r, b = req("delete", f"{BASE}/routes/{route_id}", headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201, 204)
        log("Routes", "DELETE", f"/routes/{route_id}", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Routes", "DELETE", "/routes/:id", 0, False, f"EXCEPTION: {b}")

# ─────────────────────────────────────────────────────────────────────────────
# 11. NOTIFICATIONS
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("11. NOTIFICATIONS")
print("="*60)

r, b = req("get", f"{BASE}/notifications", headers=auth_headers())
if r:
    ok = r.status_code == 200
    log("Notifications", "GET", "/notifications", r.status_code, ok,
        f"count={len(b.get('data',[]) if isinstance(b,dict) else [])}", resp_body=b)
    if ok and isinstance(b, dict):
        notifs = b.get("data", [])
        if notifs:
            notification_id = notifs[0].get("id")
            print(f"  notification_id={notification_id}")
else:
    log("Notifications", "GET", "/notifications", 0, False, f"EXCEPTION: {b}")

r, b = req("get", f"{BASE}/notifications/unread-count", headers=auth_headers())
if r:
    ok = r.status_code == 200
    log("Notifications", "GET", "/notifications/unread-count", r.status_code, ok, str(b)[:150], resp_body=b)
    print(f"  unread-count: {json.dumps(b)[:100]}")
else:
    log("Notifications", "GET", "/notifications/unread-count", 0, False, f"EXCEPTION: {b}")

if notification_id:
    r, b = req("patch", f"{BASE}/notifications/{notification_id}/read", headers=auth_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Notifications", "PATCH", f"/notifications/{notification_id}/read", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Notifications", "PATCH", "/notifications/:id/read", 0, False, f"EXCEPTION: {b}")

r, b = req("patch", f"{BASE}/notifications/read-all", headers=auth_headers())
if r:
    ok = r.status_code in (200, 201)
    log("Notifications", "PATCH", "/notifications/read-all", r.status_code, ok, str(b)[:150], resp_body=b)
else:
    log("Notifications", "PATCH", "/notifications/read-all", 0, False, f"EXCEPTION: {b}")

# ─────────────────────────────────────────────────────────────────────────────
# 12. UPLOAD
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("12. UPLOAD")
print("="*60)

# Create a minimal valid PNG in memory (1x1 red pixel)
def make_minimal_png():
    def make_chunk(chunk_type, data):
        chunk_len = len(data)
        chunk_data = chunk_type + data
        crc = zlib.crc32(chunk_data) & 0xffffffff
        return struct.pack('>I', chunk_len) + chunk_data + struct.pack('>I', crc)

    signature = b'\x89PNG\r\n\x1a\n'
    ihdr_data = struct.pack('>IIBBBBB', 1, 1, 8, 2, 0, 0, 0)
    ihdr = make_chunk(b'IHDR', ihdr_data)
    raw_data = b'\x00\xff\x00\x00'  # filter byte + RGB pixel
    compressed = zlib.compress(raw_data)
    idat = make_chunk(b'IDAT', compressed)
    iend = make_chunk(b'IEND', b'')
    return signature + ihdr + idat + iend

png_data = make_minimal_png()
r, b = req("post", f"{BASE}/upload",
           files={"file": ("test.png", io.BytesIO(png_data), "image/png")},
           headers=auth_headers())
if r:
    ok = r.status_code in (200, 201)
    log("Upload", "POST", "/upload", r.status_code, ok, str(b)[:200], resp_body=b)
    print(f"  upload response: {json.dumps(b)[:300]}")
else:
    log("Upload", "POST", "/upload", 0, False, f"EXCEPTION: {b}")

# ─────────────────────────────────────────────────────────────────────────────
# 13. ADMIN
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("13. ADMIN")
print("="*60)

admin_body = {"login": ADMIN_LOGIN, "password": ADMIN_PASS}
r, b = req("post", f"{BASE}/admin/login", json=admin_body)
if r:
    ok = r.status_code in (200, 201)
    log("Admin", "POST", "/admin/login", r.status_code, ok, str(b)[:150], req_body=admin_body, resp_body=b)
    print(f"  admin/login: {json.dumps(b)[:300]}")
    if ok and isinstance(b, dict):
        admin_data = b.get("data", {})
        admin_token = admin_data.get("access_token") or admin_data.get("token")
        print(f"  admin_token={'SET' if admin_token else 'MISSING'}")
else:
    log("Admin", "POST", "/admin/login", 0, False, f"EXCEPTION: {b}")

if admin_token:
    r, b = req("get", f"{BASE}/admin/dashboard", params={"period": "30d"}, headers=admin_headers())
    if r:
        ok = r.status_code == 200
        log("Admin", "GET", "/admin/dashboard?period=30d", r.status_code, ok, str(b)[:150], resp_body=b)
        print(f"  dashboard: {json.dumps(b)[:300]}")
    else:
        log("Admin", "GET", "/admin/dashboard", 0, False, f"EXCEPTION: {b}")

    r, b = req("get", f"{BASE}/admin/users", params={"page": 1, "per_page": 5}, headers=admin_headers())
    if r:
        ok = r.status_code == 200
        log("Admin", "GET", "/admin/users", r.status_code, ok,
            f"count={len(b.get('data',[]) if isinstance(b,dict) else [])}", resp_body=b)
        admin_user_id = None
        if ok and isinstance(b, dict):
            users = b.get("data", [])
            print(f"  Users list (first 2): {json.dumps(users[:2])[:400]}")
            for u in users:
                if u.get("phone") == PHONE1:
                    admin_user_id = u.get("id")
                    break
            if not admin_user_id and users:
                admin_user_id = users[0].get("id")
            print(f"  admin_user_id={admin_user_id}")
    else:
        log("Admin", "GET", "/admin/users", 0, False, f"EXCEPTION: {b}")
        admin_user_id = None

    if admin_user_id:
        r, b = req("get", f"{BASE}/admin/users/{admin_user_id}", headers=admin_headers())
        if r:
            ok = r.status_code == 200
            log("Admin", "GET", f"/admin/users/{admin_user_id}", r.status_code, ok, str(b)[:150], resp_body=b)
        else:
            log("Admin", "GET", "/admin/users/:id", 0, False, f"EXCEPTION: {b}")

        r, b = req("patch", f"{BASE}/admin/users/{admin_user_id}/block",
                   json={"blocked": True}, headers=admin_headers())
        if r:
            ok = r.status_code in (200, 201)
            log("Admin", "PATCH", f"/admin/users/{admin_user_id}/block (block)", r.status_code, ok, str(b)[:150], resp_body=b)
        else:
            log("Admin", "PATCH", "/admin/users/:id/block", 0, False, f"EXCEPTION: {b}")

        r, b = req("patch", f"{BASE}/admin/users/{admin_user_id}/block",
                   json={"blocked": False}, headers=admin_headers())
        if r:
            ok = r.status_code in (200, 201)
            log("Admin", "PATCH", f"/admin/users/{admin_user_id}/block (unblock)", r.status_code, ok, str(b)[:150], resp_body=b)
        else:
            log("Admin", "PATCH", "/admin/users/:id/unblock", 0, False, f"EXCEPTION: {b}")

        topup_body = {"amount": 10, "reason": "QA test topup"}
        r, b = req("post", f"{BASE}/admin/users/{admin_user_id}/topup",
                   json=topup_body, headers=admin_headers())
        if r:
            ok = r.status_code in (200, 201)
            log("Admin", "POST", f"/admin/users/{admin_user_id}/topup", r.status_code, ok, str(b)[:150], req_body=topup_body, resp_body=b)
            print(f"  topup: {json.dumps(b)[:200]}")
        else:
            log("Admin", "POST", "/admin/users/:id/topup", 0, False, f"EXCEPTION: {b}")

    r, b = req("get", f"{BASE}/admin/companies", params={"page": 1, "per_page": 5}, headers=admin_headers())
    if r:
        ok = r.status_code == 200
        log("Admin", "GET", "/admin/companies", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Admin", "GET", "/admin/companies", 0, False, f"EXCEPTION: {b}")

    r, b = req("get", f"{BASE}/admin/listings", params={"type": "cargo", "page": 1, "per_page": 5}, headers=admin_headers())
    if r:
        ok = r.status_code == 200
        log("Admin", "GET", "/admin/listings?type=cargo", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Admin", "GET", "/admin/listings", 0, False, f"EXCEPTION: {b}")

    r, b = req("get", f"{BASE}/admin/payments", params={"page": 1, "per_page": 5}, headers=admin_headers())
    if r:
        ok = r.status_code == 200
        log("Admin", "GET", "/admin/payments", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Admin", "GET", "/admin/payments", 0, False, f"EXCEPTION: {b}")

    # Pricing CRUD
    r, b = req("get", f"{BASE}/admin/pricing", headers=admin_headers())
    if r:
        ok = r.status_code == 200
        log("Admin", "GET", "/admin/pricing", r.status_code, ok, str(b)[:150], resp_body=b)
        print(f"  pricing: {json.dumps(b)[:400]}")
        existing_keys = []
        if ok and isinstance(b, dict):
            pricing_items = b.get("data", [])
            existing_keys = [p.get("key") for p in pricing_items if p.get("key")]
            print(f"  existing pricing keys: {existing_keys[:5]}")
    else:
        log("Admin", "GET", "/admin/pricing", 0, False, f"EXCEPTION: {b}")
        existing_keys = []

    test_pricing_key = "qa_test_pricing_001"
    new_pricing_body = {
        "key": test_pricing_key,
        "name": "QA Test Package",
        "price": 100000,
        "currency": "UZS",
        "tokens": 5,
        "type": "tokens"
    }
    r, b = req("post", f"{BASE}/admin/pricing", json=new_pricing_body, headers=admin_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Admin", "POST", "/admin/pricing", r.status_code, ok, str(b)[:200], req_body=new_pricing_body, resp_body=b)
        print(f"  POST pricing: {json.dumps(b)[:300]}")
    else:
        log("Admin", "POST", "/admin/pricing", 0, False, f"EXCEPTION: {b}")

    update_pricing = {"name": "QA Updated Package", "price": 120000, "tokens": 6}
    r, b = req("put", f"{BASE}/admin/pricing/{test_pricing_key}", json=update_pricing, headers=admin_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Admin", "PUT", f"/admin/pricing/{test_pricing_key}", r.status_code, ok, str(b)[:200], resp_body=b)
    else:
        log("Admin", "PUT", "/admin/pricing/:key", 0, False, f"EXCEPTION: {b}")

    r, b = req("delete", f"{BASE}/admin/pricing/{test_pricing_key}", headers=admin_headers())
    if r:
        ok = r.status_code in (200, 201, 204)
        log("Admin", "DELETE", f"/admin/pricing/{test_pricing_key}", r.status_code, ok, str(b)[:150], resp_body=b)
    else:
        log("Admin", "DELETE", "/admin/pricing/:key", 0, False, f"EXCEPTION: {b}")

    # Categories CRUD
    r, b = req("get", f"{BASE}/admin/categories", headers=admin_headers())
    if r:
        ok = r.status_code == 200
        log("Admin", "GET", "/admin/categories", r.status_code, ok, str(b)[:150], resp_body=b)
        if ok and isinstance(b, dict):
            cats = b.get("data", [])
            print(f"  categories: {json.dumps(cats[:3])[:300]}")
    else:
        log("Admin", "GET", "/admin/categories", 0, False, f"EXCEPTION: {b}")

    cat_body = {"name": "QA Test Category", "name_uz": "QA Test Kategoriya", "name_en": "QA Test Category"}
    r, b = req("post", f"{BASE}/admin/categories", json=cat_body, headers=admin_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Admin", "POST", "/admin/categories", r.status_code, ok, str(b)[:200], req_body=cat_body, resp_body=b)
        print(f"  POST category: {json.dumps(b)[:250]}")
        if ok and isinstance(b, dict):
            category_id = b.get("data", {}).get("id")
            print(f"  category_id={category_id}")
    else:
        log("Admin", "POST", "/admin/categories", 0, False, f"EXCEPTION: {b}")

    if category_id:
        update_cat = {"name": "Updated QA Category"}
        r, b = req("put", f"{BASE}/admin/categories/{category_id}", json=update_cat, headers=admin_headers())
        if r:
            ok = r.status_code in (200, 201)
            log("Admin", "PUT", f"/admin/categories/{category_id}", r.status_code, ok, str(b)[:150], resp_body=b)
        else:
            log("Admin", "PUT", "/admin/categories/:id", 0, False, f"EXCEPTION: {b}")

        r, b = req("delete", f"{BASE}/admin/categories/{category_id}", headers=admin_headers())
        if r:
            ok = r.status_code in (200, 201, 204)
            log("Admin", "DELETE", f"/admin/categories/{category_id}", r.status_code, ok, str(b)[:150], resp_body=b)
        else:
            log("Admin", "DELETE", "/admin/categories/:id", 0, False, f"EXCEPTION: {b}")

    # Moderators
    mod_body = {"login": "qamod_test_001", "password": "testpass123456"}
    r, b = req("post", f"{BASE}/admin/moderators", json=mod_body, headers=admin_headers())
    if r:
        ok = r.status_code in (200, 201)
        log("Admin", "POST", "/admin/moderators", r.status_code, ok, str(b)[:200], req_body=mod_body, resp_body=b)
        print(f"  POST moderator: {json.dumps(b)[:250]}")
        if ok and isinstance(b, dict):
            moderator_id = b.get("data", {}).get("id")
            print(f"  moderator_id={moderator_id}")
    else:
        log("Admin", "POST", "/admin/moderators", 0, False, f"EXCEPTION: {b}")

    if moderator_id:
        r, b = req("delete", f"{BASE}/admin/moderators/{moderator_id}", headers=admin_headers())
        if r:
            ok = r.status_code in (200, 201, 204)
            log("Admin", "DELETE", f"/admin/moderators/{moderator_id}", r.status_code, ok, str(b)[:150], resp_body=b)
        else:
            log("Admin", "DELETE", "/admin/moderators/:id", 0, False, f"EXCEPTION: {b}")

# ─────────────────────────────────────────────────────────────────────────────
# 14. MODERATOR FLOW
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("14. MODERATOR FLOW")
print("="*60)

# Create temp moderator for testing
mod_login = "qamod_flow_002"
mod_pass = "qapass123456"
mod_token = None
if admin_token:
    r, b = req("post", f"{BASE}/admin/moderators",
               json={"login": mod_login, "password": mod_pass}, headers=admin_headers())
    temp_mod_id = None
    if r and r.status_code in (200, 201) and isinstance(b, dict):
        temp_mod_id = b.get("data", {}).get("id")
        print(f"  Created temp moderator id={temp_mod_id}")

    # Login as moderator (using admin/login endpoint)
    r, b = req("post", f"{BASE}/admin/login", json={"login": mod_login, "password": mod_pass})
    if r:
        ok = r.status_code in (200, 201)
        log("Moderator", "POST", "/admin/login (moderator)", r.status_code, ok, str(b)[:150], resp_body=b)
        print(f"  mod login: {json.dumps(b)[:250]}")
        if ok and isinstance(b, dict):
            mod_data = b.get("data", {})
            mod_token = mod_data.get("access_token") or mod_data.get("token")
    else:
        log("Moderator", "POST", "/admin/login (moderator)", 0, False, f"EXCEPTION: {b}")

if mod_token:
    r, b = req("get", f"{BASE}/moderator/queue",
               headers={"Authorization": f"Bearer {mod_token}"})
    if r:
        ok = r.status_code == 200
        log("Moderator", "GET", "/moderator/queue", r.status_code, ok, str(b)[:150], resp_body=b)
        print(f"  queue: {json.dumps(b)[:300]}")
        if ok and isinstance(b, dict):
            queue_items = b.get("data", [])
            if queue_items:
                queue_item_id = queue_items[0].get("id")
                r2, b2 = req("get", f"{BASE}/moderator/queue/{queue_item_id}",
                             headers={"Authorization": f"Bearer {mod_token}"})
                if r2:
                    ok2 = r2.status_code == 200
                    log("Moderator", "GET", f"/moderator/queue/{queue_item_id}", r2.status_code, ok2, str(b2)[:150], resp_body=b2)
    else:
        log("Moderator", "GET", "/moderator/queue", 0, False, f"EXCEPTION: {b}")

# Cleanup temp moderator
if admin_token and 'temp_mod_id' in dir() and temp_mod_id:
    req("delete", f"{BASE}/admin/moderators/{temp_mod_id}", headers=admin_headers())
    print(f"  Cleaned up temp moderator")

# ─────────────────────────────────────────────────────────────────────────────
# 15. EDGE CASES
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("15. EDGE CASES")
print("="*60)

# Auth endpoints without token
protected_endpoints = [
    ("get", f"{BASE}/users/me"),
    ("get", f"{BASE}/users/me/stats"),
    ("get", f"{BASE}/favorites"),
    ("get", f"{BASE}/routes"),
    ("get", f"{BASE}/notifications"),
    ("get", f"{BASE}/contacts/history"),
    ("get", f"{BASE}/payments/history"),
]
print("  Testing unauthenticated access:")
for method, url in protected_endpoints:
    r, b = req(method, url)  # no auth headers
    if r:
        ok = r.status_code == 401
        path = url.replace(BASE, "")
        log("EdgeCases", method.upper(), f"{path} (no token)", r.status_code, ok,
            f"401 expected, got {r.status_code}", resp_body=b)
        if not ok:
            print(f"  WARN {method.upper()} {path} without token -> {r.status_code} (expected 401)")
    else:
        log("EdgeCases", method.upper(), url, 0, False, f"EXCEPTION: {b}")

# Admin endpoint with user token
print("  Testing user token on admin endpoint:")
if access_token:
    r, b = req("get", f"{BASE}/admin/users", headers=auth_headers())
    if r:
        ok = r.status_code == 403
        log("EdgeCases", "GET", "/admin/users (user token)", r.status_code, ok,
            f"403 expected, got {r.status_code}", resp_body=b)
        print(f"  User token on admin endpoint → {r.status_code}: {json.dumps(b)[:150]}")
    else:
        log("EdgeCases", "GET", "/admin/users (user token)", 0, False, f"EXCEPTION: {b}")

# Invalid UUIDs
print("  Testing invalid UUIDs:")
invalid_id = "not-a-valid-uuid-xyz"
for method, url, body in [
    ("get", f"{BASE}/listings/cargo/{invalid_id}", None),
    ("get", f"{BASE}/warehouses/{invalid_id}", None),
    ("get", f"{BASE}/payments/{invalid_id}", None),
]:
    kwargs = {"headers": auth_headers()}
    if body:
        kwargs["json"] = body
    r, b = req(method, url, **kwargs)
    if r:
        ok = r.status_code in (400, 404)
        path = url.replace(BASE, "")
        log("EdgeCases", method.upper(), f"{path} (invalid uuid)", r.status_code, ok,
            f"400/404 expected, got {r.status_code}", resp_body=b)
        print(f"  Invalid UUID on {path} → {r.status_code}: {json.dumps(b)[:150]}")
    else:
        log("EdgeCases", method.upper(), url, 0, False, f"EXCEPTION: {b}")

# Empty/invalid JSON
print("  Testing invalid JSON bodies:")
invalid_bodies = [
    ("post", f"{BASE}/auth/send-otp", {}),
    ("post", f"{BASE}/auth/verify-otp", {"phone": "not-a-phone", "code": "123"}),
    ("post", f"{BASE}/payments/create", {"payment_type": "invalid_type"}, True),
]
for item in invalid_bodies:
    method, url, body = item[0], item[1], item[2]
    needs_auth = len(item) > 3 and item[3]
    kwargs = {"json": body}
    if needs_auth and access_token:
        kwargs["headers"] = auth_headers()
    r, b = req(method, url, **kwargs)
    if r:
        ok = r.status_code in (400, 422, 401)
        path = url.replace(BASE, "")
        log("EdgeCases", method.upper(), f"{path} (invalid body)", r.status_code, ok,
            f"400/422 expected, got {r.status_code}", resp_body=b)
        print(f"  Invalid body on {path} → {r.status_code}: {json.dumps(b)[:200]}")
    else:
        log("EdgeCases", method.upper(), url, 0, False, f"EXCEPTION: {b}")

# Pagination edge cases
print("  Testing pagination edge cases:")
pagination_tests = [
    f"{BASE}/listings/cargo?page=0&per_page=5",
    f"{BASE}/listings/cargo?page=1&per_page=999",
    f"{BASE}/listings/cargo?page=1&per_page=-1",
]
for url in pagination_tests:
    r, b = req("get", url)
    if r:
        ok = r.status_code in (200, 400)
        path = url.replace(BASE, "")
        log("EdgeCases", "GET", path, r.status_code, ok, str(b)[:100], resp_body=b)
        print(f"  Pagination {path} → {r.status_code}")
    else:
        log("EdgeCases", "GET", url, 0, False, f"EXCEPTION: {b}")

# Lang param test
print("  Testing lang params:")
for lang in ["uz", "en"]:
    r, b = req("post", f"{BASE}/auth/send-otp",
               json={"phone": "+998909999999", "channel": "sms"},
               params={"lang": lang})
    if r:
        ok = r.status_code in (200, 201, 400)
        msg = b.get("message", "") if isinstance(b, dict) else ""
        log("EdgeCases", "POST", f"/auth/send-otp?lang={lang}", r.status_code, ok,
            f"message='{msg[:60]}'", resp_body=b)
        print(f"  lang={lang} → {r.status_code}: message='{msg[:80]}'")
    else:
        log("EdgeCases", "POST", f"/auth/send-otp?lang={lang}", 0, False, f"EXCEPTION: {b}")

# ─────────────────────────────────────────────────────────────────────────────
# REPORT
# ─────────────────────────────────────────────────────────────────────────────
print("\n" + "="*60)
print("GENERATING FINAL REPORT")
print("="*60)

passed = [r for r in results if r["ok"]]
failed = [r for r in results if not r["ok"]]

print(f"\nTotal: {len(results)}, Passed: {len(passed)}, Failed: {len(failed)}")
print("\nFailed endpoints:")
for r in failed:
    print(f"  ❌ [{r['status']}] {r['method']} {r['path']} — {r['note']}")
    if r.get('resp_body'):
        print(f"     Response: {json.dumps(r['resp_body'])[:200]}")

# Save detailed results
with open("/tmp/audit_results.json", "w", encoding="utf-8") as f:
    json.dump(results, f, ensure_ascii=False, indent=2)
print("\n[Detailed results saved to /tmp/audit_results.json]")
