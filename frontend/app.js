// DOM Elements
const pathInput = document.getElementById("path-input");
const browseBtn = document.getElementById("browse-btn");
const homeBtn = document.getElementById("home-btn");
const scanBtn = document.getElementById("scan-btn");
const stopBtn = document.getElementById("stop-btn");
const ignoreHiddenCheckbox = document.getElementById("ignore-hidden");
const vizTypeRadios = document.querySelectorAll('input[name="viz-type"]');
const progressContainer = document.getElementById("progress-container");
const progressBarFill = document.getElementById("progress-bar-fill");
const scannedItemsText = document.getElementById("scanned-items");
const totalItemsText = document.getElementById("total-items");
const progressPercentText = document.getElementById("progress-percent");
const currentPathText = document.getElementById("current-path");
const visualizationContainer = document.getElementById("visualization");
const selectedPathText = document.getElementById("selected-path");
const selectedSizeText = document.getElementById("selected-size");
const selectedTypeText = document.getElementById("selected-type");
const breadcrumbsContainer = document.getElementById("breadcrumbs");
const breadcrumbTrail = document.getElementById("breadcrumb-trail");
const previousScansContainer = document.getElementById("previous-scans-container");
const previousScansList = document.getElementById("previous-scans-list");

// Application state
// Variables to store state
let currentData = null;
let currentPath = [];
let vizType = "treemap";
let scanning = false;
let progressInterval = null;
let previousScans = [];

// File type colors
const typeColors = {
  directory: "#5b9bd5",
  image: "#e74c3c",
  video: "#9b59b6",
  audio: "#2ecc71",
  document: "#f39c12",
  archive: "#f1c40f",
  other: "#95a5a6",
};

// File type mappings
const fileTypeMappings = {
  image: ["jpg", "jpeg", "png", "gif", "svg", "webp", "bmp", "tiff", "ico", "heic"],
  video: ["mp4", "mov", "avi", "mkv", "wmv", "flv", "webm", "m4v", "mpg", "mpeg"],
  audio: ["mp3", "wav", "ogg", "flac", "aac", "m4a", "wma", "aiff"],
  document: [
    "pdf",
    "doc",
    "docx",
    "xls",
    "xlsx",
    "ppt",
    "pptx",
    "txt",
    "rtf",
    "md",
    "csv",
    "json",
    "xml",
    "html",
    "css",
    "js",
    "ts",
    "go",
    "py",
    "java",
    "c",
    "cpp",
    "h",
    "rb",
    "php",
  ],
  archive: ["zip", "rar", "7z", "tar", "gz", "bz2", "xz", "iso", "dmg"],
};

// Initialize the application
function init() {
  console.log("Initializing Storage Shower application...");
  console.log("DOM elements found:", {
    pathInput: !!pathInput,
    browseBtn: !!browseBtn,
    scanBtn: !!scanBtn,
    stopBtn: !!stopBtn,
    homeBtn: !!homeBtn,
  });

  // Set up event listeners
  homeBtn.addEventListener("click", setHomeDirectory);
  browseBtn.addEventListener("click", browseDirectory);
  scanBtn.addEventListener("click", startScan);
  stopBtn.addEventListener("click", stopScan);
  vizTypeRadios.forEach((radio) => {
    radio.addEventListener("change", (e) => {
      vizType = e.target.value;
      if (currentData) {
        renderVisualization(currentData);
      }
    });
  });

  console.log("Event listeners attached");

  // Add click event for copying file path
  selectedPathText.addEventListener("click", copyPathToClipboard);

  // Fetch previous scans
  fetchPreviousScans();

  // Get home directory on load
  setHomeDirectory();
}

// Open file dialog to select a directory
async function browseDirectory() {
  console.log("Browse button clicked");
  try {
    const response = await fetch("/api/browse");
    if (!response.ok) {
      throw new Error(`Error ${response.status}: ${await response.text()}`);
    }

    const data = await response.json();
    console.log("Selected directory:", data.path);

    // Update the path input with the selected directory
    pathInput.value = data.path;
  } catch (error) {
    console.error("Error browsing directory:", error);
    alert("Failed to open directory browser. Please manually enter the directory path.");
  }
}

