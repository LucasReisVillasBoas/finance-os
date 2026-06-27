# QA AGENT — FinanceOS Full Test Suite

## Missão
Você é um QA Engineer sênior especializado em testes automatizados.
Sua missão é testar TODA a aplicação FinanceOS em http://localhost:3000
usando Playwright, identificar TODOS os bugs, corrigi-los no código-fonte
e validar novamente até tudo funcionar.

## Regras
1. NUNCA pare — se um teste falhar, investigue, corrija e reteste
2. SEMPRE salve resultados em QA_REPORT.md
3. SEMPRE abra o DevTools (console + network) para capturar erros JS e requests HTTP
4. Após corrigir um bug, revalide o fluxo completo do início
5. Não marque como ✅ sem confirmar visualmente que funcionou

---

## FASE 0 — Setup do ambiente de testes

```bash
# Instalar Playwright se não tiver
npm init -y
npm install playwright @playwright/test
npx playwright install chromium

# Verificar se o app está rodando
curl -s http://localhost:3000 | head -20
curl -s http://localhost:8000/health || echo "API offline"

# Verificar logs da API
docker logs financeos-api --tail=50 2>&1

# Verificar logs do banco
docker logs financeos-postgres --tail=20 2>&1
```

---

## FASE 1 — Diagnóstico inicial (loading infinito)

Antes de qualquer teste, investigue o bug reportado de loading infinito após login.

```javascript
// diagnose.js — rodar com: node diagnose.js
const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false, slowMo: 500 });
  const context = await browser.newContext();
  const page = await context.newPage();

  // Capturar TODOS os erros de console
  page.on('console', msg => {
    if (msg.type() === 'error') console.log('❌ CONSOLE ERROR:', msg.text());
    else console.log('📋 CONSOLE:', msg.type(), msg.text());
  });

  // Capturar TODAS as requests de rede
  page.on('request', req => {
    console.log('📤 REQUEST:', req.method(), req.url());
  });

  page.on('response', resp => {
    const status = resp.status();
    const emoji = status >= 400 ? '❌' : '✅';
    console.log(`${emoji} RESPONSE: ${status} ${resp.url()}`);
    if (status >= 400) {
      resp.text().then(body => console.log('   BODY:', body.substring(0, 500)));
    }
  });

  page.on('requestfailed', req => {
    console.log('💥 REQUEST FAILED:', req.url(), req.failure()?.errorText);
  });

  // Acessar app
  await page.goto('http://localhost:3000');
  await page.waitForTimeout(3000);
  console.log('URL atual:', page.url());
  console.log('Título:', await page.title());

  // Fazer login
  try {
    await page.fill('[data-testid="email"], input[type="email"]', 'test@financeos.com');
    await page.fill('[data-testid="password"], input[type="password"]', 'Test@123456');
    await page.click('[data-testid="login-btn"], button[type="submit"]');
    
    // Aguardar 10 segundos e capturar estado
    await page.waitForTimeout(10000);
    console.log('URL após login:', page.url());
    
    // Capturar screenshot
    await page.screenshot({ path: 'after_login.png', fullPage: true });
    console.log('Screenshot salvo: after_login.png');
    
    // Verificar o que está na tela
    const bodyText = await page.evaluate(() => document.body.innerText.substring(0, 1000));
    console.log('Conteúdo da tela:', bodyText);
    
    // Verificar se há spinner/loading
    const hasSpinner = await page.$('.loading, .spinner, [class*="load"], [class*="spin"]');
    console.log('Spinner encontrado:', !!hasSpinner);
    
    // Verificar localStorage/sessionStorage
    const storage = await page.evaluate(() => ({
      localStorage: Object.entries(localStorage),
      sessionStorage: Object.entries(sessionStorage)
    }));
    console.log('Storage:', JSON.stringify(storage, null, 2));
    
  } catch (e) {
    console.log('❌ Erro no login:', e.message);
    await page.screenshot({ path: 'login_error.png', fullPage: true });
  }

  await browser.close();
})();
```

