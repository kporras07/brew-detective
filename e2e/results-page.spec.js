const { test, expect } = require('@playwright/test');
const { mockAPI, loginAs } = require('./helpers');

async function submitAndNavigateToResults(page) {
  await mockAPI(page);
  await page.goto('/');
  await loginAs(page);
  await page.goto('/submit');
  await page.waitForSelector('#activeCaseName:not(:has-text("Cargando"))');

  await page.fill('#orderId', 'ABC123');
  for (let i = 1; i <= 4; i++) {
    await page.selectOption(`#coffee${i}_region`, { index: 1 });
    await page.selectOption(`#coffee${i}_variety`, { index: 1 });
    await page.selectOption(`#coffee${i}_process`, { index: 1 });
  }

  await page.click('button[type="submit"]');
  await page.waitForSelector('#coffeeResultsContainer .coffee-result-card', { timeout: 10000 });
}

test.describe('Results page - per-coffee breakdown', () => {
  test('displays result cards for each coffee after submission', async ({ page }) => {
    await submitAndNavigateToResults(page);

    const cards = page.locator('.coffee-result-card');
    await expect(cards).toHaveCount(4);

    await expect(cards.nth(0).locator('.coffee-result-card__title')).toContainText('Café 1');
    await expect(cards.nth(1).locator('.coffee-result-card__title')).toContainText('Café 2');
  });

  test('shows coffee names in card titles', async ({ page }) => {
    await submitAndNavigateToResults(page);

    const firstTitle = page.locator('.coffee-result-card__title').first();
    await expect(firstTitle).toContainText('Café Tarrazú');
  });

  test('shows correct and incorrect indicators', async ({ page }) => {
    await submitAndNavigateToResults(page);

    const firstCard = page.locator('.coffee-result-card').first();
    const correctIndicators = firstCard.locator('.correct-indicator');
    const incorrectIndicators = firstCard.locator('.incorrect-indicator');

    // First coffee has 3 correct (region, process, taste_note_1) and 2 incorrect (variety, taste_note_2)
    await expect(correctIndicators).toHaveCount(3);
    await expect(incorrectIndicators).toHaveCount(2);
  });

  test('correct answers are hidden by default', async ({ page }) => {
    await submitAndNavigateToResults(page);

    const correctAnswers = page.locator('.correct-answer');
    const firstCorrectAnswer = correctAnswers.first();
    await expect(firstCorrectAnswer).not.toBeVisible();
  });

  test('reveal button shows and hides correct answers', async ({ page }) => {
    await submitAndNavigateToResults(page);

    const revealBtn = page.locator('#revealAnswersBtn');
    await expect(revealBtn).toBeVisible();
    await expect(revealBtn).toContainText('Revelar Respuestas Correctas');

    await revealBtn.click();
    const firstCorrectAnswer = page.locator('.correct-answer').first();
    await expect(firstCorrectAnswer).toBeVisible();
    await expect(revealBtn).toContainText('Ocultar Respuestas Correctas');

    await revealBtn.click();
    await expect(firstCorrectAnswer).not.toBeVisible();
    await expect(revealBtn).toContainText('Revelar Respuestas Correctas');
  });

  test('displays user answers in question rows', async ({ page }) => {
    await submitAndNavigateToResults(page);

    const firstCard = page.locator('.coffee-result-card').first();
    const answers = firstCard.locator('.question-answer');

    // First coffee answers: Tarrazú, Caturra, Lavado, Chocolate, Frutal
    await expect(answers.nth(0)).toContainText('Tarrazú');
    await expect(answers.nth(1)).toContainText('Caturra');
    await expect(answers.nth(2)).toContainText('Lavado');
  });

  test('displays score and accuracy alongside results', async ({ page }) => {
    await submitAndNavigateToResults(page);

    const finalScore = page.locator('#finalScore');
    const finalAccuracy = page.locator('#finalAccuracy');

    await expect(finalScore).toContainText('280');
    await expect(finalAccuracy).toContainText('70%');
  });
});
