# Test Migration Guide

## Overview

This document describes the strategy for migrating integration tests from Vitest (mocked) to Playwright (real E2E).

## Test Categories

### 1. Unit Tests (Keep in Vitest)
**Location:** `web/tests/*.test.js`

**What to keep:**
- Component rendering tests
- State management tests
- Utility function tests
- Event handler tests
- Tests that don't need a real server

**Examples:**
- `tests/utils.test.js` - Format helpers, date formatting
- `tests/state.test.js` - State initialization, persistence
- `tests/player.test.js` - Player controls (mocked)
- `tests/routing.test.js` - Route parsing logic

### 2. E2E Tests (Migrate to Playwright)
**Location:** `e2e/tests/*.spec.ts`

**What to migrate:**
- Tests that verify server integration
- Tests that check actual API responses
- Tests for complete user workflows
- Tests for cross-component interactions

**Already migrated:**
- ✅ `e2e/tests/navigation.spec.ts` - Basic navigation
- ✅ `e2e/tests/captions.spec.ts` - Caption functionality
- ✅ `e2e/tests/du-navigation.spec.ts` - Disk Usage
- ✅ `e2e/tests/playback.spec.ts` - Media playback
- ✅ `e2e/tests/search-filter.spec.ts` - Search and filtering
- ✅ `e2e/tests/query-integration.spec.ts` - Query integration

**To migrate from `web/tests/integration-*.test.js`:**
- ❌ Playlist drag-and-drop (needs real DnD)
- ❌ Local storage persistence across page reloads
- ❌ Real API error handling
- ❌ Multi-tab synchronization

## Migration Process

### Step 1: Identify Test to Migrate

Look for tests that:
- Mock `fetch()` extensively
- Test API parameter building
- Verify server responses
- Test cross-component state

### Step 2: Create E2E Test

```typescript
// e2e/tests/my-feature.spec.ts
import { test, expect } from '../fixtures';

test.describe('My Feature', () => {
  test('does something with real server', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    // Interact with real UI
    await page.click('#button');
    
    // Verify real server response
    await expect(page.locator('.result')).toContainText('expected');
  });
});
```

### Step 3: Update Vitest Test

Keep the Vitest test for fast feedback, but reduce scope:
- Remove server-dependent assertions
- Focus on UI logic only
- Keep mocked API tests

### Step 4: Update CI

E2E tests run in GitHub Actions automatically via `.github/workflows/e2e.yml`.

## Running Tests

### Local Development

```bash
# Fast unit tests only
make webtest

# E2E tests (slower, needs server)
make e2e

# All tests
make test-all
```

### CI/CD

- **Unit tests**: Run on every commit (fast)
- **E2E tests**: Run on PRs and merges (slower)

## Test File Naming

- **Vitest**: `*.test.js` in `web/tests/`
- **Playwright**: `*.spec.ts` in `e2e/tests/`

## Migration Checklist

- [ ] Identify integration tests that need real server
- [ ] Create corresponding E2E test in `e2e/tests/`
- [ ] Verify E2E test passes
- [ ] Simplify Vitest test (remove server-dependent checks)
- [ ] Update this document
- [ ] Add to CI if new coverage area

## Current Status

### Migration Progress

| Test File | Status | Notes |
|-----------|--------|-------|
| `integration-1.test.js` | ✅ Migrated | Search, trash, view modes, playlists |
| `integration-2.test.js` | ✅ Migrated | Playback, settings, keyboard shortcuts |
| `integration-3.test.js` | ✅ Migrated | Filter persistence, history |
| `captions.test.js` | ✅ Migrated | Caption display, jumping |
| `du.test.js` | ✅ Migrated | DU navigation, auto-skip |
| `player.test.js` | Keep | Mocked player controls |
| `utils.test.js` | Keep | Pure functions |
| `state.test.js` | Keep | State logic |
| `routing.test.js` | Keep | Route parsing |
| `settings.test.js` | ✅ Migrated | Settings persistence |
| `keyboard-shortcuts.test.js` | ✅ Migrated | Keyboard shortcuts |
| `pagination-limit.test.js` | Keep | Limit logic |
| `progress.test.js` | Keep | Progress calculations |
| `search.test.js` | ✅ Migrated | Search functionality |

### E2E Test Coverage

| Feature | E2E Tests | Location |
|---------|-----------|----------|
| Navigation | ✅ 5 tests | `navigation.spec.ts` |
| Captions | ✅ 6 tests | `captions.spec.ts` |
| Disk Usage | ✅ 6 tests | `du-navigation.spec.ts` |
| Playback | ✅ 7 tests | `playback.spec.ts` |
| Search/Filter | ✅ 13 tests | `search-filter.spec.ts`, `query-integration.spec.ts` |
| Query Integration | ✅ 11 tests | `query-integration.spec.ts` |
| Playlists | ✅ 4 tests | `settings-playback.spec.ts` |
| Settings | ✅ 7 tests | `settings-playback.spec.ts` |
| Keyboard | ✅ 6 tests | `settings-playback.spec.ts` |
| **Total** | **177 tests** | 7 test files (3 browsers each = 531 test runs) |

## Benefits

### Before (All Vitest)
- ❌ False positives (mocks don't match reality)
- ❌ Miss integration bugs
- ❌ Can't test real server behavior

### After (Hybrid)
- ✅ Fast unit tests for logic (< 5s)
- ✅ Real E2E tests for workflows (< 2min)
- ✅ Catch integration bugs before production
- ✅ Better confidence in releases
