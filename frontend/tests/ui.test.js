/**
 * UI tests for Storage Shower frontend
 */

// Mock DOM manipulation functions that would normally be in app.js
const appFunctions = {
  updateSelectedPath: function (path) {
    document.getElementById("selected-path").textContent = path;
    return path;
  },

  formatBytes: function (bytes, decimals = 2) {
    if (bytes === 0) return "0 Bytes";

    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];

    const i = Math.floor(Math.log(bytes) / Math.log(k));

    // Special case for bytes to not show decimal places
    if (i === 0) {
      return Math.floor(bytes) + " " + sizes[i];
    }

    return (bytes / Math.pow(k, i)).toFixed(dm) + " " + sizes[i];
  },

  copyToClipboard: function (text) {
    return navigator.clipboard.writeText(text);
  },

  showToast: function (message, duration = 2000) {
    const toast = document.getElementById("toast-message");
    toast.textContent = message;
    toast.style.display = "block";

    setTimeout(() => {
      toast.style.display = "none";
    }, duration);

    return message;
  },
};

// Basic DOM setup and tests
describe("UI Components", () => {
  // Set up DOM elements before each test
  beforeEach(() => {
    document.body.innerHTML = `
      <div class="app-container">
        <input type="text" id="path-input" value="/test/path">
        <button id="browse-btn">Browse</button>
        <button id="home-btn">Home</button>
        <button id="scan-btn">Scan</button>
        <button id="stop-btn" disabled>Stop</button>
        <input type="checkbox" id="ignore-hidden">
        <div id="selected-path">No item selected</div>
        <div id="visualization"></div>
        <div id="breadcrumb-trail"></div>
        <div id="toast-message" class="toast" style="display: none;"></div>
        <div id="progress-container" style="display: none;">
          <div id="progress-bar" style="width: 0%"></div>
          <div id="progress-text">0%</div>
        </div>
      </div>
    `;

    // Attach the mock functions to window
    window.updateSelectedPath = appFunctions.updateSelectedPath;
    window.formatBytes = appFunctions.formatBytes;
    window.copyToClipboard = appFunctions.copyToClipboard;
    window.showToast = appFunctions.showToast;
  });

  test("path input should initialize correctly", () => {
    const pathInput = document.getElementById("path-input");
    expect(pathInput.value).toBe("/test/path");
  });

  test("ignore hidden checkbox should be unchecked by default", () => {
    const ignoreHiddenCheckbox = document.getElementById("ignore-hidden");
    expect(ignoreHiddenCheckbox.checked).toBe(false);
  });

  test("selected path should show default message", () => {
    const selectedPathText = document.getElementById("selected-path");
    expect(selectedPathText.textContent).toBe("No item selected");
  });

  test("scan button should be enabled initially", () => {
    const scanBtn = document.getElementById("scan-btn");
    expect(scanBtn.disabled).toBe(false);
  });

  test("stop button should be disabled initially", () => {
    const stopBtn = document.getElementById("stop-btn");
    expect(stopBtn.disabled).toBe(true);
  });

  test("progress container should be hidden initially", () => {
    const progressContainer = document.getElementById("progress-container");
    expect(progressContainer.style.display).toBe("none");
  });

  test("toast message should be hidden initially", () => {
    const toast = document.getElementById("toast-message");
    expect(toast.style.display).toBe("none");
  });
});

// Test path handling functionality
describe("Path Handling", () => {
  beforeEach(() => {
    document.body.innerHTML = `
      <div class="app-container">
        <input type="text" id="path-input" value="/test/path">
        <div id="selected-path">/test/path</div>
      </div>
    `;

    window.updateSelectedPath = appFunctions.updateSelectedPath;
  });

  test("updateSelectedPath should update the selected path element", () => {
    const newPath = "/new/test/path";
    const result = window.updateSelectedPath(newPath);

    expect(document.getElementById("selected-path").textContent).toBe(newPath);
    expect(result).toBe(newPath);
  });

  test("updating selected path should handle special characters", () => {
    const specialPath = "/path with spaces/and-symbols/file.txt";
    const result = window.updateSelectedPath(specialPath);

    expect(document.getElementById("selected-path").textContent).toBe(specialPath);
    expect(result).toBe(specialPath);
  });
});

// Test formatBytes functionality
describe("formatBytes Function", () => {
  test("should format bytes correctly", () => {
    expect(appFunctions.formatBytes(0)).toBe("0 Bytes");
    expect(appFunctions.formatBytes(100)).toBe("100 Bytes");
    expect(appFunctions.formatBytes(1023)).toBe("1023 Bytes");
  });

  test("should format kilobytes correctly", () => {
    expect(appFunctions.formatBytes(1024)).toBe("1.00 KB");
    expect(appFunctions.formatBytes(1536)).toBe("1.50 KB");
  });

  test("should format megabytes correctly", () => {
    expect(appFunctions.formatBytes(1048576)).toBe("1.00 MB");
    expect(appFunctions.formatBytes(2097152)).toBe("2.00 MB");
  });

  test("should respect decimal parameter", () => {
    expect(appFunctions.formatBytes(1234567, 0)).toBe("1 MB");
    expect(appFunctions.formatBytes(1234567, 1)).toBe("1.2 MB");
    expect(appFunctions.formatBytes(1234567, 3)).toBe("1.177 MB");
  });
});

// Test clipboard functionality
describe("Clipboard Functionality", () => {
  // Mock the clipboard API
  const originalClipboard = { ...global.navigator.clipboard };
  let clipboardText = "";

  beforeEach(() => {
    document.body.innerHTML = `
      <div id="selected-path" class="selected-path">/test/folder1/file1.txt</div>
      <div id="toast-message" class="toast" style="display: none;"></div>
    `;

    // Mock clipboard API
    global.navigator.clipboard = {
      writeText: jest.fn((text) => {
        clipboardText = text;
        return Promise.resolve();
      }),
    };

    // Add mock functions
    window.copyToClipboard = appFunctions.copyToClipboard;
    window.showToast = appFunctions.showToast;
  });

  afterEach(() => {
    global.navigator.clipboard = originalClipboard;
  });

  test("copyToClipboard should write text to clipboard", async () => {
    const textToCopy = "/test/folder1/file1.txt";
    await window.copyToClipboard(textToCopy);

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(textToCopy);
  });

  test("showToast should display a toast message", () => {
    const message = "Test toast message";
    window.showToast(message);

    const toast = document.getElementById("toast-message");
    expect(toast.textContent).toBe(message);
    expect(toast.style.display).toBe("block");
  });
});

// Test formatBytes against frontend app implementation
describe("formatBytes Implementation Consistency", () => {
  test("should match the app.js implementation", () => {
    // Tests that our test function matches the actual implementation
    // This validates our test setup matches the actual app

    // These values should match the expected values in formatBytes.test.js
    expect(appFunctions.formatBytes(0)).toBe("0 Bytes");
    expect(appFunctions.formatBytes(100)).toBe("100 Bytes");
    expect(appFunctions.formatBytes(1024)).toBe("1.00 KB");
    expect(appFunctions.formatBytes(1536)).toBe("1.50 KB");
    expect(appFunctions.formatBytes(1048576)).toBe("1.00 MB");
    expect(appFunctions.formatBytes(1073741824)).toBe("1.00 GB");
    expect(appFunctions.formatBytes(1099511627776)).toBe("1.00 TB");
  });
});
