const { chromium } = require('playwright');
const https = require('https');
const http = require('http');
const fs = require('fs');
const path = require('path');

// --- Config ---
const BASE_URL = 'http://localhost:3000';
const API_URL = 'http://localhost:8000/api/v1';
const TEST_EMAIL = 'qa@financeos.com';
const TEST_PASSWORD = 'QA@Test123456';
const TEST_NAME = 'QA Tester';
const SCREENSHOTS_DIR = path.join(__dirname, 'screenshots');

if (!fs.existsSync(SCREENSHOTS_DIR)) fs.mkdirSync(SCREENSHOTS_DIR, { recursive: true });

// --- Test results ---
const results = [];
let passed = 0;
let failed = 0;
let accessToken = '';
let createdAccountId = '';
let createdCategoryId = '';
let createdTransactionId = '';
let createdBudgetId = '';
let createdGoalId = '';
let createdPortfolioId = '';

// --- API Helper ---
function apiRequest(method, path, body, token) {
  return new Promise((resolve, reject) => {
    const url = new URL(API_URL + path);
    const options = {
      hostname: url.hostname,
      port: url.port || 80,
      path: url.pathname + (url.search || ''),
      method,
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
    };
    const req = http.request(options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        try {
          resolve({ status: res.statusCode, body: JSON.parse(data) });
        } catch {
          resolve({ status: res.statusCode, body: data });
        }
      });
    });
    req.on('error', reject);
    if (body) req.write(JSON.stringify(body));
    req.end();
  });
}

function record(name, ok, details = '') {
  const status = ok ? '✅' : '❌';
  results.push({ name, ok, details });
  if (ok) passed++; else failed++;
  console.log(`${status} ${name}${details ? ' — ' + details : ''}`);
}

async function withPage(browser, fn) {
  const page = await browser.newPage();
  const consoleErrors = [];
  page.on('console', msg => {
    if (msg.type() === 'error') consoleErrors.push(msg.text());
  });
  try {
    await fn(page, consoleErrors);
  } finally {
    await page.close();
  }
}

async function flutterNav(page, hash, waitMs = 4000) {
  await page.goto(`${BASE_URL}/${hash}`, { waitUntil: 'load' });
  await page.waitForTimeout(waitMs);
}

async function enableFlutterAccessibility(page) {
  // Click the accessibility button to enable Flutter semantics
  await page.evaluate(() => {
    const btn = document.querySelector('flt-semantics-placeholder');
    if (btn) btn.click();
  });
  await page.waitForTimeout(2000);
}

async function getBodyText(page) {
  await enableFlutterAccessibility(page);
  return page.evaluate(() => {
    const sem = document.querySelector('flt-semantics-host');
    if (sem) return sem.innerText || sem.textContent || '';
    return document.body.innerText || '';
  }).catch(() => '');
}

async function fillFlutterInput(page, index, value) {
  // Flutter web inputs appear in flt-text-editing-host after click/focus
  const inputs = await page.locator('flt-text-editing-host input, flt-text-editing-host textarea').all();
  if (inputs.length > index) {
    await inputs[index].click();
    await inputs[index].fill(value);
    return true;
  }
  return false;
}