Após rodar o diagnóstico:
1. Analise TODOS os erros de console
2. Analise TODOS os requests com status 4xx/5xx
3. Identifique a causa raiz do loading infinito
4. Corrija no código-fonte (Flutter ou Go conforme necessário)
5. Rebuild: `cd apps/web && flutter build web && docker restart financeos-nginx`

---

## FASE 2 — Criar usuário de teste via API

```bash
# Criar usuário de teste diretamente na API
curl -s -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "QA Tester",
    "email": "qa@financeos.com",
    "password": "QA@Test123456"
  }' | jq .

# Fazer login e salvar token
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "qa@financeos.com", "password": "QA@Test123456"}' \
  | jq -r '.data.access_token')

echo "Token: $TOKEN"

# Testar endpoint protegido
curl -s http://localhost:8000/api/v1/accounts \
  -H "Authorization: Bearer $TOKEN" | jq .
```

Se os endpoints da API retornarem erro, investigue e corrija o Go antes de continuar com Playwright.

---

## FASE 3 — Suite de testes Playwright completa

Após corrigir o loading infinito, rode a suite completa:

```javascript
// qa_suite.js
const { chromium } = require('playwright');
const fs = require('fs');

const BASE_URL = 'http://localhost:3000';
const API_URL = 'http://localhost:8000';
const TEST_USER = { name: 'QA Tester', email: 'qa@financeos.com', password: 'QA@Test123456' };

const report = { passed: [], failed: [], total: 0 };

async function test(name, fn, page) {
  report.total++;
  try {
    await fn(page);
    report.passed.push(name);
    console.log(`✅ ${name}`);
  } catch (e) {
    report.failed.push({ name, error: e.message });
    console.log(`❌ ${name}: ${e.message}`);
    await page.screenshot({ path: `screenshots/fail_${Date.now()}.png`, fullPage: true });
  }
}

async function waitForNavigate(page, path, timeout = 10000) {
  await page.waitForURL(`**${path}*`, { timeout });
}

async function waitForNoLoading(page, timeout = 15000) {
  // Aguardar spinner sumir
  try {
    await page.waitForFunction(
      () => !document.querySelector('.loading, .spinner, [class*="CircularProgress"]'),
      { timeout }
    );
  } catch (e) {
    // ignore se não houver spinner
  }
  await page.waitForTimeout(1000);
}

(async () => {
  fs.mkdirSync('screenshots', { recursive: true });
  const browser = await chromium.launch({ headless: false, slowMo: 300 });
  const context = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await context.newPage();

  // Capturar erros silenciosos
  const errors = [];
  page.on('console', msg => { if (msg.type() === 'error') errors.push(msg.text()); });
  page.on('pageerror', e => errors.push(e.message));

  // ──────────────────────────────────────────────
  // BLOCO 1: AUTENTICAÇÃO
  // ──────────────────────────────────────────────

  await test('1.1 — Splash Screen carrega', async (page) => {
    await page.goto(BASE_URL);
    await page.waitForTimeout(2000);
    const title = await page.title();
    if (!title) throw new Error('Título vazio');
  }, page);

  await test('1.2 — Redireciona para onboarding/login sem sessão', async (page) => {
    await page.goto(BASE_URL);
    await page.waitForTimeout(3000);
    const url = page.url();
    if (!url.includes('login') && !url.includes('onboarding') && !url.includes('register')) {
      throw new Error(`URL inesperada: ${url}`);
    }
  }, page);

  await test('1.3 — Cadastro de novo usuário', async (page) => {
    await page.goto(`${BASE_URL}/register`);
    await page.waitForTimeout(1000);
    
    const nameField = await page.$('input[name="name"], input[placeholder*="nome"], input[placeholder*="Name"]');
    if (nameField) await nameField.fill(TEST_USER.name);
    
    await page.fill('input[type="email"]', TEST_USER.email);
    await page.fill('input[type="password"]', TEST_USER.password);
    
    const confirmField = await page.$('input[placeholder*="confirm"], input[name="confirmPassword"]');
    if (confirmField) await confirmField.fill(TEST_USER.password);
    
    await page.click('button[type="submit"]');
    await page.waitForTimeout(3000);
    await page.screenshot({ path: 'screenshots/after_register.png' });
  }, page);

  await test('1.4 — Login com credenciais válidas', async (page) => {
    await page.goto(`${BASE_URL}/login`);
    await page.waitForTimeout(1000);
    await page.fill('input[type="email"]', TEST_USER.email);
    await page.fill('input[type="password"]', TEST_USER.password);
    await page.click('button[type="submit"]');
    await waitForNoLoading(page, 15000);
    
    const url = page.url();
    if (url.includes('login')) throw new Error('Ainda na tela de login após submit');
    await page.screenshot({ path: 'screenshots/after_login.png' });
  }, page);

  await test('1.5 — Dashboard carrega após login (sem loading infinito)', async (page) => {
    await waitForNoLoading(page, 20000);
    const url = page.url();
    const body = await page.evaluate(() => document.body.innerText);
    if (body.length < 50) throw new Error('Página parece vazia: ' + body.substring(0, 100));
    await page.screenshot({ path: 'screenshots/dashboard.png' });
  }, page);

  await test('1.6 — Login com credenciais inválidas mostra erro', async (page) => {
    await page.goto(`${BASE_URL}/login`);
    await page.fill('input[type="email"]', 'wrong@email.com');
    await page.fill('input[type="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');
    await page.waitForTimeout(3000);
    
    const errorVisible = await page.$('[class*="error"], [class*="Error"], .snackbar');
    if (!errorVisible) throw new Error('Mensagem de erro não exibida para credenciais inválidas');
  }, page);

  // Re-login para continuar testes
  await page.goto(`${BASE_URL}/login`);
  await page.fill('input[type="email"]', TEST_USER.email);
  await page.fill('input[type="password"]', TEST_USER.password);
  await page.click('button[type="submit"]');
  await waitForNoLoading(page, 15000);

  // ──────────────────────────────────────────────
  // BLOCO 2: CONTAS BANCÁRIAS
  // ──────────────────────────────────────────────

  await test('2.1 — Navegar para Contas', async (page) => {
    const accountsLink = await page.$('a[href*="account"], [data-testid*="account"], nav >> text=Conta');
    if (accountsLink) {
      await accountsLink.click();
    } else {
      await page.goto(`${BASE_URL}/accounts`);
    }
    await waitForNoLoading(page);
    await page.screenshot({ path: 'screenshots/accounts.png' });
  }, page);

  await test('2.2 — Criar conta bancária', async (page) => {
    const fab = await page.$('[data-testid="fab"], button >> text=+, [aria-label*="add"], [aria-label*="criar"]');
    if (fab) await fab.click();
    else await page.goto(`${BASE_URL}/accounts/new`);
    
    await page.waitForTimeout(1000);
    await page.fill('input[name="name"], input[placeholder*="nome"]', 'Conta Nubank');
    
    await page.click('button[type="submit"], button >> text=Salvar, button >> text=Criar');
    await page.waitForTimeout(2000);
    await page.screenshot({ path: 'screenshots/create_account.png' });
  }, page);

  await test('2.3 — Conta aparece na lista', async (page) => {
    const content = await page.evaluate(() => document.body.innerText);
    if (!content.includes('Nubank')) throw new Error('Conta criada não aparece na lista');
  }, page);

  // ──────────────────────────────────────────────
  // BLOCO 3: TRANSAÇÕES
  // ──────────────────────────────────────────────

  await test('3.1 — Navegar para Transações', async (page) => {
    const link = await page.$('a[href*="transaction"], nav >> text=Transaç');
    if (link) await link.click();
    else await page.goto(`${BASE_URL}/transactions`);
    await waitForNoLoading(page);
    await page.screenshot({ path: 'screenshots/transactions.png' });
  }, page);

  await test('3.2 — Criar transação de despesa', async (page) => {
    const fab = await page.$('[data-testid="fab"], button >> text=+');
    if (fab) await fab.click();
    else await page.goto(`${BASE_URL}/transactions/new`);
    
    await page.waitForTimeout(1000);
    
    // Preencher valor
    const amountField = await page.$('input[name="amount"], input[placeholder*="valor"], input[placeholder*="0,00"]');
    if (amountField) await amountField.fill('150.00');
    
    // Descrição
    const descField = await page.$('input[name="description"], input[placeholder*="descrição"]');
    if (descField) await descField.fill('Supermercado teste');
    
    await page.click('button[type="submit"], button >> text=Salvar');
    await page.waitForTimeout(2000);
    await page.screenshot({ path: 'screenshots/create_transaction.png' });
  }, page);

  await test('3.3 — Criar transação de receita', async (page) => {
    const fab = await page.$('[data-testid="fab"], button >> text=+');
    if (fab) await fab.click();
    else await page.goto(`${BASE_URL}/transactions/new`);
    
    await page.waitForTimeout(1000);
    
    // Selecionar tipo receita
    const incomeBtn = await page.$('button >> text=Receita, [value="income"]');
    if (incomeBtn) await incomeBtn.click();
    
    const amountField = await page.$('input[name="amount"], input[placeholder*="valor"]');
    if (amountField) await amountField.fill('5000.00');
    
    const descField = await page.$('input[name="description"]');
    if (descField) await descField.fill('Salário teste');
    
    await page.click('button[type="submit"], button >> text=Salvar');
    await page.waitForTimeout(2000);
    await page.screenshot({ path: 'screenshots/create_income.png' });
  }, page);

  await test('3.4 — Transações aparecem na lista', async (page) => {
    await page.goto(`${BASE_URL}/transactions`);
    await waitForNoLoading(page);
    const content = await page.evaluate(() => document.body.innerText);
    if (!content.includes('Supermercado') && !content.includes('Salário')) {
      throw new Error('Transações criadas não aparecem na lista');
    }
  }, page);

  // ──────────────────────────────────────────────
  // BLOCO 4: ORÇAMENTO
  // ──────────────────────────────────────────────

  await test('4.1 — Navegar para Orçamentos', async (page) => {
    const link = await page.$('a[href*="budget"], nav >> text=Orçamento');
    if (link) await link.click();
    else await page.goto(`${BASE_URL}/budgets`);
    await waitForNoLoading(page);
    await page.screenshot({ path: 'screenshots/budgets.png' });
  }, page);

  await test('4.2 — Criar orçamento', async (page) => {
    const fab = await page.$('[data-testid="fab"], button >> text=+');
    if (fab) await fab.click();
    else await page.goto(`${BASE_URL}/budgets/new`);
    
    await page.waitForTimeout(1000);
    
    const amountField = await page.$('input[name="amount"], input[placeholder*="valor"]');
    if (amountField) await amountField.fill('1000.00');
    
    await page.click('button[type="submit"], button >> text=Salvar');
    await page.waitForTimeout(2000);
    await page.screenshot({ path: 'screenshots/create_budget.png' });
  }, page);

  // ──────────────────────────────────────────────
  // BLOCO 5: INVESTIMENTOS
  // ──────────────────────────────────────────────

  await test('5.1 — Navegar para Investimentos', async (page) => {
    const link = await page.$('a[href*="invest"], nav >> text=Invest');
    if (link) await link.click();
    else await page.goto(`${BASE_URL}/investments`);
    await waitForNoLoading(page);
    await page.screenshot({ path: 'screenshots/investments.png' });
  }, page);

  await test('5.2 — Adicionar ativo (ação BR)', async (page) => {
    const fab = await page.$('[data-testid="fab"], button >> text=+');
    if (fab) await fab.click();
    else await page.goto(`${BASE_URL}/investments/new`);
    
    await page.waitForTimeout(1000);
    
    // Buscar PETR4
    const searchField = await page.$('input[placeholder*="ticker"], input[placeholder*="buscar"], input[placeholder*="ativo"]');
    if (searchField) {
      await searchField.fill('PETR4');
      await page.waitForTimeout(2000);
      
      const result = await page.$('text=PETR4');
      if (result) await result.click();
    }
    
    const qtyField = await page.$('input[name="quantity"], input[placeholder*="quantidade"]');
    if (qtyField) await qtyField.fill('10');
    
    const priceField = await page.$('input[name="price"], input[placeholder*="preço"]');
    if (priceField) await priceField.fill('38.50');
    
    await page.click('button[type="submit"], button >> text=Salvar');
    await page.waitForTimeout(2000);
    await page.screenshot({ path: 'screenshots/add_investment.png' });
  }, page);

  // ──────────────────────────────────────────────
  // BLOCO 6: METAS
  // ──────────────────────────────────────────────

  await test('6.1 — Navegar para Metas', async (page) => {
    const link = await page.$('a[href*="goal"], nav >> text=Meta');
    if (link) await link.click();
    else await page.goto(`${BASE_URL}/goals`);
    await waitForNoLoading(page);
    await page.screenshot({ path: 'screenshots/goals.png' });
  }, page);

  await test('6.2 — Criar meta financeira', async (page) => {
    const fab = await page.$('[data-testid="fab"], button >> text=+');
    if (fab) await fab.click();
    else await page.goto(`${BASE_URL}/goals/new`);
    
    await page.waitForTimeout(1000);
    
    const nameField = await page.$('input[name="name"], input[placeholder*="nome"]');
    if (nameField) await nameField.fill('Reserva de emergência');
    
    const amountField = await page.$('input[name="target_amount"], input[placeholder*="valor"]');
    if (amountField) await amountField.fill('30000.00');
    
    await page.click('button[type="submit"], button >> text=Salvar');
    await page.waitForTimeout(2000);
    await page.screenshot({ path: 'screenshots/create_goal.png' });
  }, page);

  // ──────────────────────────────────────────────
  // BLOCO 7: DASHBOARD
  // ──────────────────────────────────────────────

  await test('7.1 — Dashboard mostra saldo total', async (page) => {
    await page.goto(`${BASE_URL}/dashboard`);
    await waitForNoLoading(page, 20000);
    const content = await page.evaluate(() => document.body.innerText);
    if (content.length < 100) throw new Error('Dashboard parece vazio');
    await page.screenshot({ path: 'screenshots/dashboard_full.png' });
  }, page);

  await test('7.2 — Gráficos renderizam (sem área em branco)', async (page) => {
    const charts = await page.$$('canvas, svg[class*="chart"], [class*="Chart"]');
    console.log(`  Gráficos encontrados: ${charts.length}`);
    if (charts.length === 0) throw new Error('Nenhum gráfico encontrado no dashboard');
  }, page);

  // ──────────────────────────────────────────────
  // BLOCO 8: CONFIGURAÇÕES E PERFIL
  // ──────────────────────────────────────────────

  await test('8.1 — Navegar para Configurações', async (page) => {
    const link = await page.$('a[href*="setting"], a[href*="profile"], nav >> text=Config, [aria-label*="config"]');
    if (link) await link.click();
    else await page.goto(`${BASE_URL}/settings`);
    await waitForNoLoading(page);
    await page.screenshot({ path: 'screenshots/settings.png' });
  }, page);

  await test('8.2 — Logout funciona', async (page) => {
    const logoutBtn = await page.$('button >> text=Sair, button >> text=Logout, [data-testid="logout"]');
    if (logoutBtn) {
      await logoutBtn.click();
      await page.waitForTimeout(2000);
      const url = page.url();
      if (!url.includes('login') && !url.includes('onboarding')) {
        throw new Error('Após logout não redirecionou para login');
      }
    } else {
      console.log('  ⚠️  Botão de logout não encontrado — pulando');
    }
  }, page);

  // ──────────────────────────────────────────────
  // BLOCO 9: RESPONSIVIDADE
  // ──────────────────────────────────────────────

  await test('9.1 — Layout mobile (375px)', async (page) => {
    await page.setViewportSize({ width: 375, height: 812 });
    await page.goto(`${BASE_URL}/login`);
    await page.waitForTimeout(1000);
    await page.screenshot({ path: 'screenshots/mobile_login.png' });
    const overflow = await page.evaluate(() => document.body.scrollWidth > window.innerWidth);
    if (overflow) throw new Error('Overflow horizontal no mobile');
  }, page);

  await test('9.2 — Layout tablet (768px)', async (page) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto(`${BASE_URL}/login`);
    await page.waitForTimeout(1000);
    await page.screenshot({ path: 'screenshots/tablet_login.png' });
  }, page);

  await test('9.3 — Layout desktop (1440px)', async (page) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto(`${BASE_URL}/login`);
    await page.waitForTimeout(1000);
    await page.screenshot({ path: 'screenshots/desktop_login.png' });
  }, page);

  // ──────────────────────────────────────────────
  // RELATÓRIO FINAL
  // ──────────────────────────────────────────────

  await browser.close();

  const passRate = ((report.passed.length / report.total) * 100).toFixed(1);
  
  const reportContent = `# QA Report — FinanceOS
