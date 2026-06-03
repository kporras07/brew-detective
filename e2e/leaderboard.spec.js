const { test, expect } = require('@playwright/test');
const { mockAPI, loginAs } = require('./helpers');

test.describe('Leaderboard', () => {
  test.beforeEach(async ({ page }) => {
    await mockAPI(page);
  });

  test('displays current case leaderboard entries', async ({ page }) => {
    await page.goto('/leaderboard');

    const currentBoard = page.locator('#currentCaseLeaderboard');
    await expect(currentBoard.locator('.leaderboard-entry')).toHaveCount(2);
    await expect(currentBoard.locator('.leaderboard-entry').first()).toContainText('Test User');
    await expect(currentBoard.locator('.leaderboard-entry').first()).toContainText('80 pts');
  });

  test('displays global leaderboard entries', async ({ page }) => {
    await page.goto('/leaderboard');

    const globalBoard = page.locator('#globalLeaderboard');
    await expect(globalBoard.locator('.leaderboard-entry')).toHaveCount(3);
    await expect(globalBoard.locator('.leaderboard-entry').first()).toContainText('Test User');
    await expect(globalBoard.locator('.leaderboard-entry').first()).toContainText('150 pts');
  });

  test('shows case info for current leaderboard', async ({ page }) => {
    await page.goto('/leaderboard');

    await expect(page.locator('#currentCaseInfo')).toContainText('Caso de Prueba');
    await expect(page.locator('#currentCaseInfo')).toContainText('2 detectives');
  });

  test('shows empty state when no leaderboard data', async ({ page }) => {
    await page.route('https://api.brewdetective.coffee/api/v1/leaderboard', route =>
      route.fulfill({ json: { leaderboard: [] } })
    );
    await page.route('https://api.brewdetective.coffee/api/v1/leaderboard/current', route =>
      route.fulfill({ json: { leaderboard: [] } })
    );

    await page.goto('/leaderboard');

    await expect(page.locator('#globalLeaderboard')).toContainText('No hay detectives');
    await expect(page.locator('#currentCaseLeaderboard')).toContainText('No hay detectives');
  });

  test('shows error state on API failure', async ({ page }) => {
    await page.route('https://api.brewdetective.coffee/api/v1/leaderboard', route =>
      route.fulfill({ status: 500, json: { error: 'Internal Server Error' } })
    );

    await page.goto('/leaderboard');

    await expect(page.locator('#globalLeaderboard')).toContainText('Error');
  });
});

test.describe('Submit Form', () => {
  test.beforeEach(async ({ page }) => {
    await mockAPI(page);
  });

  test('redirects unauthenticated users away from submit', async ({ page }) => {
    await page.goto('/submit');

    // Should show notification and redirect to home
    await expect(page.locator('#home')).toHaveClass(/active/, { timeout: 5000 });
  });

  test('loads active case info for authenticated users', async ({ page }) => {
    await page.goto('/');
    await loginAs(page);
    await page.reload();

    await page.click('a[href="/submit"]');

    await expect(page.locator('#activeCaseName')).toHaveText('Caso Misterio #1', { timeout: 5000 });
    await expect(page.locator('#activeCaseDescription')).toHaveText('Descubre los origenes de estos 4 cafes');
  });

  test('populates catalog dropdowns from API', async ({ page }) => {
    await page.goto('/');
    await loginAs(page);
    await page.reload();

    await page.click('a[href="/submit"]');

    // Wait for dropdowns to be populated
    await page.waitForFunction(() => {
      const select = document.getElementById('coffee1_region');
      return select && select.options.length > 1;
    }, { timeout: 5000 });

    const regionOptions = await page.locator('#coffee1_region option').allTextContents();
    expect(regionOptions).toContain('Tarrazú');
    expect(regionOptions).toContain('Valle Occidental');
  });

  test('successful submission redirects to thank you page', async ({ page }) => {
    await page.goto('/');
    await loginAs(page);
    await page.reload();

    await page.click('a[href="/submit"]');

    // Wait for form to load
    await page.waitForFunction(() => {
      const select = document.getElementById('coffee1_region');
      return select && select.options.length > 1;
    }, { timeout: 5000 });

    // Fill out the form
    await page.fill('#orderId', 'TEST-001');

    for (let i = 1; i <= 4; i++) {
      await page.selectOption(`#coffee${i}_region`, 'tarrazu');
      await page.selectOption(`#coffee${i}_variety`, 'caturra');
      await page.selectOption(`#coffee${i}_process`, 'washed');
      await page.fill(`#coffee${i}_note1`, 'chocolate');
      await page.fill(`#coffee${i}_note2`, 'frutal');
    }

    await page.fill('#favorite_coffee', 'Cafe 1');
    await page.selectOption('#brewing_method', 'v60');

    await page.click('#submitForm button[type="submit"]');

    // Should redirect to thank you page
    await expect(page.locator('#thankyou')).toHaveClass(/active/, { timeout: 5000 });
  });
});
