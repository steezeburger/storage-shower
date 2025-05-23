package utils

import (
	"time"
)

// StallDetector provides improved stall detection with thresholds and grace periods
type StallDetector struct {
	// The time of the last activity
	lastActivityTime time.Time
	
	// The number of items processed at the last check
	lastItemCount int
	
	// Threshold in seconds before considering a scan stalled
	stallThresholdSeconds int
	
	// Minimum number of items that must be processed before stall detection activates
	minimumItemsForStallCheck int
	
	// Grace period in seconds during startup
	startupGracePeriodSeconds int
	
	// Time when the scan started
	scanStartTime time.Time
}

// NewStallDetector creates a new stall detector with default settings
func NewStallDetector() *StallDetector {
	return &StallDetector{
		lastActivityTime:          time.Now(),
		lastItemCount:             0,
		stallThresholdSeconds:     30,
		minimumItemsForStallCheck: 500,
		startupGracePeriodSeconds: 10,
		scanStartTime:             time.Now(),
	}
}

// UpdateActivity records new activity, preventing stall detection
func (d *StallDetector) UpdateActivity(itemCount int) {
	// Only update if progress was made
	if itemCount > d.lastItemCount {
		d.lastActivityTime = time.Now()
		d.lastItemCount = itemCount
	}
}

// Reset resets the stall detector for a new scan
func (d *StallDetector) Reset() {
	d.lastActivityTime = time.Now()
	d.lastItemCount = 0
	d.scanStartTime = time.Now()
}

// IsStalled determines if the scan is stalled based on activity time and thresholds
func (d *StallDetector) IsStalled() bool {
	// Early in the scan, don't report stalls (grace period)
	if time.Since(d.scanStartTime).Seconds() < float64(d.startupGracePeriodSeconds) {
		return false
	}
	
	// If we haven't processed enough items, don't consider it stalled
	if d.lastItemCount < d.minimumItemsForStallCheck {
		return false
	}
	
	// Check if we've passed the stall threshold
	return time.Since(d.lastActivityTime).Seconds() > float64(d.stallThresholdSeconds)
}

// GetLastActivityTime returns the time of the last activity
func (d *StallDetector) GetLastActivityTime() time.Time {
	return d.lastActivityTime
}

// SetStallThreshold sets the threshold in seconds before considering a scan stalled
func (d *StallDetector) SetStallThreshold(seconds int) {
	if seconds > 0 {
		d.stallThresholdSeconds = seconds
	}
}

// SetMinimumItemsForStallCheck sets the minimum number of items that must be processed
// before stall detection activates
func (d *StallDetector) SetMinimumItemsForStallCheck(count int) {
	if count > 0 {
		d.minimumItemsForStallCheck = count
	}
}

// SetStartupGracePeriod sets the grace period in seconds during startup
func (d *StallDetector) SetStartupGracePeriod(seconds int) {
	if seconds > 0 {
		d.startupGracePeriodSeconds = seconds
	}
}

