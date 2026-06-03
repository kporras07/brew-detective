const { test, expect } = require('@playwright/test');
const { mockAPI } = require('./helpers');

test.describe('Navigation and Routing', () => {
  test.beforeEach(async ({ page }) => {
    await mockAPI(page);
  });

  test('home page loads by default', async ({ page }) => {
    await page.goto('/');
    const homePage = page.locator('#home');
    await expect(homePage).toHaveClass(/active/);
    await expect(page).toHaveTitle(/Brew Detective/);
  });

  test('navigating to order page', async ({ page }) => {
    await page.goto('/');
    await page.click('a[href="/order"]');
    await expect(page.locator('#order')).toHaveClass(/active/);
    await expect(page.locator('#home')).not.toHaveClass(/active/);
    await expect(page).toHaveTitle(/Ordenar Caja/);
  });

  test('navigating to leaderboard page', async ({ page }) => {
    await page.goto('/');
    await page.click('a[href="/leaderboard"]');
    await expect(page.locator('#leaderboard')).toHaveClass(/active/);
    await expect(page).toHaveTitle(/Ranking/);
  });

  test('direct URL navigation works', async ({ page }) => {
    await page.goto('/leaderboard');
    await expect(page.locator('#leaderboard')).toHaveClass(/active/);
  });

  test('browser back button works', async ({ page }) => {
    await page.goto('/');
    await page.click('a[href="/order"]');
    await expect(page.locator('#order')).toHaveClass(/active/);

    await page.goBack();
    await expect(page.locator('#home')).toHaveClass(/active/);
  });

  test('unknown route falls back to home', async ({ page }) => {
    await page.goto('/nonexistent');
    await expect(page.locator('#home')).toHaveClass(/active/);
  });

  test('CTA button navigates to order page', async ({ page }) => {
    await page.goto('/');
    await page.click('.hero .cta-button');
    await expect(page.locator('#order')).toHaveClass(/active/);
  });

  test('mobile menu toggles', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/');

    const mobileNav = page.locator('#mobileNav');
    await expect(mobileNav).not.toHaveClass(/active/);

    await page.click('.mobile-menu-toggle');
    await expect(mobileNav).toHaveClass(/active/);

    // Navigate via mobile menu closes it
    await page.click('#mobileNav a[href="/order"]');
    await expect(mobileNav).not.toHaveClass(/active/);
    await expect(page.locator('#order')).toHaveClass(/active/);
  });
});
