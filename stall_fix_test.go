package main

import (
	"testing"
	"time"
)

func TestStallDetector_IsStalled(t *testing.T) {
	detector := NewStallDetector()
	
	// Initially, it shouldn't be stalled
	if detector.IsStalled() {
		t.Error("New detector should not report stalled state")
	}
	
	// Manually set startup time to bypass grace period
	detector.scanStartTime = time.Now().Add(-time.Duration(detector.startupGracePeriodSeconds+1) * time.Second)
	
	// Still shouldn't be stalled because not enough items processed
	if detector.IsStalled() {
		t.Error("Detector should not report stalled with insufficient items")
	}
	
	// Update with enough items to enable stall detection
	detector.UpdateActivity(detector.minimumItemsForStallCheck + 10)
	
	// Still not stalled because activity was just updated
	if detector.IsStalled() {
		t.Error("Detector should not report stalled immediately after activity")
	}
	
	// Manually set the last activity time to trigger stall
	detector.lastActivityTime = time.Now().Add(-time.Duration(detector.stallThresholdSeconds+5) * time.Second)
	
	// Now it should be stalled
	if !detector.IsStalled() {
		t.Error("Detector should report stalled after threshold exceeded")
	}
	
	// Reset should clear stalled state
	detector.Reset()
	if detector.IsStalled() {
		t.Error("Detector should not report stalled after reset")
	}
}

func TestStallDetector_UpdateActivity(t *testing.T) {
	detector := NewStallDetector()
	
	// Record the initial activity time
	initialTime := detector.lastActivityTime
	
	// Small delay
	time.Sleep(10 * time.Millisecond)
	
	// Update with same item count (should not update time)
	detector.UpdateActivity(0)
	if !detector.lastActivityTime.Equal(initialTime) {
		t.Error("Activity time should not update when item count is unchanged")
	}
	
	// Update with higher item count (should update time)
	detector.UpdateActivity(10)
	if detector.lastActivityTime.Equal(initialTime) {
		t.Error("Activity time should update when item count increases")
	}
	if detector.lastItemCount != 10 {
		t.Errorf("Item count should be updated to 10, got %d", detector.lastItemCount)
	}
}

func TestStallDetector_SetParameters(t *testing.T) {
	detector := NewStallDetector()
	
	// Test setting stall threshold
	detector.SetStallThreshold(60)
	if detector.stallThresholdSeconds != 60 {
		t.Errorf("Stall threshold should be 60, got %d", detector.stallThresholdSeconds)
	}
	
	// Invalid value should not change the setting
	detector.SetStallThreshold(-10)
	if detector.stallThresholdSeconds != 60 {
		t.Errorf("Stall threshold should still be 60 after invalid input, got %d", detector.stallThresholdSeconds)
	}
	
	// Test setting minimum items
	detector.SetMinimumItemsForStallCheck(1000)
	if detector.minimumItemsForStallCheck != 1000 {
		t.Errorf("Minimum items should be 1000, got %d", detector.minimumItemsForStallCheck)
	}
	
	// Test setting grace period
	detector.SetStartupGracePeriod(20)
	if detector.startupGracePeriodSeconds != 20 {
		t.Errorf("Grace period should be 20, got %d", detector.startupGracePeriodSeconds)
	}
}

func TestStallDetector_GracePeriod(t *testing.T) {
	detector := NewStallDetector()
	
	// Set a short grace period for testing
	detector.SetStartupGracePeriod(1)
	
	// Update with enough items to enable stall detection
	detector.UpdateActivity(detector.minimumItemsForStallCheck + 10)
	
	// Set last activity to be beyond stall threshold
	detector.lastActivityTime = time.Now().Add(-time.Duration(detector.stallThresholdSeconds+5) * time.Second)
	
	// Should not be stalled during grace period
	if detector.IsStalled() {
		t.Error("Detector should not report stalled during grace period")
	}
	
	// Wait for grace period to end
	time.Sleep(time.Duration(detector.startupGracePeriodSeconds+1) * time.Second)
	
	// Now it should be stalled
	if !detector.IsStalled() {
		t.Error("Detector should report stalled after grace period ends")
	}
}
