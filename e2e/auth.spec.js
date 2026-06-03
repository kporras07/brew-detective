const { test, expect } = require('@playwright/test');
const { mockAPI, loginAs, createFakeJWT } = require('./helpers');

test.describe('Authentication UI', () => {
  test.beforeEach(async ({ page }) => {
    await mockAPI(page);
  });

  test('shows login button when not authenticated', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('#loginBtn')).toBeVisible();
    await expect(page.locator('#userDropdown')).toHaveClass(/hidden/);
  });

  test('shows user dropdown when authenticated', async ({ page }) => {
    await page.goto('/');
    await loginAs(page);
    await page.reload();

    await expect(page.locator('#loginBtn')).toBeHidden();
    await expect(page.locator('#userDropdown')).not.toHaveClass(/hidden/);
    await expect(page.locator('#userName')).toHaveText('Test User');
  });

  test('submit menu item hidden when not authenticated', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('#submitMenuItem')).toBeHidden();
  });

  test('submit menu item visible when authenticated', async ({ page }) => {
    await page.goto('/');
    await loginAs(page);
    await page.reload();

    await expect(page.locator('#submitMenuItem')).toBeVisible();
  });

  test('admin menu hidden for regular users', async ({ page }) => {
    await page.goto('/');
    await loginAs(page, { type: 'regular' });
    await page.reload();

    await expect(page.locator('#adminMenuItem')).toBeHidden();
  });

  test('admin menu visible for admin users', async ({ page }) => {
    // Mock profile to return admin type before any navigation
    await page.route('https://api.brewdetective.coffee/api/v1/profile', route =>
      route.fulfill({ json: { id: 'user1', name: 'Admin', email: 'admin@test.com', type: 'admin', picture: '' } })
    );

    await page.goto('/');
    await loginAs(page, { type: 'admin' });
    await page.reload();

    // The admin menu item is inside the user dropdown - open it first
    await page.click('#userMenuToggle');
    await expect(page.locator('#adminMenuItem')).toBeVisible();
  });

  test('handles token from URL hash on callback', async ({ page }) => {
    const token = createFakeJWT();
    await page.goto(`/#token=${token}`);

    // Should store token and show authenticated state
    await page.waitForFunction(() => localStorage.getItem('auth_token') !== null);
    const storedToken = await page.evaluate(() => localStorage.getItem('auth_token'));
    expect(storedToken).toBe(token);
  });

  test('user menu dropdown toggles', async ({ page }) => {
    await page.goto('/');
    await loginAs(page);
    await page.reload();

    const userMenu = page.locator('#userMenu');
    await expect(userMenu).not.toHaveClass(/active/);

    await page.click('#userMenuToggle');
    await expect(userMenu).toHaveClass(/active/);

    // Click outside closes menu
    await page.click('body', { position: { x: 10, y: 10 } });
    await expect(userMenu).not.toHaveClass(/active/);
  });

  test('expired token shows login button', async ({ page }) => {
    const expiredToken = createFakeJWT({ exp: Math.floor(Date.now() / 1000) - 3600 });

    await page.goto('/');
    await page.evaluate((token) => {
      localStorage.setItem('auth_token', token);
      localStorage.setItem('user_data', JSON.stringify({ id: 'user1', name: 'Test', email: 'test@test.com' }));
    }, expiredToken);
    await page.reload();

    await expect(page.locator('#loginBtn')).toBeVisible();
    await expect(page.locator('#userDropdown')).toHaveClass(/hidden/);
  });
});
