package poller

import (
	"math"
	"testing"

	"github.com/gosnmp/gosnmp"

	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/models"
)

func target(version string) models.SnmpTarget {
	return models.SnmpTarget{ID: 1, IPAddress: "127.0.0.1", SNMPVersion: version}
}

func TestToFloat(t *testing.T) {
	tests := []struct {
		name    string
		pdu     gosnmp.SnmpPDU
		want    float64
		wantErr bool
	}{
		{"integer", gosnmp.SnmpPDU{Type: gosnmp.Integer, Value: 42}, 42, false},
		{"gauge32", gosnmp.SnmpPDU{Type: gosnmp.Gauge32, Value: uint(87)}, 87, false},
		{"timeticks", gosnmp.SnmpPDU{Type: gosnmp.TimeTicks, Value: uint32(123456)}, 123456, false},
		{"counter32", gosnmp.SnmpPDU{Type: gosnmp.Counter32, Value: uint(4294967295)}, 4294967295, false},
		// Above MaxInt64 — must stay positive, not wrap.
		{"counter64 huge", gosnmp.SnmpPDU{Type: gosnmp.Counter64, Value: uint64(math.MaxUint64)}, math.MaxUint64, false},
		{"negative integer", gosnmp.SnmpPDU{Type: gosnmp.Integer, Value: -12}, -12, false},

		{"octet string numeric", gosnmp.SnmpPDU{Type: gosnmp.OctetString, Value: []byte(" -3.25 ")}, -3.25, false},
		{"octet string text", gosnmp.SnmpPDU{Type: gosnmp.OctetString, Value: []byte("Cisco IOS")}, 0, true},

		{"opaque float", gosnmp.SnmpPDU{Type: gosnmp.OpaqueFloat, Value: float32(1.5)}, 1.5, false},
		{"opaque double", gosnmp.SnmpPDU{Type: gosnmp.OpaqueDouble, Value: 2.25}, 2.25, false},

		{"no such object", gosnmp.SnmpPDU{Type: gosnmp.NoSuchObject}, 0, true},
		{"no such instance", gosnmp.SnmpPDU{Type: gosnmp.NoSuchInstance}, 0, true},
		{"null", gosnmp.SnmpPDU{Type: gosnmp.Null}, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := toFloat(tc.pdu)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestSnmpClient(t *testing.T) {
	// v3 must be refused, not silently polled as v2c with a community string.
	if _, err := snmpClient(target("3"), 0, 0); err == nil {
		t.Error("snmpv3 should be rejected: the table has no v3 credentials")
	}
	if _, err := snmpClient(target("nonsense"), 0, 0); err == nil {
		t.Error("unknown version should be rejected")
	}

	c, err := snmpClient(target("2c"), 0, 0)
	if err != nil {
		t.Fatalf("v2c: %v", err)
	}
	// Zero values in the row must not produce a client dialling port 0.
	if c.Port != 161 || c.Community != "public" {
		t.Errorf("want port 161 / public fallback, got %d / %q", c.Port, c.Community)
	}
}
