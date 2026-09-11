/** One row of GET /api/metrics/latest. Mirrors Go's models.LatestMetric. */
export interface LatestMetric {
  target_id: number
  device_name: string
  ip_address: string
  oid: string
  metric_name: string
  unit: string
  /** null when the target is registered but has never been polled. */
  value: number | null
  timestamp: string | null
}

export interface DeviceGroup {
  deviceName: string
  ipAddress: string
  metrics: LatestMetric[]
}

/** Matches the backend's POLL_INTERVAL default. */
export const POLL_INTERVAL_MS = 60_000

/**
 * How much an operator can trust a reading.
 *
 * 'silent' covers both "never polled" and "not seen in a long time" — from
 * the floor they mean the same thing: this number tells you nothing.
 */
export type Freshness = 'live' | 'stale' | 'silent'

export function freshness(timestamp: string | null, now = Date.now()): Freshness {
  if (!timestamp) return 'silent'

  const age = now - new Date(timestamp).getTime()
  if (Number.isNaN(age)) return 'silent'

  // One missed poll is normal jitter; two is late; ten is gone.
  if (age <= 2 * POLL_INTERVAL_MS) return 'live'
  if (age <= 10 * POLL_INTERVAL_MS) return 'stale'
  return 'silent'
}

const relative = new Intl.RelativeTimeFormat('en', { numeric: 'auto', style: 'narrow' })

/** Human age of a reading. Native Intl — no date library. */
export function relativeAge(timestamp: string | null, now = Date.now()): string {
  if (!timestamp) return 'never polled'

  const seconds = Math.round((now - new Date(timestamp).getTime()) / 1000)
  if (Number.isNaN(seconds)) return 'never polled'
  if (seconds < 5) return 'just now'

  if (seconds < 60) return relative.format(-seconds, 'second')
  if (seconds < 3600) return relative.format(-Math.round(seconds / 60), 'minute')
  if (seconds < 86_400) return relative.format(-Math.round(seconds / 3600), 'hour')
  return relative.format(-Math.round(seconds / 86_400), 'day')
}

/**
 * Collapses the flat API rows into one entry per device.
 * A Map preserves insertion order, and the backend already sorts by
 * device_name then metric_name — so the display order is the query order.
 */
export function groupByDevice(rows: LatestMetric[]): DeviceGroup[] {
  const groups = new Map<string, DeviceGroup>()

  for (const row of rows) {
    const existing = groups.get(row.device_name)
    if (existing) {
      existing.metrics.push(row)
    } else {
      groups.set(row.device_name, {
        deviceName: row.device_name,
        ipAddress: row.ip_address,
        metrics: [row],
      })
    }
  }

  return [...groups.values()]
}

/** Readings currently worth trusting, for the "n of m reporting" counts. */
export function countLive(rows: LatestMetric[], now = Date.now()): number {
  return rows.filter((row) => freshness(row.timestamp, now) === 'live').length
}

/**
 * Formats a value for display. Keeps whole numbers whole and caps the noise
 * on floats — an SNMP gauge scaled by 0.01 otherwise renders 14 decimals.
 */
export function formatValue(value: number): string {
  return Number.isInteger(value) ? value.toString() : value.toFixed(2)
}