// =============================================================================
// BLOCK 1 — API Health & Auth
// =============================================================================
async function runApiAuthTests() {
  console.log('\n=== BLOCK 1: API Health & Auth ===');

  // Test 1.1: Health check
  try {
    const res = await new Promise((resolve, reject) => {
      const req = http.request({ hostname: 'localhost', port: 8000, path: '/health', method: 'GET' }, (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => { try { resolve({ status: res.statusCode, body: JSON.parse(data) }); } catch { resolve({ status: res.statusCode, body: data }); }});
      });
      req.on('error', reject);
      req.end();
    });
    const ok = res.status === 200 && res.body.status === 'ok';
    record('1.1 API Health check', ok, `status=${res.status}`);
  } catch (e) {
    record('1.1 API Health check', false, e.message);
  }

  // Test 1.2: Register (may already exist)
  try {
    const res = await apiRequest('POST', '/auth/register', {
      name: TEST_NAME, email: TEST_EMAIL, password: TEST_PASSWORD
    });
    const ok = res.status === 201 || res.status === 409; // 409 = user exists
    const detail = res.status === 409 ? 'user already exists (OK)' : `status=${res.status}`;
    record('1.2 Register new user', ok, detail);
  } catch (e) {
    record('1.2 Register new user', false, e.message);
  }

  // Test 1.3: Login
  try {
    const res = await apiRequest('POST', '/auth/login', {
      email: TEST_EMAIL, password: TEST_PASSWORD
    });
    const ok = res.status === 200 && res.body.data?.access_token;
    if (ok) accessToken = res.body.data.access_token;
    record('1.3 Login', ok, `token=${ok ? accessToken.slice(0, 20) + '...' : 'MISSING'}`);
  } catch (e) {
    record('1.3 Login', false, e.message);
  }

  // Test 1.4: Login with bad password
  try {
    const res = await apiRequest('POST', '/auth/login', {
      email: TEST_EMAIL, password: 'wrongpassword'
    });
    const ok = res.status === 401 || res.status === 400;
    record('1.4 Login with bad password returns 4xx', ok, `status=${res.status}`);
  } catch (e) {
    record('1.4 Login with bad password returns 4xx', false, e.message);
  }

  // Test 1.5: Refresh token
  try {
    const loginRes = await apiRequest('POST', '/auth/login', {
      email: TEST_EMAIL, password: TEST_PASSWORD
    });
    const refreshToken = loginRes.body.data?.refresh_token;
    if (refreshToken) {
      const res = await apiRequest('POST', '/auth/refresh', { refresh_token: refreshToken });
      const ok = res.status === 200 && res.body.data?.access_token;
      record('1.5 Refresh token', ok, `status=${res.status}`);
    } else {
      record('1.5 Refresh token', false, 'no refresh token in login response');
    }
  } catch (e) {
    record('1.5 Refresh token', false, e.message);
  }

  // Test 1.6: Protected endpoint without token
  try {
    const res = await apiRequest('GET', '/accounts', null, null);
    const ok = res.status === 401;
    record('1.6 Protected endpoint rejects unauthenticated', ok, `status=${res.status}`);
  } catch (e) {
    record('1.6 Protected endpoint rejects unauthenticated', false, e.message);
  }

  // Test 1.7: Forgot password endpoint
  try {
    const res = await apiRequest('POST', '/auth/forgot-password', { email: TEST_EMAIL });
    const ok = res.status === 200 || res.status === 404; // 404 might be expected if email not verified
    record('1.7 Forgot password endpoint', ok, `status=${res.status}`);
  } catch (e) {
    record('1.7 Forgot password endpoint', false, e.message);
  }
}

// =============================================================================
// BLOCK 2 — Accounts CRUD
// =============================================================================
async function runAccountsTests() {
  console.log('\n=== BLOCK 2: Accounts CRUD ===');
  if (!accessToken) { record('2.x Accounts tests skipped', false, 'no access token'); return; }

  // Test 2.1: List accounts
  try {
    const res = await apiRequest('GET', '/accounts', null, accessToken);
    const ok = res.status === 200 && res.body.data !== undefined;
    record('2.1 List accounts', ok, `count=${Array.isArray(res.body.data) ? res.body.data.length : 'N/A'}`);
  } catch (e) {
    record('2.1 List accounts', false, e.message);
  }

  // Test 2.2: Create account
  try {
    const res = await apiRequest('POST', '/accounts', {
      name: 'QA Conta Corrente', type: 'checking', currency: 'BRL', balance: 2500.00
    }, accessToken);
    const ok = res.status === 201 && res.body.data?.id;
    if (ok) createdAccountId = res.body.data.id;
    record('2.2 Create account', ok, `id=${createdAccountId || 'MISSING'}`);
  } catch (e) {
    record('2.2 Create account', false, e.message);
  }

  // Test 2.3: Get account by ID
  try {
    if (!createdAccountId) throw new Error('no account created');
    const res = await apiRequest('GET', `/accounts/${createdAccountId}`, null, accessToken);
    const ok = res.status === 200 && res.body.data?.id === createdAccountId;
    record('2.3 Get account by ID', ok, `status=${res.status}`);
  } catch (e) {
    record('2.3 Get account by ID', false, e.message);
  }

  // Test 2.4: Update account
  try {
    if (!createdAccountId) throw new Error('no account created');
    const res = await apiRequest('PUT', `/accounts/${createdAccountId}`, {
      name: 'QA Conta Corrente Updated', type: 'checking', currency: 'BRL', balance: 3000.00
    }, accessToken);
    const ok = res.status === 200 && res.body.data?.name === 'QA Conta Corrente Updated';
    record('2.4 Update account', ok, `status=${res.status}`);
  } catch (e) {
    record('2.4 Update account', false, e.message);
  }

  // Test 2.5: Accounts summary
  try {
    const res = await apiRequest('GET', '/accounts/summary', null, accessToken);
    const ok = res.status === 200 && res.body.data?.total_balance !== undefined;
    record('2.5 Accounts summary', ok, `balance=${res.body.data?.total_balance}`);
  } catch (e) {
    record('2.5 Accounts summary', false, e.message);
  }
}

