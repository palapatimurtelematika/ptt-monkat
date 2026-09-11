package models

import "time"

// SnmpTarget mirrors db/001_mariadb_init.sql:snmp_targets 1:1.
// One row = one OID polled on one device. The poller is device-agnostic:
// everything it needs to reach the device and interpret the value is here.
type SnmpTarget struct {
	ID         uint64  `db:"id"          json:"id"`
	DeviceName string  `db:"device_name" json:"device_name"`
	IPAddress  string  `db:"ip_address"  json:"ip_address"`
	OID        string  `db:"oid"         json:"oid"`
	MetricName string  `db:"metric_name" json:"metric_name"`
	Unit       string  `db:"unit"        json:"unit"`
	Multiplier float64 `db:"multiplier"  json:"multiplier"`
	IsActive   bool    `db:"is_active"   json:"is_active"`

	SNMPCommunity   string `db:"snmp_community"    json:"snmp_community"`
	SNMPVersion     string `db:"snmp_version"      json:"snmp_version"` // "1" | "2c" | "3"
	SNMPPort        uint16 `db:"snmp_port"         json:"snmp_port"`
	PollIntervalSec uint32 `db:"poll_interval_sec" json:"poll_interval_sec"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
