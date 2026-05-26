import { test, expect } from '@playwright/test';

test.describe('NektoKZ Realtime Matchmaking & Chat E2E', () => {
  test('two independent users can match and chat with each other successfully', async ({ browser }) => {
    // 1. Create two isolated browser contexts and inject a unique device ID to bypass fingerprint collision in headless envs
    const devIdA = 'device_user_a_' + Math.random().toString(36).substring(2) + Date.now().toString(36);
    const contextA = await browser.newContext();
    await contextA.addInitScript((id) => {
      localStorage.setItem('device_id', id);
    }, devIdA);

    const devIdB = 'device_user_b_' + Math.random().toString(36).substring(2) + Date.now().toString(36);
    const contextB = await browser.newContext();
    await contextB.addInitScript((id) => {
      localStorage.setItem('device_id', id);
    }, devIdB);

    const pageA = await contextA.newPage();
    const pageB = await contextB.newPage();

    pageA.on('console', msg => console.log('BROWSER A:', msg.text()));
    pageB.on('console', msg => console.log('BROWSER B:', msg.text()));

    console.log('User A navigating to search settings...');
    await pageA.goto('http://localhost:3000/search');
    // Select Gender: Male, Target Gender: Any, Mode: Text
    await pageA.getByRole('button', { name: 'Парень' }).first().click();
    await pageA.getByRole('button', { name: 'Любой' }).click();
    await pageA.getByRole('button', { name: 'Текст' }).click();

    console.log('User B navigating to search settings...');
    await pageB.goto('http://localhost:3000/search');
    // Select Gender: Female, Target Gender: Any, Mode: Text
    await pageB.getByRole('button', { name: 'Девушка' }).first().click();
    await pageB.getByRole('button', { name: 'Любой' }).click();
    await pageB.getByRole('button', { name: 'Текст' }).click();

    console.log('Starting matchmaking search for User A and User B...');
    await pageA.getByRole('button', { name: 'Найти собеседника' }).click();
    await pageB.getByRole('button', { name: 'Найти собеседника' }).click();

    // 2. Wait for match to succeed (chat input gets visible on both pages)
    console.log('Waiting for matchmaking to complete...');
    await expect(pageA.getByPlaceholder('Напишите сообщение...')).toBeVisible({ timeout: 20000 });
    await expect(pageB.getByPlaceholder('Напишите сообщение...')).toBeVisible({ timeout: 20000 });
    console.log('Successfully matched both users!');

    // 3. Test text messaging
    console.log('User A sending message: "Привет, как дела?"');
    await pageA.getByPlaceholder('Напишите сообщение...').fill('Привет, как дела?');
    await pageA.keyboard.press('Enter');

    console.log('Checking if User B received the message...');
    await expect(pageB.locator('text=Привет, как дела?')).toBeVisible({ timeout: 10000 });
    console.log('User B successfully received the message!');

    console.log('User B replying: "Привет! Всё отлично, тестируем!"');
    await pageB.getByPlaceholder('Напишите сообщение...').fill('Привет! Всё отлично, тестируем!');
    await pageB.keyboard.press('Enter');

    console.log('Checking if User A received the reply...');
    await expect(pageA.locator('text=Привет! Всё отлично, тестируем!')).toBeVisible({ timeout: 10000 });
    console.log('User A successfully received the reply!');

    // 4. Test "Next" (Следующий) functionality
    console.log('User A clicks "Следующий" (Next)...');
    await pageA.getByRole('button', { name: 'Следующий' }).click();

    console.log('Verifying both User A and User B are redirected back and automatically start searching again...');
    await expect(pageA).toHaveURL('http://localhost:3000/search', { timeout: 10000 });
    await expect(pageB).toHaveURL('http://localhost:3000/search', { timeout: 10000 });

    // Assert that both enter the matchmaking search loop immediately
    await expect(pageA.getByRole('heading', { name: 'Поиск собеседника' })).toBeVisible({ timeout: 10000 });
    await expect(pageB.getByRole('heading', { name: 'Поиск собеседника' })).toBeVisible({ timeout: 10000 });
    console.log('Both users successfully and cleanly returned to search page and auto-started searching!');

    // Cleanup
    await contextA.close();
    await contextB.close();
  });
});