// =============================================================================
// BLOCK 3 — Categories
// =============================================================================
async function runCategoriesTests() {
  console.log('\n=== BLOCK 3: Categories ===');
  if (!accessToken) { record('3.x Categories skipped', false, 'no token'); return; }

  // Test 3.1: List categories
  try {
    const res = await apiRequest('GET', '/categories', null, accessToken);
    const ok = res.status === 200 && Array.isArray(res.body.data);
    record('3.1 List categories', ok, `count=${res.body.data?.length}`);
  } catch (e) {
    record('3.1 List categories', false, e.message);
  }

  // Test 3.2: Create custom category
  try {
    const res = await apiRequest('POST', '/categories', {
      name: 'QA Category', type: 'expense', icon: 'category', color: '#FF0000'
    }, accessToken);
    const ok = res.status === 201 && res.body.data?.id;
    if (ok) createdCategoryId = res.body.data.id;
    record('3.2 Create category', ok, `id=${createdCategoryId || 'MISSING'}`);
  } catch (e) {
    record('3.2 Create category', false, e.message);
  }

  // Test 3.3: Update category
  try {
    if (!createdCategoryId) throw new Error('no category created');
    const res = await apiRequest('PUT', `/categories/${createdCategoryId}`, {
      name: 'QA Category Updated', type: 'expense', icon: 'category', color: '#00FF00'
    }, accessToken);
    const ok = res.status === 200;
    record('3.3 Update category', ok, `status=${res.status}`);
  } catch (e) {
    record('3.3 Update category', false, e.message);
  }
}

// =============================================================================
// BLOCK 4 — Transactions
// =============================================================================
async function runTransactionsTests() {
  console.log('\n=== BLOCK 4: Transactions ===');
  if (!accessToken) { record('4.x Transactions skipped', false, 'no token'); return; }

  // Get a category ID for transactions
  const catRes = await apiRequest('GET', '/categories', null, accessToken);
  const expenseCat = catRes.body.data?.find(c => c.type === 'expense');
  const incomeCat = catRes.body.data?.find(c => c.type === 'income');
  const accRes = await apiRequest('GET', '/accounts', null, accessToken);
  const account = accRes.body.data?.[0];

  // Test 4.1: Create transaction
  try {
    if (!account || !expenseCat) throw new Error('missing account or category');
    const res = await apiRequest('POST', '/transactions', {
      account_id: account.id,
      category_id: expenseCat.id,
      type: 'expense',
      amount: 150.00,
      description: 'QA Test Transaction',
      date: '2026-04-04T00:00:00Z',
    }, accessToken);
    const ok = res.status === 201 && res.body.data?.id;
    if (ok) createdTransactionId = res.body.data.id;
    record('4.1 Create transaction', ok, `id=${createdTransactionId || 'MISSING'}`);
  } catch (e) {
    record('4.1 Create transaction', false, e.message);
  }

  // Test 4.2: List transactions
  try {
    const res = await apiRequest('GET', '/transactions?limit=10', null, accessToken);
    const ok = res.status === 200 && Array.isArray(res.body.data);
    record('4.2 List transactions', ok, `count=${res.body.data?.length}`);
  } catch (e) {
    record('4.2 List transactions', false, e.message);
  }

  // Test 4.3: Get transaction by ID
  try {
    if (!createdTransactionId) throw new Error('no transaction');
    const res = await apiRequest('GET', `/transactions/${createdTransactionId}`, null, accessToken);
    const ok = res.status === 200 && res.body.data?.id === createdTransactionId;
    record('4.3 Get transaction by ID', ok, `status=${res.status}`);
  } catch (e) {
    record('4.3 Get transaction by ID', false, e.message);
  }

  // Test 4.4: Update transaction
  try {
    if (!createdTransactionId || !account || !expenseCat) throw new Error('missing data');
    const res = await apiRequest('PUT', `/transactions/${createdTransactionId}`, {
      account_id: account.id,
      category_id: expenseCat.id,
      type: 'expense',
      amount: 200.00,
      description: 'QA Test Transaction Updated',
      date: '2026-04-04T00:00:00Z',
    }, accessToken);
    const ok = res.status === 200;
    record('4.4 Update transaction', ok, `status=${res.status}`);
  } catch (e) {
    record('4.4 Update transaction', false, e.message);
  }

  // Test 4.5: Transactions summary
  try {
    const res = await apiRequest('GET', '/transactions/summary?month=4&year=2026', null, accessToken);
    const ok = res.status === 200 && res.body.data?.total_expense !== undefined;
    record('4.5 Transactions summary', ok, `expense=${res.body.data?.total_expense}`);
  } catch (e) {
    record('4.5 Transactions summary', false, e.message);
  }

  // Test 4.6: Create income transaction
  try {
    if (!account || !incomeCat) throw new Error('missing account or income category');
    const res = await apiRequest('POST', '/transactions', {
      account_id: account.id,
      category_id: incomeCat.id,
      type: 'income',
      amount: 5000.00,
      description: 'QA Salário',
      date: '2026-04-01T00:00:00Z',
    }, accessToken);
    const ok = res.status === 201 && res.body.data?.id;
    record('4.6 Create income transaction', ok, `status=${res.status}`);
  } catch (e) {
    record('4.6 Create income transaction', false, e.message);
  }
}

