// DOM Elements
const pathInput = document.getElementById("path-input");
const searchInput = document.getElementById("search-input");
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
const breadcrumbTrail = document.getElementById("breadcrumb-trail");
const previousScansContainer = document.getElementById("previous-scans-container");
const previousScansList = document.getElementById("previous-scans-list");
const searchResultsContainer = document.getElementById("search-results-container");
const searchResultsCount = document.getElementById("search-results-count");
const searchResultsList = document.getElementById("search-results-list");
const colorLegend = document.getElementById("color-legend");
const legendItems = document.querySelector(".legend-items");
const zoomControls = document.getElementById("zoom-controls");
const zoomInBtn = document.getElementById("zoom-in-btn");
const zoomOutBtn = document.getElementById("zoom-out-btn");
const zoomResetBtn = document.getElementById("zoom-reset-btn");

// Application state
let currentData = null;
let currentPath = [];
let vizType = "treemap";
// scanning state is managed by UI updates
let progressInterval = null;
let previousScans = [];
let currentZoom = null;

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

// Map file extensions to types
const fileTypeMappings = {
  // Images
  jpg: "image",
  jpeg: "image",
  png: "image",
  gif: "image",
  webp: "image",
  svg: "image",
  bmp: "image",
  tiff: "image",
  ico: "image",
  heic: "image",

  // Videos
  mp4: "video",
  avi: "video",
  mov: "video",
  wmv: "video",
  mkv: "video",
  webm: "video",
  flv: "video",
  m4v: "video",
  mpg: "video",
  mpeg: "video",

  // Audio
  mp3: "audio",
  wav: "audio",
  ogg: "audio",
  flac: "audio",
  aac: "audio",
  m4a: "audio",
  wma: "audio",
  opus: "audio",

  // Documents
  pdf: "document",
  doc: "document",
  docx: "document",
  xls: "document",
  xlsx: "document",
  ppt: "document",
  pptx: "document",
  txt: "document",
  rtf: "document",
  md: "document",
  csv: "document",
  json: "document",
  xml: "document",
  html: "document",
  htm: "document",

  // Archives
  zip: "archive",
  rar: "archive",
  "7z": "archive",
  tar: "archive",
  gz: "archive",
  bz2: "archive",
  xz: "archive",
  iso: "archive",
  dmg: "archive",
};

// Initialize the application
function init() {
  // Set up event listeners
  browseBtn.addEventListener("click", browseDirectory);
  homeBtn.addEventListener("click", setHomeDirectory);
  scanBtn.addEventListener("click", startScan);
  stopBtn.addEventListener("click", stopScan);

  // Set up search input event listener
  searchInput.addEventListener("input", handleSearchInput);

  // Listen for visualization type changes
  vizTypeRadios.forEach((radio) => {
    radio.addEventListener("change", (e) => {
      vizType = e.target.value;

      // Show/hide zoom controls based on visualization type
      if (vizType === "sunburst") {
        zoomControls.style.display = "flex";
      } else {
        zoomControls.style.display = "none";
      }

      if (currentData) {
        renderVisualization(currentData);
      }
    });
  });

  // Set up zoom control event listeners
  zoomInBtn.addEventListener("click", () => {
    if (currentZoom) {
      const svg = d3.select(visualizationContainer).select("svg");
      svg.transition().call(currentZoom.scaleBy, 1.5);
    }
  });

  zoomOutBtn.addEventListener("click", () => {
    if (currentZoom) {
      const svg = d3.select(visualizationContainer).select("svg");
      svg.transition().call(currentZoom.scaleBy, 1 / 1.5);
    }
  });

  zoomResetBtn.addEventListener("click", () => {
    if (currentZoom) {
      const svg = d3.select(visualizationContainer).select("svg");
      svg.transition().call(currentZoom.transform, d3.zoomIdentity);
    }
  });

  // Set up click handler for path text to copy
  selectedPathText.addEventListener("click", copyPathToClipboard);

  // Set initial home directory
  setHomeDirectory();

  // Fetch previous scans
  fetchPreviousScans();

  // Initialize color legend
  initializeColorLegend();

  // Set up keyboard shortcuts
  document.addEventListener("keydown", (e) => {
    // Escape key navigates up one level
    if (e.key === "Escape" && currentPath.length > 0) {
      currentPath.pop();
      renderVisualization(currentData);
    }

    // Enter key starts scan when path input is focused
    if (e.key === "Enter" && document.activeElement === pathInput) {
      startScan();
    }
  });
}

