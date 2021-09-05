package main

import (
	"bytes"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var (
	listenAddress = flag.String("web.listen-address", ":9860", "Address to listen on for web interface.")
	metricPath    = flag.String("web.metrics-path", "/metrics", "Path under which to expose metrics.")
	commandsPath  = flag.String("web.commands-path", "/commands", "Path under which to expose metrics.")
)

func main() {
	log.Fatal(serverMetrics(*listenAddress, *metricPath))
}

func serverMetrics(listenAddress, metricsPath string) error {
	http.Handle(metricsPath, promhttp.Handler())
	http.HandleFunc("/commands", commandsHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
            <html>
            <head><title>Command Exporter Metrics</title></head>
            <body>
            <h1>ConfigMap Reload</h1>
            <p><a href='` + metricsPath + `'>Metrics</a></p>
            </body>
            </html>
        `))
	})
	return http.ListenAndServe(listenAddress, nil)
}

func commandsHandler(w http.ResponseWriter, r *http.Request) {

	commandName := r.URL.Query().Get("target")
	var commandArgs []string
	var stdout, stderr string
	var exit int
	if commandName == "" {
		//TODO:
		http.Error(w, "Command name is missing", http.StatusBadRequest)
	}

	if strings.Index(commandName, " ") != -1 {
		split := strings.Split(commandName, " ")
		commandName = split[0]
		commandArgs = split[1:]
	}

	start := time.Now()

	registry := prometheus.NewRegistry()

	stdout, stderr, exit = RunCommand(commandName, commandArgs...)

	stdout = trimStrings(stdout) + "..."
	stderr = trimStrings(stderr) + "..."

	commandErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem:   commandName + "_" + strings.Join(commandArgs, "_"),
		Name:        "command_errors_total",
		ConstLabels: prometheus.Labels{"command": commandName, "args": strings.Join(commandArgs, " "), "stdout": stdout, "stderr": stderr},
		Help:        "number of total errors on commands.",
	})
	commandDuration := prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem:   commandName + "_" + strings.Join(commandArgs, "_"),
		Name:        "command_processing_duration_seconds",
		Help:        "number of total duration of commands running",
		ConstLabels: prometheus.Labels{"command": commandName, "args": strings.Join(commandArgs, " "), "stdout": stdout, "stderr": stderr},
	})
	commandStatus := prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem:   commandName + "_" + strings.Join(commandArgs, "_"),
		Name:        "command_exit_with",
		Help:        "return of the command's exit code.",
		ConstLabels: prometheus.Labels{"command": commandName, "args": strings.Join(commandArgs, " "), "stdout": stdout, "stderr": stderr},
	})
	commandGlobalStatus := prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem:   "command",
		Name:        "exit_with",
		Help:        "return of the command's exit code..",
		ConstLabels: prometheus.Labels{"command": commandName, "args": strings.Join(commandArgs, " "), "stdout": stdout, "stderr": stderr},
	})

	registry.MustRegister(commandStatus, commandDuration, commandErrors, commandGlobalStatus)

	duration := time.Since(start).Seconds()

	commandStatus.Set(float64(exit))
	commandGlobalStatus.Set(float64(exit))
	commandDuration.Set(duration)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func RunCommand(name string, args ...string) (string, string, int) {

	var stdout string
	var stderr string
	var exitcode int

	var outbuf, errbuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout = outbuf.String()
	stderr = errbuf.String()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitcode = ws.ExitStatus()
		} else {
			exitcode = 1
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitcode = ws.ExitStatus()
	}

	return stdout, stderr, exitcode
}

func trimStrings(string2 string) string {
	var trimWord = []string{
		"\"",
		"\n",
		"\r",
		"\r\n",
	}
	for _, i := range trimWord {
		string2 = strings.Replace(string2, i, "", -1)
	}
	if len(string2) > 50 {
		string2 = string2[:50]
	}
	return string2
}