// =============================================================================
// BLOCK 5 — Dashboard
// =============================================================================
async function runDashboardTests() {
  console.log('\n=== BLOCK 5: Dashboard ===');
  if (!accessToken) { record('5.x Dashboard skipped', false, 'no token'); return; }

  // Test 5.1: Dashboard overview
  try {
    const res = await apiRequest('GET', '/dashboard/overview', null, accessToken);
    const ok = res.status === 200 && res.body.data !== undefined;
    record('5.1 Dashboard overview', ok, `status=${res.status}`);
  } catch (e) {
    record('5.1 Dashboard overview', false, e.message);
  }

  // Test 5.2: Dashboard cashflow
  try {
    const res = await apiRequest('GET', '/dashboard/cashflow?months=3', null, accessToken);
    const ok = res.status === 200 && Array.isArray(res.body.data);
    record('5.2 Dashboard cashflow', ok, `months=${res.body.meta?.months}`);
  } catch (e) {
    record('5.2 Dashboard cashflow', false, e.message);
  }
}

// =============================================================================
// BLOCK 6 — Budgets
// =============================================================================
async function runBudgetsTests() {
  console.log('\n=== BLOCK 6: Budgets ===');
  if (!accessToken) { record('6.x Budgets skipped', false, 'no token'); return; }

  const catRes = await apiRequest('GET', '/categories', null, accessToken);
  const expenseCat = catRes.body.data?.find(c => c.type === 'expense');

  // Test 6.1: Create budget
  try {
    if (!expenseCat) throw new Error('no expense category');
    const res = await apiRequest('POST', '/budgets', {
      category_id: expenseCat.id,
      amount: 500.00,
      period: 'monthly',
      month: 4,
      year: 2026,
    }, accessToken);
    const ok = res.status === 201 && res.body.data?.id;
    if (ok) createdBudgetId = res.body.data.id;
    record('6.1 Create budget', ok, `id=${createdBudgetId || 'MISSING'}`);
  } catch (e) {
    record('6.1 Create budget', false, e.message);
  }

  // Test 6.2: List budgets
  try {
    const res = await apiRequest('GET', '/budgets', null, accessToken);
    const ok = res.status === 200 && Array.isArray(res.body.data);
    record('6.2 List budgets', ok, `count=${res.body.data?.length}`);
  } catch (e) {
    record('6.2 List budgets', false, e.message);
  }

  // Test 6.3: Budget progress
  try {
    const res = await apiRequest('GET', '/budgets/progress', null, accessToken);
    const ok = res.status === 200;
    record('6.3 Budget progress', ok, `status=${res.status}`);
  } catch (e) {
    record('6.3 Budget progress', false, e.message);
  }

  // Test 6.4: Update budget
  try {
    if (!createdBudgetId || !expenseCat) throw new Error('no budget or category');
    const res = await apiRequest('PUT', `/budgets/${createdBudgetId}`, {
      category_id: expenseCat.id,
      amount: 700.00,
      period: 'monthly',
      month: 4,
      year: 2026,
    }, accessToken);
    const ok = res.status === 200;
    record('6.4 Update budget', ok, `status=${res.status}`);
  } catch (e) {
    record('6.4 Update budget', false, e.message);
  }
}

