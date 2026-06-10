const { chromium } = require('playwright');

async function check() {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();

  console.log('=== Checking Login Page Semantics ===');
  await page.goto('http://localhost:3000/#/login', { waitUntil: 'load' });
  await page.waitForTimeout(6000);

  const info = await page.evaluate(() => {
    const host = document.querySelector('flt-semantics-host');
    const body = document.body;
    return {
      bodyInnerText: body.innerText?.slice(0, 500) || '',
      bodyTextContent: body.textContent?.slice(0, 500) || '',
      semanticsHTML: host?.innerHTML?.slice(0, 1000) || 'no semantics host',
      semanticsText: host?.innerText?.slice(0, 500) || '',
      allText: document.documentElement.textContent?.slice(0, 1000) || '',
      fltSemNodes: document.querySelectorAll('flt-semantics').length,
      fltSemTexts: Array.from(document.querySelectorAll('[aria-label]')).map(e => e.getAttribute('aria-label')).slice(0, 20),
    };
  });

  console.log('Body innerText:', info.bodyInnerText);
  console.log('Body textContent (first 200):', info.bodyTextContent.slice(0, 200));
  console.log('Semantics host innerText:', info.semanticsText);
  console.log('flt-semantics nodes:', info.fltSemNodes);
  console.log('aria-labels:', JSON.stringify(info.fltSemTexts));

  // Try clicking the page to enable Flutter accessibility
  await page.click('body');
  await page.waitForTimeout(2000);

  const info2 = await page.evaluate(() => {
    const accessBtn = document.querySelector('flt-semantics-placeholder');
    return {
      accessibilityBtn: accessBtn?.getAttribute('aria-label') || 'no button',
      fltSemNodes: document.querySelectorAll('flt-semantics').length,
    };
  });
  console.log('After click:', JSON.stringify(info2));

  // Try enabling accessibility
  await page.evaluate(() => {
    const btn = document.querySelector('flt-semantics-placeholder');
    if (btn) btn.click();
  });
  await page.waitForTimeout(3000);

  const info3 = await page.evaluate(() => {
    const host = document.querySelector('flt-semantics-host');
    return {
      semanticsText: host?.innerText?.slice(0, 500) || '',
      fltSemNodes: document.querySelectorAll('flt-semantics').length,
      ariaLabels: Array.from(document.querySelectorAll('[aria-label]')).map(e => e.getAttribute('aria-label')).slice(0, 30),
    };
  });
  console.log('After accessibility enable:', JSON.stringify(info3, null, 2));

  await page.screenshot({ path: '/Users/lucasreis/Documents/projects/personal/FinanceOS/qa/screenshots/semantics_check.png' });
  await browser.close();
}

check().catch(console.error);
