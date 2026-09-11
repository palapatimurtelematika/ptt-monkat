import { useEffect, useState } from 'react'
import { Moon, Sun } from 'lucide-react'
import { MetricCard } from '@/components/MetricCard'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { useSnmpMetrics } from '@/hooks/useSnmpMetrics'
import { useTheme } from '@/hooks/useTheme'
import { countLive, groupByDevice } from '@/lib/metrics'

/**
 * Re-renders the age lines between refetches. Without this "12s ago" would sit
 * frozen on screen for a full minute — worse than showing nothing, because it
 * reads as current.
 */
function useNow(intervalMs = 10_000) {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), intervalMs)
    return () => clearInterval(id)
  }, [intervalMs])

  return now
}

export function Dashboard() {
  const { data, isPending, isError, error, isFetching, refetch } = useSnmpMetrics()
  const { theme, toggle } = useTheme()
  const now = useNow()

  const metrics = data ?? []
  const groups = groupByDevice(metrics)
  const live = countLive(metrics, now)

  return (
    <div className="mx-auto w-full max-w-[1100px] px-5 py-10 sm:px-8">
      <header className="mb-10 flex items-baseline justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">ptt-monkat</h1>
          <p className="text-muted-foreground mt-1 text-sm">SNMP device monitoring</p>
        </div>

        <div className="flex items-center gap-4">
          {metrics.length > 0 && (
            <p
              className={`tabular text-muted-foreground text-sm transition-opacity ${
                isFetching ? 'opacity-50' : 'opacity-100'
              }`}
            >
              {live} of {metrics.length} reporting
            </p>
          )}
          <Button
            variant="ghost"
            size="icon"
            onClick={toggle}
            aria-label={theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'}
          >
            {theme === 'dark' ? <Sun className="size-4" /> : <Moon className="size-4" />}
          </Button>
        </div>
      </header>

      {isPending ? (
        <LoadingGrid />
      ) : isError ? (
        <ApiUnreachable message={error.message} onRetry={() => refetch()} />
      ) : groups.length === 0 ? (
        <NoTargets />
      ) : (
        <div className="space-y-9">
          {groups.map((group) => (
            <section key={group.deviceName}>
              <div className="border-border mb-3 flex items-baseline justify-between gap-4 border-b pb-2">
                <div className="flex items-baseline gap-3">
                  <h2 className="font-medium">{group.deviceName}</h2>
                  <span className="tabular text-muted-foreground text-sm">{group.ipAddress}</span>
                </div>
                <span className="tabular text-muted-foreground text-xs">
                  {countLive(group.metrics, now)} of {group.metrics.length} reporting
                </span>
              </div>

              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {group.metrics.map((metric) => (
                  <MetricCard key={metric.target_id} metric={metric} now={now} />
                ))}
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  )
}

function LoadingGrid() {
  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3" aria-label="Loading readings">
      {Array.from({ length: 6 }, (_, i) => (
        <Card
          key={i}
          className="bg-muted/40 border-l-state-silent h-[100px] animate-pulse rounded-none rounded-r-sm border-0 border-l-[3px] shadow-none"
        />
      ))}
    </div>
  )
}

function ApiUnreachable({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <Card className="border-l-state-fault bg-muted/40 gap-0 rounded-none rounded-r-sm border-0 border-l-[3px] p-5 shadow-none">
      <h2 className="font-medium">Can't reach the monitoring API.</h2>
      <p className="text-muted-foreground mt-1.5 text-sm">
        The Go server on port 3000 may be down. The poller keeps collecting either way, so nothing
        is lost.
      </p>
      <p className="text-muted-foreground mt-3 text-xs">{message}</p>
      <div className="mt-4">
        <Button variant="outline" size="sm" onClick={onRetry}>
          Try again
        </Button>
      </div>
    </Card>
  )
}

function NoTargets() {
  return (
    <Card className="border-l-state-silent bg-muted/40 gap-0 rounded-none rounded-r-sm border-0 border-l-[3px] p-5 shadow-none">
      <h2 className="font-medium">No active targets.</h2>
      <p className="text-muted-foreground mt-1.5 text-sm">
        Add rows to <span className="tabular">snmp_targets</span> with{' '}
        <span className="tabular">is_active = 1</span> and they'll appear here within a minute.
      </p>
    </Card>
  )
}
