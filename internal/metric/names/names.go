package names

type Gauge float64
type Counter int64

type GaugeMetric struct {
	name  string
	value Gauge
}

func (gm GaugeMetric) Name() string {
	return gm.name
}

func (gm GaugeMetric) Value() Gauge {
	return gm.value
}

func NewGaugeMetric(name string, value Gauge) *GaugeMetric {
	m := &GaugeMetric{name: name, value: value}
	return m
}

type CounterMetric struct {
	name  string
	value Counter
}

func (cm CounterMetric) Name() string {
	return cm.name
}

func (cm CounterMetric) Value() Counter {
	return cm.value
}

func NewCounterMetric(name string, value Counter) *CounterMetric {
	m := &CounterMetric{name: name, value: value}
	return m
}

var (
	Alloc         = "Alloc"
	BuckHashSys   = "BuckHashSys"
	Frees         = "Frees"
	GCCPUFraction = "GCCPUFraction"
	GCSys         = "GCSys"
	HeapAlloc     = "HeapAlloc"
	HeapIdle      = "HeapIdle"
	HeapInuse     = "HeapInuse"
	HeapObjects   = "HeapObjects"
	HeapReleased  = "HeapReleased"
	HeapSys       = "HeapSys"
	LastGC        = "LastGC"
	Lookups       = "Lookups"
	MCacheInuse   = "MCacheInuse"
	MCacheSys     = "MCacheSys"
	MSpanInuse    = "MSpanInuse"
	MSpanSys      = "MSpanSys"
	Mallocs       = "Mallocs"
	NextGC        = "NextGC"
	NumForcedGC   = "NumForcedGC"
	NumGC         = "NumGC"
	OtherSys      = "OtherSys"
	PauseTotalNs  = "PauseTotalNs"
	StackInuse    = "StackInuse"
	StackSys      = "StackSys"
	Sys           = "Sys"
	TotalAlloc    = "TotalAlloc"
	RandomValue   = "RandomValue"
	PollCounter   = "PollCounter"
)

var GaugeMetrics = []string{
	Alloc,
	BuckHashSys,
	Frees,
	GCCPUFraction,
	GCSys,
	HeapAlloc,
	HeapIdle,
	HeapInuse,
	HeapObjects,
	HeapReleased,
	HeapSys,
	LastGC,
	Lookups,
	MCacheInuse,
	MCacheSys,
	MSpanInuse,
	MSpanSys,
	Mallocs,
	NextGC,
	NumForcedGC,
	NumGC,
	OtherSys,
	PauseTotalNs,
	StackInuse,
	StackSys,
	Sys,
	TotalAlloc,
	RandomValue,
}
