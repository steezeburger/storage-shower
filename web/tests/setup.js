// setup.js - Jest setup file for frontend tests

// Mock required browser globals that might be missing in the test environment
global.d3 = {
  select: jest.fn(),
  hierarchy: jest.fn(),
  treemap: jest.fn().mockReturnValue({
    size: jest.fn().mockReturnThis(),
    padding: jest.fn().mockReturnThis(),
  }),
  partition: jest.fn().mockReturnValue({
    size: jest.fn().mockReturnThis(),
  }),
  arc: jest.fn().mockReturnValue(jest.fn()),
};

// Mock browser features like Element.prototype.classList
if (!("classList" in Element.prototype)) {
  Object.defineProperty(Element.prototype, "classList", {
    value: {
      add: jest.fn(),
      remove: jest.fn(),
      toggle: jest.fn(),
      contains: jest.fn(),
    },
  });
}

// Setup custom matchers if needed
expect.extend({
  toHaveBeenCalledWithPath(received, expected) {
    const pass = received.mock.calls.some(
      (call) => call[0] && call[0].includes && call[0].includes(expected)
    );

    return {
      pass,
      message: () => `expected ${received.mock.calls} to include path containing "${expected}"`,
    };
  },
});

// Clean up after tests
afterEach(() => {
  jest.clearAllMocks();
});
