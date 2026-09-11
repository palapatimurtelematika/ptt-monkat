package repository

import (
	"context"
	"fmt"
	"strconv"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"

	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/models"
)

// latestLookback bounds the Flux range. A device silent for longer than this
// drops out of the response and reads as never-polled.
// ponytail: fixed 24h window, promote to config if a longer memory is needed.
const latestLookback = "-24h"

type MetricRepository struct {
	query  api.QueryAPI
	write  api.WriteAPI // non-blocking: WritePoint buffers and returns
	bucket string
}

func NewMetricRepository(client influxdb2.Client, org, bucket string) *MetricRepository {
	return &MetricRepository{
		query:  client.QueryAPI(org),
		write:  client.WriteAPI(org, bucket),
		bucket: bucket,
	}
}

// LatestAll returns the newest point per series, keyed by target id.
//
// One bulk query rather than one per target: last() runs per series, so a
// single round trip covers every device no matter how many rows MariaDB holds.
// No measurement filter — measurements are metric names and therefore
// open-ended; the value field is what identifies our points.
func (r *MetricRepository) LatestAll(ctx context.Context) (map[uint64]models.MetricPoint, error) {
	// bucket comes from config, never from a request — no injection path.
	flux := fmt.Sprintf(`
from(bucket: %q)
  |> range(start: %s)
  |> filter(fn: (r) => r._field == %q)
  |> last()`,
		r.bucket, latestLookback, models.FieldValue)

	result, err := r.query.Query(ctx, flux)
	if err != nil {
		return nil, fmt.Errorf("influx query: %w", err)
	}
	defer result.Close()

	points := make(map[uint64]models.MetricPoint)
	for result.Next() {
		rec := result.Record()

		targetID, err := strconv.ParseUint(tag(rec.ValueByKey(models.TagTargetID)), 10, 64)
		if err != nil {
			// Not one of ours (or written before the target_id tag existed).
			continue
		}

		value, ok := rec.Value().(float64)
		if !ok {
			// A non-float landed in the value field — skip rather than kill
			// the whole dashboard over one bad series.
			continue
		}

		points[targetID] = models.MetricPoint{
			TargetID:   targetID,
			DeviceName: tag(rec.ValueByKey(models.TagDeviceName)),
			IPAddress:  tag(rec.ValueByKey(models.TagIPAddress)),
			MetricName: rec.Measurement(),
			Unit:       tag(rec.ValueByKey(models.TagUnit)),
			Value:      value,
			Time:       rec.Time(),
		}
	}
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("influx stream: %w", err)
	}

	return points, nil
}

// Write buffers a point. It never blocks and never returns an error —
// failures surface on Errors(), which the caller must drain.
func (r *MetricRepository) Write(p models.MetricPoint) {
	point := influxdb2.NewPointWithMeasurement(p.MetricName).
		AddTag(models.TagTargetID, strconv.FormatUint(p.TargetID, 10)).
		AddTag(models.TagDeviceName, p.DeviceName).
		AddTag(models.TagIPAddress, p.IPAddress).
		AddTag(models.TagUnit, p.Unit).
		AddField(models.FieldValue, p.Value).
		SetTime(p.Time)

	r.write.WritePoint(point)
}

// Errors surfaces async write failures. An undrained channel means writes
// fail silently, so whoever calls Write must consume this.
func (r *MetricRepository) Errors() <-chan error { return r.write.Errors() }

// Flush blocks until the buffer is written. Call on shutdown.
func (r *MetricRepository) Flush() { r.write.Flush() }

func tag(v interface{}) string {
	s, _ := v.(string)
	return s
}
