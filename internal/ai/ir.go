package ai

import (
	"errors"
	"time"

	"github.com/jaeger-ai-assist-prototype/internal"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

type SearchIR struct {
	Service       *string           `json:"service"`
	Operation     *string           `json:"operation"`
	MinDurationMs *string           `json:"min_duration_ms"`
	MaxDurationMs *string           `json:"max_duration_ms"`
	StartTime     *string           `json:"start_time"`
	EndTime       *string           `json:"end_time"`
	Tags          map[string]string `json:"tags"`
}

func MapIRToQueryParams(ir SearchIR) (internal.TraceQueryParams, error) {
	var qp internal.TraceQueryParams

	if ir.Service != nil {
		qp.ServiceName = *ir.Service
	}

	if ir.Operation != nil {
		qp.OperationName = *ir.Operation
	}

	if ir.MinDurationMs != nil {
		d, err := time.ParseDuration(*ir.MinDurationMs)
		if err != nil {
			return qp, err
		}
		qp.DurationMin = d
	}

	if ir.MaxDurationMs != nil {
		d, err := time.ParseDuration(*ir.MaxDurationMs)
		if err != nil {
			return qp, err
		}
		qp.DurationMax = d
	}

	if ir.StartTime != nil {
		t, err := time.Parse(time.RFC3339, *ir.StartTime)
		if err != nil {
			return qp, err
		}
		qp.StartTimeMin = t
	}

	if ir.EndTime != nil {
		t, err := time.Parse(time.RFC3339, *ir.EndTime)
		if err != nil {
			return qp, err
		}
		qp.StartTimeMax = t
	}

	if len(ir.Tags) > 0 {
		qp.Attributes = pcommon.NewMap()
		for k, v := range ir.Tags {
			qp.Attributes.PutStr(k, v)
		}
	}

	return qp, nil
}

func ValidateSearchIR(ir SearchIR) error {
	var minDur, maxDur time.Duration
	var err error

	if ir.MinDurationMs != nil {
		minDur, err = time.ParseDuration(*ir.MinDurationMs)
		if err != nil {
			return errors.New("min_duration_ms must be a valid duration string (e.g. '300ms', '1.5s')")
		}
		if minDur < 0 {
			return errors.New("min_duration_ms cannot be negative")
		}
	}

	if ir.MaxDurationMs != nil {
		maxDur, err = time.ParseDuration(*ir.MaxDurationMs)
		if err != nil {
			return errors.New("max_duration_ms must be a valid duration string (e.g. '300ms', '1.5s')")
		}
		if maxDur < 0 {
			return errors.New("max_duration_ms cannot be negative")
		}
	}

	if ir.MinDurationMs != nil && ir.MaxDurationMs != nil {
		if minDur > maxDur {
			return errors.New("min_duration_ms cannot be greater than max_duration_ms")
		}
	}

	if ir.StartTime != nil {
		if _, err := time.Parse(time.RFC3339, *ir.StartTime); err != nil {
			return errors.New("start_time must be RFC3339")
		}
	}

	if ir.EndTime != nil {
		if _, err := time.Parse(time.RFC3339, *ir.EndTime); err != nil {
			return errors.New("end_time must be RFC3339")
		}
	}

	for k, v := range ir.Tags {
		if k == "" || v == "" {
			return errors.New("tag keys and values must be non-empty")
		}
	}

	return nil
}