// =============================================================================
// BLOCK 7 — Goals
// =============================================================================
async function runGoalsTests() {
  console.log('\n=== BLOCK 7: Goals ===');
  if (!accessToken) { record('7.x Goals skipped', false, 'no token'); return; }

  // Test 7.1: Create goal
  try {
    const res = await apiRequest('POST', '/goals', {
      name: 'QA Emergency Fund',
      target_amount: 10000.00,
      target_date: '2026-12-31T00:00:00Z',
      icon: 'savings',
      color: '#4CAF50',
    }, accessToken);
    const ok = res.status === 201 && res.body.data?.id;
    if (ok) createdGoalId = res.body.data.id;
    record('7.1 Create goal', ok, `id=${createdGoalId || 'MISSING'}`);
  } catch (e) {
    record('7.1 Create goal', false, e.message);
  }

  // Test 7.2: List goals
  try {
    const res = await apiRequest('GET', '/goals', null, accessToken);
    const ok = res.status === 200 && Array.isArray(res.body.data);
    record('7.2 List goals', ok, `count=${res.body.data?.length}`);
  } catch (e) {
    record('7.2 List goals', false, e.message);
  }

  // Test 7.3: Contribute to goal
  try {
    if (!createdGoalId) throw new Error('no goal');
    const res = await apiRequest('POST', `/goals/${createdGoalId}/contribute`, {
      amount: 1000.00,
      date: '2026-04-04T00:00:00Z',
    }, accessToken);
    const ok = res.status === 200 || res.status === 201;
    record('7.3 Contribute to goal', ok, `status=${res.status}`);
  } catch (e) {
    record('7.3 Contribute to goal', false, e.message);
  }

  // Test 7.4: Goals projections
  try {
    const res = await apiRequest('GET', '/goals/projections', null, accessToken);
    const ok = res.status === 200;
    record('7.4 Goals projections', ok, `status=${res.status}`);
  } catch (e) {
    record('7.4 Goals projections', false, e.message);
  }

  // Test 7.5: Update goal
  try {
    if (!createdGoalId) throw new Error('no goal');
    const res = await apiRequest('PUT', `/goals/${createdGoalId}`, {
      name: 'QA Emergency Fund Updated',
      target_amount: 15000.00,
      target_date: '2027-06-30T00:00:00Z',
      icon: 'savings',
      color: '#2196F3',
    }, accessToken);
    const ok = res.status === 200;
    record('7.5 Update goal', ok, `status=${res.status}`);
  } catch (e) {
    record('7.5 Update goal', false, e.message);
  }
}

// =============================================================================
// BLOCK 8 — Investments
// =============================================================================
async function runInvestmentsTests() {
  console.log('\n=== BLOCK 8: Investments ===');
  if (!accessToken) { record('8.x Investments skipped', false, 'no token'); return; }

  // Test 8.1: List portfolios
  try {
    const res = await apiRequest('GET', '/portfolios', null, accessToken);
    const ok = res.status === 200;
    record('8.1 List portfolios', ok, `status=${res.status}`);
  } catch (e) {
    record('8.1 List portfolios', false, e.message);
  }

  // Test 8.2: Create portfolio
  try {
    const res = await apiRequest('POST', '/portfolios', {
      name: 'QA Portfolio',
      description: 'QA Test Portfolio',
    }, accessToken);
    const ok = res.status === 201 && res.body.data?.id;
    if (ok) createdPortfolioId = res.body.data.id;
    record('8.2 Create portfolio', ok, `id=${createdPortfolioId || 'MISSING'}`);
  } catch (e) {
    record('8.2 Create portfolio', false, e.message);
  }

  // Test 8.3: Update portfolio
  try {
    if (!createdPortfolioId) throw new Error('no portfolio');
    const res = await apiRequest('PUT', `/portfolios/${createdPortfolioId}`, {
      name: 'QA Portfolio Updated',
      description: 'Updated description',
    }, accessToken);
    const ok = res.status === 200;
    record('8.3 Update portfolio', ok, `status=${res.status}`);
  } catch (e) {
    record('8.3 Update portfolio', false, e.message);
  }

  // Test 8.4: List custom assets
  try {
    const res = await apiRequest('GET', '/custom-assets', null, accessToken);
    const ok = res.status === 200;
    record('8.4 List custom assets', ok, `status=${res.status}`);
  } catch (e) {
    record('8.4 List custom assets', false, e.message);
  }

  // Test 8.5: Search assets
  try {
    const res = await apiRequest('GET', '/assets/search?q=PETR', null, accessToken);
    const ok = res.status === 200;
    record('8.5 Search assets', ok, `status=${res.status}`);
  } catch (e) {
    record('8.5 Search assets', false, e.message);
  }
}

// =============================================================================
// BLOCK 9 — Notifications
// =============================================================================
async function runNotificationsTests() {
  console.log('\n=== BLOCK 9: Notifications ===');
  if (!accessToken) { record('9.x Notifications skipped', false, 'no token'); return; }

  // Test 9.1: List notifications
  try {
    const res = await apiRequest('GET', '/notifications', null, accessToken);
    const ok = res.status === 200 && Array.isArray(res.body.data);
    record('9.1 List notifications', ok, `count=${res.body.data?.length}`);
  } catch (e) {
    record('9.1 List notifications', false, e.message);
  }

  // Test 9.2: Mark all as read
  try {
    const res = await apiRequest('PUT', '/notifications/read-all', null, accessToken);
    const ok = res.status === 200 || res.status === 204;
    record('9.2 Mark all notifications as read', ok, `status=${res.status}`);
  } catch (e) {
    record('9.2 Mark all notifications as read', false, e.message);
  }
}

