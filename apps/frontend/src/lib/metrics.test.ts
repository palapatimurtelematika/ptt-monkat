import { describe, expect, it } from 'vitest'
import {
  POLL_INTERVAL_MS,
  countLive,
  formatValue,
  freshness,
  groupByDevice,
  relativeAge,
  type LatestMetric,
} from './metrics'

const NOW = Date.parse('2026-09-11T12:00:00Z')
const ago = (ms: number) => new Date(NOW - ms).toISOString()

function row(overrides: Partial<LatestMetric> = {}): LatestMetric {
  return {
    target_id: 1,
    device_name: 'core-rtr-01',
    ip_address: '10.0.0.1',
    oid: '1.3.6.1.2',
    metric_name: 'CPU',
    unit: '%',
    value: 37.5,
    timestamp: ago(1000),
    ...overrides,
  }
}

describe('freshness', () => {
  it('is live inside two poll intervals', () => {
    expect(freshness(ago(0), NOW)).toBe('live')
    expect(freshness(ago(2 * POLL_INTERVAL_MS), NOW)).toBe('live')
  })

  it('is stale past two intervals, up to ten', () => {
    expect(freshness(ago(2 * POLL_INTERVAL_MS + 1), NOW)).toBe('stale')
    expect(freshness(ago(10 * POLL_INTERVAL_MS), NOW)).toBe('stale')
  })

  it('is silent past ten intervals', () => {
    expect(freshness(ago(10 * POLL_INTERVAL_MS + 1), NOW)).toBe('silent')
  })

  it('is silent when never polled, not live', () => {
    // The dangerous failure: a null timestamp reading as healthy.
    expect(freshness(null, NOW)).toBe('silent')
    expect(freshness('not a date', NOW)).toBe('silent')
  })
})

describe('relativeAge', () => {
  it('names the never-polled case in words, not null', () => {
    expect(relativeAge(null, NOW)).toBe('never polled')
  })

  it('scales the unit with the age', () => {
    expect(relativeAge(ago(1000), NOW)).toBe('just now')
    expect(relativeAge(ago(30_000), NOW)).toBe('30s ago')
    expect(relativeAge(ago(5 * 60_000), NOW)).toBe('5m ago')
    expect(relativeAge(ago(3 * 3_600_000), NOW)).toBe('3h ago')
    expect(relativeAge(ago(2 * 86_400_000), NOW)).toBe('2d ago')
  })
})

describe('groupByDevice', () => {
  it('groups without reordering', () => {
    const groups = groupByDevice([
      row({ target_id: 1, metric_name: 'CPU' }),
      row({ target_id: 2, metric_name: 'Temp' }),
      row({ target_id: 3, device_name: 'dwdm-01', ip_address: '10.0.0.5', metric_name: 'RxPower' }),
    ])

    expect(groups.map((g) => g.deviceName)).toEqual(['core-rtr-01', 'dwdm-01'])
    expect(groups[0].metrics.map((m) => m.metric_name)).toEqual(['CPU', 'Temp'])
    expect(groups[1].ipAddress).toBe('10.0.0.5')
  })

  it('returns nothing for no rows', () => {
    expect(groupByDevice([])).toEqual([])
  })
})

describe('countLive', () => {
  it('counts only readings inside the live window', () => {
    const rows = [
      row({ timestamp: ago(1000) }),
      row({ timestamp: ago(20 * POLL_INTERVAL_MS) }),
      row({ timestamp: null, value: null }),
    ]
    expect(countLive(rows, NOW)).toBe(1)
  })
})

describe('formatValue', () => {
  it('keeps whole numbers whole and trims float noise', () => {
    expect(formatValue(42)).toBe('42')
    expect(formatValue(41.19000000000001)).toBe('41.19')
    expect(formatValue(-3.256)).toBe('-3.26')
  })
})
