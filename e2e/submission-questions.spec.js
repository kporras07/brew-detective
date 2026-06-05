const { test, expect } = require('@playwright/test');
const { mockAPI, mockAPIWithPerCoffeeQuestions, loginAs } = require('./helpers');

test.describe('Submission form with per-coffee questions', () => {
  test('per-coffee question visibility - override hides non-enabled fields', async ({ page }) => {
    await mockAPIWithPerCoffeeQuestions(page);
    await page.goto('/');
    await loginAs(page);
    await page.goto('/submit');
    await page.waitForSelector('#activeCaseName:not(:has-text("Cargando"))');

    // Coffee 1 has override: region only. Variety/process should be hidden.
    await expect(page.locator('#coffee1_region').locator('..')).toBeVisible();
    await expect(page.locator('#coffee1_variety').locator('..')).toBeHidden();
    await expect(page.locator('#coffee1_process').locator('..')).toBeHidden();

    // Coffee 2 uses case-level defaults (all enabled)
    await expect(page.locator('#coffee2_region').locator('..')).toBeVisible();
    await expect(page.locator('#coffee2_variety').locator('..')).toBeVisible();
    await expect(page.locator('#coffee2_process').locator('..')).toBeVisible();
  });

  test('additional questions render for the correct coffee only', async ({ page }) => {
    await mockAPIWithPerCoffeeQuestions(page);
    await page.goto('/');
    await loginAs(page);
    await page.goto('/submit');
    await page.waitForSelector('#activeCaseName:not(:has-text("Cargando"))');

    // Coffee 1 should have the additional question dropdown
    const aq1 = page.locator('#coffee1_additional_aq1');
    await expect(aq1).toBeVisible();
    const options = aq1.locator('option');
    await expect(options).toHaveCount(4); // default + 3 options

    // Coffee 2 should NOT have additional question fields
    await expect(page.locator('#coffee2_additional_aq1')).toHaveCount(0);
  });

  test('submission includes additional answers in request body', async ({ page }) => {
    await mockAPIWithPerCoffeeQuestions(page);
    await page.goto('/');
    await loginAs(page);
    await page.goto('/submit');
    await page.waitForSelector('#activeCaseName:not(:has-text("Cargando"))');

    await page.fill('#orderId', 'ABC123');
    await page.selectOption('#coffee1_region', { index: 1 });
    await page.selectOption('#coffee1_additional_aq1', 'Alta');

    for (let i = 2; i <= 4; i++) {
      await page.selectOption(`#coffee${i}_region`, { index: 1 });
      await page.selectOption(`#coffee${i}_variety`, { index: 1 });
      await page.selectOption(`#coffee${i}_process`, { index: 1 });
    }

    let submittedBody;
    await page.route('**/api/v1/submissions*', async route => {
      if (route.request().method() === 'POST') {
        submittedBody = route.request().postDataJSON();
        await route.fulfill({ json: { score: 100, accuracy: 0.5 } });
      } else {
        await route.fulfill({ json: { submissions: [] } });
      }
    });

    await page.click('button[type="submit"]');
    await page.waitForFunction(
      () => document.getElementById('submitSuccess')?.style.display === 'block',
      { timeout: 5000 }
    );

    expect(submittedBody).toBeDefined();
    expect(submittedBody.coffee_answers[0].additional_answers).toBeDefined();
    expect(submittedBody.coffee_answers[0].additional_answers).toHaveLength(1);
    expect(submittedBody.coffee_answers[0].additional_answers[0]).toEqual({
      question_id: 'aq1',
      answer: 'Alta',
    });
  });

  test('scoring info shows additional question points', async ({ page }) => {
    await mockAPIWithPerCoffeeQuestions(page);
    await page.goto('/');
    await loginAs(page);
    await page.goto('/submit');
    await page.waitForSelector('#activeCaseName:not(:has-text("Cargando"))');

    const maxScore = page.locator('#maxScoreDisplay');
    const text = await maxScore.textContent();
    // base 100*4=400 + additional 15 + bonus 100 = 515
    expect(text).toContain('515');
  });

  test('backward compat - no overrides renders all questions identically', async ({ page }) => {
    await mockAPI(page);
    await page.goto('/');
    await loginAs(page);
    await page.goto('/submit');
    await page.waitForSelector('#activeCaseName:not(:has-text("Cargando"))');

    for (let i = 1; i <= 4; i++) {
      await expect(page.locator(`#coffee${i}_region`).locator('..')).toBeVisible();
      await expect(page.locator(`#coffee${i}_variety`).locator('..')).toBeVisible();
      await expect(page.locator(`#coffee${i}_process`).locator('..')).toBeVisible();
    }

    const additionalFields = page.locator('.additional-question-group');
    await expect(additionalFields).toHaveCount(0);
  });
});