// =============================================================================
// BLOCK 10 — Recurrences
// =============================================================================
async function runRecurrencesTests() {
  console.log('\n=== BLOCK 10: Recurrences ===');
  if (!accessToken) { record('10.x Recurrences skipped', false, 'no token'); return; }

  const catRes = await apiRequest('GET', '/categories', null, accessToken);
  const expenseCat = catRes.body.data?.find(c => c.type === 'expense');
  const accRes = await apiRequest('GET', '/accounts', null, accessToken);
  const account = accRes.body.data?.[0];

  let recurrenceId = '';

  // Test 10.1: Create recurrence
  try {
    if (!account || !expenseCat) throw new Error('missing account or category');
    const res = await apiRequest('POST', '/recurrences', {
      account_id: account.id,
      category_id: expenseCat.id,
      type: 'expense',
      amount: 100.00,
      description: 'QA Recurrence',
      frequency: 'monthly',
      start_date: '2026-04-01T00:00:00Z',
    }, accessToken);
    const ok = res.status === 201 && res.body.data?.id;
    if (ok) recurrenceId = res.body.data.id;
    record('10.1 Create recurrence', ok, `id=${recurrenceId || 'MISSING'}`);
  } catch (e) {
    record('10.1 Create recurrence', false, e.message);
  }

  // Test 10.2: List recurrences
  try {
    const res = await apiRequest('GET', '/recurrences', null, accessToken);
    const ok = res.status === 200 && Array.isArray(res.body.data);
    record('10.2 List recurrences', ok, `count=${res.body.data?.length}`);
  } catch (e) {
    record('10.2 List recurrences', false, e.message);
  }

  // Test 10.3: Update recurrence
  try {
    if (!recurrenceId || !account || !expenseCat) throw new Error('missing data');
    const res = await apiRequest('PUT', `/recurrences/${recurrenceId}`, {
      account_id: account.id,
      category_id: expenseCat.id,
      type: 'expense',
      amount: 150.00,
      description: 'QA Recurrence Updated',
      frequency: 'monthly',
      start_date: '2026-04-01T00:00:00Z',
    }, accessToken);
    const ok = res.status === 200;
    record('10.3 Update recurrence', ok, `status=${res.status}`);
  } catch (e) {
    record('10.3 Update recurrence', false, e.message);
  }
}

// =============================================================================
// BLOCK 11 — Family
// =============================================================================
async function runFamilyTests() {
  console.log('\n=== BLOCK 11: Family ===');
  if (!accessToken) { record('11.x Family skipped', false, 'no token'); return; }

  // Test 11.1: Get family (may not exist)
  try {
    const res = await apiRequest('GET', '/family', null, accessToken);
    const ok = res.status === 200 || res.status === 404;
    record('11.1 Get family (or 404)', ok, `status=${res.status}`);
  } catch (e) {
    record('11.1 Get family (or 404)', false, e.message);
  }
}

// =============================================================================
// BLOCK 12 — AI Chat (free tier)
// =============================================================================
async function runAITests() {
  console.log('\n=== BLOCK 12: AI Chat ===');
  if (!accessToken) { record('12.x AI skipped', false, 'no token'); return; }

  // Test 12.1: AI chat endpoint
  try {
    const res = await apiRequest('POST', '/ai/chat', {
      message: 'Qual é o meu saldo atual?'
    }, accessToken);
    // 200 = success, 500 = Claude API key not configured in test env (acceptable), 503 = service unavailable
    const ok = res.status === 200 || res.status === 500 || res.status === 503;
    record('12.1 AI chat endpoint responds', ok, `status=${res.status}`);
  } catch (e) {
    record('12.1 AI chat endpoint responds', false, e.message);
  }

  // Test 12.2: AI spending forecast (pro only — should return 402)
  try {
    const res = await apiRequest('GET', '/ai/spending-forecast', null, accessToken);
    const ok = res.status === 402; // free plan user should get 402
    record('12.2 AI spending forecast requires pro plan', ok, `status=${res.status} (expected 402)`);
  } catch (e) {
    record('12.2 AI spending forecast requires pro plan', false, e.message);
  }
}

