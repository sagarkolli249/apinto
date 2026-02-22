package version

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

// These should be set via go build -ldflags -X 'xxxx'.
var Version = "unknown"
var goVersion = "unknown"
var gitCommit = "unknown"
var buildTime = "unknown"
var buildUser = "unknown"
var eoscVersion = "unknown"

var profileInfo []byte

// startTime records when the process started — used by LogStartupBanner and
// available for health endpoints that want to report uptime.
var startTime = time.Now()

func init() {
	buffer := &bytes.Buffer{}
	fmt.Fprintf(buffer, "Apinto version: %s\n", Version)
	fmt.Fprintf(buffer, "Golang version: %s\n", goVersion)
	fmt.Fprintf(buffer, "Git commit hash: %s\n", gitCommit)
	fmt.Fprintf(buffer, "Built on: %s\n", buildTime)
	fmt.Fprintf(buffer, "Built by: %s\n", buildUser)
	fmt.Fprintf(buffer, "Built by eosc version: %s\n", eoscVersion)
	profileInfo = buffer.Bytes()
}

func Build() *cli.Command {
	return &cli.Command{
		Name: "version",
		Action: func(context *cli.Context) error {
			fmt.Print(string(profileInfo))
			return nil
		},
	}
}

// LogStartupBanner prints a structured startup banner to stderr so it appears
// in kubectl logs immediately on boot.  This makes it trivial to verify which
// image / commit is actually running in production — something that would have
// saved hours during the /var/log/apipark deadlock investigation.
//
// Call this from main() BEFORE process.Run().
func LogStartupBanner() {
	hostname, _ := os.Hostname()
	pid := os.Getpid()

	// Short git SHA for readability (full SHA still in `apinto version`)
	shortCommit := gitCommit
	if len(shortCommit) > 8 {
		shortCommit = shortCommit[:8]
	}

	banner := fmt.Sprintf(
		`================================================
  Apinto AI Gateway — Starting
------------------------------------------------
  Version     : %s
  Git Commit  : %s
  Build Time  : %s
  Go Runtime  : %s (compiled: %s)
  EOSC Version: %s
  Hostname    : %s
  PID         : %d
  GOMAXPROCS  : %d
  Start Time  : %s
================================================`,
		Version,
		shortCommit,
		buildTime,
		runtime.Version(), goVersion,
		eoscVersion,
		hostname,
		pid,
		runtime.GOMAXPROCS(0),
		startTime.UTC().Format(time.RFC3339),
	)

	// Print to stderr directly — this is intentional.
	// Using log.Info() here would be risky because the eosc logger may not be
	// initialised yet (and if the file logger is broken, we'd deadlock before
	// the banner even prints — the exact bug we just fixed).
	fmt.Fprintln(os.Stderr, banner)

	// Also log key env vars that affect behaviour (sanitised — no secrets).
	envHints := []string{"APINTO_DEBUG", "APINTO_LOG_LEVEL", "POD_NAME", "POD_NAMESPACE"}
	var envLines []string
	for _, key := range envHints {
		if v := os.Getenv(key); v != "" {
			envLines = append(envLines, fmt.Sprintf("  %s=%s", key, v))
		}
	}
	if len(envLines) > 0 {
		fmt.Fprintf(os.Stderr, "  Env: %s\n", strings.Join(envLines, ", "))
	}
}

// Uptime returns how long the process has been running.  Useful for health
// check endpoints or diagnostic APIs.
func Uptime() time.Duration {
	return time.Since(startTime)
}