async function setHomeDirectory() {
  try {
    const response = await fetch("/api/home");
    const data = await response.json();

    pathInput.value = data.home;
    console.log("Home directory set:", data.home);
  } catch (error) {
    console.error("Error getting home directory:", error);
  }
}

// Start scanning the specified directory
async function startScan() {
  const path = pathInput.value.trim();
  console.log("startScan called with path:", path);

  if (!path) {
    console.log("Scan aborted: empty path");
    alert("Please enter a valid path");
    return;
  }

  try {
    console.log("Starting scan...", { path, ignoreHidden: ignoreHiddenCheckbox.checked });
    scanning = true;
    updateScanningUI(true);

    const response = await fetch("/api/scan", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        path: path,
        ignoreHidden: ignoreHiddenCheckbox.checked,
      }),
    });

    console.log("Scan API response status:", response.status);

    if (!response.ok) {
      const error = await response.text();
      console.error("Scan API error:", error);
      throw new Error(error);
    }

    const result = await response.json();
    console.log("Scan started successfully:", result);

    // Start polling for scan progress
    console.log("Starting progress polling...");
    progressInterval = setInterval(pollScanProgress, 500);
  } catch (error) {
    console.error("Error starting scan:", error);
    alert("Error starting scan: " + error.message);
    scanning = false;
    updateScanningUI(false);
  }
}

// Poll for scan progress
async function pollScanProgress() {
  try {
    const response = await fetch("/api/scan/status");
    const data = await response.json();

    if (!data.inProgress) {
      // Scan is complete or was stopped
      console.log("Scan completed, stopping progress polling");
      clearInterval(progressInterval);
      progressInterval = null;
      scanning = false;
      updateScanningUI(false);

      // Wait a moment to ensure the data is ready, then fetch the result
      console.log("Fetching scan results in 1 second...");
      setTimeout(() => fetchScanResult(), 1000);
      return;
    }

    // Update progress UI
    const progress = data.progress;
    scannedItemsText.textContent = data.progress.scannedItems;
    totalItemsText.textContent = data.progress.totalItems;
    const percentage = (progress.progress * 100).toFixed(1);
    progressPercentText.textContent = percentage + "%";
    progressBarFill.style.width = percentage + "%";
    currentPathText.textContent = progress.currentPath;

    // Check if scan is stalled
    if (data.stalled) {
      console.warn("Scan appears to be stalled");
      currentPathText.innerHTML = `<span style="color: orange; font-weight: bold;">STALLED: ${progress.currentPath}</span>`;

      // Show option to stop scan if stalled
      if (!document.getElementById("stall-warning")) {
        const warningEl = document.createElement("div");
        warningEl.id = "stall-warning";
        warningEl.className = "stall-warning";
        warningEl.innerHTML = `
                    <p>The scan appears to be stalled processing large files or directories.</p>
                    <p>You can <button id="restart-scan-btn" class="btn btn-warning">Stop Scan</button> or wait for it to complete.</p>
                `;
        progressContainer.appendChild(warningEl);
        document.getElementById("restart-scan-btn").addEventListener("click", stopScan);
      }
    } else if (document.getElementById("stall-warning")) {
      // Remove stall warning if scan is no longer stalled
      document.getElementById("stall-warning").remove();
    }
  } catch (error) {
    console.error("Error polling scan progress:", error);
  }
}

// Stop the current scan
async function stopScan() {
  console.log("Stop scan requested");
  try {
    const response = await fetch("/api/scan/stop", { method: "POST" });
    console.log("Stop scan response:", response.status);
    // The poll will detect when the scan is stopped
  } catch (error) {
    console.error("Error stopping scan:", error);
  }
}

