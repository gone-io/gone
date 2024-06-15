package tracer

import "github.com/gone-io/gone"

// Tracer is used to assign a unified traceId to the same call link to facilitate log tracking
// Deprecated use gone.Tracer instead
type Tracer = gone.Tracer