Data: ${new Date().toISOString()}

## Resultado: ${report.passed.length}/${report.total} (${passRate}%)

## ✅ Passou (${report.passed.length})
${report.passed.map(t => `- ${t}`).join('\n')}

## ❌ Falhou (${report.failed.length})
${report.failed.map(t => `- ${t.name}\n  Erro: ${t.error}`).join('\n')}

## Erros de Console capturados
${errors.length === 0 ? 'Nenhum' : errors.map(e => `- ${e}`).join('\n')}

## Screenshots
Salvos em: ./screenshots/
`;

  fs.writeFileSync('QA_REPORT.md', reportContent);
  console.log('\n' + reportContent);
  console.log('📸 Screenshots em ./screenshots/');
  console.log('📄 Relatório em QA_REPORT.md');
})();
```

---

## FASE 4 — Após rodar os testes, corrija TODOS os bugs

Para cada item ❌ no relatório:

1. Identifique o arquivo afetado (Flutter `.dart` ou Go `.go`)
2. Leia o código atual
3. Corrija o bug
4. Rebuild apenas o necessário:
   - Flutter: `cd apps/web && flutter build web && docker restart financeos-nginx`
   - Go: `cd apps/api && go build ./... && docker restart financeos-api`
5. Rode o teste específico novamente para confirmar

## FASE 5 — Investigar loading infinito especificamente

Se o loading infinito persistir após os testes, execute:

```bash
# Ver logs da API em tempo real enquanto faz login
docker logs financeos-api -f &

# Testar endpoint de profile/me diretamente
curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"qa@financeos.com","password":"QA@Test123456"}' | jq .

# Com o token retornado, testar endpoints que o dashboard chama
TOKEN="SEU_TOKEN_AQUI"
curl -s http://localhost:8000/api/v1/dashboard/overview -H "Authorization: Bearer $TOKEN" | jq .
curl -s http://localhost:8000/api/v1/accounts -H "Authorization: Bearer $TOKEN" | jq .
curl -s http://localhost:8000/api/v1/transactions -H "Authorization: Bearer $TOKEN" | jq .
```

Se algum endpoint retornar 500 ou timeout, corrija o handler Go correspondente.

## FASE 6 — Loop até 100%

Continue corrigindo e retestando até o relatório mostrar 100% de pass rate.
Só encerre quando QA_REPORT.md mostrar todos os testes ✅.