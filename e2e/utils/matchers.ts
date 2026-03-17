import { expect, Locator, Page } from '@playwright/test';

/**
 * Custom Playwright matchers for Discoteca E2E tests
 * Extended automatically via fixtures
 */

// Use Playwright's Response type
type PlaywrightResponse = import('@playwright/test').Response;

/**
 * Assert media card count
 */
export async function toHaveMediaCount(
  locator: Locator,
  expected: number
): Promise<{ pass: boolean; message: () => string }> {
  const actual = await locator.count();
  return {
    pass: actual === expected,
    message: () => `Expected media count to be ${expected}, got ${actual}`
  };
}

/**
 * Assert page is in specific mode (URL hash)
 */
export async function toBeInMode(
  page: Page,
  expectedMode: string
): Promise<{ pass: boolean; message: () => string }> {
  const url = page.url();
  const currentMode = url.split('#')[1] || '';
  return {
    pass: currentMode === expectedMode,
    message: () => `Expected mode "${expectedMode}", got "${currentMode}" (URL: ${url})`
  };
}

/**
 * Assert JSON response has expected keys
 */
export async function toHaveJsonOutput(
  response: Response,
  expectedKeys: string[]
): Promise<{ pass: boolean; message: () => string }> {
  try {
    const json = await response.json();
    const missingKeys = expectedKeys.filter(key => !(key in json));
    return {
      pass: missingKeys.length === 0,
      message: () => missingKeys.length
        ? `Missing keys: ${missingKeys.join(', ')}`
        : 'JSON has expected keys'
    };
  } catch (e) {
    return {
      pass: false,
      message: () => `Invalid JSON: ${e}`
    };
  }
}

/**
 * Assert media card has progress indicator
 */
export async function toHaveProgress(
  locator: Locator,
  expectedProgress: number,
  tolerance: number = 5
): Promise<{ pass: boolean; message: () => string }> {
  const progressBar = locator.locator('.progress-bar');
  if (!await progressBar.isVisible()) {
    return { pass: false, message: () => 'Progress bar not found' };
  }
  const width = await progressBar.evaluate((el: HTMLElement) => {
    const style = window.getComputedStyle(el);
    return parseFloat(style.width) || 0;
  });
  const pass = Math.abs(width - expectedProgress) <= tolerance;
  return {
    pass,
    message: () => `Expected ~${expectedProgress}%, got ${width}%`
  };
}

/**
 * Assert media is playing
 */
export async function toBePlaying(locator: Locator): Promise<{ pass: boolean; message: () => string }> {
  const isPlaying = await locator.evaluate((el: HTMLMediaElement) => !el.paused);
  return {
    pass: isPlaying,
    message: () => isPlaying ? 'Media is playing' : 'Media is not playing'
  };
}

/**
 * Assert media is paused
 */
export async function toBePaused(locator: Locator): Promise<{ pass: boolean; message: () => string }> {
  const isPaused = await locator.evaluate((el: HTMLMediaElement) => el.paused);
  return {
    pass: isPaused,
    message: () => isPaused ? 'Media is paused' : 'Media is not paused'
  };
}

/**
 * Assert element has data attribute
 */
export async function toHaveDataAttribute(
  locator: Locator,
  attr: string,
  value: string
): Promise<{ pass: boolean; message: () => string }> {
  const actualValue = await locator.getAttribute(attr);
  return {
    pass: actualValue === value,
    message: () => `Expected ${attr}="${value}", got "${actualValue}"`
  };
}

/**
 * Assert toast notification appears with text
 */
export async function toHaveToast(
  page: Page,
  expectedText: string,
  timeout: number = 5000
): Promise<{ pass: boolean; message: () => string }> {
  try {
    const toast = page.locator('#toast');
    await toast.waitFor({ state: 'visible', timeout });
    const toastText = await toast.textContent();
    const pass = toastText?.includes(expectedText) ?? false;
    return {
      pass,
      message: () => pass
        ? 'Toast found'
        : `Expected toast "${expectedText}", got "${toastText}"`
    };
  } catch (e) {
    return {
      pass: false,
      message: () => `Toast did not appear within ${timeout}ms`
    };
  }
}

/**
 * Assert no error toast appears
 */
export async function toHaveNoErrorToast(
  page: Page,
  timeout: number = 2000
): Promise<{ pass: boolean; message: () => string }> {
  const toast = page.locator('#toast');
  if (await toast.isVisible()) {
    const toastText = await toast.textContent();
    const isError = toastText?.includes('Error') || toastText?.includes('Failed') || toastText?.includes('⚠️');
    if (isError) {
      return { pass: false, message: () => `Error toast: "${toastText}"` };
    }
  }
  return { pass: true, message: () => 'No error toast' };
}

// Extend Playwright expect with custom matchers
export const customExpect = expect.extend({
  toHaveMediaCount,
  toBeInMode,
  toHaveJsonOutput,
  toHaveProgress,
  toBePlaying,
  toBePaused,
  toHaveDataAttribute,
  toHaveToast,
  toHaveNoErrorToast,
});

// Re-export expect
export { expect };

/**
 * Wait for API response matching URL pattern
 */
export async function waitForApiResponse(
  page: Page,
  urlPattern: string,
  timeout: number = 5000
): Promise<PlaywrightResponse> {
  const [response] = await Promise.all([
    page.waitForResponse(resp => resp.url().includes(urlPattern), { timeout }),
  ]);
  return response;
}

/**
 * Wait for API request matching URL pattern
 */
export async function waitForApiRequest(
  page: Page,
  urlPattern: string,
  timeout: number = 5000
): Promise<void> {
  await page.waitForRequest(req => req.url().includes(urlPattern), { timeout });
}
