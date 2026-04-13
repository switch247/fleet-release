/**
 * Browser UI tests (Playwright page context).
 * Requires FRONTEND_BASE_URL to point to a running frontend (default http://127.0.0.1:5173).
 * Run via: npx playwright test --project=ui
 */
const { test, expect } = require('@playwright/test');

test.use({ baseURL: process.env.FRONTEND_BASE_URL || 'http://127.0.0.1:5173' });

test('login page is reachable and renders a username input', async ({ page }) => {
  await page.goto('/');
  // The app should show a login form or redirect to /login.
  await page.waitForLoadState('networkidle');
  const usernameInput = page.locator('input[type="text"], input[placeholder*="sername" i], input[name="username"]');
  await expect(usernameInput.first()).toBeVisible({ timeout: 10000 });
});

test('invalid login shows an error message', async ({ page }) => {
  await page.goto('/');
  await page.waitForLoadState('networkidle');

  const usernameInput = page.locator('input[type="text"], input[placeholder*="sername" i], input[name="username"]').first();
  const passwordInput = page.locator('input[type="password"]').first();
  await usernameInput.fill('nonexistent');
  await passwordInput.fill('wrongpass');

  const submitBtn = page.locator('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")').first();
  await submitBtn.click();

  // Expect an error message to appear
  const errorMsg = page.locator('[role="alert"], .text-red-400, .text-rose-400, [class*="error"]');
  await expect(errorMsg.first()).toBeVisible({ timeout: 8000 });
});

test('successful login navigates away from login page', async ({ page }) => {
  await page.goto('/');
  await page.waitForLoadState('networkidle');

  const usernameInput = page.locator('input[type="text"], input[placeholder*="sername" i], input[name="username"]').first();
  const passwordInput = page.locator('input[type="password"]').first();

  // Use seeded credentials only available if BOOTSTRAP_SEED was run.
  // These are provided via env for the test environment.
  const user = process.env.TEST_UI_USERNAME || 'customer';
  const pass = process.env.TEST_UI_PASSWORD || 'Customer1234!';

  await usernameInput.fill(user);
  await passwordInput.fill(pass);

  const submitBtn = page.locator('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")').first();
  await submitBtn.click();

  // After login the URL should change away from login
  await page.waitForURL((url) => !url.pathname.includes('login'), { timeout: 10000 });
  expect(page.url()).not.toContain('login');
});

test('authenticated user sees navigation links for their role', async ({ page }) => {
  await page.goto('/');
  await page.waitForLoadState('networkidle');

  const usernameInput = page.locator('input[type="text"], input[placeholder*="sername" i], input[name="username"]').first();
  const passwordInput = page.locator('input[type="password"]').first();
  await usernameInput.fill(process.env.TEST_UI_USERNAME || 'customer');
  await passwordInput.fill(process.env.TEST_UI_PASSWORD || 'Customer1234!');

  const submitBtn = page.locator('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")').first();
  await submitBtn.click();
  await page.waitForURL((url) => !url.pathname.includes('login'), { timeout: 10000 });

  // Customer should see Bookings nav link
  const bookingsNav = page.locator('a[href*="booking" i], nav >> text=Bookings');
  await expect(bookingsNav.first()).toBeVisible({ timeout: 8000 });
});

test('bookings page renders the new booking button', async ({ page }) => {
  await page.goto('/');
  await page.waitForLoadState('networkidle');

  const usernameInput = page.locator('input[type="text"], input[placeholder*="sername" i], input[name="username"]').first();
  const passwordInput = page.locator('input[type="password"]').first();
  await usernameInput.fill(process.env.TEST_UI_USERNAME || 'customer');
  await passwordInput.fill(process.env.TEST_UI_PASSWORD || 'Customer1234!');

  await page.locator('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")').first().click();
  await page.waitForURL((url) => !url.pathname.includes('login'), { timeout: 10000 });

  await page.goto('/bookings');
  await page.waitForLoadState('networkidle');

  const newBookingBtn = page.locator('button:has-text("New Booking")');
  await expect(newBookingBtn).toBeVisible({ timeout: 8000 });
});

