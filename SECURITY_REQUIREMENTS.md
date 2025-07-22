# Security Requirements for Brew Detective

## Critical Security Issue Identified

The current implementation exposes case answers to users through the API response, allowing skilled users to cheat by viewing correct answers in the Network tab or browser console.

## Required Backend Changes

### 1. Create Secure Public Endpoint

**Endpoint:** `GET /api/v1/cases/active/public`

**Purpose:** Provide case information for regular users without revealing answers

**Response Structure:**
```json
{
  "case": {
    "id": "case-uuid",
    "name": "Case Name",
    "description": "Case description",
    "is_active": true,
    "enabled_questions": {
      "region": true,
      "variety": true,
      "process": true,
      "taste_note_1": true,
      "taste_note_2": true,
      "favorite_coffee": true,
      "brewing_method": true
    },
    "coffee_ids": ["coffee-1-uuid", "coffee-2-uuid", "coffee-3-uuid", "coffee-4-uuid"],
    "coffee_count": 4
  }
}
```

**Security Requirements:**
- ❌ **NEVER** include coffee details (region, variety, process, etc.)
- ❌ **NEVER** include taste notes or any answer data
- ✅ **ONLY** include coffee IDs needed for submission
- ✅ **ONLY** include case metadata and question configuration

### 2. Restrict Admin Endpoint

**Endpoint:** `GET /api/v1/cases/active` (existing)

**Purpose:** Full case data for admin users only

**Security Requirements:**
- ✅ Require admin authentication (JWT with admin role)
- ✅ Return complete case data including answers
- ❌ Block non-admin users with 403 Forbidden

### 3. Server-Side Answer Validation

**Critical:** All answer validation MUST happen on the server

**Requirements:**
- Client submits answers with coffee IDs only
- Server looks up correct answers from database
- Server calculates scores and accuracy
- Client never receives correct answers

## Frontend Security Measures Implemented

### 1. Data Sanitization
- Added `sanitizeCaseData()` function to strip sensitive information
- Frontend tries secure endpoint first, falls back to sanitized data
- Removes coffee answer data from any API responses

### 2. Global Variable Protection
- Replaced `window.activeCase` with `window.activeCaseData`
- Only stores safe data (IDs, metadata, no answers)
- Explicitly deletes any unsafe global variables

### 3. Console Log Cleanup
- Removed all console.log statements exposing case answers
- Removed debug functions from window object
- Sanitized authentication logging

### 4. Access Control Verification
- Admin functions properly protected by UI visibility
- Non-admin users cannot access admin interface
- Backend should enforce actual permissions

## Implementation Priority

1. **IMMEDIATE:** Implement secure public endpoint `/api/v1/cases/active/public`
2. **IMMEDIATE:** Add admin-only restriction to existing `/api/v1/cases/active`
3. **HIGH:** Ensure server-side answer validation for submissions
4. **MEDIUM:** Review all other endpoints for similar security issues

## Testing Security

After implementing backend changes:

1. **Network Tab Test:**
   - Load case as regular user
   - Verify no coffee answers in `/api/v1/cases/active/public` response
   - Verify admin endpoint returns 403 for non-admin users

2. **Console Test:**
   - Check `window.activeCaseData` only contains safe data
   - Verify `window.activeCase` is undefined
   - Confirm no coffee answers accessible anywhere

3. **Submission Test:**
   - Submit answers and verify scoring happens server-side
   - Confirm client never receives correct answers in response

## Current Status

✅ **Frontend security measures implemented**
⏳ **Backend secure endpoints needed**
⏳ **Server-side answer validation required**

The game is now secure from client-side cheating once the backend endpoints are updated.