import { Card } from '@/components/ui/card'
import {
  formatValue,
  freshness,
  relativeAge,
  type Freshness,
  type LatestMetric,
} from '@/lib/metrics'

/** The state channel: a 2px rule down the left edge carrying the freshness. */
const channel: Record<Freshness, string> = {
  live: 'border-l-state-live',
  stale: 'border-l-state-stale',
  silent: 'border-l-state-silent',
}

const ageTone: Record<Freshness, string> = {
  live: 'text-muted-foreground',
  stale: 'text-state-stale',
  silent: 'text-muted-foreground',
}

export function MetricCard({ metric, now }: { metric: LatestMetric; now: number }) {
  const state = freshness(metric.timestamp, now)

  return (
    <Card
      className={`bg-muted/40 gap-0 rounded-none rounded-r-sm border-0 border-l-[3px] py-3.5 pr-4 pl-3.5 shadow-none ${channel[state]}`}
    >
      <p className="text-foreground/70 text-[0.8125rem] leading-none">{metric.metric_name}</p>

      {/* Both branches occupy the same height so a grid row does not jump
          when a silent target starts reporting. */}
      <div className="mt-2 flex h-8 items-baseline gap-1.5">
        {metric.value === null ? (
          <span className="text-muted-foreground self-center text-sm">waiting for first poll</span>
        ) : (
          <>
            <span className="tabular text-[2rem] leading-8 font-semibold tracking-tight">
              {formatValue(metric.value)}
            </span>
            {metric.unit && <span className="text-muted-foreground text-sm">{metric.unit}</span>}
          </>
        )}
      </div>

      <p className={`tabular mt-2.5 text-xs leading-none ${ageTone[state]}`}>
        {relativeAge(metric.timestamp, now)}
      </p>
    </Card>
  )
}
