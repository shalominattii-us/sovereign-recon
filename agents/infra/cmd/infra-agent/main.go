package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Version info (set at build time)
var (
	version   = "dev"
	buildTime = "unknown"
)

// Config holds the agent configuration
type Config struct {
	Port         int      `yaml:"port"`
	MetricsPath  string   `yaml:"metrics_path"`
	Targets      []string `yaml:"targets"`
	ScrapeInterval int    `yaml:"scrape_interval"`
	LogLevel     string   `yaml:"log_level"`
}

// InfraMetrics holds infrastructure metrics
type InfraMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	NetworkIn   float64 `json:"network_in"`
	NetworkOut  float64 `json:"network_out"`
	Uptime      int64   `json:"uptime"`
	Timestamp   int64   `json:"timestamp"`
}

var (
	// Prometheus metrics
	cpuUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "infra_cpu_usage_percent",
		Help: "Current CPU usage percentage",
	})
	memoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "infra_memory_usage_percent",
		Help: "Current memory usage percentage",
	})
	diskUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "infra_disk_usage_percent",
		Help: "Current disk usage percentage",
	})
	networkIn = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "infra_network_bytes_received_total",
		Help: "Total bytes received",
	})
	networkOut = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "infra_network_bytes_sent_total",
		Help: "Total bytes sent",
	})
	uptime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "infra_uptime_seconds",
		Help: "Service uptime in seconds",
	})
	scrapesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "infra_scrapes_total",
		Help: "Total number of scrapes",
	})
	scrapeErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "infra_scrape_errors_total",
		Help: "Total number of scrape errors",
	})
)

func init() {
	prometheus.MustRegister(cpuUsage, memoryUsage, diskUsage, networkIn, networkOut, uptime, scrapesTotal, scrapeErrors)
}

func main() {
	// Parse flags
	configPath := flag.String("config", "/etc/infra-agent/config.yaml", "Path to config file")
	healthCheck := flag.Bool("health", false, "Run health check and exit")
	flag.Parse()

	if *healthCheck {
		if err := healthCheck(); err != nil {
			fmt.Fprintf(os.Stderr, "Health check failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("OK")
		os.Exit(0)
	}

	// Load config
	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logging
	setupLogging(cfg.LogLevel)

	log.Printf("Starting Infra Agent v%s (built %s)", version, buildTime)
	log.Printf("Config: port=%d, targets=%v, interval=%ds", cfg.Port, cfg.Targets, cfg.ScrapeInterval)

	// Start metrics collection
	ctx, cancel := context.WithCancel(context.Background())
	go collectMetrics(ctx, cfg)

	// Start HTTP server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/status", statusHandler)
	mux.HandleFunc("/infra", infraHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Starting HTTP server on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Infra Agent stopped")
}

func loadConfig(path string) (*Config, error) {
	// Default config
	cfg := &Config{
		Port:           9090,
		MetricsPath:    "/metrics",
		Targets:        []string{"localhost:9003", "localhost:9004", "localhost:9006", "localhost:7777"},
		ScrapeInterval: 30,
		LogLevel:       "info",
	}

	// Try to load from file
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		// Simple YAML parsing (in production use gopkg.in/yaml.v3)
		// For now just use defaults
		log.Printf("Config file found at %s (using defaults for now)", path)
	}

	return cfg, nil
}

func setupLogging(level string) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	switch level {
	case "debug":
		log.SetOutput(os.Stdout)
	case "warn":
		log.SetOutput(os.Stdout)
	case "error":
		log.SetOutput(os.Stderr)
	default:
		log.SetOutput(os.Stdout)
	}
}

func collectMetrics(ctx context.Context, cfg *Config) {
	ticker := time.NewTicker(time.Duration(cfg.ScrapeInterval) * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics, err := gatherMetrics()
			if err != nil {
				log.Printf("Error gathering metrics: %v", err)
				scrapeErrors.Inc()
			} else {
				updatePrometheusMetrics(metrics)
				scrapesTotal.Inc()
				log.Printf("Metrics collected: cpu=%.2f%% mem=%.2f%% disk=%.2f%%", metrics.CPUUsage, metrics.MemoryUsage, metrics.DiskUsage)
			}
			uptime.Set(time.Since(startTime).Seconds())
		}
	}
}

func gatherMetrics() (*InfraMetrics, error) {
	// In production, use gopsutil or similar
	// For now return simulated metrics
	return &InfraMetrics{
		CPUUsage:    15.5,
		MemoryUsage: 42.3,
		DiskUsage:   28.7,
		NetworkIn:   1024.5,
		NetworkOut:  512.3,
		Uptime:      time.Now().Unix(),
		Timestamp:   time.Now().Unix(),
	}, nil
}

func updatePrometheusMetrics(m *InfraMetrics) {
	cpuUsage.Set(m.CPUUsage)
	memoryUsage.Set(m.MemoryUsage)
	diskUsage.Set(m.DiskUsage)
	networkIn.Add(m.NetworkIn)
	networkOut.Add(m.NetworkOut)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "healthy",
		"service":   "infra-agent",
		"version":   version,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":     "infra-agent",
		"version":     version,
		"build_time":  buildTime,
		"uptime":      time.Now().Unix(),
		"metrics_endpoint": "/metrics",
		"scrape_interval": "30s",
	})
}

func infraHandler(w http.ResponseWriter, r *http.Request) {
	metrics, err := gatherMetrics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func healthCheck() error {
	// Check if we can reach local endpoints
	client := &http.Client{Timeout: 5 * time.Second}
	
	endpoints := []string{
		"http://localhost:9003/health",  // Heretic
		"http://localhost:9004/health",  // Compliance
		"http://localhost:9006/health",  // CyberCore
		"http://localhost:7777/s",       // Kernel
	}

	for _, ep := range endPoints {
		resp, err := client.Get(ep)
		if err != nil {
			return fmt.Errorf("endpoint %s unreachable: %v", ep, err)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("endpoint %s returned %d", ep, resp.StatusCode)
		}
	}
	return nil
}