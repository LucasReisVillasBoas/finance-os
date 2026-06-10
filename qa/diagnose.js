const { chromium } = require('playwright');

async function diagnose() {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  const page = await context.newPage();

  const consoleErrors = [];
  const networkErrors = [];
  const apiCalls = [];

  page.on('console', msg => {
    if (msg.type() === 'error') {
      consoleErrors.push(msg.text());
    }
  });

  page.on('requestfailed', req => {
    networkErrors.push(`FAILED: ${req.method()} ${req.url()} - ${req.failure().errorText}`);
  });

  page.on('response', async res => {
    const url = res.url();
    if (url.includes('/api/')) {
      let body = '';
      try { body = await res.text(); } catch(e) {}
      apiCalls.push({ status: res.status(), url, body: body.slice(0, 500) });
    }
  });

  console.log('=== PHASE 1: Open homepage ===');
  await page.goto('http://localhost:3000/', { waitUntil: 'networkidle', timeout: 30000 });
  await page.waitForTimeout(3000);

  const url1 = page.url();
  const bodyText1 = await page.evaluate(() => document.body.innerText).catch(() => '');
  console.log('URL after load:', url1);
  console.log('Body text (first 500):', bodyText1.slice(0, 500));

  console.log('\n=== PHASE 2: Navigate to login ===');
  await page.goto('http://localhost:3000/#/login', { waitUntil: 'networkidle', timeout: 30000 });
  await page.waitForTimeout(3000);

  const url2 = page.url();
  const bodyText2 = await page.evaluate(() => document.body.innerText).catch(() => '');
  console.log('URL after login nav:', url2);
  console.log('Body text (first 500):', bodyText2.slice(0, 500));

  // Take screenshot
  await page.screenshot({ path: '/Users/lucasreis/Documents/projects/personal/FinanceOS/qa/screenshots/login_page.png', fullPage: true });

  console.log('\n=== PHASE 3: Try registering ===');
  await page.goto('http://localhost:3000/#/register', { waitUntil: 'networkidle', timeout: 30000 });
  await page.waitForTimeout(3000);
  const url3 = page.url();
  const bodyText3 = await page.evaluate(() => document.body.innerText).catch(() => '');
  console.log('URL after register nav:', url3);
  console.log('Body text (first 500):', bodyText3.slice(0, 500));
  await page.screenshot({ path: '/Users/lucasreis/Documents/projects/personal/FinanceOS/qa/screenshots/register_page.png', fullPage: true });

  console.log('\n=== PHASE 4: Try onboarding ===');
  await page.goto('http://localhost:3000/#/onboarding', { waitUntil: 'networkidle', timeout: 30000 });
  await page.waitForTimeout(3000);
  const url4 = page.url();
  const bodyText4 = await page.evaluate(() => document.body.innerText).catch(() => '');
  console.log('URL after onboarding nav:', url4);
  console.log('Body text (first 500):', bodyText4.slice(0, 500));
  await page.screenshot({ path: '/Users/lucasreis/Documents/projects/personal/FinanceOS/qa/screenshots/onboarding_page.png', fullPage: true });

  console.log('\n=== PHASE 5: Check Flutter loaded ===');
  const flutterLoaded = await page.evaluate(() => {
    return {
      hasFlutter: typeof window._flutter !== 'undefined',
      hasFlutterConfig: typeof window.flutterConfiguration !== 'undefined',
      glassPanes: document.querySelectorAll('flt-glass-pane').length,
      canvasCount: document.querySelectorAll('canvas').length,
      bodyHTML: document.body.innerHTML.slice(0, 1000)
    };
  });
  console.log('Flutter info:', JSON.stringify(flutterLoaded, null, 2));

  console.log('\n=== PHASE 6: Try to interact with login form ===');
  await page.goto('http://localhost:3000/#/login', { waitUntil: 'networkidle', timeout: 30000 });
  await page.waitForTimeout(5000);

  // Try clicking on the page center (Flutter canvas interaction)
  const viewport = page.viewportSize();
  await page.mouse.click(viewport.width / 2, viewport.height / 2);
  await page.waitForTimeout(1000);

  // Try finding text input
  const inputs = await page.locator('input').all();
  console.log('Found inputs:', inputs.length);

  const fltInputs = await page.locator('flt-text-editing-host input').all();
  console.log('Found flt inputs:', fltInputs.length);

  if (fltInputs.length === 0) {
    // Try tabbing to focus Flutter inputs
    await page.keyboard.press('Tab');
    await page.waitForTimeout(500);
    const fltInputsAfterTab = await page.locator('flt-text-editing-host input').all();
    console.log('Found flt inputs after Tab:', fltInputsAfterTab.length);
  }

  await page.screenshot({ path: '/Users/lucasreis/Documents/projects/personal/FinanceOS/qa/screenshots/login_interact.png', fullPage: true });

  console.log('\n=== SUMMARY ===');
  console.log('Console errors:', consoleErrors.length);
  consoleErrors.forEach(e => console.log(' - ', e));
  console.log('Network errors:', networkErrors.length);
  networkErrors.forEach(e => console.log(' - ', e));
  console.log('API calls:', apiCalls.length);
  apiCalls.forEach(c => console.log(` - ${c.status} ${c.url}: ${c.body.slice(0, 200)}`));

  await browser.close();
}

diagnose().catch(console.error);
