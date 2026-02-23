import { describe, it, expect, beforeEach } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Utility Functions', () => {
  beforeEach(async () => {
    await setupTestEnvironment();
  });

  describe('formatSize', () => {
    it('formats bytes correctly', () => {
      const { formatSize } = window.disco;
      expect(formatSize(0)).toBe('-');
      expect(formatSize(1024)).toBe('1.0 KB');
      expect(formatSize(1024 * 1024)).toBe('1.0 MB');
      expect(formatSize(1024 * 1024 * 1024)).toBe('1.0 GB');
    });
  });

  describe('formatDuration', () => {
    it('formats seconds correctly', () => {
      const { formatDuration } = window.disco;
      expect(formatDuration(0)).toBe('');
      expect(formatDuration(59)).toBe('0:59');
      expect(formatDuration(60)).toBe('1:00');
      expect(formatDuration(3600)).toBe('1:00:00');
      expect(formatDuration(3661)).toBe('1:01:01');
    });
  });

  describe('getIcon', () => {
    it('returns correct icons for types', () => {
      const { getIcon } = window.disco;
      expect(getIcon('video/mp4')).toBe('🎬');
      expect(getIcon('audio/mpeg')).toBe('🎵');
      expect(getIcon('image/jpeg')).toBe('🖼️');
      expect(getIcon('application/pdf')).toBe('📚');
      expect(getIcon('unknown')).toBe('📄');
      expect(getIcon('')).toBe('📄');
    });
  });

  describe('formatRelativeDate', () => {
    it('formats timestamps correctly', () => {
      const { formatRelativeDate } = window.disco;
      const now = Math.floor(Date.now() / 1000);
      expect(formatRelativeDate(0)).toBe('-');
      expect(formatRelativeDate(now - 10)).toBe('just now');
      expect(formatRelativeDate(now - 70)).toBe('1m ago');
      expect(formatRelativeDate(now - 3700)).toBe('1h ago');
      expect(formatRelativeDate(now - 90000)).toBe('1d ago');
    });
  });

  describe('truncateString', () => {
    it('truncates long strings', () => {
      const { truncateString } = window.disco;
      expect(truncateString('Short string')).toBe('Short string');
      const longString = 'A'.repeat(60);
      expect(truncateString(longString)).toBe('A'.repeat(52) + '...');
    });
  });

  describe('formatDisplayPath', () => {
    it('formats paths correctly', () => {
      const { formatDisplayPath } = window.disco;
      expect(formatDisplayPath('/home/user/media/video.mp4')).toBe('media/video.mp4');
      expect(formatDisplayPath('video.mp4')).toBe('video.mp4');
    });
  });
});
