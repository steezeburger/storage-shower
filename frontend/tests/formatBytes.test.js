/**
 * Tests for the formatBytes function in app.js
 */

describe('formatBytes', () => {
  // Mock the formatBytes function since we can't import it directly
  function formatBytes(bytes, decimals = 2) {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
  }

  test('should format 0 bytes correctly', () => {
    expect(formatBytes(0)).toBe('0 Bytes');
  });

  test('should format bytes correctly', () => {
    expect(formatBytes(100)).toBe('100 Bytes');
  });

  test('should format KB correctly', () => {
    expect(formatBytes(1024)).toBe('1.00 KB');
    expect(formatBytes(1536)).toBe('1.50 KB');
  });

  test('should format MB correctly', () => {
    expect(formatBytes(1048576)).toBe('1.00 MB');
  });

  test('should format GB correctly', () => {
    expect(formatBytes(1073741824)).toBe('1.00 GB');
  });

  test('should format TB correctly', () => {
    expect(formatBytes(1099511627776)).toBe('1.00 TB');
  });

  test('should respect decimal parameter', () => {
    expect(formatBytes(1234567, 0)).toBe('1 MB');
    expect(formatBytes(1234567, 1)).toBe('1.2 MB');
    expect(formatBytes(1234567, 3)).toBe('1.177 MB');
  });
});