// Fetch scan result data
async function fetchScanResult(resultId = null) {
  console.log("Fetching scan results...");
  try {
    let url = "/api/results";
    if (resultId) {
      url += `?id=${resultId}`;
    }

    const response = await fetch(url);
    if (!response.ok) {
      // If no results are available yet, try again after a short delay
      if (response.status === 404) {
        console.log("Results not ready yet, retrying in 500ms...");
        setTimeout(fetchScanResult, 500);
        return;
      }
      const errorText = await response.text();
      console.error("Failed to fetch results:", errorText);
      throw new Error(`Error ${response.status}: ${errorText}`);
    }

    const data = await response.json();
    console.log("Scan results received:", data);
    // Log root size and children sizes for debugging
    console.log("Root size:", data.size);
    console.log("Root children:", data.children ? data.children.length : 0);
    if (data.children) {
      data.children.forEach((child) => {
        console.log(`Child: ${child.name}, Size: ${child.size}, IsDir: ${child.isDir}`);
      });
    }
    currentData = data;
    currentPath = [];
    renderVisualization(currentData);

    // Refresh previous scans list if we just completed a new scan
    if (!resultId) {
      fetchPreviousScans();
    }
  } catch (error) {
    console.error("Error fetching scan result:", error);
    alert("Error fetching scan results: " + error.message);
  }
}

// Update UI elements based on scanning state
function updateScanningUI(isScanning) {
  console.log("Updating scanning UI, isScanning:", isScanning);
  scanBtn.disabled = isScanning;
  stopBtn.disabled = !isScanning;
  progressContainer.classList.toggle("hidden", !isScanning);

  // Hide previous scans container during scanning
  if (isScanning) {
    previousScansContainer.classList.add("hidden");

    // Reset progress indicators
    progressBarFill.style.width = "0%";
    scannedItemsText.textContent = "0";
    totalItemsText.textContent = "0";
    progressPercentText.textContent = "0.0%";
    currentPathText.textContent = "";

    // Remove any stall warnings that might exist
    const stallWarning = document.getElementById("stall-warning");
    if (stallWarning) {
      stallWarning.remove();
    }
  }
}

// Render visualization based on data and current visualization type
function renderVisualization(data) {
  // Clear the visualization container
  visualizationContainer.innerHTML = "";

  // Select the right visualization based on the vizType
  if (vizType === "treemap") {
    renderTreemap(data);
  } else {
    renderSunburst(data);
  }

  // Update breadcrumb trail
  updateBreadcrumbs();
}

// Render treemap visualization
function renderTreemap(data) {
  // Get dimensions
  const width = visualizationContainer.clientWidth;
  const height = visualizationContainer.clientHeight;

  console.log("Rendering treemap with data:", data);
  console.log("Root size before hierarchy:", data.size);

  // Create a hierarchy from the data
  const hierarchy = d3
    .hierarchy(data)
    .sum((d) => {
      console.log(`Summing: ${d.name}, Size: ${d.size}, IsDir: ${d.isDir}`);
      return d.size > 0 ? d.size : 0; // Ensure we use all sizes, not just files
    })
    .sort((a, b) => b.value - a.value);

  console.log("Hierarchy after sum:", hierarchy);

  // If we're navigating to a subdirectory, filter the data
  if (currentPath.length > 0) {
    let currentNode = hierarchy;
    for (const segment of currentPath) {
      const nextNode = currentNode.children?.find((child) => child.data.name === segment);
      if (!nextNode) break;
      currentNode = nextNode;
    }
    hierarchy = currentNode;
  }

  // Create treemap layout
  const treemap = d3
    .treemap()
    .size([width, height])
    .paddingOuter(3)
    .paddingTop(19)
    .paddingInner(1)
    .round(true);

  // Generate the treemap data
  const root = treemap(hierarchy);

  // Create SVG element
  const svg = d3
    .select(visualizationContainer)
    .append("svg")
    .attr("width", width)
    .attr("height", height)
    .style("font-family", "sans-serif");

  // Create cells for each data point
  const cell = svg
    .selectAll("g")
    .data(root.descendants())
    .enter()
    .append("g")
    .attr("transform", (d) => `translate(${d.x0},${d.y0})`)
    .on("click", function (event, d) {
      // Navigate deeper on click if it's a directory
      if (d.data.isDir && d.children && d.depth > 0) {
        currentPath.push(d.data.name);
        renderVisualization(data);
      }

      // Update details panel
      updateDetailsPanel(d.data);
    });

  // Add rectangles for each cell - either solid color or multi-colored pattern
  cell.each(function(d) {
    const cellGroup = d3.select(this);
    const width = d.x1 - d.x0;
    const height = d.y1 - d.y0;
    
    if (d.data.isDir && d.data.fileTypes) {
      // Create multi-colored rectangle for directories with file type data
      renderMultiColoredBox(cellGroup, width, height, d.data.fileTypes);
    } else {
      // Single color rectangle for files or directories without file type data
      cellGroup
        .append("rect")
        .attr("class", "node")
        .attr("width", width)
        .attr("height", height)
        .attr("fill", d.data.isDir ? typeColors.directory : getFileTypeColor(d.data.extension));
    }
  });

  // Add labels for cells that are large enough
  cell
    .append("text")
    .attr("class", "node-label")
    .selectAll("tspan")
    .data((d) => {
      const width = d.x1 - d.x0;
      const height = d.y1 - d.y0;
      if (width < 30 || height < 20) return []; // Skip small cells

      return [d.data.name, formatBytes(d.value)];
    })
    .enter()
    .append("tspan")
    .attr("x", 4)
    .attr("y", (d, i) => 13 + i * 10)
    .text((d) => d);
}

