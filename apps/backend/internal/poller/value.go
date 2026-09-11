package poller

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
)

// toFloat flattens any SNMP variable type to a float64.
//
// This is what makes the poller universal: the database says which OID to
// read, and whatever type the device answers with lands as one float.
func toFloat(pdu gosnmp.SnmpPDU) (float64, error) {
	switch pdu.Type {
	case gosnmp.Null, gosnmp.NoSuchObject, gosnmp.NoSuchInstance, gosnmp.EndOfMibView:
		// Almost always a wrong OID or an unsupported MIB on that device —
		// name it so the log is diagnosable.
		return 0, fmt.Errorf("oid returned %s", pdu.Type)

	case gosnmp.OctetString:
		// Optical and UPS gear routinely reports numbers as strings.
		raw, ok := pdu.Value.([]byte)
		if !ok {
			return 0, fmt.Errorf("octet string carried %T", pdu.Value)
		}
		s := strings.TrimSpace(string(raw))
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, fmt.Errorf("octet string %q is not numeric", s)
		}
		return v, nil

	case gosnmp.OpaqueFloat:
		v, ok := pdu.Value.(float32)
		if !ok {
			return 0, fmt.Errorf("opaque float carried %T", pdu.Value)
		}
		return float64(v), nil

	case gosnmp.OpaqueDouble:
		v, ok := pdu.Value.(float64)
		if !ok {
			return 0, fmt.Errorf("opaque double carried %T", pdu.Value)
		}
		return v, nil

	default:
		// Integer, Counter32, Counter64, Gauge32, TimeTicks, Uinteger32.
		// Via big.Float so a Counter64 above MaxInt64 does not wrap negative.
		n := gosnmp.ToBigInt(pdu.Value)
		if n == nil {
			return 0, fmt.Errorf("unsupported snmp type %s (%T)", pdu.Type, pdu.Value)
		}
		v, _ := new(big.Float).SetInt(n).Float64()
		return v, nil
	}
}
