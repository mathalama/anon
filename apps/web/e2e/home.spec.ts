import { test, expect } from '@playwright/test';

test.describe('NektoKZ Matchmaking Flow', () => {
  test('has title and can see search form', async ({ page }) => {
    // Note: This relies on the local dev server running and API being reachable
    // Mocking API requests or waiting for network idle is recommended for stable tests
    await page.goto('/search');

    // Expect a title "to contain" a substring.
    await expect(page).toHaveTitle(/NektoKZ/i);

    // Expect the header text to be visible
    const heading = page.locator('h1', { hasText: 'Настройки поиска' });
    await expect(heading).toBeVisible();

    // The start button should be disabled initially
    const startBtn = page.getByRole('button', { name: 'Заполни все поля' });
    await expect(startBtn).toBeVisible();
    await expect(startBtn).toBeDisabled();

    // Select settings
    await page.getByRole('button', { name: 'Парень' }).first().click(); // My gender
    await page.getByRole('button', { name: 'Любой' }).click(); // Partner gender
    await page.getByRole('button', { name: 'Текст' }).click(); // Mode

    // Now the button should be active
    const activeStartBtn = page.getByRole('button', { name: 'Найти собеседника' });
    await expect(activeStartBtn).toBeVisible();
    await expect(activeStartBtn).toBeEnabled();
  });
});