test('notifications page or inbox link is reachable for authenticated user', async ({ page }) => {
  await page.goto('/');
  await page.waitForLoadState('networkidle');

  const usernameInput = page.locator('input[type="text"], input[placeholder*="sername" i], input[name="username"]').first();
  const passwordInput = page.locator('input[type="password"]').first();
  await usernameInput.fill(process.env.TEST_UI_USERNAME || 'customer');
  await passwordInput.fill(process.env.TEST_UI_PASSWORD || 'Customer1234!');
  await page.locator('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")').first().click();
  await page.waitForURL((url) => !url.pathname.includes('login'), { timeout: 10000 });

  const notificationsLink = page.locator('a[href*="notification" i], nav >> text=Notifications, nav >> text=Inbox');
  await expect(notificationsLink.first()).toBeVisible({ timeout: 8000 });
});

// ── Role-gated navigation ────────────────────────────────────────────────────

/**
 * Helper: log in and return after redirect.
 */
async function loginAs(page, username, password) {
  await page.goto('/');
  await page.waitForLoadState('networkidle');
  await page.locator('input[type="text"], input[placeholder*="sername" i], input[name="username"]').first().fill(username);
  await page.locator('input[type="password"]').first().fill(password);
  await page.locator('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")').first().click();
  await page.waitForURL((url) => !url.pathname.includes('login'), { timeout: 10000 });
}

test('customer does NOT see admin-only nav links (Admin Users / Admin Catalog)', async ({ page }) => {
  await loginAs(page, process.env.TEST_UI_USERNAME || 'customer', process.env.TEST_UI_PASSWORD || 'Customer1234!');

  // Admin-only links should not be present in the sidebar for a customer.
  await expect(page.locator('a[href="/admin/users"]')).not.toBeVisible();
  await expect(page.locator('a[href="/admin/catalog"]')).not.toBeVisible();
});

test('admin user sees admin-only nav links', async ({ page }) => {
  const adminUser = process.env.TEST_UI_ADMIN_USERNAME || 'admin';
  const adminPass = process.env.TEST_UI_ADMIN_PASSWORD || 'Admin1234!Pass';
  await loginAs(page, adminUser, adminPass);

  await expect(page.locator('a[href="/admin/users"]')).toBeVisible({ timeout: 8000 });
  await expect(page.locator('a[href="/admin/catalog"]')).toBeVisible({ timeout: 8000 });
  await expect(page.locator('a[href="/admin/notifications"]')).toBeVisible({ timeout: 8000 });
});

test('customer does NOT see Inspections nav link', async ({ page }) => {
  await loginAs(page, process.env.TEST_UI_USERNAME || 'customer', process.env.TEST_UI_PASSWORD || 'Customer1234!');

  // Inspections is only shown to provider / csa / admin roles.
  await expect(page.locator('a[href="/inspections"]')).not.toBeVisible();
});

test('complaints page is accessible and shows Open Complaint form for customer', async ({ page }) => {
  await loginAs(page, process.env.TEST_UI_USERNAME || 'customer', process.env.TEST_UI_PASSWORD || 'Customer1234!');

  await page.goto('/complaints');
  await page.waitForLoadState('networkidle');

  // The form card heading
  await expect(page.locator('text=Open Complaint')).toBeVisible({ timeout: 8000 });
  // Submit button
  await expect(page.locator('button:has-text("Submit")')).toBeVisible({ timeout: 8000 });
});

test('unauthenticated user is redirected to login when navigating to a protected route', async ({ page }) => {
  // Navigate directly without logging in
  await page.goto('/bookings');
  await page.waitForLoadState('networkidle');

  // Should end up on the login page (URL contains login or the username input is visible)
  const onLoginPage =
    page.url().includes('login') ||
    (await page.locator('input[type="text"], input[placeholder*="sername" i]').first().isVisible());
  expect(onLoginPage).toBe(true);
});
