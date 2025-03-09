package system

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/edgetainer/edgetainer/internal/shared/logging"
)

// SystemMetrics represents various system metrics
type SystemMetrics struct {
	CPUUsage    float64            `json:"cpu_usage"`    // percentage
	MemoryUsage float64            `json:"memory_usage"` // percentage
	MemoryTotal int64              `json:"memory_total"` // bytes
	MemoryFree  int64              `json:"memory_free"`  // bytes
	DiskUsage   map[string]float64 `json:"disk_usage"`   // percentage by mount point
	DiskTotal   map[string]int64   `json:"disk_total"`   // bytes by mount point
	DiskFree    map[string]int64   `json:"disk_free"`    // bytes by mount point
	Uptime      int64              `json:"uptime"`       // seconds
	LoadAvg     [3]float64         `json:"load_avg"`     // 1, 5, 15 min load averages
	Timestamp   time.Time          `json:"timestamp"`
}

// Monitor collects system metrics and reports them
type Monitor struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	interval   time.Duration
	logger     *logging.Logger
	metrics    *SystemMetrics
	done       chan struct{}
}

// NewMonitor creates a new system monitor
func NewMonitor(ctx context.Context) (*Monitor, error) {
	monitorCtx, cancel := context.WithCancel(ctx)

	return &Monitor{
		ctx:        monitorCtx,
		cancelFunc: cancel,
		interval:   30 * time.Second, // Default to 30s
		logger:     logging.WithComponent("system-monitor"),
		metrics:    &SystemMetrics{},
		done:       make(chan struct{}),
	}, nil
}

// Start begins the monitoring process
func (m *Monitor) Start() {
	m.logger.Info("System monitor starting")

	// Do an initial collection
	m.collectMetrics()

	// Start the collection loop
	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()
		defer close(m.done)

		for {
			select {
			case <-ticker.C:
				m.collectMetrics()
			case <-m.ctx.Done():
				m.logger.Info("System monitor stopping")
				return
			}
		}
	}()
}

// Stop halts the monitoring process
func (m *Monitor) Stop() {
	m.cancelFunc()
	<-m.done
}

// GetMetrics returns the current system metrics
func (m *Monitor) GetMetrics() *SystemMetrics {
	// Return a copy to avoid race conditions
	metrics := *m.metrics
	return &metrics
}

// collectMetrics gathers system information
func (m *Monitor) collectMetrics() {
	metrics := &SystemMetrics{
		DiskUsage: make(map[string]float64),
		DiskTotal: make(map[string]int64),
		DiskFree:  make(map[string]int64),
		Timestamp: time.Now(),
	}

	// Collection methods depend on the OS
	var err error
	switch runtime.GOOS {
	case "linux":
		err = m.collectLinuxMetrics(metrics)
	case "darwin":
		err = m.collectDarwinMetrics(metrics)
	default:
		err = fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err != nil {
		m.logger.Error(fmt.Sprintf("Failed to collect metrics: %v", err), err)
		return
	}

	// Update the metrics
	m.metrics = metrics

	m.logger.Debug(fmt.Sprintf("Collected system metrics: CPU: %.1f%%, Mem: %.1f%%",
		metrics.CPUUsage, metrics.MemoryUsage))
}

// collectLinuxMetrics gathers system metrics on Linux
func (m *Monitor) collectLinuxMetrics(metrics *SystemMetrics) error {
	// Simplified implementation - in a real agent, use proper Linux stats APIs
	// or libraries like github.com/shirou/gopsutil

	// Simulate CPU usage collection
	cmd := exec.Command("bash", "-c", "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1}'")
	output, err := cmd.Output()
	if err == nil {
		fmt.Sscanf(strings.TrimSpace(string(output)), "%f", &metrics.CPUUsage)
	}

	// Simulate memory usage collection
	cmd = exec.Command("bash", "-c", "free | grep Mem | awk '{print $3/$2 * 100.0, $2, $4}'")
	output, err = cmd.Output()
	if err == nil {
		fmt.Sscanf(strings.TrimSpace(string(output)), "%f %d %d",
			&metrics.MemoryUsage, &metrics.MemoryTotal, &metrics.MemoryFree)
	}

	// Simulate disk usage collection
	cmd = exec.Command("bash", "-c", "df -P | grep -v Filesystem")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}

			var device, mountpoint string
			var total, used, available int64
			var usePct float64

			fmt.Sscanf(line, "%s %d %d %d %f%% %s",
				&device, &total, &used, &available, &usePct, &mountpoint)

			metrics.DiskUsage[mountpoint] = usePct
			metrics.DiskTotal[mountpoint] = total * 1024 // df reports in KB
			metrics.DiskFree[mountpoint] = available * 1024
		}
	}

	// Simulate uptime collection
	cmd = exec.Command("bash", "-c", "cat /proc/uptime | awk '{print $1}'")
	output, err = cmd.Output()
	if err == nil {
		var uptime float64
		fmt.Sscanf(strings.TrimSpace(string(output)), "%f", &uptime)
		metrics.Uptime = int64(uptime)
	}

	// Simulate load average collection
	cmd = exec.Command("bash", "-c", "cat /proc/loadavg | awk '{print $1, $2, $3}'")
	output, err = cmd.Output()
	if err == nil {
		fmt.Sscanf(strings.TrimSpace(string(output)), "%f %f %f",
			&metrics.LoadAvg[0], &metrics.LoadAvg[1], &metrics.LoadAvg[2])
	}

	return nil
}

