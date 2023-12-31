package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/carlmjohnson/versioninfo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	overall    [][]byte
	memPtr     *int
	fastPtr    *bool
	maxMem     *int
	listenPort *string

	version = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "version",
		Help: "Version information about this binary",
		ConstLabels: map[string]string{
			"version": versioninfo.Revision,
		},
	})

	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Count of all HTTP requests",
	}, []string{"function"})

	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of all HTTP requests",
	}, []string{"code", "handler", "method"})

	opsProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of simulated processed ops.",
	})

	capName = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "CAP_MYAPP_PROCESSED",
		Help: "The total number of simulated processed ops in CAPS",
	})
)

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			capName.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

func main() {

	memPtr = flag.Int("memory", 50, "how much memory to consume")
	maxMem = flag.Int("maxmemory", 1000, "dont consume more than this")
	fastPtr = flag.Bool("fast", true, "build up memory usage quickly")
	listenPort = flag.String("listen", ":8080", "port to listen on")
	versioninfo.AddFlag(nil)

	flag.Parse()

	r := prometheus.NewRegistry()
	r.MustRegister(httpRequestsTotal)
	r.MustRegister(httpRequestDuration)
	r.MustRegister(version)
	r.MustRegister(opsProcessed)
	r.MustRegister(capName)
	r.MustRegister(collectors.NewGoCollector())

	recordMetrics()

	// enable signal trapping
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c,
			syscall.SIGINT,  // Ctrl+C
			syscall.SIGTERM, // Termination Request
			syscall.SIGSEGV, // FullDerp
			syscall.SIGABRT, // Abnormal termination
			syscall.SIGILL,  // illegal instruction
			syscall.SIGFPE)  // floating point
		sig := <-c
		log.Println("-----------------------------------------")
		log.Printf("Signal (%v) Detected, Shutting Down\n", sig)
		log.Println("Final Memory usage when killed:")
		log.Println(ReturnMemUsage())
		os.Exit(1)
	}()

	http.HandleFunc("/", HelloServer)
	http.HandleFunc("/consumemem", ConsumeMemory)
	http.HandleFunc("/clearmem", ClearMemory)
	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	log.Printf("Starting k8s_memuser_advanced revision: %v", versioninfo.Short())
	log.Printf("Listening on port: %v", *listenPort)
	log.Fatal(http.ListenAndServe(*listenPort, nil))

}

// ReturnMemUsage returns a string showing the current memory usage.
func ReturnMemUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return fmt.Sprintf("Alloc = %v MiB\tSys = %v MiB\tNumGC = %v\n", bToMb(m.Alloc), bToMb(m.Sys), m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// AllocateMemory will allocate memory on the shared byte array
func AllocateMemory() {

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	if bToMb(m.Alloc) < uint64(*maxMem) {
		for i := 0; i < *memPtr; i++ {

			// Allocate memory using make() and append to overall (so it doesn't get
			// garbage collected). This is to create an ever increasing memory usage
			// which we can track. We're just using []int as an example.
			a := make([]byte, 1048576)
			_, err := rand.Read(a)
			if err != nil {
				fmt.Println("Error getting random data.")
				return
			}
			//for j := 0; j < 1024; j++ {
			overall = append(overall, a)
			//}

			if !*fastPtr {
				time.Sleep(time.Second)
			}
		}
	}
	fmt.Println(ReturnMemUsage())
}

// HelloServer will return a hello world, and will consume some more memory in the process
func HelloServer(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("Hello User. My current memory usage is:\n %v", ReturnMemUsage())
	httpRequestsTotal.With(prometheus.Labels{"function": "HelloServer"}).Inc()
	fmt.Fprint(w, message)
}

// ConsumeMemory will return a hello world, and will consume some more memory in the process
func ConsumeMemory(w http.ResponseWriter, r *http.Request) {
	AllocateMemory()
	message := fmt.Sprintf("Hello User. My current memory usage is:\n %v", ReturnMemUsage())
	httpRequestsTotal.With(prometheus.Labels{"function": "ConsumeMemory"}).Inc()
	fmt.Fprint(w, message)

}

// ClearMemory will return a hello world, and will consume some more memory in the process
func ClearMemory(w http.ResponseWriter, r *http.Request) {
	overall = nil
	runtime.GC()
	debug.FreeOSMemory()
	message := fmt.Sprintf("Memory has been cleared.\n %v", ReturnMemUsage())
	httpRequestsTotal.With(prometheus.Labels{"function": "ClearMemory"}).Inc()
	fmt.Fprint(w, message)

}
