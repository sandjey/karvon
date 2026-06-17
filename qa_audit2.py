#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""KARVON API Full E2E Audit - v2"""

import requests
import json
import io
import struct
import zlib
import sys

BASE = "https://fan.sarbon.me/api/v1"
HEALTH_URL = "https://fan.sarbon.me/health"
OTP = "136092"
PHONE1 = "+998901234567"
PHONE2 = "+998901234568"
ADMIN_LOGIN = "karvonadmin"
ADMIN_PASS = "karvon321321"

results = []
state = {
    "access_token": None,
    "refresh_token": None,
    "admin_token": None,
    "mod_token": None,
    "user2_token": None,
    "user_id": None,
    "company_id": None,
    "my_cargo_id": None,
    "pub_cargo_id": None,
    "my_warehouse_id": None,
    "pub_warehouse_id": None,
    "payment_id": None,
    "route_id": None,
    "category_id": None,
    "moderator_id": None,
}

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
    icon = "OK  " if ok else "FAIL"
    s = f"  [{icon}] [{status:3d}] {method:6s} {path}"
    if note:
        s += f" => {note}"
    print(s)

def req(method, url, **kwargs):
    try:
        r = getattr(requests, method)(url, timeout=25, **kwargs)
        try:
            body = r.json()
        except Exception:
            body = r.text[:1000]
        return r, body
    except Exception as e:
        return None, f"NET_ERROR: {e}"

def ah():
    """auth headers"""
    t = state["access_token"]
    return {"Authorization": f"Bearer {t}"} if t else {}

def adm():
    """admin headers"""
    t = state["admin_token"]
    return {"Authorization": f"Bearer {t}"} if t else {}

def show(label, b, maxlen=300):
    if isinstance(b, dict):
        print(f"    {label}: {json.dumps(b, ensure_ascii=False)[:maxlen]}")
    else:
        print(f"    {label}: {str(b)[:maxlen]}")

# ================================================================================
print("\n" + "="*70)
print("1. PUBLIC / HEALTH")
print("="*70)

r, b = req("get", HEALTH_URL)
if r:
    log("Public", "GET", "/health", r.status_code, r.status_code < 400, str(b)[:80], resp_body=b)
else:
    log("Public", "GET", "/health", 0, False, b)

r, b = req("get", f"{BASE}/config")
if r:
    keys = list(b.get("data", {}).keys()) if isinstance(b, dict) else []
    log("Public", "GET", "/config", r.status_code, r.status_code == 200, f"keys={keys}", resp_body=b)
else:
    log("Public", "GET", "/config", 0, False, b)