// Render sunburst visualization
function renderSunburst(data) {
  // Get dimensions
  const width = visualizationContainer.clientWidth;
  const height = visualizationContainer.clientHeight;
  const radius = Math.min(width, height) / 2;

  // Create a hierarchy from the data
  const hierarchy = d3
    .hierarchy(data)
    .sum((d) => (d.isDir ? 0 : d.size))
    .sort((a, b) => b.value - a.value);

  // Create the arc generator
  const arc = d3
    .arc()
    .startAngle((d) => d.x0)
    .endAngle((d) => d.x1)
    .padAngle(0.01)
    .padRadius(radius / 3)
    .innerRadius((d) => Math.sqrt(d.y0) * radius)
    .outerRadius((d) => Math.sqrt(d.y1) * radius - 1);

  // Create partition layout
  const partition = d3.partition().size([2 * Math.PI, 1]);

  // Generate the partition data
  const root = partition(hierarchy);

  // Create SVG element
  const svg = d3
    .select(visualizationContainer)
    .append("svg")
    .attr("width", width)
    .attr("height", height)
    .append("g")
    .attr("transform", `translate(${width / 2},${height / 2})`);

  // Create paths for each data point
  const path = svg
    .selectAll("path")
    .data(root.descendants().filter((d) => d.depth))
    .enter()
    .append("path")
    .attr("class", "sunburst-path")
    .attr("fill", (d) => {
      if (d.data.isDir) return typeColors.directory;
      return getFileTypeColor(d.data.extension);
    })
    .attr("d", arc)
    .on("click", function (event, d) {
      // Update details panel
      updateDetailsPanel(d.data);

      // Navigate deeper on click if it's a directory
      if (d.data.isDir && d.children) {
        // Calculate the path based on ancestors
        currentPath = [];
        let current = d;
        while (current.parent && current.parent.data.name) {
          currentPath.unshift(current.data.name);
          current = current.parent;
        }
        renderVisualization(data);
      }
    });

  // Add labels for larger arcs
  const label = svg
    .selectAll("text")
    .data(root.descendants().filter((d) => d.depth && (d.y1 - d.y0) * (d.x1 - d.x0) > 0.03))
    .enter()
    .append("text")
    .attr("class", "node-label")
    .attr("transform", (d) => {
      const x = (((d.x0 + d.x1) / 2) * 180) / Math.PI;
      const y = ((Math.sqrt(d.y0) + Math.sqrt(d.y1)) / 2) * radius;
      return `rotate(${x - 90}) translate(${y},0) rotate(${x < 180 ? 0 : 180})`;
    })
    .attr("dy", "0.35em")
    .text((d) => d.data.name);
}

