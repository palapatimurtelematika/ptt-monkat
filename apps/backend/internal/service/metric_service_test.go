package service

import (
	"testing"
	"time"

	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/models"
)

func TestJoinLatest(t *testing.T) {
	polled := time.Date(2026, 9, 11, 6, 30, 0, 0, time.UTC)

	targets := []models.SnmpTarget{
		{ID: 1, DeviceName: "core-rtr-01", IPAddress: "10.0.0.1", OID: "1.3.6.1.2", MetricName: "CPU", Unit: "%"},
		{ID: 2, DeviceName: "core-rtr-01", IPAddress: "10.0.0.1", OID: "1.3.6.1.3", MetricName: "Temp", Unit: "C"},
	}
	points := map[string]models.MetricPoint{
		models.MetricKey("10.0.0.1", "CPU"): {Value: 37.5, Time: polled},
		// stale series for a target that is gone / inactive
		models.MetricKey("10.9.9.9", "CPU"): {Value: 99, Time: polled},
	}

	got := joinLatest(targets, points)

	if len(got) != 2 {
		t.Fatalf("want 2 rows (one per active target), got %d", len(got))
	}

	// target with a point carries value + timestamp
	if got[0].Value == nil || *got[0].Value != 37.5 {
		t.Errorf("CPU value: want 37.5, got %v", got[0].Value)
	}
	if got[0].Timestamp == nil || !got[0].Timestamp.Equal(polled) {
		t.Errorf("CPU timestamp: want %v, got %v", polled, got[0].Timestamp)
	}
	if got[0].OID != "1.3.6.1.2" {
		t.Errorf("OID must come from MariaDB, got %q", got[0].OID)
	}

	// target never polled stays in the list with nulls
	if got[1].Value != nil || got[1].Timestamp != nil {
		t.Errorf("unpolled target: want nil value/timestamp, got %v/%v", got[1].Value, got[1].Timestamp)
	}
	if got[1].MetricName != "Temp" {
		t.Errorf("want Temp row, got %q", got[1].MetricName)
	}
}

func TestJoinLatestAliasing(t *testing.T) {
	// Guards the classic loop-variable/pointer bug: each row must own its float.
	targets := []models.SnmpTarget{
		{ID: 1, IPAddress: "10.0.0.1", MetricName: "CPU"},
		{ID: 2, IPAddress: "10.0.0.2", MetricName: "CPU"},
	}
	points := map[string]models.MetricPoint{
		models.MetricKey("10.0.0.1", "CPU"): {Value: 1},
		models.MetricKey("10.0.0.2", "CPU"): {Value: 2},
	}

	got := joinLatest(targets, points)

	if *got[0].Value == *got[1].Value {
		t.Fatalf("rows share a pointer: both read %v", *got[0].Value)
	}
}
