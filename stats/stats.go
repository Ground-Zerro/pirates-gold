package stats

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Stats struct {
	Checked atomic.Int64
	Used    atomic.Int64
	Found   atomic.Int64
	start   time.Time
	file    *os.File
	stop    chan struct{}
	done    chan struct{}
}

type Total struct {
	Sessions int
	Checked  int64
	Used     int64
	Found    int64
	Seconds  int64
}

func New(dir string) (*Stats, error) {
	f, err := os.OpenFile(filepath.Join(dir, "stats.txt"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	s := &Stats{
		start: time.Now(),
		file:  f,
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}

	fmt.Fprintf(f, "\n=== Session %s ===\n", s.start.Format("02.01.2006 15:04"))

	go s.loop()
	return s, nil
}

func (s *Stats) loop() {
	defer close(s.done)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.writeLine()
		case <-s.stop:
			return
		}
	}
}

func (s *Stats) writeLine() {
	elapsed := time.Since(s.start).Round(time.Second)
	h := int(elapsed.Hours())
	m := int(elapsed.Minutes()) % 60
	sec := int(elapsed.Seconds()) % 60
	fmt.Fprintf(s.file, "[%s] Time: %dh %02dm %02ds | Checked: %d | Used: %d | Found: %d\n",
		time.Now().Format("15:04"),
		h, m, sec,
		s.Checked.Load(),
		s.Used.Load(),
		s.Found.Load(),
	)
}

func (s *Stats) Close() {
	close(s.stop)
	<-s.done
	s.writeLine()
	fmt.Fprintf(s.file, "=== End of session %s ===\n",
		time.Now().Format("02.01.2006 15:04"))
	s.file.Close()
}

func ParseTotal(dir string) (Total, error) {
	data, err := os.ReadFile(filepath.Join(dir, "stats.txt"))
	if err != nil {
		return Total{}, err
	}

	var total Total
	var lastLine string
	inSession := false

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "=== Session"):
			inSession = true
			total.Sessions++
			lastLine = ""
		case inSession && strings.HasPrefix(line, "[") && strings.Contains(line, "Checked:"):
			lastLine = line
		case strings.HasPrefix(line, "=== End of session"):
			if lastLine != "" {
				accumulateLine(lastLine, &total)
				lastLine = ""
			}
			inSession = false
		}
	}
	if inSession && lastLine != "" {
		accumulateLine(lastLine, &total)
	}

	return total, nil
}

func accumulateLine(line string, t *Total) {
	parts := strings.Split(line, "|")
	if len(parts) < 4 {
		return
	}
	t.Checked += parseValue(parts[1], "Checked:")
	t.Used += parseValue(parts[2], "Used:")
	t.Found += parseValue(parts[3], "Found:")
	t.Seconds += parseRuntime(parts[0])
}

func parseValue(s, prefix string) int64 {
	idx := strings.Index(s, prefix)
	if idx < 0 {
		return 0
	}
	v, _ := strconv.ParseInt(strings.TrimSpace(s[idx+len(prefix):]), 10, 64)
	return v
}

func parseRuntime(s string) int64 {
	idx := strings.Index(s, "Time:")
	if idx < 0 {
		return 0
	}
	s = strings.TrimSpace(s[idx+len("Time:"):])
	var h, m, sec int64
	fmt.Sscanf(s, "%dh %dm %ds", &h, &m, &sec)
	return h*3600 + m*60 + sec
}
