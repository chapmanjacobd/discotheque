import { Page, Locator } from '@playwright/test';

/**
 * Page Object Model for media viewer/player
 * Handles playback controls, theatre mode, and viewer navigation
 */
export class ViewerPage {
  readonly page: Page;
  readonly playerContainer: Locator;
  readonly mediaTitle: Locator;
  readonly videoElement: Locator;
  readonly audioElement: Locator;
  readonly closeBtn: Locator;
  readonly theatreBtn: Locator;
  readonly speedBtn: Locator;
  readonly speedMenu: Locator;
  readonly nextBtn: Locator;
  readonly prevBtn: Locator;
  readonly fullscreenBtn: Locator;
  readonly pipBtn: Locator;
  readonly streamTypeBtn: Locator;
  readonly channelSurfBtn: Locator;
  readonly slideshowBtn: Locator;
  readonly queueContainer: Locator;

  constructor(page: Page) {
    this.page = page;
    this.playerContainer = page.locator('#pip-player');
    this.mediaTitle = page.locator('#media-title');
    this.videoElement = page.locator('#pip-player video');
    this.audioElement = page.locator('#pip-player audio');
    this.closeBtn = page.locator('.close-pip, #pip-player .player-close, button:has-text("Close")').first();
    this.theatreBtn = page.locator('#pip-theatre');
    this.speedBtn = page.locator('#pip-speed');
    this.speedMenu = page.locator('#pip-speed-menu');
    this.nextBtn = page.locator('#pip-next, button:has-text("Next")').first();
    this.prevBtn = page.locator('#pip-previous, button:has-text("Previous")').first();
    this.fullscreenBtn = page.locator('#pip-fullscreen, button:has-text("Fullscreen")').first();
    this.pipBtn = page.locator('#pip-native, button:has-text("PiP")').first();
    this.streamTypeBtn = page.locator('#pip-stream-type');
    this.channelSurfBtn = page.locator('#channel-surf-btn');
    this.slideshowBtn = page.locator('#pip-slideshow');
    this.queueContainer = page.locator('#queue-container');
  }

  /**
   * Wait for player to be visible
   */
  async waitForPlayer(timeout: number = 10000): Promise<void> {
    await this.playerContainer.waitFor({ state: 'visible', timeout });
  }

  /**
   * Check if player is open/visible
   */
  async isOpen(): Promise<boolean> {
    return await this.playerContainer.isVisible();
  }

  /**
   * Check if player is in theatre mode
   */
  async isInTheatreMode(): Promise<boolean> {
    const classes = await this.playerContainer.getAttribute('class');
    return classes?.includes('theatre') ?? false;
  }

  /**
   * Toggle theatre mode
   */
  async toggleTheatreMode(): Promise<void> {
    await this.theatreBtn.click();
  }

  /**
   * Enter theatre mode
   */
  async enterTheatreMode(): Promise<void> {
    if (!await this.isInTheatreMode()) {
      await this.toggleTheatreMode();
      await this.playerContainer.waitFor({ state: 'visible' });
    }
  }

  /**
   * Exit theatre mode
   */
  async exitTheatreMode(): Promise<void> {
    if (await this.isInTheatreMode()) {
      await this.toggleTheatreMode();
      await this.page.waitForTimeout(300);
    }
  }

  /**
   * Close the player
   */
  async close(): Promise<void> {
    if (await this.closeBtn.isVisible()) {
      await this.closeBtn.click();
      await this.playerContainer.waitFor({ state: 'hidden' });
    }
  }

  /**
   * Get current playback speed
   */
  async getPlaybackSpeed(): Promise<string> {
    return await this.speedBtn.textContent() || '1x';
  }

  /**
   * Set playback speed
   */
  async setPlaybackSpeed(speed: string): Promise<void> {
    await this.speedBtn.click();
    await this.speedMenu.waitFor({ state: 'visible' });
    const option = this.speedMenu.locator(`button:has-text("${speed}")`);
    await option.click();
    await this.speedMenu.waitFor({ state: 'hidden' });
  }

  /**
   * Play media
   */
  async play(): Promise<void> {
    const media = this.videoElement.or(this.audioElement);
    await media.evaluate((el: HTMLMediaElement) => el.play());
  }

