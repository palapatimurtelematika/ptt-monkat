# InfluxDB v2 — bucket + point schema

Influx v2 is schemaless, so there is no DDL. What has to be agreed up front is
the **point shape**, because stage 2 writes it and stage 3 queries it.

## Setup

```bash
influx setup --org ptt --bucket snmp_metrics --retention 90d \
  --username admin --password '<choose>' --force

influx auth list      # copy the token into apps/backend/.env as INFLUX_TOKEN
```

## Point convention

| part               | value                                            |
| ------------------ | ------------------------------------------------ |
| measurement        | `snmp`                                           |
| tag `device_name`  | `snmp_targets.device_name`                       |
| tag `ip_address`   | `snmp_targets.ip_address`                        |
| tag `metric_name`  | `CPU`, `Temp`, `RxPower`, …                      |
| tag `unit`         | `%`, `C`, `dBm`                                  |
| field `value`      | float, **already multiplied** by `multiplier`    |
| timestamp          | poll time, ns                                    |

One measurement with `metric_name` as a tag — rather than a measurement per
metric — keeps Flux queries identical regardless of device type, which is the
whole point of the system.

The raw OID is deliberately **not** a tag: it lives in MariaDB and would only
inflate series cardinality.

## Sanity check

```bash
influx bucket list --org ptt

influx query 'from(bucket:"snmp_metrics")
  |> range(start: -1h)
  |> filter(fn: (r) => r._measurement == "snmp")'
```