// Update the details panel with information about the selected item
function updateDetailsPanel(item) {
  selectedPathText.textContent = item.path;
  selectedPathText.title = "Click to copy path to clipboard";
  selectedSizeText.textContent = formatBytes(item.size);

  if (item.isDir) {
    let typeText = "Directory";
    if (item.fileTypes) {
      const totalSize = item.fileTypes.image + item.fileTypes.video + item.fileTypes.audio + 
                       item.fileTypes.document + item.fileTypes.archive + item.fileTypes.other;
      if (totalSize > 0) {
        const breakdown = [];
        if (item.fileTypes.image > 0) breakdown.push(`Images: ${formatBytes(item.fileTypes.image)}`);
        if (item.fileTypes.video > 0) breakdown.push(`Videos: ${formatBytes(item.fileTypes.video)}`);
        if (item.fileTypes.audio > 0) breakdown.push(`Audio: ${formatBytes(item.fileTypes.audio)}`);
        if (item.fileTypes.document > 0) breakdown.push(`Documents: ${formatBytes(item.fileTypes.document)}`);
        if (item.fileTypes.archive > 0) breakdown.push(`Archives: ${formatBytes(item.fileTypes.archive)}`);
        if (item.fileTypes.other > 0) breakdown.push(`Other: ${formatBytes(item.fileTypes.other)}`);
        
        if (breakdown.length > 0) {
          typeText += " - " + breakdown.join(", ");
        }
      }
    }
    selectedTypeText.textContent = typeText;
  } else {
    selectedTypeText.textContent = item.extension ? `File (.${item.extension})` : "File";
  }
}

// Update breadcrumb trail
function updateBreadcrumbs() {
  breadcrumbTrail.innerHTML = "";

  // Add home breadcrumb
  const homeBreadcrumb = document.createElement("span");
  homeBreadcrumb.className = "breadcrumb-item";
  homeBreadcrumb.textContent = "Root";
  homeBreadcrumb.addEventListener("click", () => {
    currentPath = [];
    renderVisualization(currentData);
  });
  breadcrumbTrail.appendChild(homeBreadcrumb);

  // Add path segments
  let pathSoFar = [];
  currentPath.forEach((segment, index) => {
    pathSoFar.push(segment);

    const breadcrumb = document.createElement("span");
    breadcrumb.className = "breadcrumb-item";
    breadcrumb.textContent = segment;

    // Make a copy of the current path up to this point
    const pathCopy = [...pathSoFar];

    breadcrumb.addEventListener("click", () => {
      currentPath = pathCopy;
      renderVisualization(currentData);
    });

    breadcrumbTrail.appendChild(breadcrumb);
  });
}

// Get color for a file type based on its extension
function getFileTypeColor(extension) {
  if (!extension) return typeColors.other;

  const ext = extension.toLowerCase();

  for (const [type, extensions] of Object.entries(fileTypeMappings)) {
    if (extensions.includes(ext)) {
      return typeColors[type];
    }
  }

  return typeColors.other;
}

// Render a multi-colored box showing file type proportions
function renderMultiColoredBox(parentGroup, width, height, fileTypes) {
  const totalSize = fileTypes.image + fileTypes.video + fileTypes.audio + 
                   fileTypes.document + fileTypes.archive + fileTypes.other;
  
  if (totalSize === 0) {
    // If no files, render as directory color
    parentGroup
      .append("rect")
      .attr("class", "node")
      .attr("width", width)
      .attr("height", height)
      .attr("fill", typeColors.directory);
    return;
  }

  // Calculate proportions and create segments
  const segments = [];
  const types = [
    { name: 'image', size: fileTypes.image, color: typeColors.image },
    { name: 'video', size: fileTypes.video, color: typeColors.video },
    { name: 'audio', size: fileTypes.audio, color: typeColors.audio },
    { name: 'document', size: fileTypes.document, color: typeColors.document },
    { name: 'archive', size: fileTypes.archive, color: typeColors.archive },
    { name: 'other', size: fileTypes.other, color: typeColors.other }
  ];

  // Filter out types with zero size and calculate cumulative widths
  let currentX = 0;
  for (const type of types) {
    if (type.size > 0) {
      const segmentWidth = (type.size / totalSize) * width;
      segments.push({
        x: currentX,
        width: segmentWidth,
        color: type.color,
        type: type.name,
        size: type.size
      });
      currentX += segmentWidth;
    }
  }

  // Render each segment as a rectangle
  segments.forEach(segment => {
    parentGroup
      .append("rect")
      .attr("class", "node file-type-segment")
      .attr("x", segment.x)
      .attr("y", 0)
      .attr("width", segment.width)
      .attr("height", height)
      .attr("fill", segment.color)
      .attr("data-file-type", segment.type)
      .attr("data-file-size", segment.size);
  });
}