  /**
   * Pause media
   */
  async pause(): Promise<void> {
    const media = this.videoElement.or(this.audioElement);
    await media.evaluate((el: HTMLMediaElement) => el.pause());
  }

  /**
   * Check if media is playing
   */
  async isPlaying(): Promise<boolean> {
    const media = this.videoElement.or(this.audioElement);
    return await media.evaluate((el: HTMLMediaElement) => !el.paused);
  }

  /**
   * Get current playback position in seconds
   */
  async getCurrentTime(): Promise<number> {
    const media = this.videoElement.or(this.audioElement);
    return await media.evaluate((el: HTMLMediaElement) => el.currentTime);
  }

  /**
   * Get media duration in seconds
   */
  async getDuration(): Promise<number> {
    const media = this.videoElement.or(this.audioElement);
    return await media.evaluate((el: HTMLMediaElement) => el.duration);
  }

  /**
   * Seek to specific time
   */
  async seekTo(time: number): Promise<void> {
    const media = this.videoElement.or(this.audioElement);
    await media.evaluate((el: HTMLMediaElement, t: number) => {
      el.currentTime = t;
    }, time);
  }

  /**
   * Go to next media
   */
  async next(): Promise<void> {
    if (await this.nextBtn.isVisible()) {
      const isDisabled = await this.nextBtn.isDisabled();
      if (!isDisabled) {
        await this.nextBtn.click();
      }
    }
  }

  /**
   * Go to previous media
   */
  async previous(): Promise<void> {
    if (await this.prevBtn.isVisible()) {
      const isDisabled = await this.prevBtn.isDisabled();
      if (!isDisabled) {
        await this.prevBtn.click();
      }
    }
  }

  /**
   * Get media title
   */
  async getTitle(): Promise<string> {
    return await this.mediaTitle.textContent() || '';
  }

  /**
   * Check if queue is enabled and visible
   */
  async isQueueVisible(): Promise<boolean> {
    return await this.queueContainer.isVisible();
  }

  /**
   * Get queue item count
   */
  async getQueueCount(): Promise<number> {
    if (!await this.isQueueVisible()) {
      return 0;
    }
    return await this.queueContainer.locator('.queue-item').count();
  }

  /**
   * Check if slideshow mode is active
   */
  async isSlideshowActive(): Promise<boolean> {
    return await this.slideshowBtn.isVisible() && 
           !await this.slideshowBtn.classList().then(classes => classes?.includes('hidden'));
  }

  /**
   * Start/stop slideshow
   */
  async toggleSlideshow(): Promise<void> {
    await this.slideshowBtn.click();
  }

  /**
   * Get current stream type (Direct/HLS)
   */
  async getStreamType(): Promise<string> {
    return await this.streamTypeBtn.textContent() || '';
  }

  /**
   * Wait for media to finish loading
   */
  async waitForMediaLoaded(timeout: number = 10000): Promise<void> {
    const media = this.videoElement.or(this.audioElement);
    await media.waitFor({ state: 'visible', timeout });
    await this.page.waitForFunction(
      () => {
        const video = document.querySelector('#pip-player video') as HTMLVideoElement;
        const audio = document.querySelector('#pip-player audio') as HTMLAudioElement;
        const media = video || audio;
        return media && !media.readyState;
      },
      { timeout }
    );
  }

  /**
   * Take a screenshot of the current frame (for video)
   */
  async captureFrame(): Promise<Buffer> {
    return await this.playerContainer.screenshot();
  }

  /**
   * Check if media has ended
   */
  async hasEnded(): Promise<boolean> {
    const media = this.videoElement.or(this.audioElement);
    return await media.evaluate((el: HTMLMediaElement) => el.ended);
  }

  /**
   * Wait for media to end
   */
  async waitForEnd(timeout: number = 60000): Promise<void> {
    await this.page.waitForFunction(
      () => {
        const video = document.querySelector('#pip-player video') as HTMLVideoElement;
        const audio = document.querySelector('#pip-player audio') as HTMLAudioElement;
        const media = video || audio;
        return media && media.ended;
      },
      { timeout }
    );
  }
}
