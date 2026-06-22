// Shared test helpers for Playwright E2E tests

const API_BASE = 'https://api.brewdetective.coffee';

/**
 * Mock all API endpoints with default responses.
 * Individual tests can override specific routes after calling this.
 */
async function mockAPI(page) {
  // Auth - Google login
  await page.route(`${API_BASE}/auth/google`, route =>
    route.fulfill({ json: { auth_url: 'https://accounts.google.com/o/oauth2/v2/auth?fake=1' } })
  );

  // Profile
  await page.route(`${API_BASE}/api/v1/profile`, route =>
    route.fulfill({ json: { id: 'user1', name: 'Test User', email: 'test@example.com', type: 'regular', picture: '', points: 150, cases_count: 3, accuracy: 0.75 } })
  );

  // Leaderboard (global)
  await page.route(`${API_BASE}/api/v1/leaderboard`, route =>
    route.fulfill({
      json: {
        leaderboard: [
          { user_id: 'user1', detective_name: 'Test User', points: 150, cases_count: 3, accuracy: 0.75 },
          { user_id: 'user2', detective_name: 'Detective Dos', points: 120, cases_count: 2, accuracy: 0.60 },
          { user_id: 'user3', detective_name: 'Detective Tres', points: 80, cases_count: 1, accuracy: 0.50 },
        ],
      },
    })
  );

  // Leaderboard (current case)
  await page.route(`${API_BASE}/api/v1/leaderboard/current`, route =>
    route.fulfill({
      json: {
        case_name: 'Caso de Prueba',
        leaderboard: [
          { user_id: 'user1', detective_name: 'Test User', points: 80, accuracy: 0.80 },
          { user_id: 'user2', detective_name: 'Detective Dos', points: 60, accuracy: 0.60 },
        ],
      },
    })
  );

  // Active case (public)
  await page.route(`${API_BASE}/api/v1/cases/active/public`, route =>
    route.fulfill({
      json: {
        case: {
          id: 'case1',
          name: 'Caso Misterio #1',
          description: 'Descubre los origenes de estos 4 cafes',
          is_active: true,
          coffee_ids: ['c1', 'c2', 'c3', 'c4'],
          coffees: [
            { id: 'c1' },
            { id: 'c2' },
            { id: 'c3' },
            { id: 'c4' },
          ],
          enabled_questions: {
            region: true, variety: true, process: true,
            taste_note_1: true, taste_note_2: true,
            favorite_coffee: true, brewing_method: true,
          },
        },
      },
    })
  );

  // Catalog
  await page.route(`${API_BASE}/api/v1/catalog`, route =>
    route.fulfill({
      json: {
        catalog: {
          region: [
            { value: 'tarrazu', label: 'Tarrazú' },
            { value: 'west_valley', label: 'Valle Occidental' },
          ],
          variety: [
            { value: 'caturra', label: 'Caturra' },
            { value: 'catuai', label: 'Catuaí' },
          ],
          process: [
            { value: 'washed', label: 'Lavado' },
            { value: 'natural', label: 'Natural' },
          ],
          brewing_method: [
            { value: 'v60', label: 'V60' },
            { value: 'chemex', label: 'Chemex' },
          ],
        },
      },
    })
  );

  // Submissions
  await page.route(`${API_BASE}/api/v1/submissions*`, route => {
    if (route.request().method() === 'POST') {
      return route.fulfill({ json: {
        score: 280,
        accuracy: 0.70,
        status: 'completed',
        coffee_results: [
          {
            coffee_id: 'c1', coffee_name: 'Café Tarrazú',
            results: {
              region: { answer: 'Tarrazú', correct: 'Tarrazú', is_correct: true },
              variety: { answer: 'Caturra', correct: 'Catuaí', is_correct: false },
              process: { answer: 'Lavado', correct: 'Lavado', is_correct: true },
              taste_note_1: { answer: 'Chocolate', correct: 'Chocolate, Caramelo', is_correct: true },
              taste_note_2: { answer: 'Frutal', correct: 'Chocolate, Caramelo', is_correct: false },
            },
          },
          {
            coffee_id: 'c2', coffee_name: 'Café Valle Occidental',
            results: {
              region: { answer: 'Valle Occidental', correct: 'Valle Central', is_correct: false },
              variety: { answer: 'Catuaí', correct: 'Catuaí', is_correct: true },
              process: { answer: 'Natural', correct: 'Natural', is_correct: true },
              taste_note_1: { answer: 'Miel', correct: 'Nuez, Miel', is_correct: true },
              taste_note_2: { answer: 'Cítrico', correct: 'Nuez, Miel', is_correct: false },
            },
          },
          {
            coffee_id: 'c3', coffee_name: 'Café Brunca',
            results: {
              region: { answer: 'Tarrazú', correct: 'Brunca', is_correct: false },
              variety: { answer: 'Caturra', correct: 'Caturra', is_correct: true },
              process: { answer: 'Lavado', correct: 'Honey', is_correct: false },
              taste_note_1: { answer: 'Frutal', correct: 'Frutal, Floral', is_correct: true },
              taste_note_2: { answer: 'Chocolate', correct: 'Frutal, Floral', is_correct: false },
            },
          },
          {
            coffee_id: 'c4', coffee_name: 'Café Orosi',
            results: {
              region: { answer: 'Tarrazú', correct: 'Orosi', is_correct: false },
              variety: { answer: 'Catuaí', correct: 'Geisha', is_correct: false },
              process: { answer: 'Natural', correct: 'Natural', is_correct: true },
              taste_note_1: { answer: 'Floral', correct: 'Floral, Jazmín', is_correct: true },
              taste_note_2: { answer: 'Miel', correct: 'Floral, Jazmín', is_correct: false },
            },
          },
        ],
      }});
    }
    return route.fulfill({ json: { submissions: [] } });
  });

  // Cases public
  await page.route(`${API_BASE}/api/v1/cases/public`, route =>
    route.fulfill({ json: { cases: [] } })
  );
}

