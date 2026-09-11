package models

import "time"

// SnmpTarget mirrors db/001_mariadb_init.sql:snmp_targets 1:1.
// One row = one OID polled on one device. The poller is device-agnostic:
// everything it needs to reach the device and interpret the value is here.
//
// Columns are tagged explicitly rather than left to GORM's naming strategy —
// "SNMP" and "OID" are not in its initialism list and would be guessed wrong.
type SnmpTarget struct {
	ID         uint64  `gorm:"column:id;primaryKey"  json:"id"`
	DeviceName string  `gorm:"column:device_name"    json:"device_name"`
	IPAddress  string  `gorm:"column:ip_address"     json:"ip_address"`
	OID        string  `gorm:"column:oid"            json:"oid"`
	MetricName string  `gorm:"column:metric_name"    json:"metric_name"`
	Unit       string  `gorm:"column:unit"           json:"unit"`
	Multiplier float64 `gorm:"column:multiplier"     json:"multiplier"`
	IsActive   bool    `gorm:"column:is_active"      json:"is_active"`

	SNMPCommunity   string `gorm:"column:snmp_community"    json:"snmp_community"`
	SNMPVersion     string `gorm:"column:snmp_version"      json:"snmp_version"` // "1" | "2c" | "3"
	SNMPPort        uint16 `gorm:"column:snmp_port"         json:"snmp_port"`
	PollIntervalSec uint32 `gorm:"column:poll_interval_sec" json:"poll_interval_sec"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName pins the table; the schema is owned by db/001_mariadb_init.sql,
// never by AutoMigrate.
func (SnmpTarget) TableName() string { return "snmp_targets" }