// collectDarwinMetrics gathers system metrics on macOS
func (m *Monitor) collectDarwinMetrics(metrics *SystemMetrics) error {
	// Simplified implementation - in a real agent, use proper macOS stats APIs
	// or libraries like github.com/shirou/gopsutil

	// Set dummy values for testing
	metrics.CPUUsage = 25.0
	metrics.MemoryUsage = 50.0
	metrics.MemoryTotal = 16 * 1024 * 1024 * 1024 // 16GB
	metrics.MemoryFree = 8 * 1024 * 1024 * 1024   // 8GB
	metrics.DiskUsage["/"] = 45.0
	metrics.DiskTotal["/"] = 500 * 1024 * 1024 * 1024 // 500GB
	metrics.DiskFree["/"] = 275 * 1024 * 1024 * 1024  // 275GB
	metrics.Uptime = 3600 * 24 * 2                    // 2 days
	metrics.LoadAvg = [3]float64{1.5, 1.2, 0.9}

	return nil
}

// GetOSInfo returns information about the operating system
func GetOSInfo() (map[string]string, error) {
	info := make(map[string]string)

	// Get hostname
	cmd := exec.Command("hostname")
	output, err := cmd.Output()
	if err == nil {
		info["hostname"] = strings.TrimSpace(string(output))
	}

	// OS specific information
	switch runtime.GOOS {
	case "linux":
		// Get OS version (e.g., Ubuntu 20.04)
		cmd = exec.Command("bash", "-c", "cat /etc/os-release | grep PRETTY_NAME | cut -d '\"' -f 2")
		output, err = cmd.Output()
		if err == nil {
			info["os_version"] = strings.TrimSpace(string(output))
		}

		// Get kernel version
		cmd = exec.Command("uname", "-r")
		output, err = cmd.Output()
		if err == nil {
			info["kernel_version"] = strings.TrimSpace(string(output))
		}

	case "darwin":
		// Get macOS version
		cmd = exec.Command("sw_vers", "-productVersion")
		output, err = cmd.Output()
		if err == nil {
			info["os_version"] = "macOS " + strings.TrimSpace(string(output))
		}

		// Get kernel version
		cmd = exec.Command("uname", "-r")
		output, err = cmd.Output()
		if err == nil {
			info["kernel_version"] = strings.TrimSpace(string(output))
		}
	}

	info["architecture"] = runtime.GOARCH
	info["os"] = runtime.GOOS
	info["cpu_count"] = fmt.Sprintf("%d", runtime.NumCPU())

	return info, nil
}
