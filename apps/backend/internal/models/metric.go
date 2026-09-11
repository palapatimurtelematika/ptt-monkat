package models

import "time"

// Influx schema constants — see db/influx_setup.md. Shared by the poller's
// writer and the API's reader so the two cannot drift.
//
// The measurement is the target's metric_name (CPU, Temp, RxPower, ...), so
// there is no fixed measurement constant. target_id is the join key back to
// MariaDB: it is the primary key, so renaming or re-IPing a device does not
// orphan its history.
const (
	FieldValue = "value"

	TagTargetID   = "target_id"
	TagDeviceName = "device_name"
	TagIPAddress  = "ip_address"
	TagUnit       = "unit"
)

// MetricPoint is the generic InfluxDB payload: one float reading for one
// target. Device-type-agnostic by design — a UPS and a DWDM shelf produce
// this shape, differing only by the rows in snmp_targets.
type MetricPoint struct {
	TargetID   uint64
	DeviceName string
	IPAddress  string
	MetricName string
	Unit       string
	Value      float64
	Time       time.Time
}

// LatestMetric is one element of the GET /api/metrics/latest response:
// master data from MariaDB plus the newest point from InfluxDB.
// Value/Timestamp are pointers — null means "registered but never polled",
// which is a state the dashboard has to be able to show.
type LatestMetric struct {
	TargetID   uint64     `json:"target_id"`
	DeviceName string     `json:"device_name"`
	IPAddress  string     `json:"ip_address"`
	OID        string     `json:"oid"`
	MetricName string     `json:"metric_name"`
	Unit       string     `json:"unit"`
	Value      *float64   `json:"value"`
	Timestamp  *time.Time `json:"timestamp"`
}