// Browse for a directory
async function browseDirectory() {
  try {
    const response = await fetch("/api/browse");
    if (!response.ok) {
      throw new Error(`Server responded with ${response.status}: ${response.statusText}`);
    }

    const result = await response.json();
    if (result.path) {
      pathInput.value = result.path;
    }
  } catch (error) {
    alert("Error browsing for directory: " + error.message);
  }
}

// Set home directory in the path input
async function setHomeDirectory() {
  try {
    const response = await fetch("/api/home");
    if (!response.ok) {
      throw new Error(`Server responded with ${response.status}: ${response.statusText}`);
    }

    const result = await response.json();
    if (result.home) {
      pathInput.value = result.home;
    }
  } catch (error) {
    alert("Error setting home directory: " + error.message);
  }
}

// Start a new scan
async function startScan() {
  // Validate path
  const path = pathInput.value.trim();
  if (!path) {
    alert("Please enter a directory path");
    return;
  }

  // Prepare request data
  const requestData = {
    path: path,
    ignoreHidden: ignoreHiddenCheckbox.checked,
    searchTerm: searchInput.value.trim(),
  };

  try {
    // Update UI
    updateScanningUI(true);

    // Send scan request
    const response = await fetch("/api/scan", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestData),
    });

    if (!response.ok) {
      throw new Error(`Server responded with ${response.status}: ${response.statusText}`);
    }

    await response.json();

    // Start polling for scan progress
    progressInterval = setInterval(pollScanProgress, 500);
  } catch (error) {
    // Handle error when starting scan
    alert("Error starting scan: " + error.message);

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
      clearInterval(progressInterval);
      progressInterval = null;
      updateScanningUI(false);

      // Wait a moment to ensure the data is ready, then fetch the result
      setTimeout(fetchScanResult, 500);
      return;
    }

    // Update progress UI
    const progress = data.progress;
    scannedItemsText.textContent = progress.scannedItems;
    totalItemsText.textContent = progress.totalItems;
    const percentage = (progress.progress * 100).toFixed(1);
    progressPercentText.textContent = percentage + "%";
    progressBarFill.style.width = percentage + "%";
    currentPathText.textContent = progress.currentPath;

    // Update search results if search term is provided
    if (progress.searchTerm && progress.searchResults) {
      updateSearchResults(progress.searchResults, progress.searchTerm);
    }

    // Check if scan is stalled
    if (data.stalled) {
      // Alert user that scan appears to be stalled
      const stalledMsg = `STALLED: ${progress.currentPath}`;
      const styledMsg = `<span style="color: orange; font-weight: bold;">${stalledMsg}</span>`;
      currentPathText.innerHTML = styledMsg;

      // Show option to stop scan if stalled
      if (!document.getElementById("stall-warning")) {
        const warningEl = document.createElement("div");
        warningEl.id = "stall-warning";
        warningEl.className = "stall-warning";
        warningEl.innerHTML = `<p>Scan stalled processing large files or directories.</p>
           <p><button id="restart-scan-btn" class="btn btn-warning">Stop Scan</button> or wait</p>`;
        progressContainer.appendChild(warningEl);
        document.getElementById("restart-scan-btn").addEventListener("click", stopScan);
      }
    } else if (document.getElementById("stall-warning")) {
      // Remove stall warning if scan is no longer stalled
      document.getElementById("stall-warning").remove();
    }
  } catch (error) {
    // Handle error when polling for scan progress
  }
}