/**
 * Create a fake JWT token with the given payload.
 * This produces a structurally valid JWT (header.payload.signature) that
 * Auth.isAuthenticated() will accept (it only checks expiry).
 */
function createFakeJWT(payload = {}) {
  const header = { alg: 'HS256', typ: 'JWT' };
  const defaultPayload = {
    user_id: 'user1',
    email: 'test@example.com',
    name: 'Test User',
    exp: Math.floor(Date.now() / 1000) + 3600, // 1 hour from now
    iat: Math.floor(Date.now() / 1000),
  };
  const finalPayload = { ...defaultPayload, ...payload };

  const b64 = (obj) => Buffer.from(JSON.stringify(obj)).toString('base64url');
  return `${b64(header)}.${b64(finalPayload)}.fake-signature`;
}

/**
 * Set up an authenticated session by injecting token + user data into localStorage.
 */
async function loginAs(page, user = {}) {
  const defaultUser = { id: 'user1', name: 'Test User', email: 'test@example.com', type: 'regular', picture: '' };
  const finalUser = { ...defaultUser, ...user };
  const token = createFakeJWT({ user_id: finalUser.id, email: finalUser.email, name: finalUser.name });

  await page.evaluate(({ token, user }) => {
    localStorage.setItem('auth_token', token);
    localStorage.setItem('user_data', JSON.stringify(user));
  }, { token, user: finalUser });
}

/**
 * Mock API with per-coffee question overrides and additional questions.
 * Coffee 1: override (region only) + 1 additional question
 * Coffees 2-4: case-level defaults, no additional questions
 */
async function mockAPIWithPerCoffeeQuestions(page) {
  await mockAPI(page);

  // Override active case with per-coffee data
  await page.route(`${API_BASE}/api/v1/cases/active/public`, route =>
    route.fulfill({
      json: {
        case: {
          id: 'case1',
          name: 'Caso Misterio #1',
          description: 'Descubre los origenes de estos 4 cafes',
          is_active: true,
          coffee_ids: ['c1', 'c2', 'c3', 'c4'],
          coffees: [
            {
              id: 'c1',
              enabled_questions: {
                region: true, variety: false, process: false,
                taste_note_1: false, taste_note_2: false,
              },
              additional_questions: [
                { id: 'aq1', question: 'Altitud del cultivo?', options: ['Alta', 'Media', 'Baja'], points: 15 },
              ],
            },
            { id: 'c2' },
            { id: 'c3' },
            { id: 'c4' },
          ],
          enabled_questions: {
            region: true, variety: true, process: true,
            taste_note_1: true, taste_note_2: true,
            favorite_coffee: true, brewing_method: true,
          },
        },
      },
    }),
    { times: 1 } // override the earlier route once
  );
}

module.exports = { API_BASE, mockAPI, mockAPIWithPerCoffeeQuestions, createFakeJWT, loginAs };