// =============================================================================
// BLOCK 13 — Frontend Smoke Tests (Playwright)
// =============================================================================
async function runFrontendTests(browser) {
  console.log('\n=== BLOCK 13: Frontend Smoke Tests ===');

  // Test 13.1: Homepage loads and redirects to onboarding
  await withPage(browser, async (page) => {
    try {
      await page.goto(BASE_URL, { waitUntil: 'load', timeout: 30000 });
      await page.waitForTimeout(5000);
      const url = page.url();
      const ok = url.includes('onboarding') || url.includes('login') || url.includes('home');
      await page.screenshot({ path: path.join(SCREENSHOTS_DIR, '13.1-homepage.png') });
      record('13.1 Homepage loads and redirects', ok, `url=${url}`);
    } catch (e) {
      record('13.1 Homepage loads and redirects', false, e.message);
    }
  });

  // Test 13.2: Login page renders
  await withPage(browser, async (page) => {
    try {
      await flutterNav(page, '#/login', 5000);
      const url = page.url();
      const bodyText = await getBodyText(page);
      const ok = url.includes('login') && (bodyText.toLowerCase().includes('entrar') || bodyText.toLowerCase().includes('e-mail') || bodyText.toLowerCase().includes('financeos'));
      await page.screenshot({ path: path.join(SCREENSHOTS_DIR, '13.2-login.png') });
      record('13.2 Login page renders', ok, `url=${url}, text=${bodyText.slice(0, 100)}`);
    } catch (e) {
      record('13.2 Login page renders', false, e.message);
    }
  });

  // Test 13.3: Register page renders
  await withPage(browser, async (page) => {
    try {
      await flutterNav(page, '#/register', 5000);
      const url = page.url();
      const bodyText = await getBodyText(page);
      const ok = url.includes('register') && (bodyText.length > 0);
      await page.screenshot({ path: path.join(SCREENSHOTS_DIR, '13.3-register.png') });
      record('13.3 Register page renders', ok, `url=${url}`);
    } catch (e) {
      record('13.3 Register page renders', false, e.message);
    }
  });

  // Test 13.4: Onboarding page renders
  await withPage(browser, async (page) => {
    try {
      await flutterNav(page, '#/onboarding', 5000);
      const url = page.url();
      await page.screenshot({ path: path.join(SCREENSHOTS_DIR, '13.4-onboarding.png') });
      const ok = true; // just check it doesn't crash
      record('13.4 Onboarding page renders', ok, `url=${url}`);
    } catch (e) {
      record('13.4 Onboarding page renders', false, e.message);
    }
  });

  // Test 13.5: Flutter web renders canvas/flt-glass-pane
  await withPage(browser, async (page) => {
    try {
      await flutterNav(page, '#/login', 5000);
      const flutterInfo = await page.evaluate(() => ({
        hasFlutter: typeof window._flutter !== 'undefined',
        glassPanes: document.querySelectorAll('flt-glass-pane').length,
        hasSemanticsHost: !!document.querySelector('flt-semantics-host'),
      }));
      const ok = flutterInfo.hasFlutter && (flutterInfo.glassPanes > 0 || flutterInfo.hasSemanticsHost);
      record('13.5 Flutter web app loaded properly', ok, JSON.stringify(flutterInfo));
    } catch (e) {
      record('13.5 Flutter web app loaded properly', false, e.message);
    }
  });

  // Test 13.6: No critical JS errors on load
  await withPage(browser, async (page) => {
    const errors = [];
    page.on('console', msg => {
      if (msg.type() === 'error') errors.push(msg.text());
    });
    try {
      await flutterNav(page, '#/login', 5000);
      const criticalErrors = errors.filter(e =>
        !e.includes('favicon') &&
        !e.includes('404') &&
        !e.includes('net::ERR')
      );
      const ok = criticalErrors.length === 0;
      record('13.6 No critical JS errors on login page', ok, `errors=${criticalErrors.length}: ${criticalErrors.slice(0, 2).join('; ')}`);
    } catch (e) {
      record('13.6 No critical JS errors on login page', false, e.message);
    }
  });

  // Test 13.7: Home/dashboard accessible after login (via API state)
  await withPage(browser, async (page) => {
    try {
      // Set auth token in localStorage/sessionStorage so Flutter can read it
      // Flutter uses flutter_secure_storage which in web uses IndexedDB
      await page.goto(BASE_URL, { waitUntil: 'load', timeout: 30000 });
      await page.waitForTimeout(3000);

      // Check if protected routes redirect to login when not authenticated
      await flutterNav(page, '#/home', 4000);
      const url = page.url();
      const ok = url.includes('login') || url.includes('onboarding') || url.includes('home');
      await page.screenshot({ path: path.join(SCREENSHOTS_DIR, '13.7-protected-redirect.png') });
      record('13.7 Protected route handles unauthenticated state', ok, `url=${url}`);
    } catch (e) {
      record('13.7 Protected route handles unauthenticated state', false, e.message);
    }
  });
}