// Stop an in-progress scan
async function stopScan() {
  try {
    await fetch("/api/scan/stop", { method: "POST" });

    // Clear the progress interval
    if (progressInterval) {
      clearInterval(progressInterval);
      progressInterval = null;
    }
  } catch (error) {
    alert("Error stopping scan: " + error.message);
  }
}

// Fetch scan result data
async function fetchScanResult(resultId = null) {
  try {
    let url = "/api/results";
    if (resultId) {
      url += `?id=${resultId}`;
    }

    const response = await fetch(url);

    if (!response.ok) {
      if (response.status === 404) {
        alert("No scan results available. Please run a scan first.");
      } else {
        throw new Error(`Server responded with ${response.status}: ${response.statusText}`);
      }
      return;
    }

    const result = await response.json();

    // Reset current path
    currentPath = [];

    // Store the data
    currentData = result;

    // Render the visualization
    renderVisualization(result);

    // Show the visualization container
    visualizationContainer.style.display = "block";

    // Update details panel with root directory
    updateDetailsPanel(result);

    // Show the breadcrumb trail
    breadcrumbTrail.style.display = "block";

    // Update breadcrumbs
    updateBreadcrumbs();

    // If this is a new scan, refresh the previous scans list
    if (!resultId) {
      fetchPreviousScans();
    }
  } catch (error) {
    // Handle error when fetching scan results
    alert("Error fetching scan results: " + error.message);
  }
}

// Update UI during scanning
function updateScanningUI(isScanning) {
  if (isScanning) {
    // Update button states
    scanBtn.disabled = true;
    stopBtn.disabled = false;

    // Show progress container
    progressContainer.classList.remove("hidden");

    // Reset progress indicators
    progressBarFill.style.width = "0%";
    scannedItemsText.textContent = "0";
    totalItemsText.textContent = "0";
    progressPercentText.textContent = "0.0%";
    currentPathText.textContent = "Starting scan...";

    // Show search results container if search term is provided
    const searchTerm = searchInput.value.trim();
    if (searchTerm) {
      searchResultsContainer.classList.remove("hidden");
      searchResultsList.innerHTML = "";
      searchResultsCount.textContent = "0 files found";
    } else {
      searchResultsContainer.classList.add("hidden");
    }
  } else {
    // Update button states
    scanBtn.disabled = false;
    stopBtn.disabled = true;

    // Hide progress container
    progressContainer.classList.add("hidden");
  }
}