r, b = req("get", f"{BASE}/geo/countries", params={"q": "Узб"})
if r:
    data = b.get("data", []) if isinstance(b, dict) else []
    log("Public", "GET", "/geo/countries?q=Узб", r.status_code, r.status_code == 200,
        f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
    show("countries", b)
else:
    log("Public", "GET", "/geo/countries", 0, False, b)

r, b = req("get", f"{BASE}/geo/cities", params={"q": "Таш"})
if r:
    data = b.get("data", []) if isinstance(b, dict) else []
    log("Public", "GET", "/geo/cities?q=Таш", r.status_code, r.status_code == 200,
        f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
else:
    log("Public", "GET", "/geo/cities", 0, False, b)

r, b = req("get", f"{BASE}/payments/packages")
if r:
    log("Public", "GET", "/payments/packages", r.status_code, r.status_code == 200,
        str(b)[:80], resp_body=b)
    show("packages", b)
else:
    log("Public", "GET", "/payments/packages", 0, False, b)

r, b = req("get", f"{BASE}/listings/cargo")
if r:
    data = b.get("data", []) if isinstance(b, dict) else []
    if isinstance(data, list) and data:
        state["pub_cargo_id"] = data[0].get("id")
    log("Public", "GET", "/listings/cargo", r.status_code, r.status_code == 200,
        f"count={len(data) if isinstance(data, list) else '?'}, first_id={state['pub_cargo_id']}", resp_body=b)
else:
    log("Public", "GET", "/listings/cargo", 0, False, b)

r, b = req("get", f"{BASE}/warehouses")
if r:
    data = b.get("data", []) if isinstance(b, dict) else []
    if isinstance(data, list) and data:
        state["pub_warehouse_id"] = data[0].get("id")
    log("Public", "GET", "/warehouses", r.status_code, r.status_code == 200,
        f"count={len(data) if isinstance(data, list) else '?'}, first_id={state['pub_warehouse_id']}", resp_body=b)
else:
    log("Public", "GET", "/warehouses", 0, False, b)

r, b = req("get", f"{BASE}/search", params={"type": "cargo"})
if r:
    log("Public", "GET", "/search?type=cargo", r.status_code, r.status_code == 200, str(b)[:80], resp_body=b)
else:
    log("Public", "GET", "/search?type=cargo", 0, False, b)

r, b = req("get", f"{BASE}/search", params={"type": "warehouse"})
if r:
    log("Public", "GET", "/search?type=warehouse", r.status_code, r.status_code == 200, str(b)[:80], resp_body=b)
else:
    log("Public", "GET", "/search?type=warehouse", 0, False, b)

r, b = req("get", f"{BASE}/search/cities", params={"q": "Та"})
if r:
    data = b.get("data", []) if isinstance(b, dict) else []
    log("Public", "GET", "/search/cities?q=Та", r.status_code, r.status_code == 200,
        f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
    show("first cities", data[:3] if isinstance(data, list) else data, 200)
else:
    log("Public", "GET", "/search/cities", 0, False, b)

# ================================================================================
print("\n" + "="*70)
print("2. AUTH FLOW")
print("="*70)

body = {"phone": PHONE1, "channel": "whatsapp"}
r, b = req("post", f"{BASE}/auth/send-otp", json=body)
if r:
    ok = r.status_code in (200, 201)
    log("Auth", "POST", "/auth/send-otp", r.status_code, ok,
        b.get("message","?") if isinstance(b, dict) else str(b)[:80],
        req_body=body, resp_body=b)
else:
    log("Auth", "POST", "/auth/send-otp", 0, False, b)

body = {"phone": PHONE1, "code": OTP}
r, b = req("post", f"{BASE}/auth/verify-otp", json=body)
if r:
    ok = r.status_code in (200, 201)
    data = b.get("data", {}) if isinstance(b, dict) else {}
    state["access_token"] = data.get("access_token")
    state["refresh_token"] = data.get("refresh_token")
    requires_reg = data.get("requires_registration", "N/A")
    log("Auth", "POST", "/auth/verify-otp", r.status_code, ok,
        f"requires_registration={requires_reg}, access_token={'SET' if state['access_token'] else 'MISSING'}",
        req_body=body, resp_body=b)
    show("verify-otp", b)
else:
    log("Auth", "POST", "/auth/verify-otp", 0, False, b)

# NOTE: requires_registration flag check
if isinstance(b, dict):
    data = b.get("data", {})
    if "requires_registration" not in data:
        print("    NOTE: 'requires_registration' flag is ABSENT from verify-otp response data!")

body = {"name": "Test User", "role": "shipper", "city": "Ташкент"}
r, b = req("post", f"{BASE}/auth/complete-registration", json=body, headers=ah())
if r:
    ok = r.status_code in (200, 201)
    log("Auth", "POST", "/auth/complete-registration", r.status_code, ok,
        b.get("message","?") if isinstance(b, dict) else str(b)[:80],
        req_body=body, resp_body=b)
else:
    log("Auth", "POST", "/auth/complete-registration", 0, False, b)

r, b = req("get", f"{BASE}/users/me", headers=ah())
if r:
    ok = r.status_code == 200
    data = b.get("data", {}) if isinstance(b, dict) else {}
    state["user_id"] = data.get("id")
    log("Auth", "GET", "/users/me", r.status_code, ok,
        f"id={state['user_id']}, role={data.get('role')}, tokens={data.get('token_balance')}",
        resp_body=b)
    show("profile", data)
else:
    log("Auth", "GET", "/users/me", 0, False, b)

if state["refresh_token"]:
    body = {"refresh_token": state["refresh_token"]}
    r, b = req("post", f"{BASE}/auth/refresh", json=body)
    if r:
        ok = r.status_code in (200, 201)
        data = b.get("data", {}) if isinstance(b, dict) else {}
        if ok and data.get("access_token"):
            state["access_token"] = data["access_token"]
            state["refresh_token"] = data.get("refresh_token", state["refresh_token"])
        log("Auth", "POST", "/auth/refresh", r.status_code, ok,
            f"new_token={'SET' if data.get('access_token') else 'MISSING'}",
            req_body=body, resp_body=b)
    else:
        log("Auth", "POST", "/auth/refresh", 0, False, b)

# Second user
print("  [Second user auth]")
req("post", f"{BASE}/auth/send-otp", json={"phone": PHONE2, "channel": "whatsapp"})
r2, b2 = req("post", f"{BASE}/auth/verify-otp", json={"phone": PHONE2, "code": OTP})
if r2 and r2.status_code in (200, 201) and isinstance(b2, dict):
    d2 = b2.get("data", {})
    state["user2_token"] = d2.get("access_token")
    if state["user2_token"]:
        req("post", f"{BASE}/auth/complete-registration",
            json={"name": "Test User 2", "role": "carrier", "city": "Самарканд"},
            headers={"Authorization": f"Bearer {state['user2_token']}"})
        print(f"    user2_token=SET")

r, b = req("post", f"{BASE}/auth/logout", headers=ah())
if r:
    ok = r.status_code in (200, 201, 204)
    log("Auth", "POST", "/auth/logout", r.status_code, ok,
        b.get("message","?") if isinstance(b, dict) else str(b)[:80], resp_body=b)
else:
    log("Auth", "POST", "/auth/logout", 0, False, b)

# Re-auth after logout
r, b = req("post", f"{BASE}/auth/verify-otp", json={"phone": PHONE1, "code": OTP})
if r and r.status_code in (200, 201) and isinstance(b, dict):
    data = b.get("data", {})
    state["access_token"] = data.get("access_token")
    state["refresh_token"] = data.get("refresh_token")
    print(f"    Re-auth: access_token={'SET' if state['access_token'] else 'MISSING'}")

# ================================================================================
print("\n" + "="*70)
print("3. USER PROFILE")
print("="*70)

body = {"name": "Updated Name", "email": "test@test.com"}
r, b = req("put", f"{BASE}/users/me", json=body, headers=ah())
if r:
    ok = r.status_code in (200, 201)
    log("Profile", "PUT", "/users/me", r.status_code, ok,
        b.get("message","?") if isinstance(b, dict) else str(b)[:80],
        req_body=body, resp_body=b)
    show("updated profile", b.get("data",b) if isinstance(b,dict) else b)
else:
    log("Profile", "PUT", "/users/me", 0, False, b)

r, b = req("get", f"{BASE}/users/me/stats", headers=ah())
if r:
    ok = r.status_code == 200
    log("Profile", "GET", "/users/me/stats", r.status_code, ok,
        str(b.get("data","?"))[:100] if isinstance(b,dict) else str(b)[:80], resp_body=b)
    show("stats", b)
else:
    log("Profile", "GET", "/users/me/stats", 0, False, b)

r, b = req("get", f"{BASE}/users/me/events", headers=ah())
if r:
    ok = r.status_code == 200
    log("Profile", "GET", "/users/me/events", r.status_code, ok,
        str(b)[:100] if not ok else f"count={len(b.get('data',[]) if isinstance(b,dict) else [])}", resp_body=b)
    if not ok:
        show("events error", b)
else:
    log("Profile", "GET", "/users/me/events", 0, False, b)

r, b = req("get", f"{BASE}/users/me/tokens", headers=ah())
if r:
    ok = r.status_code == 200
    data = b.get("data", {}) if isinstance(b, dict) else {}
    log("Profile", "GET", "/users/me/tokens", r.status_code, ok,
        f"balance={data.get('token_balance','?')}", resp_body=b)
else:
    log("Profile", "GET", "/users/me/tokens", 0, False, b)

# ================================================================================
print("\n" + "="*70)
print("4. COMPANIES")
print("="*70)

company_body = {
    "name": "Test QA Company",
    "type": "llc",
    "inn": "123456789",
    "phone": "+998901111111",
    "address": "Ташкент, ул. Тест 1"
}
r, b = req("post", f"{BASE}/companies", json=company_body, headers=ah())
if r:
    ok = r.status_code in (200, 201)
    data = b.get("data", {}) if isinstance(b, dict) else {}
    if ok:
        state["company_id"] = data.get("id")
    log("Companies", "POST", "/companies", r.status_code, ok,
        f"id={state['company_id']}" if ok else (b.get("error",{}).get("message","?") if isinstance(b,dict) else str(b)[:80]),
        req_body=company_body, resp_body=b)
    show("create company", b)
else:
    log("Companies", "POST", "/companies", 0, False, b)

r, b = req("get", f"{BASE}/companies", headers=ah())
if r:
    ok = r.status_code == 200
    data = b.get("data", []) if isinstance(b, dict) else []
    if not state["company_id"] and isinstance(data, list) and data:
        state["company_id"] = data[0].get("id")
        print(f"    Using company from list: {state['company_id']}")
    log("Companies", "GET", "/companies", r.status_code, ok,
        f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
else:
    log("Companies", "GET", "/companies", 0, False, b)

if state["company_id"]:
    r, b = req("get", f"{BASE}/companies/{state['company_id']}", headers=ah())
    if r:
        ok = r.status_code == 200
        log("Companies", "GET", f"/companies/:id", r.status_code, ok,
            str(b.get("data",{}).get("name","?") if isinstance(b,dict) else b)[:80], resp_body=b)
    else:
        log("Companies", "GET", "/companies/:id", 0, False, b)

    update_body = {"name": "Updated QA Company", "phone": "+998902222222"}
    r, b = req("put", f"{BASE}/companies/{state['company_id']}", json=update_body, headers=ah())
    if r:
        ok = r.status_code in (200, 201)
        log("Companies", "PUT", f"/companies/:id", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80],
            req_body=update_body, resp_body=b)
        show("updated company", b)
    else:
        log("Companies", "PUT", "/companies/:id", 0, False, b)

# Empty body validation
r, b = req("post", f"{BASE}/companies", json={}, headers=ah())
if r:
    ok = r.status_code in (400, 422)
    log("Companies", "POST", "/companies (empty body)", r.status_code, ok,
        f"got {r.status_code}, expected 400/422" + (" - VALIDATION OK" if ok else " - VALIDATION MISSING"),
        req_body={}, resp_body=b)
    show("empty body resp", b)
else:
    log("Companies", "POST", "/companies (empty body)", 0, False, b)

# ================================================================================
print("\n" + "="*70)
print("5. CARGO")
print("="*70)

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
r, b = req("post", f"{BASE}/listings/cargo", json=cargo_body, headers=ah())
if r:
    ok = r.status_code in (200, 201)
    data = b.get("data", {}) if isinstance(b, dict) else {}
    if ok and data.get("id"):
        state["my_cargo_id"] = data["id"]
    log("Cargo", "POST", "/listings/cargo", r.status_code, ok,
        f"id={state['my_cargo_id']}" if ok else (b.get("error",{}).get("message","?") if isinstance(b,dict) else str(b)[:80]),
        req_body=cargo_body, resp_body=b)
    if not ok:
        show("cargo create error", b)
else:
    log("Cargo", "POST", "/listings/cargo", 0, False, b)

# Filtered list
r, b = req("get", f"{BASE}/listings/cargo",
           params={"from_city": "Ташкент", "to_city": "Самарканд", "page": 1, "per_page": 5})
if r:
    data = b.get("data", []) if isinstance(b, dict) else []
    log("Cargo", "GET", "/listings/cargo (filtered)", r.status_code, r.status_code == 200,
        f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
else:
    log("Cargo", "GET", "/listings/cargo (filtered)", 0, False, b)

# Use my cargo if available, else fall back to public
test_cargo_id = state["my_cargo_id"] or state["pub_cargo_id"]
if test_cargo_id:
    r, b = req("get", f"{BASE}/listings/cargo/{test_cargo_id}")
    if r:
        ok = r.status_code == 200
        data = b.get("data", {}) if isinstance(b, dict) else {}
        log("Cargo", "GET", "/listings/cargo/:id", r.status_code, ok,
            f"from={data.get('from_city','?')} to={data.get('to_city','?')}", resp_body=b)
    else:
        log("Cargo", "GET", "/listings/cargo/:id", 0, False, b)

if state["my_cargo_id"]:
    cid = state["my_cargo_id"]

    update_cargo = {"price": 600000, "description": "Updated QA cargo"}
    r, b = req("put", f"{BASE}/listings/cargo/{cid}", json=update_cargo, headers=ah())
    if r:
        ok = r.status_code in (200, 201)
        log("Cargo", "PUT", "/listings/cargo/:id", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
        if not ok: show("PUT cargo error", b)
    else:
        log("Cargo", "PUT", "/listings/cargo/:id", 0, False, b)

    r, b = req("patch", f"{BASE}/listings/cargo/{cid}/status",
               json={"status": "archived"}, headers=ah())
    if r:
        ok = r.status_code in (200, 201)
        log("Cargo", "PATCH", "/listings/cargo/:id/status", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
        if not ok: show("status error", b)
    else:
        log("Cargo", "PATCH", "/listings/cargo/:id/status", 0, False, b)

    r, b = req("get", f"{BASE}/listings/cargo/{cid}/stats", headers=ah())
    if r:
        ok = r.status_code == 200
        log("Cargo", "GET", "/listings/cargo/:id/stats", r.status_code, ok,
            str(b.get("data","?"))[:80] if isinstance(b,dict) else str(b)[:80], resp_body=b)
    else:
        log("Cargo", "GET", "/listings/cargo/:id/stats", 0, False, b)

    r, b = req("post", f"{BASE}/listings/cargo/{cid}/duplicate", headers=ah())
    if r:
        ok = r.status_code in (200, 201)
        data = b.get("data", {}) if isinstance(b, dict) else {}
        log("Cargo", "POST", "/listings/cargo/:id/duplicate", r.status_code, ok,
            f"new_id={data.get('id','?')}" if ok else (b.get("error",{}).get("message","?") if isinstance(b,dict) else str(b)[:80]),
            resp_body=b)
    else:
        log("Cargo", "POST", "/listings/cargo/:id/duplicate", 0, False, b)

    r, b = req("post", f"{BASE}/listings/cargo/{cid}/template",
               json={"name": "My QA Template"}, headers=ah())
    if r:
        ok = r.status_code in (200, 201)
        log("Cargo", "POST", "/listings/cargo/:id/template", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
        if not ok: show("template error", b)
    else:
        log("Cargo", "POST", "/listings/cargo/:id/template", 0, False, b)
else:
    print("  SKIPPED: cargo create/edit/delete (no owned cargo)")

r, b = req("get", f"{BASE}/listings/cargo/templates", headers=ah())
if r:
    ok = r.status_code == 200
    data = b.get("data", []) if isinstance(b, dict) else []
    log("Cargo", "GET", "/listings/cargo/templates", r.status_code, ok,
        f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
else:
    log("Cargo", "GET", "/listings/cargo/templates", 0, False, b)

# Cross-user access
if state["user2_token"] and test_cargo_id:
    r, b = req("put", f"{BASE}/listings/cargo/{test_cargo_id}",
               json={"price": 999},
               headers={"Authorization": f"Bearer {state['user2_token']}"})
    if r:
        ok = r.status_code == 403
        log("Cargo", "PUT", "/listings/cargo/:id (user2 = other owner)", r.status_code, ok,
            f"403 expected, got {r.status_code}" + (" - ACCESS CONTROL OK" if ok else " - ACCESS CONTROL BROKEN!"),
            resp_body=b)
        show("cross-user resp", b)
    else:
        log("Cargo", "PUT", "/listings/cargo/:id (user2)", 0, False, b)

if state["my_cargo_id"]:
    r, b = req("delete", f"{BASE}/listings/cargo/{state['my_cargo_id']}", headers=ah())
    if r:
        ok = r.status_code in (200, 201, 204)
        log("Cargo", "DELETE", "/listings/cargo/:id", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
    else:
        log("Cargo", "DELETE", "/listings/cargo/:id", 0, False, b)

# ================================================================================
print("\n" + "="*70)
print("6. WAREHOUSES")
print("="*70)

wh_body = {
    "city": "Ташкент",
    "address": "ул. Тест, 1",
    "area": 500,
    "price_per_day": 10000,
    "currency": "UZS",
    "description": "QA Test warehouse"
}
r, b = req("post", f"{BASE}/warehouses", json=wh_body, headers=ah())
if r:
    ok = r.status_code in (200, 201)
    data = b.get("data", {}) if isinstance(b, dict) else {}
    if ok and data.get("id"):
        state["my_warehouse_id"] = data["id"]
    log("Warehouses", "POST", "/warehouses", r.status_code, ok,
        f"id={state['my_warehouse_id']}" if ok else (b.get("error",{}).get("message","?") if isinstance(b,dict) else str(b)[:80]),
        req_body=wh_body, resp_body=b)
    if not ok: show("wh create error", b)
else:
    log("Warehouses", "POST", "/warehouses", 0, False, b)

r, b = req("get", f"{BASE}/warehouses")
if r:
    data = b.get("data", []) if isinstance(b, dict) else []
    log("Warehouses", "GET", "/warehouses", r.status_code, r.status_code == 200,
        f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
else:
    log("Warehouses", "GET", "/warehouses", 0, False, b)

test_wh_id = state["my_warehouse_id"] or state["pub_warehouse_id"]
if test_wh_id:
    r, b = req("get", f"{BASE}/warehouses/{test_wh_id}")
    if r:
        ok = r.status_code == 200
        data = b.get("data", {}) if isinstance(b, dict) else {}
        log("Warehouses", "GET", "/warehouses/:id", r.status_code, ok,
            f"city={data.get('city','?')}", resp_body=b)
    else:
        log("Warehouses", "GET", "/warehouses/:id", 0, False, b)

if state["my_warehouse_id"]:
    wid = state["my_warehouse_id"]

    r, b = req("put", f"{BASE}/warehouses/{wid}",
               json={"price_per_day": 12000, "description": "Updated QA warehouse"}, headers=ah())
    if r:
        ok = r.status_code in (200, 201)
        log("Warehouses", "PUT", "/warehouses/:id", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
        if not ok: show("PUT wh error", b)
    else:
        log("Warehouses", "PUT", "/warehouses/:id", 0, False, b)

    r, b = req("patch", f"{BASE}/warehouses/{wid}/status",
               json={"status": "archived"}, headers=ah())
    if r:
        ok = r.status_code in (200, 201)
        log("Warehouses", "PATCH", "/warehouses/:id/status", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
    else:
        log("Warehouses", "PATCH", "/warehouses/:id/status", 0, False, b)

    r, b = req("get", f"{BASE}/warehouses/{wid}/stats", headers=ah())
    if r:
        ok = r.status_code == 200
        log("Warehouses", "GET", "/warehouses/:id/stats", r.status_code, ok,
            str(b.get("data","?"))[:80] if isinstance(b,dict) else str(b)[:80], resp_body=b)
    else:
        log("Warehouses", "GET", "/warehouses/:id/stats", 0, False, b)

    r, b = req("delete", f"{BASE}/warehouses/{wid}", headers=ah())
    if r:
        ok = r.status_code in (200, 201, 204)
        log("Warehouses", "DELETE", "/warehouses/:id", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
    else:
        log("Warehouses", "DELETE", "/warehouses/:id", 0, False, b)
else:
    print("  SKIPPED: warehouse edit/delete (no owned warehouse)")

# ================================================================================
print("\n" + "="*70)
print("7. CONTACTS & TOKENS")
print("="*70)

if state["pub_cargo_id"]:
    body = {"listing_type": "cargo", "listing_id": state["pub_cargo_id"]}
    r, b = req("post", f"{BASE}/contacts/view", json=body, headers=ah())
    if r:
        ok = r.status_code in (200, 201, 402)
        if r.status_code == 402:
            note = "INSUFFICIENT_TOKENS (expected if balance=0)"
        elif r.status_code == 200:
            note = f"contact revealed: {b.get('data',{}).get('phone','?') if isinstance(b,dict) else '?'}"
        else:
            note = f"got {r.status_code}: {b.get('error',{}).get('message','?') if isinstance(b,dict) else str(b)[:80]}"
        log("Contacts", "POST", "/contacts/view", r.status_code, ok, note, req_body=body, resp_body=b)
        show("contact view", b)
    else:
        log("Contacts", "POST", "/contacts/view", 0, False, b)

r, b = req("get", f"{BASE}/contacts/history", headers=ah())
if r:
    data = b.get("data", []) if isinstance(b, dict) else []
    log("Contacts", "GET", "/contacts/history", r.status_code, r.status_code == 200,
        f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
else:
    log("Contacts", "GET", "/contacts/history", 0, False, b)

r, b = req("get", f"{BASE}/users/me/tokens", headers=ah())
if r:
    ok = r.status_code == 200
    data = b.get("data", {}) if isinstance(b, dict) else {}
    log("Contacts", "GET", "/users/me/tokens (post-view)", r.status_code, ok,
        f"balance={data.get('token_balance','?')}", resp_body=b)
else:
    log("Contacts", "GET", "/users/me/tokens", 0, False, b)

# ================================================================================
print("\n" + "="*70)
print("8. PAYMENTS")
print("="*70)

r, b = req("get", f"{BASE}/payments/packages")
if r:
    ok = r.status_code == 200
    log("Payments", "GET", "/payments/packages", r.status_code, ok, str(b)[:80], resp_body=b)
    pricing_key = None
    if ok and isinstance(b, dict):
        pkgs = b.get("data", {})
        pkg_list = pkgs.get("packages", []) if isinstance(pkgs, dict) else (pkgs if isinstance(pkgs, list) else [])
        if pkg_list:
            first = pkg_list[0] if isinstance(pkg_list[0], dict) else {}
            pricing_key = first.get("key")
            print(f"    pricing_key={pricing_key}, packages={len(pkg_list)}")
            show("all packages", pkg_list)
else:
    log("Payments", "GET", "/payments/packages", 0, False, b)
    pricing_key = None

pay_body = {"payment_type": "tokens", "pricing_key": pricing_key or "tokens_basic", "currency": "UZS"}
r, b = req("post", f"{BASE}/payments/create", json=pay_body, headers=ah())
if r:
    ok = r.status_code in (200, 201)
    data = b.get("data", {}) if isinstance(b, dict) else {}
    if ok:
        state["payment_id"] = data.get("id")
    log("Payments", "POST", "/payments/create", r.status_code, ok,
        f"payment_id={state['payment_id']}, payment_url={str(data.get('payment_url',''))[:60]}" if ok
        else (b.get("error",{}).get("message","?") if isinstance(b,dict) else str(b)[:80]),
        req_body=pay_body, resp_body=b)
    show("payment create", b)
else:
    log("Payments", "POST", "/payments/create", 0, False, b)

if state["payment_id"]:
    r, b = req("get", f"{BASE}/payments/{state['payment_id']}", headers=ah())
    if r:
        ok = r.status_code == 200
        data = b.get("data", {}) if isinstance(b, dict) else {}
        log("Payments", "GET", "/payments/:id", r.status_code, ok,
            f"status={data.get('status','?')}, amount={data.get('amount','?')}", resp_body=b)
        show("payment detail", data)
    else:
        log("Payments", "GET", "/payments/:id", 0, False, b)

r, b = req("get", f"{BASE}/payments/history", headers=ah())
if r:
    data = b.get("data", []) if isinstance(b, dict) else []
    log("Payments", "GET", "/payments/history", r.status_code, r.status_code == 200,
        f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
else:
    log("Payments", "GET", "/payments/history", 0, False, b)

r, b = req("get", f"{BASE}/subscriptions/active", headers=ah())
if r:
    ok = r.status_code == 200
    log("Payments", "GET", "/subscriptions/active", r.status_code, ok, str(b)[:100], resp_body=b)
else:
    log("Payments", "GET", "/subscriptions/active", 0, False, b)

# ================================================================================
print("\n" + "="*70)
print("9. FAVORITES")
print("="*70)

fav_id = state["pub_cargo_id"]
if fav_id:
    body = {"listing_type": "cargo", "listing_id": fav_id}
    r, b = req("post", f"{BASE}/favorites", json=body, headers=ah())
    if r:
        ok = r.status_code in (200, 201)
        log("Favorites", "POST", "/favorites", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], req_body=body, resp_body=b)
        show("add fav", b)
    else:
        log("Favorites", "POST", "/favorites", 0, False, b)

r, b = req("get", f"{BASE}/favorites", headers=ah())
if r:
    data = b.get("data", []) if isinstance(b, dict) else []
    log("Favorites", "GET", "/favorites", r.status_code, r.status_code == 200,
        f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
else:
    log("Favorites", "GET", "/favorites", 0, False, b)

if fav_id:
    body = {"listing_type": "cargo", "listing_id": fav_id}
    r, b = req("delete", f"{BASE}/favorites", json=body, headers=ah())
    if r:
        ok = r.status_code in (200, 201, 204)
        log("Favorites", "DELETE", "/favorites", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], req_body=body, resp_body=b)
    else:
        log("Favorites", "DELETE", "/favorites", 0, False, b)

# ================================================================================
print("\n" + "="*70)
print("10. ROUTES")
print("="*70)

body = {"from_city": "Ташкент", "to_city": "Самарканд"}
r, b = req("post", f"{BASE}/routes", json=body, headers=ah())
if r:
    ok = r.status_code in (200, 201)
    data = b.get("data", {}) if isinstance(b, dict) else {}
    if ok:
        state["route_id"] = data.get("id")
    log("Routes", "POST", "/routes", r.status_code, ok,
        f"id={state['route_id']}" if ok else (b.get("error",{}).get("message","?") if isinstance(b,dict) else str(b)[:80]),
        req_body=body, resp_body=b)
    show("route", b)
else:
    log("Routes", "POST", "/routes", 0, False, b)

r, b = req("get", f"{BASE}/routes", headers=ah())
if r:
    data = b.get("data", []) if isinstance(b, dict) else []
    if not state["route_id"] and isinstance(data, list) and data:
        state["route_id"] = data[0].get("id")
    log("Routes", "GET", "/routes", r.status_code, r.status_code == 200,
        f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
else:
    log("Routes", "GET", "/routes", 0, False, b)

if state["route_id"]:
    r, b = req("patch", f"{BASE}/routes/{state['route_id']}/notifications",
               json={"enabled": False}, headers=ah())
    if r:
        ok = r.status_code in (200, 201)
        log("Routes", "PATCH", "/routes/:id/notifications", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
    else:
        log("Routes", "PATCH", "/routes/:id/notifications", 0, False, b)

    r, b = req("delete", f"{BASE}/routes/{state['route_id']}", headers=ah())
    if r:
        ok = r.status_code in (200, 201, 204)
        log("Routes", "DELETE", "/routes/:id", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
    else:
        log("Routes", "DELETE", "/routes/:id", 0, False, b)

# ================================================================================
print("\n" + "="*70)
print("11. NOTIFICATIONS")
print("="*70)

r, b = req("get", f"{BASE}/notifications", headers=ah())
if r:
    ok = r.status_code == 200
    data = b.get("data", []) if isinstance(b, dict) else []
    notif_id = data[0].get("id") if isinstance(data, list) and data else None
    log("Notifications", "GET", "/notifications", r.status_code, ok,
        f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
else:
    log("Notifications", "GET", "/notifications", 0, False, b)
    notif_id = None

r, b = req("get", f"{BASE}/notifications/unread-count", headers=ah())
if r:
    ok = r.status_code == 200
    log("Notifications", "GET", "/notifications/unread-count", r.status_code, ok,
        str(b.get("data","?"))[:80] if isinstance(b,dict) else str(b)[:80], resp_body=b)
    show("unread-count", b)
else:
    log("Notifications", "GET", "/notifications/unread-count", 0, False, b)

if notif_id:
    r, b = req("patch", f"{BASE}/notifications/{notif_id}/read", headers=ah())
    if r:
        ok = r.status_code in (200, 201)
        log("Notifications", "PATCH", "/notifications/:id/read", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
    else:
        log("Notifications", "PATCH", "/notifications/:id/read", 0, False, b)
else:
    print("  SKIPPED: /notifications/:id/read (no notifications)")

r, b = req("patch", f"{BASE}/notifications/read-all", headers=ah())
if r:
    ok = r.status_code in (200, 201)
    log("Notifications", "PATCH", "/notifications/read-all", r.status_code, ok,
        b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
else:
    log("Notifications", "PATCH", "/notifications/read-all", 0, False, b)

# ================================================================================
print("\n" + "="*70)
print("12. UPLOAD")
print("="*70)

def make_minimal_png():
    def make_chunk(chunk_type, data):
        chunk_data = chunk_type + data
        crc = zlib.crc32(chunk_data) & 0xffffffff
        return struct.pack('>I', len(data)) + chunk_data + struct.pack('>I', crc)
    signature = b'\x89PNG\r\n\x1a\n'
    ihdr_data = struct.pack('>IIBBBBB', 1, 1, 8, 2, 0, 0, 0)
    ihdr = make_chunk(b'IHDR', ihdr_data)
    raw_data = b'\x00\xff\x00\x00'
    idat = make_chunk(b'IDAT', zlib.compress(raw_data))
    iend = make_chunk(b'IEND', b'')
    return signature + ihdr + idat + iend

png_data = make_minimal_png()
r, b = req("post", f"{BASE}/upload",
           files={"file": ("test.png", io.BytesIO(png_data), "image/png")},
           headers=ah())
if r:
    ok = r.status_code in (200, 201)
    data = b.get("data", {}) if isinstance(b, dict) else {}
    log("Upload", "POST", "/upload", r.status_code, ok,
        f"url={data.get('url','?')}" if ok else (b.get("error",{}).get("message","?") if isinstance(b,dict) else str(b)[:80]),
        resp_body=b)
    show("upload resp", b)
else:
    log("Upload", "POST", "/upload", 0, False, b)

# ================================================================================
print("\n" + "="*70)
print("13. ADMIN")
print("="*70)

r, b = req("post", f"{BASE}/admin/login",
           json={"login": ADMIN_LOGIN, "password": ADMIN_PASS})
if r:
    ok = r.status_code in (200, 201)
    data = b.get("data", {}) if isinstance(b, dict) else {}
    state["admin_token"] = data.get("access_token") or data.get("token")
    log("Admin", "POST", "/admin/login", r.status_code, ok,
        f"admin_token={'SET' if state['admin_token'] else 'MISSING'}", resp_body=b)
    show("admin login", b)
else:
    log("Admin", "POST", "/admin/login", 0, False, b)

if state["admin_token"]:
    r, b = req("get", f"{BASE}/admin/dashboard", params={"period": "30d"}, headers=adm())
    if r:
        ok = r.status_code == 200
        log("Admin", "GET", "/admin/dashboard?period=30d", r.status_code, ok,
            str(list(b.get("data",{}).keys()))[:80] if isinstance(b,dict) else str(b)[:80], resp_body=b)
        show("dashboard", b.get("data",b) if isinstance(b,dict) else b, 400)
    else:
        log("Admin", "GET", "/admin/dashboard", 0, False, b)

    r, b = req("get", f"{BASE}/admin/users", params={"page": 1, "per_page": 5}, headers=adm())
    found_user_id = None
    if r:
        ok = r.status_code == 200
        data = b.get("data", []) if isinstance(b, dict) else []
        log("Admin", "GET", "/admin/users", r.status_code, ok,
            f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
        if ok and isinstance(data, list):
            for u in data:
                if u.get("phone") == PHONE1:
                    found_user_id = u.get("id")
                    break
            if not found_user_id and data:
                found_user_id = data[0].get("id")
            print(f"    Using user_id for admin tests: {found_user_id}")
            show("users sample", data[:2])
    else:
        log("Admin", "GET", "/admin/users", 0, False, b)

    if found_user_id:
        r, b = req("get", f"{BASE}/admin/users/{found_user_id}", headers=adm())
        if r:
            ok = r.status_code == 200
            log("Admin", "GET", "/admin/users/:id", r.status_code, ok,
                str(b.get("data",{}).get("name","?") if isinstance(b,dict) else b)[:80], resp_body=b)
        else:
            log("Admin", "GET", "/admin/users/:id", 0, False, b)

        r, b = req("patch", f"{BASE}/admin/users/{found_user_id}/block",
                   json={"blocked": True}, headers=adm())
        if r:
            ok = r.status_code in (200, 201)
            log("Admin", "PATCH", "/admin/users/:id/block (block)", r.status_code, ok,
                b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
        else:
            log("Admin", "PATCH", "/admin/users/:id/block", 0, False, b)

        r, b = req("patch", f"{BASE}/admin/users/{found_user_id}/block",
                   json={"blocked": False}, headers=adm())
        if r:
            ok = r.status_code in (200, 201)
            log("Admin", "PATCH", "/admin/users/:id/block (unblock)", r.status_code, ok,
                b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
        else:
            log("Admin", "PATCH", "/admin/users/:id/unblock", 0, False, b)

        topup_body = {"amount": 10, "reason": "QA test topup"}
        r, b = req("post", f"{BASE}/admin/users/{found_user_id}/topup",
                   json=topup_body, headers=adm())
        if r:
            ok = r.status_code in (200, 201)
            log("Admin", "POST", "/admin/users/:id/topup", r.status_code, ok,
                b.get("message","?") if isinstance(b,dict) else str(b)[:80],
                req_body=topup_body, resp_body=b)
            show("topup", b)
        else:
            log("Admin", "POST", "/admin/users/:id/topup", 0, False, b)

    r, b = req("get", f"{BASE}/admin/companies", params={"page": 1, "per_page": 5}, headers=adm())
    if r:
        ok = r.status_code == 200
        data = b.get("data", []) if isinstance(b, dict) else []
        log("Admin", "GET", "/admin/companies", r.status_code, ok,
            f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
    else:
        log("Admin", "GET", "/admin/companies", 0, False, b)

    r, b = req("get", f"{BASE}/admin/listings",
               params={"type": "cargo", "page": 1, "per_page": 5}, headers=adm())
    if r:
        ok = r.status_code == 200
        data = b.get("data", []) if isinstance(b, dict) else []
        log("Admin", "GET", "/admin/listings?type=cargo", r.status_code, ok,
            f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
    else:
        log("Admin", "GET", "/admin/listings", 0, False, b)

    r, b = req("get", f"{BASE}/admin/payments",
               params={"page": 1, "per_page": 5}, headers=adm())
    if r:
        ok = r.status_code == 200
        data = b.get("data", []) if isinstance(b, dict) else []
        log("Admin", "GET", "/admin/payments", r.status_code, ok,
            f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
    else:
        log("Admin", "GET", "/admin/payments", 0, False, b)

    # Pricing CRUD
    r, b = req("get", f"{BASE}/admin/pricing", headers=adm())
    if r:
        ok = r.status_code == 200
        data = b.get("data", []) if isinstance(b, dict) else []
        log("Admin", "GET", "/admin/pricing", r.status_code, ok,
            f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
        show("pricing items", data)
    else:
        log("Admin", "GET", "/admin/pricing", 0, False, b)

    test_key = "qa_test_pricing_xyz"
    r, b = req("post", f"{BASE}/admin/pricing",
               json={"key": test_key, "name": "QA Test", "price": 100000, "currency": "UZS", "tokens": 5, "type": "tokens"},
               headers=adm())
    if r:
        ok = r.status_code in (200, 201)
        log("Admin", "POST", "/admin/pricing", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
        show("pricing create", b)
    else:
        log("Admin", "POST", "/admin/pricing", 0, False, b)

    r, b = req("put", f"{BASE}/admin/pricing/{test_key}",
               json={"name": "QA Updated", "price": 120000, "tokens": 6},
               headers=adm())
    if r:
        ok = r.status_code in (200, 201)
        log("Admin", "PUT", "/admin/pricing/:key", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
        if not ok: show("pricing update error", b)
    else:
        log("Admin", "PUT", "/admin/pricing/:key", 0, False, b)

    r, b = req("delete", f"{BASE}/admin/pricing/{test_key}", headers=adm())
    if r:
        ok = r.status_code in (200, 201, 204)
        log("Admin", "DELETE", "/admin/pricing/:key", r.status_code, ok,
            b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
    else:
        log("Admin", "DELETE", "/admin/pricing/:key", 0, False, b)

    # Categories
    r, b = req("get", f"{BASE}/admin/categories", headers=adm())
    if r:
        ok = r.status_code == 200
        data = b.get("data", []) if isinstance(b, dict) else []
        log("Admin", "GET", "/admin/categories", r.status_code, ok,
            f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
        show("categories", data[:3] if isinstance(data, list) else data)
    else:
        log("Admin", "GET", "/admin/categories", 0, False, b)

    r, b = req("post", f"{BASE}/admin/categories",
               json={"name": "QA Test Category", "name_uz": "QA Test Kategoriya", "name_en": "QA Test Category"},
               headers=adm())
    if r:
        ok = r.status_code in (200, 201)
        data = b.get("data", {}) if isinstance(b, dict) else {}
        cat_id = data.get("id") if ok else None
        log("Admin", "POST", "/admin/categories", r.status_code, ok,
            f"id={cat_id}" if ok else (b.get("error",{}).get("message","?") if isinstance(b,dict) else str(b)[:80]),
            resp_body=b)
        show("create cat", b)
    else:
        log("Admin", "POST", "/admin/categories", 0, False, b)
        cat_id = None

    if cat_id:
        r, b = req("put", f"{BASE}/admin/categories/{cat_id}",
                   json={"name": "Updated QA Category"}, headers=adm())
        if r:
            ok = r.status_code in (200, 201)
            log("Admin", "PUT", "/admin/categories/:id", r.status_code, ok,
                b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
        else:
            log("Admin", "PUT", "/admin/categories/:id", 0, False, b)

        r, b = req("delete", f"{BASE}/admin/categories/{cat_id}", headers=adm())
        if r:
            ok = r.status_code in (200, 201, 204)
            log("Admin", "DELETE", "/admin/categories/:id", r.status_code, ok,
                b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
        else:
            log("Admin", "DELETE", "/admin/categories/:id", 0, False, b)

    # Moderators
    mod_login = "qamod_test_x001"
    mod_pass = "qapass123456789"
    r, b = req("post", f"{BASE}/admin/moderators",
               json={"login": mod_login, "password": mod_pass},
               headers=adm())
    if r:
        ok = r.status_code in (200, 201)
        data = b.get("data", {}) if isinstance(b, dict) else {}
        state["moderator_id"] = data.get("id") if ok else None
        log("Admin", "POST", "/admin/moderators", r.status_code, ok,
            f"id={state['moderator_id']}" if ok else (b.get("error",{}).get("message","?") if isinstance(b,dict) else str(b)[:80]),
            resp_body=b)
        show("create mod", b)
    else:
        log("Admin", "POST", "/admin/moderators", 0, False, b)
        mod_login, mod_pass = None, None

    if state["moderator_id"]:
        r, b = req("delete", f"{BASE}/admin/moderators/{state['moderator_id']}", headers=adm())
        if r:
            ok = r.status_code in (200, 201, 204)
            log("Admin", "DELETE", "/admin/moderators/:id", r.status_code, ok,
                b.get("message","?") if isinstance(b,dict) else str(b)[:80], resp_body=b)
        else:
            log("Admin", "DELETE", "/admin/moderators/:id", 0, False, b)

# ================================================================================
print("\n" + "="*70)
print("14. MODERATOR FLOW")
print("="*70)

temp_mod_login = "qamod_flow_x99"
temp_mod_pass = "qapass9999999"
temp_mod_id = None
state["mod_token"] = None

if state["admin_token"]:
    r, b = req("post", f"{BASE}/admin/moderators",
               json={"login": temp_mod_login, "password": temp_mod_pass},
               headers=adm())
    if r and r.status_code in (200, 201) and isinstance(b, dict):
        temp_mod_id = b.get("data", {}).get("id")
        print(f"    Created temp mod: {temp_mod_id}")

    r, b = req("post", f"{BASE}/admin/login",
               json={"login": temp_mod_login, "password": temp_mod_pass})
    if r:
        ok = r.status_code in (200, 201)
        data = b.get("data", {}) if isinstance(b, dict) else {}
        state["mod_token"] = data.get("access_token") or data.get("token")
        log("Moderator", "POST", "/admin/login (mod)", r.status_code, ok,
            f"role={data.get('role','?')}, token={'SET' if state['mod_token'] else 'MISSING'}",
            resp_body=b)
        show("mod login", b)
    else:
        log("Moderator", "POST", "/admin/login (mod)", 0, False, b)

if state["mod_token"]:
    mod_h = {"Authorization": f"Bearer {state['mod_token']}"}
    r, b = req("get", f"{BASE}/moderator/queue", headers=mod_h)
    if r:
        ok = r.status_code == 200
        data = b.get("data", []) if isinstance(b, dict) else []
        log("Moderator", "GET", "/moderator/queue", r.status_code, ok,
            f"count={len(data) if isinstance(data, list) else '?'}", resp_body=b)
        show("queue", b)
        if ok and isinstance(data, list) and data:
            qid = data[0].get("id")
            r2, b2 = req("get", f"{BASE}/moderator/queue/{qid}", headers=mod_h)
            if r2:
                ok2 = r2.status_code == 200
                log("Moderator", "GET", "/moderator/queue/:id", r2.status_code, ok2,
                    str(b2)[:80], resp_body=b2)
            else:
                log("Moderator", "GET", "/moderator/queue/:id", 0, False, b2)
        else:
            print("  SKIPPED: /moderator/queue/:id (queue empty)")
    else:
        log("Moderator", "GET", "/moderator/queue", 0, False, b)

# Cleanup
if temp_mod_id and state["admin_token"]:
    req("delete", f"{BASE}/admin/moderators/{temp_mod_id}", headers=adm())
    print(f"    Cleaned up temp mod: {temp_mod_id}")

# ================================================================================
print("\n" + "="*70)
print("15. EDGE CASES")
print("="*70)

print("\n  [a] Unauthenticated access to protected endpoints (expect 401)")
protected = [
    ("get", f"{BASE}/users/me"),
    ("get", f"{BASE}/users/me/stats"),
    ("get", f"{BASE}/favorites"),
    ("get", f"{BASE}/routes"),
    ("get", f"{BASE}/notifications"),
    ("get", f"{BASE}/contacts/history"),
    ("get", f"{BASE}/payments/history"),
    ("post", f"{BASE}/contacts/view"),
    ("post", f"{BASE}/payments/create"),
]
for method, url in protected:
    r, b = req(method, url)  # no auth
    if r:
        ok = r.status_code == 401
        path = url.replace(BASE, "")
        log("EdgeCase-Auth", method.upper(), f"{path} [no token]",
            r.status_code, ok,
            "401 OK" if ok else f"GOT {r.status_code} instead of 401 - PROBLEM!",
            resp_body=b)
    else:
        log("EdgeCase-Auth", method.upper(), url, 0, False, b)

print("\n  [b] User token on admin endpoint (expect 403)")
if state["access_token"]:
    r, b = req("get", f"{BASE}/admin/users", headers=ah())
    if r:
        ok = r.status_code == 403
        log("EdgeCase-Authz", "GET", "/admin/users [user token]", r.status_code, ok,
            "403 OK" if ok else f"GOT {r.status_code} instead of 403 - PROBLEM!",
            resp_body=b)
        show("admin with user token", b)
    else:
        log("EdgeCase-Authz", "GET", "/admin/users [user token]", 0, False, b)

    r, b = req("get", f"{BASE}/admin/dashboard", headers=ah())
    if r:
        ok = r.status_code == 403
        log("EdgeCase-Authz", "GET", "/admin/dashboard [user token]", r.status_code, ok,
            "403 OK" if ok else f"GOT {r.status_code} instead of 403", resp_body=b)
    else:
        log("EdgeCase-Authz", "GET", "/admin/dashboard [user token]", 0, False, b)

print("\n  [c] Invalid UUID params (expect 400 or 404)")
bad_uuid = "not-valid-uuid-xyz"
for method, url, use_auth in [
    ("get", f"{BASE}/listings/cargo/{bad_uuid}", False),
    ("get", f"{BASE}/warehouses/{bad_uuid}", False),
    ("get", f"{BASE}/payments/{bad_uuid}", True),
    ("get", f"{BASE}/companies/{bad_uuid}", True),
]:
    kwargs = {"headers": ah()} if use_auth else {}
    r, b = req(method, url, **kwargs)
    if r:
        ok = r.status_code in (400, 404)
        path = url.replace(BASE, "")
        log("EdgeCase-UUID", method.upper(), f"{path} [bad uuid]", r.status_code, ok,
            "400/404 OK" if ok else f"GOT {r.status_code} - should be 400/404",
            resp_body=b)
    else:
        log("EdgeCase-UUID", method.upper(), url, 0, False, b)

print("\n  [d] Invalid request bodies (expect 400/422)")
invalid_cases = [
    ("post", f"{BASE}/auth/send-otp", {}, False, "empty body"),
    ("post", f"{BASE}/auth/send-otp", {"phone": "123"}, False, "bad phone format"),
    ("post", f"{BASE}/auth/verify-otp", {"phone": "not-phone", "code": "xxx"}, False, "bad phone+code"),
    ("post", f"{BASE}/payments/create", {"payment_type": "fake_type"}, True, "invalid payment_type"),
    ("post", f"{BASE}/routes", {}, True, "empty route body"),
    ("post", f"{BASE}/favorites", {"listing_type": "invalid"}, True, "invalid listing_type"),
]
for method, url, body, needs_auth, desc in invalid_cases:
    kwargs = {"json": body}
    if needs_auth and state["access_token"]:
        kwargs["headers"] = ah()
    r, b = req(method, url, **kwargs)
    if r:
        ok = r.status_code in (400, 422)
        path = url.replace(BASE, "")
        log("EdgeCase-Validation", method.upper(), f"{path} [{desc}]", r.status_code, ok,
            "validation OK" if ok else f"GOT {r.status_code} - should be 400/422",
            req_body=body, resp_body=b)
    else:
        log("EdgeCase-Validation", method.upper(), url, 0, False, b)

print("\n  [e] Pagination edge cases")
pagination_tests = [
    (f"{BASE}/listings/cargo?page=0&per_page=5", "page=0"),
    (f"{BASE}/listings/cargo?page=1&per_page=999", "per_page=999"),
    (f"{BASE}/listings/cargo?page=1&per_page=-1", "per_page=-1"),
    (f"{BASE}/warehouses?page=0", "page=0"),
]
for url, desc in pagination_tests:
    r, b = req("get", url)
    if r:
        ok = r.status_code in (200, 400)
        path = url.replace(BASE, "")
        note = f"got {r.status_code}"
        if r.status_code == 200 and isinstance(b, dict):
            data = b.get("data", [])
            note += f", count={len(data) if isinstance(data, list) else '?'}"
        elif not ok:
            note += " - unexpected status"
        log("EdgeCase-Pagination", "GET", f"[{desc}]", r.status_code, ok, note, resp_body=b)
    else:
        log("EdgeCase-Pagination", "GET", url, 0, False, b)

print("\n  [f] i18n lang param test")
for lang in ["uz", "en", "ru"]:
    r, b = req("post", f"{BASE}/auth/send-otp",
               json={"phone": "+998909999991", "channel": "sms"},
               params={"lang": lang})
    if r:
        ok = r.status_code in (200, 201, 400)
        msg = b.get("message", "?") if isinstance(b, dict) else str(b)[:60]
        log("EdgeCase-i18n", "POST", f"/auth/send-otp?lang={lang}", r.status_code, ok,
            f"message='{msg[:80]}'", resp_body=b)
    else:
        log("EdgeCase-i18n", "POST", f"/auth/send-otp?lang={lang}", 0, False, b)

# ================================================================================
print("\n" + "="*70)
print("SUMMARY")
print("="*70)

passed = [r for r in results if r["ok"]]
failed = [r for r in results if not r["ok"]]

print(f"\nTotal tested: {len(results)}")
print(f"Passed:       {len(passed)}")
print(f"Failed:       {len(failed)}")

print("\n--- FAILED ENDPOINTS ---")
for item in failed:
    print(f"  [{item['status']:3d}] {item['method']:6s} {item['path']}")
    if item.get("resp_body"):
        body_str = json.dumps(item["resp_body"], ensure_ascii=False)[:200] if isinstance(item["resp_body"], dict) else str(item["resp_body"])[:200]
        print(f"         response: {body_str}")

# Save full results
import datetime
ts = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
out_path = f"C:/Users/ibrok/Desktop/karvon/qa_audit_results_{ts}.json"
with open(out_path, "w", encoding="utf-8") as f:
    json.dump(results, f, ensure_ascii=False, indent=2)
print(f"\nFull results saved to: {out_path}")
