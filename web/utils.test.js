import { describe, it, expect } from 'vitest';
import fs from 'fs';
import path from 'path';

// Read app.js content
const appJsPath = path.resolve(__dirname, 'app.js');
const appJsContent = fs.readFileSync(appJsPath, 'utf8');

// Extract the functions from the bottom of the file using simpler regexes
function extractFunction(name, content) {
    const startIdx = content.indexOf(`function ${name}`);
    if (startIdx === -1) return null;
    
    // Find the end of the function by matching braces
    let braceCount = 0;
    let started = false;
    for (let i = startIdx; i < content.length; i++) {
        if (content[i] === '{') {
            braceCount++;
            started = true;
        } else if (content[i] === '}') {
            braceCount--;
        }
        
        if (started && braceCount === 0) {
            const funcStr = content.substring(startIdx, i + 1);
            return new Function('arg', funcStr + ` return ${name}(arg);`);
        }
    }
    return null;
}

const formatSize = extractFunction('formatSize', appJsContent);
const formatDuration = extractFunction('formatDuration', appJsContent);
const getIcon = extractFunction('getIcon', appJsContent);

describe('Utility Functions', () => {
  describe('formatSize', () => {
    it('formats bytes correctly', () => {
      expect(formatSize(0)).toBe('-');
      expect(formatSize(1024)).toBe('1.0 KB');
      expect(formatSize(1024 * 1024)).toBe('1.0 MB');
      expect(formatSize(1024 * 1024 * 1024)).toBe('1.0 GB');
    });
  });

  describe('formatDuration', () => {
    it('formats seconds correctly', () => {
      expect(formatDuration(0)).toBe('');
      expect(formatDuration(59)).toBe('0:59');
      expect(formatDuration(60)).toBe('1:00');
      expect(formatDuration(3600)).toBe('1:00:00');
      expect(formatDuration(3661)).toBe('1:01:01');
    });
  });

  describe('getIcon', () => {
    it('returns correct icons for types', () => {
      expect(getIcon('video/mp4')).toBe('🎬');
      expect(getIcon('audio/mpeg')).toBe('🎵');
      expect(getIcon('image/jpeg')).toBe('🖼️');
      expect(getIcon('application/pdf')).toBe('📚');
      expect(getIcon('unknown')).toBe('📄');
      expect(getIcon('')).toBe('📄');
    });
  });
});