// Render visualization based on data and current visualization type
function renderVisualization(data) {
  // Clear the visualization container
  visualizationContainer.innerHTML = "";

  // Reset zoom reference
  currentZoom = null;

  // Show/hide zoom controls based on visualization type
  if (vizType === "sunburst") {
    zoomControls.style.display = "flex";
  } else {
    zoomControls.style.display = "none";
  }

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

  // Create a hierarchy from the data
  const hierarchy = d3
    .hierarchy(data)
    .sum(
      (d) => (d.size > 0 ? d.size : 0) // Ensure we use all sizes, not just files
    )
    .sort((a, b) => b.value - a.value);

  // Find the current node if navigating into a subdirectory
  let currentHierarchy = hierarchy;
  if (currentPath.length > 0) {
    let currentNode = hierarchy;
    for (const segment of currentPath) {
      const nextNode = currentNode.children?.find((child) => child.data.name === segment);
      if (!nextNode) {
        break;
      }
      currentNode = nextNode;
    }
    // Create a new hierarchy from the current node's data
    currentHierarchy = d3
      .hierarchy(currentNode.data)
      .sum((d) => (d.size > 0 ? d.size : 0))
      .sort((a, b) => b.value - a.value);
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
  const root = treemap(currentHierarchy);

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
    .attr("transform", (d) => {
      const x = isNaN(d.x0) ? 0 : d.x0;
      const y = isNaN(d.y0) ? 0 : d.y0;
      return `translate(${x},${y})`;
    })
    .on("click", function (event, d) {
      // Navigate deeper on click if it's a directory
      if (d.data.isDir && d.children && d.depth > 0) {
        // Create a new array with a copy of the current path plus the new item
        const newPath = [...currentPath];
        newPath.push(d.data.name);
        currentPath = newPath;
        renderVisualization(data);
      }

      // Update details panel
      updateDetailsPanel(d.data);
    });

  // Add rectangles for each cell - show direct children (depth 1)
  cell
    .filter((d) => d.depth === 1)
    .append("rect")
    .attr("id", (d) => `rect-${d.data.name.replace(/\s+/g, "-")}`)
    .attr("width", (d) => {
      const width = d.x1 - d.x0;
      return isNaN(width) || width < 0 ? 0 : width;
    })
    .attr("height", (d) => {
      const height = d.y1 - d.y0;
      return isNaN(height) || height < 0 ? 0 : height;
    })
    .attr("fill", (d) => {
      if (d.data.isDir) {
        // For directories, use a standard color
        return typeColors.directory;
      } else if (d.data.fileTypes) {
        // For directories with file type stats, use a multi-colored pattern
        const boxWidth = d.x1 - d.x0;
        const boxHeight = d.y1 - d.y0;
        if (!isNaN(boxWidth) && !isNaN(boxHeight) && boxWidth > 0 && boxHeight > 0) {
          renderMultiColoredBox(
            `rect-${d.data.name.replace(/\s+/g, "-")}`,
            boxWidth,
            boxHeight,
            d.data.fileTypes
          );
        }
        return `url(#pattern-${d.data.name.replace(/\s+/g, "-")})`;
      } else {
        // For files, color by extension
        return getFileTypeColor(d.data.extension);
      }
    });

  // Add title for each cell
  cell
    .filter((d) => d.depth === 1)
    .append("text")
    .attr("x", 3)
    .attr("y", 14)
    .text((d) => {
      let name = d.data.name;
      // Truncate long names
      const cellWidth = d.x1 - d.x0;
      const maxLength = Math.floor((!isNaN(cellWidth) && cellWidth > 0 ? cellWidth : 50) / 8); // Approximate chars that fit
      if (name.length > maxLength) {
        name = name.substring(0, maxLength - 3) + "...";
      }
      return name;
    })
    .attr("fill", "white")
    .attr("font-weight", "bold")
    .attr("font-size", "12px")
    .attr("pointer-events", "none");

  // Add file size for each cell if there's room
  cell
    .filter((d) => d.depth === 1 && d.y1 - d.y0 > 30)
    .append("text")
    .attr("x", 3)
    .attr("y", 30)
    .text((d) => formatBytes(d.value))
    .attr("fill", "white")
    .attr("font-size", "10px")
    .attr("pointer-events", "none");
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

  // If we're navigating to a subdirectory, filter the data
  let currentHierarchy = hierarchy;
  if (currentPath.length > 0) {
    let currentNode = hierarchy;
    for (const segment of currentPath) {
      const nextNode = currentNode.children?.find((child) => child.data.name === segment);
      if (!nextNode) {
        break;
      }
      currentNode = nextNode;
    }
    // Create a new hierarchy from the current node's data, making it the new root
    currentHierarchy = d3
      .hierarchy(currentNode.data)
      .sum((d) => (d.size > 0 ? d.size : 0))
      .sort((a, b) => b.value - a.value);
  }

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
  const root = partition(currentHierarchy);

  // Create SVG element
  const svg = d3
    .select(visualizationContainer)
    .append("svg")
    .attr("width", width)
    .attr("height", height);

  // Create a group for the sunburst that will be transformed by zoom
  const sunburstGroup = svg.append("g").attr("transform", `translate(${width / 2},${height / 2})`);

  // Create zoom behavior
  const zoom = d3
    .zoom()
    .scaleExtent([0.5, 10])
    .on("zoom", function (event) {
      const { x, y, k } = event.transform;
      sunburstGroup.attr("transform", `translate(${width / 2 + x},${height / 2 + y}) scale(${k})`);
    });

  // Store zoom reference for controls
  currentZoom = zoom;

  // Apply zoom to SVG
  svg.call(zoom);

  // Create paths for each data point
  sunburstGroup
    .selectAll("path")
    .data(root.descendants().filter((d) => d.depth))
    .enter()
    .append("path")
    .attr("class", "sunburst-path")
    .attr("fill", (d) => {
      if (d.data.isDir) {
        return typeColors.directory;
      }
      return getFileTypeColor(d.data.extension);
    })
    .attr("d", arc)
    .on("click", function (event, d) {
      // Stop zoom propagation
      event.stopPropagation();

      // Update details panel
      updateDetailsPanel(d.data);

      // Navigate deeper on click if it's a directory
      if (d.data.isDir && d.children) {
        // Calculate the path based on ancestors
        const newPath = [];
        let current = d;
        while (current.parent && current.parent.data.name) {
          newPath.unshift(current.data.name);
          current = current.parent;
        }
        currentPath = newPath;
        renderVisualization(data);
      }
    });

  // Add labels for larger segments
  sunburstGroup
    .selectAll("text")
    .data(
      root
        .descendants()
        .filter((d) => d.depth && (d.y1 - d.y0) * (d.x1 - d.x0) > 0.03 && d.data.name.length < 15)
    )
    .enter()
    .append("text")
    .attr("transform", (d) => {
      const x = (d.x0 + d.x1) / 2;
      const rotation = (x - Math.PI / 2) * (180 / Math.PI);
      return `translate(${arc.centroid(d)}) rotate(${rotation})`;
    })
    .attr("text-anchor", "middle")
    .text((d) => d.data.name)
    .attr("fill", "white")
    .attr("font-size", "10px")
    .attr("pointer-events", "none");
}

// Update details panel with selected item
function updateDetailsPanel(item) {
  if (!item) {
    return;
  }

  // Set path
  selectedPathText.textContent = item.path;
  selectedPathText.title = "Click to copy path to clipboard: " + item.path;

  // Set size
  selectedSizeText.textContent = formatBytes(item.size);

  // Set type information
  if (item.isDir) {
    let typeText = "Directory";
    if (item.fileTypes) {
      // Add breakdown of file types if available
      const breakdown = [];
      if (item.fileTypes.image > 0) {
        breakdown.push(`Images: ${formatBytes(item.fileTypes.image)}`);
      }
      if (item.fileTypes.video > 0) {
        breakdown.push(`Videos: ${formatBytes(item.fileTypes.video)}`);
      }
      if (item.fileTypes.audio > 0) {
        breakdown.push(`Audio: ${formatBytes(item.fileTypes.audio)}`);
      }
      if (item.fileTypes.document > 0) {
        breakdown.push(`Documents: ${formatBytes(item.fileTypes.document)}`);
      }
      if (item.fileTypes.archive > 0) {
        breakdown.push(`Archives: ${formatBytes(item.fileTypes.archive)}`);
      }
      if (item.fileTypes.other > 0) {
        breakdown.push(`Other: ${formatBytes(item.fileTypes.other)}`);
      }

      if (breakdown.length > 0) {
        typeText += " - " + breakdown.join(", ");
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
    currentPath = []; // This is fine as we're assigning a new empty array
    renderVisualization(currentData);
  });
  breadcrumbTrail.appendChild(homeBreadcrumb);

  // Add path segments
  const pathSoFar = [];
  currentPath.forEach((segment) => {
    pathSoFar.push(segment);

    // Add separator before breadcrumb item (except for the first one after Root)
    const separator = document.createElement("span");
    separator.className = "breadcrumb-separator";
    separator.textContent = ">";
    breadcrumbTrail.appendChild(separator);

    const breadcrumb = document.createElement("span");
    breadcrumb.className = "breadcrumb-item";
    breadcrumb.textContent = segment;

    // Make a copy of the current path up to this point
    const pathCopy = [...pathSoFar];

    breadcrumb.addEventListener("click", () => {
      // pathCopy is already a new array created with [...pathSoFar]
      currentPath = pathCopy;
      renderVisualization(currentData);
    });

    breadcrumbTrail.appendChild(breadcrumb);
  });
}

// Get color for file type based on extension
function getFileTypeColor(extension) {
  if (!extension) {
    return typeColors.other;
  }

  const lowerExt = extension.toLowerCase();
  const type = fileTypeMappings[lowerExt];

  if (type && typeColors[type]) {
    return typeColors[type];
  }

  return typeColors.other;
}

// Render a multi-colored box representing file type distribution
function renderMultiColoredBox(rectId, width, height, fileTypes) {
  // Calculate total size
  const total =
    fileTypes.image +
    fileTypes.video +
    fileTypes.audio +
    fileTypes.document +
    fileTypes.archive +
    fileTypes.other;

  if (total === 0) {
    return;
  }

  // Calculate proportions
  const imageProp = fileTypes.image / total;
  const videoProp = fileTypes.video / total;
  const audioProp = fileTypes.audio / total;
  const documentProp = fileTypes.document / total;
  const archiveProp = fileTypes.archive / total;
  const otherProp = fileTypes.other / total;

  // Create segments
  const segments = [];
  let currentPosition = 0;

  if (imageProp > 0) {
    segments.push({
      color: typeColors.image,
      start: currentPosition,
      end: currentPosition + imageProp,
    });
    currentPosition += imageProp;
  }

  if (videoProp > 0) {
    segments.push({
      color: typeColors.video,
      start: currentPosition,
      end: currentPosition + videoProp,
    });
    currentPosition += videoProp;
  }

  if (audioProp > 0) {
    segments.push({
      color: typeColors.audio,
      start: currentPosition,
      end: currentPosition + audioProp,
    });
    currentPosition += audioProp;
  }

  if (documentProp > 0) {
    segments.push({
      color: typeColors.document,
      start: currentPosition,
      end: currentPosition + documentProp,
    });
    currentPosition += documentProp;
  }

  if (archiveProp > 0) {
    segments.push({
      color: typeColors.archive,
      start: currentPosition,
      end: currentPosition + archiveProp,
    });
    currentPosition += archiveProp;
  }

  if (otherProp > 0) {
    segments.push({
      color: typeColors.other,
      start: currentPosition,
      end: 1,
    });
  }

  // Create pattern definition with stripes
  const svg = d3.select("svg");
  const patternId = `pattern-${rectId.replace("rect-", "")}`;

  const defs = svg.append("defs");

  const pattern = defs
    .append("pattern")
    .attr("id", patternId)
    .attr("width", width)
    .attr("height", height)
    .attr("patternUnits", "userSpaceOnUse");

  // Add colored rectangles
  segments.forEach((segment) => {
    pattern
      .append("rect")
      .attr("x", 0)
      .attr("y", height * segment.start)
      .attr("width", width)
      .attr("height", height * (segment.end - segment.start))
      .attr("fill", segment.color);
  });
}

// Format bytes to human readable format
function formatBytes(bytes) {
  if (bytes === 0) {
    return "0 Bytes";
  }

  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB"];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

// Copy path to clipboard
function copyPathToClipboard() {
  const path = selectedPathText.textContent;

  if (!path || path === "No item selected") {
    return;
  }

  navigator.clipboard
    .writeText(path)
    .then(() => {
      // Show temporary confirmation
      const originalText = selectedPathText.textContent;
      selectedPathText.textContent = "Copied to clipboard!";
      setTimeout(() => {
        selectedPathText.textContent = originalText;
      }, 1500);
    })
    .catch((err) => {
      console.error("Failed to copy path:", err);
    });
}

// Fetch and display previous scans
async function fetchPreviousScans() {
  try {
    const response = await fetch("/api/previous-scans");
    if (!response.ok) {
      throw new Error(`Server responded with ${response.status}: ${response.statusText}`);
    }

    const scans = await response.json();

    // Store scans in state
    previousScans = scans;

    // Display the scans
    displayPreviousScans();
  } catch (error) {
    // Handle error when fetching previous scans
  }
}

// Display previous scans in the UI
function displayPreviousScans() {
  previousScansList.innerHTML = "";

  if (previousScans.length > 0) {
    previousScansContainer.classList.remove("hidden");

    previousScans.forEach((scan) => {
      const scanItem = document.createElement("div");
      scanItem.className = "previous-scan-item";

      // Format date
      const scanDate = new Date(scan.timestamp);
      const formattedDate = scanDate.toLocaleString();

      scanItem.innerHTML = `
        <div class="scan-path">${scan.path}</div>
        <div class="scan-info">${formattedDate} - ${formatBytes(scan.size)}</div>
      `;

      scanItem.addEventListener("click", () => {
        fetchScanResult(scan.resultId);
      });

      previousScansList.appendChild(scanItem);
    });
  } else {
    previousScansContainer.classList.add("hidden");
  }
}

// Handle search input changes
function handleSearchInput() {
  const searchTerm = searchInput.value.trim();

  if (!currentData) {
    return;
  }

  if (searchTerm === "") {
    // Hide search results if no search term
    searchResultsContainer.classList.add("hidden");
    return;
  }

  // Show search results container
  searchResultsContainer.classList.remove("hidden");

  // Search through current data
  const searchResults = [];

  function searchInData(data, path = "") {
    if (data.name && data.name.toLowerCase().includes(searchTerm.toLowerCase())) {
      searchResults.push({
        name: data.name,
        path: path + "/" + data.name,
        size: data.size || 0,
        type: data.type || "unknown",
      });
    }

    if (data.children) {
      data.children.forEach((child) => {
        searchInData(child, path + "/" + data.name);
      });
    }
  }

  // Start search from root
  if (currentData.children) {
    currentData.children.forEach((child) => {
      searchInData(child, "");
    });
  }

  // Update search results display
  updateSearchResults(searchResults, searchTerm);
}

// Update search results display
function updateSearchResults(searchResults, searchTerm) {
  // Update count
  searchResultsCount.textContent = `${searchResults.length} files found matching "${searchTerm}"`;

  // Clear and populate results list
  searchResultsList.innerHTML = "";

  // Limit to last 50 results to avoid performance issues
  const displayResults = searchResults.slice(-50);

  displayResults.forEach((result) => {
    const resultItem = document.createElement("div");
    resultItem.className = "search-result-item";

    const fileName = document.createElement("div");
    fileName.className = "search-result-name";
    fileName.textContent = result.name;

    const filePath = document.createElement("div");
    filePath.className = "search-result-path";
    filePath.textContent = result.path;
    filePath.title = "Click to copy path";

    const fileSize = document.createElement("div");
    fileSize.className = "search-result-size";
    fileSize.textContent = formatBytes(result.size);

    // Add click handler to copy path
    filePath.addEventListener("click", () => {
      navigator.clipboard.writeText(result.path).then(() => {
        const original = filePath.textContent;
        filePath.textContent = "Copied!";
        setTimeout(() => {
          filePath.textContent = original;
        }, 1000);
      });
    });

    resultItem.appendChild(fileName);
    resultItem.appendChild(filePath);
    resultItem.appendChild(fileSize);
    searchResultsList.appendChild(resultItem);
  });

  // If we're showing only a subset, add a note
  if (searchResults.length > 50) {
    const moreItem = document.createElement("div");
    moreItem.className = "search-result-more";
    moreItem.textContent = `... and ${searchResults.length - 50} more files (showing latest 50)`;
    searchResultsList.appendChild(moreItem);
  }
}

// Initialize the color legend
function initializeColorLegend() {
  // Clear existing legend items
  legendItems.innerHTML = "";

  // Create legend items for each file type
  Object.entries(typeColors).forEach(([type, color]) => {
    const legendItem = document.createElement("div");
    legendItem.className = "legend-item";

    const colorBox = document.createElement("div");
    colorBox.className = "legend-color";
    colorBox.style.backgroundColor = color;

    const label = document.createElement("span");
    label.className = "legend-label";
    label.textContent = type;

    legendItem.appendChild(colorBox);
    legendItem.appendChild(label);
    legendItems.appendChild(legendItem);
  });
}

// Initialize the application when the page loads
document.addEventListener("DOMContentLoaded", init);