// Format bytes to human-readable format
function formatBytes(bytes, decimals = 2) {
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
}

function copyPathToClipboard() {
  if (selectedPathText.textContent && selectedPathText.textContent !== "No item selected") {
    navigator.clipboard
      .writeText(selectedPathText.textContent)
      .then(() => {
        // Visual feedback
        selectedPathText.classList.add("copied");
        setTimeout(() => {
          selectedPathText.classList.remove("copied");
        }, 1500);
      })
      .catch((err) => {
        console.error("Could not copy text: ", err);
      });
  }
}

// Fetch previous scans from the server
async function fetchPreviousScans() {
  try {
    const response = await fetch("/api/previous-scans");
    if (!response.ok) {
      throw new Error(`Error ${response.status}: ${await response.text()}`);
    }

    previousScans = await response.json();
    if (previousScans && previousScans.length > 0) {
      displayPreviousScans();
      previousScansContainer.classList.remove("hidden");
    } else {
      previousScansContainer.classList.add("hidden");
    }
  } catch (error) {
    console.error("Error fetching previous scans:", error);
  }
}

// Display previous scans in the UI
function displayPreviousScans() {
  previousScansList.innerHTML = "";

  previousScans.forEach((scan) => {
    const scanItem = document.createElement("div");
    scanItem.className = "previous-scan-item";
    scanItem.dataset.resultId = scan.resultId;

    const scanDate = new Date(scan.timestamp);
    const formattedDate = scanDate.toLocaleString();

    scanItem.innerHTML = `
            <div class="previous-scan-path">${scan.path}</div>
            <div class="previous-scan-info">
                <span>${formatBytes(scan.size)}</span>
                <span>${formattedDate}</span>
            </div>
        `;

    scanItem.addEventListener("click", () => {
      fetchScanResult(scan.resultId);
    });

    previousScansList.appendChild(scanItem);
  });
}

// Mock data generation for testing
function generateMockChildren(parentPath, depth) {
  if (depth <= 0) return [];

  const types = [
    { name: "Documents", size: 220000000, ext: null },
    { name: "Pictures", size: 350000000, ext: null },
    { name: "Music", size: 180000000, ext: null },
    { name: "Videos", size: 420000000, ext: null },
    { name: "report.pdf", size: 5000000, ext: "pdf" },
    { name: "presentation.pptx", size: 15000000, ext: "pptx" },
    { name: "vacation.jpg", size: 3000000, ext: "jpg" },
    { name: "movie.mp4", size: 150000000, ext: "mp4" },
    { name: "archive.zip", size: 80000000, ext: "zip" },
    { name: "song.mp3", size: 8000000, ext: "mp3" },
  ];

  // Randomize the number of items
  const numItems = Math.floor(Math.random() * 5) + 2;
  const children = [];

  for (let i = 0; i < numItems; i++) {
    const typeIndex = Math.floor(Math.random() * types.length);
    const type = types[typeIndex];

    const isDir = type.ext === null;
    const size = type.size * (0.5 + Math.random());

    const child = {
      name: type.name,
      path: parentPath + "/" + type.name,
      size: size,
      isDir: isDir,
      extension: type.ext,
    };

    if (isDir && depth > 1) {
      child.children = generateMockChildren(child.path, depth - 1);
    }

    children.push(child);
  }

  return children;
}

// Initialize the application
document.addEventListener("DOMContentLoaded", function () {
  console.log("DOM Content Loaded, starting initialization...");
  init();
});