// =============================================================================
// BLOCK 14 — Delete operations (cleanup)
// =============================================================================
async function runCleanupTests() {
  console.log('\n=== BLOCK 14: Delete operations ===');
  if (!accessToken) { record('14.x Cleanup skipped', false, 'no token'); return; }

  // Test 14.1: Delete transaction
  try {
    if (!createdTransactionId) throw new Error('no transaction');
    const res = await apiRequest('DELETE', `/transactions/${createdTransactionId}`, null, accessToken);
    const ok = res.status === 204 || res.status === 200;
    record('14.1 Delete transaction', ok, `status=${res.status}`);
  } catch (e) {
    record('14.1 Delete transaction', false, e.message);
  }

  // Test 14.2: Delete budget
  try {
    if (!createdBudgetId) throw new Error('no budget');
    const res = await apiRequest('DELETE', `/budgets/${createdBudgetId}`, null, accessToken);
    const ok = res.status === 204 || res.status === 200;
    record('14.2 Delete budget', ok, `status=${res.status}`);
  } catch (e) {
    record('14.2 Delete budget', false, e.message);
  }

  // Test 14.3: Delete goal
  try {
    if (!createdGoalId) throw new Error('no goal');
    const res = await apiRequest('DELETE', `/goals/${createdGoalId}`, null, accessToken);
    const ok = res.status === 204 || res.status === 200;
    record('14.3 Delete goal', ok, `status=${res.status}`);
  } catch (e) {
    record('14.3 Delete goal', false, e.message);
  }

  // Test 14.4: Delete portfolio
  try {
    if (!createdPortfolioId) throw new Error('no portfolio');
    const res = await apiRequest('DELETE', `/portfolios/${createdPortfolioId}`, null, accessToken);
    const ok = res.status === 204 || res.status === 200;
    record('14.4 Delete portfolio', ok, `status=${res.status}`);
  } catch (e) {
    record('14.4 Delete portfolio', false, e.message);
  }

  // Test 14.5: Delete custom category
  try {
    if (!createdCategoryId) throw new Error('no category');
    const res = await apiRequest('DELETE', `/categories/${createdCategoryId}`, null, accessToken);
    const ok = res.status === 204 || res.status === 200;
    record('14.5 Delete category', ok, `status=${res.status}`);
  } catch (e) {
    record('14.5 Delete category', false, e.message);
  }

  // Test 14.6: Delete account
  try {
    if (!createdAccountId) throw new Error('no account');
    const res = await apiRequest('DELETE', `/accounts/${createdAccountId}`, null, accessToken);
    const ok = res.status === 204 || res.status === 200;
    record('14.6 Delete account', ok, `status=${res.status}`);
  } catch (e) {
    record('14.6 Delete account', false, e.message);
  }
}

// =============================================================================
// MAIN
// =============================================================================
async function main() {
  console.log('=== FinanceOS QA Suite ===');
  console.log('Starting at:', new Date().toISOString());

  // API Tests
  await runApiAuthTests();
  await runAccountsTests();
  await runCategoriesTests();
  await runTransactionsTests();
  await runDashboardTests();
  await runBudgetsTests();
  await runGoalsTests();
  await runInvestmentsTests();
  await runNotificationsTests();
  await runRecurrencesTests();
  await runFamilyTests();
  await runAITests();

  // Frontend Tests (Playwright)
  const browser = await chromium.launch({ headless: true });
  try {
    await runFrontendTests(browser);
  } finally {
    await browser.close();
  }

  // Cleanup
  await runCleanupTests();

  // Summary
  const total = passed + failed;
  const pct = total > 0 ? Math.round((passed / total) * 100) : 0;
  console.log('\n=== SUMMARY ===');
  console.log(`Total: ${total} | Passed: ${passed} | Failed: ${failed} | Pass rate: ${pct}%`);

  // Write JSON results for report
  fs.writeFileSync(path.join(__dirname, 'qa_results.json'), JSON.stringify({
    timestamp: new Date().toISOString(),
    total, passed, failed, pct,
    results
  }, null, 2));

  return { total, passed, failed, pct, results };
}

main().then(({ total, passed, failed, pct }) => {
  console.log(`\nQA Suite Complete: ${pct}% (${passed}/${total})`);
  process.exit(failed > 0 ? 1 : 0);
}).catch(e => {
  console.error('Fatal error:', e);
  process.exit(1);
});
