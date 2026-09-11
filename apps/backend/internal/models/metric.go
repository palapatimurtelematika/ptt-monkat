package models

import "time"

// Influx schema constants — see db/influx_setup.md. Shared by the reader here
// and by the stage 3 writer so the two cannot drift.
const (
	MeasurementSNMP = "snmp"
	FieldValue      = "value"

	TagDeviceName = "device_name"
	TagIPAddress  = "ip_address"
	TagMetricName = "metric_name"
	TagUnit       = "unit"
)

// MetricPoint is the generic InfluxDB payload: measurement "snmp", tags
// identifying the device and metric, one float field "value".
// Device-type-agnostic by design — a UPS and a DWDM shelf produce this shape.
type MetricPoint struct {
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

// MetricKey joins the two stores. (ip_address, metric_name) is the only pair
// present on both sides — the OID lives in MariaDB alone.
func MetricKey(ip, metric string) string { return ip + "|" + metric }
