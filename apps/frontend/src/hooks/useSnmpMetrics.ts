import { useQuery } from '@tanstack/react-query'
import { POLL_INTERVAL_MS, type LatestMetric } from '@/lib/metrics'

/**
 * Native fetch — no Axios anywhere in this project. The signal comes from
 * TanStack so a request still in flight is cancelled when the query is.
 */
async function fetchLatest(signal: AbortSignal): Promise<LatestMetric[]> {
  const response = await fetch('/api/metrics/latest', { signal })

  if (!response.ok) {
    throw new Error(`API responded ${response.status}`)
  }

  return response.json() as Promise<LatestMetric[]>
}

/**
 * Every active target with its newest reading.
 *
 * Refetches on the poller's own cadence: asking more often than the backend
 * collects just returns the same numbers.
 */
export function useSnmpMetrics() {
  return useQuery({
    queryKey: ['snmp', 'metrics', 'latest'],
    queryFn: ({ signal }) => fetchLatest(signal),
    refetchInterval: POLL_INTERVAL_MS,
  })
}
