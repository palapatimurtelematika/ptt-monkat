package repository

import (
	"context"
	"fmt"

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
	bucket string
}

func NewMetricRepository(client influxdb2.Client, org, bucket string) *MetricRepository {
	return &MetricRepository{query: client.QueryAPI(org), bucket: bucket}
}

// LatestAll returns the newest point per series, keyed by models.MetricKey.
//
// One bulk query rather than one per target: last() runs per series, so a
// single round trip covers every device no matter how many rows MariaDB holds.
func (r *MetricRepository) LatestAll(ctx context.Context) (map[string]models.MetricPoint, error) {
	// bucket comes from config, never from a request — no injection path.
	flux := fmt.Sprintf(`
from(bucket: %q)
  |> range(start: %s)
  |> filter(fn: (r) => r._measurement == %q and r._field == %q)
  |> last()`,
		r.bucket, latestLookback, models.MeasurementSNMP, models.FieldValue)

	result, err := r.query.Query(ctx, flux)
	if err != nil {
		return nil, fmt.Errorf("influx query: %w", err)
	}
	defer result.Close()

	points := make(map[string]models.MetricPoint)
	for result.Next() {
		rec := result.Record()

		value, ok := rec.Value().(float64)
		if !ok {
			// A non-float landed in the value field — skip rather than kill
			// the whole dashboard over one bad series.
			continue
		}

		p := models.MetricPoint{
			DeviceName: tag(rec.ValueByKey(models.TagDeviceName)),
			IPAddress:  tag(rec.ValueByKey(models.TagIPAddress)),
			MetricName: tag(rec.ValueByKey(models.TagMetricName)),
			Unit:       tag(rec.ValueByKey(models.TagUnit)),
			Value:      value,
			Time:       rec.Time(),
		}
		points[models.MetricKey(p.IPAddress, p.MetricName)] = p
	}
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("influx stream: %w", err)
	}

	return points, nil
}

func tag(v interface{}) string {
	s, _ := v.(string)
	return s
}
