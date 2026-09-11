# InfluxDB v2 — bucket + point schema

Influx v2 is schemaless, so there is no DDL. What has to be agreed up front is
the **point shape**, because the poller writes it and the API queries it.

## Setup

The dev instance comes from `docker-compose.yml` (host port **8087** — 8086 is
already taken on this machine) and self-initialises the org, bucket and token.

For a standalone instance:

```bash
influx setup --org ptt --bucket snmp_metrics --retention 90d \
  --username admin --password '<choose>' --force

influx auth list      # copy the token into apps/backend/.env as INFLUX_TOKEN
```

## Point convention

| part               | value                                                    |
| ------------------ | -------------------------------------------------------- |
| measurement        | `snmp_targets.metric_name` — `CPU`, `Temp`, `RxPower`, …  |
| tag `target_id`    | `snmp_targets.id`, as a string — **the join key**         |
| tag `device_name`  | `snmp_targets.device_name`                                |
| tag `ip_address`   | `snmp_targets.ip_address`                                 |
| tag `unit`         | `%`, `C`, `dBm`                                           |
| field `value`      | float, **already multiplied** by `multiplier`             |
| timestamp          | poll time                                                 |

Written by `internal/poller` via `repository.MetricRepository.Write`, read back
by `LatestAll`. The tag names are constants in `internal/models/metric.go` so
the two sides cannot drift.

`target_id` is the join key rather than `(ip_address, metric_name)` because it
is the MariaDB primary key: renaming a device or re-IPing it keeps its history
attached.

The raw OID is deliberately **not** a tag — it lives in MariaDB and would only
inflate series cardinality.

## Sanity check

```bash
influx bucket list --org ptt

# every series' newest point
influx query 'from(bucket:"snmp_metrics")
  |> range(start: -1h)
  |> filter(fn: (r) => r._field == "value")
  |> last()'
```
