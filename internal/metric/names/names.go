package names

type Gauge float64
type Counter int64

type GaugeMetric struct {
	Name  string
	Value Gauge
}

type CounterMetric struct {
	Name  string
	Value Counter
}

var (
	Alloc         string = "Alloc"
	BuckHashSys   string = "BuckHashSys"
	Frees         string = "Frees"
	GCCPUFraction string = "GCCPUFraction"
	GCSys         string = "GCSys"
	HeapAlloc     string = "HeapAlloc"
	HeapIdle      string = "HeapIdle"
	HeapInuse     string = "HeapInuse"
	HeapObjects   string = "HeapObjects"
	HeapReleased  string = "HeapReleased"
	HeapSys       string = "HeapSys"
	LastGC        string = "LastGC"
	Lookups       string = "Lookups"
	MCacheInuse   string = "MCacheInuse"
	MCacheSys     string = "MCacheSys"
	MSpanInuse    string = "MSpanInuse"
	MSpanSys      string = "MSpanSys"
	Mallocs       string = "Mallocs"
	NextGC        string = "NextGC"
	NumForcedGC   string = "NumForcedGC"
	NumGC         string = "NumGC"
	OtherSys      string = "OtherSys"
	PauseTotalNs  string = "PauseTotalNs"
	StackInuse    string = "StackInuse"
	StackSys      string = "StackSys"
	Sys           string = "Sys"
	TotalAlloc    string = "TotalAlloc"
	RandomValue   string = "RandomValue"
	PollCounter   string = "PollCounter"
)

var GauageMetrics []string = []string{
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

var CounterMetrics []string = []string{PollCounter}
