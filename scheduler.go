package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type scheduleType int

const (
	scheduleInterval scheduleType = iota
	scheduleDaily
	scheduleWeekly
	scheduleMonthly
)

type schedule struct {
	Type     scheduleType
	Interval time.Duration
	Weekdays []time.Weekday
	AtHour   int
	AtMinute int
}

type Scheduler struct {
	mu      sync.Mutex
	store   *CronStore
	timers  map[string]chan struct{}
	running bool
}

func NewScheduler(store *CronStore) *Scheduler {
	return &Scheduler{
		store:  store,
		timers: make(map[string]chan struct{}),
	}
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	entries := s.store.List()
	for _, entry := range entries {
		if entry.State == "active" {
			s.scheduleEntry(entry)
		}
	}
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = false
	for name, stop := range s.timers {
		close(stop)
		delete(s.timers, name)
	}
}

func (s *Scheduler) Schedule(entry CronEntry) {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	if entry.State == "active" {
		s.scheduleEntry(entry)
	}
}

func (s *Scheduler) Unschedule(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stop, ok := s.timers[name]; ok {
		close(stop)
		delete(s.timers, name)
	}
}

func (s *Scheduler) scheduleEntry(entry CronEntry) {
	sched, err := parseSchedule(entry.Every, entry.At)
	if err != nil {
		fmt.Fprintf(defaultStderr, "failed to parse schedule for %s: %v\n", entry.Name, err)
		return
	}

	stop := make(chan struct{})

	s.mu.Lock()
	// Stop existing timer if any
	if oldStop, ok := s.timers[entry.Name]; ok {
		close(oldStop)
	}
	s.timers[entry.Name] = stop
	s.mu.Unlock()

	go s.runSchedule(entry.Name, entry.Run, sched, stop)
}

func (s *Scheduler) runSchedule(name, command string, sched *schedule, stop chan struct{}) {
	switch sched.Type {
	case scheduleInterval:
		s.runInterval(name, command, sched.Interval, stop)
	case scheduleDaily, scheduleWeekly, scheduleMonthly:
		s.runCalendar(name, command, sched, stop)
	}
}

func (s *Scheduler) runInterval(name, command string, interval time.Duration, stop chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			s.trigger(name, command)
		}
	}
}

func (s *Scheduler) runCalendar(name, command string, sched *schedule, stop chan struct{}) {
	for {
		next := nextOccurrence(sched)
		waitDuration := time.Until(next)
		if waitDuration < 0 {
			waitDuration = 0
		}

		timer := time.NewTimer(waitDuration)
		select {
		case <-stop:
			timer.Stop()
			return
		case <-timer.C:
			s.trigger(name, command)
		}
	}
}

func (s *Scheduler) trigger(name, command string) {
	now := time.Now().UTC().Format(time.RFC3339)

	jobID := ""
	status := "TRIGGERED"

	cmd := exec.Command("aux4", "jobs", "run", command)
	output, err := cmd.Output()
	if err != nil {
		status = "FAILED"
		fmt.Fprintf(defaultStderr, "cron %s: failed to run job: %v\n", name, err)
	} else {
		var result map[string]interface{}
		if jsonErr := json.Unmarshal(output, &result); jsonErr == nil {
			if id, ok := result["id"]; ok {
				jobID = fmt.Sprintf("%v", id)
			}
		}
	}

	entry := HistoryEntry{
		Name:      name,
		JobID:     jobID,
		Timestamp: now,
		Status:    status,
	}

	if histErr := s.store.AddHistory(entry); histErr != nil {
		fmt.Fprintf(defaultStderr, "cron %s: failed to save history: %v\n", name, histErr)
	}
}

func nextOccurrence(sched *schedule) time.Time {
	now := time.Now()
	target := time.Date(now.Year(), now.Month(), now.Day(), sched.AtHour, sched.AtMinute, 0, 0, now.Location())

	switch sched.Type {
	case scheduleDaily:
		if !target.After(now) {
			target = target.AddDate(0, 0, 1)
		}
		return target

	case scheduleWeekly:
		// Find the next matching weekday
		best := time.Time{}
		for _, wd := range sched.Weekdays {
			candidate := target
			daysAhead := int(wd) - int(now.Weekday())
			if daysAhead < 0 {
				daysAhead += 7
			}
			candidate = candidate.AddDate(0, 0, daysAhead)
			if !candidate.After(now) {
				candidate = candidate.AddDate(0, 0, 7)
			}
			if best.IsZero() || candidate.Before(best) {
				best = candidate
			}
		}
		return best

	case scheduleMonthly:
		// First of next month
		target = time.Date(now.Year(), now.Month(), 1, sched.AtHour, sched.AtMinute, 0, 0, now.Location())
		if !target.After(now) {
			target = target.AddDate(0, 1, 0)
		}
		return target
	}

	// fallback: 1 hour from now
	return now.Add(time.Hour)
}

var (
	intervalRegex = regexp.MustCompile(`^(\d+)\s*(s|sec|secs|second|seconds|m|min|mins|minute|minutes|h|hr|hrs|hour|hours|d|day|days)$`)
	monthRegex    = regexp.MustCompile(`^(\d+)\s*(month|months)$`)
)

func parseSchedule(every, at string) (*schedule, error) {
	every = strings.TrimSpace(strings.ToLower(every))

	atHour, atMinute := 0, 0
	if at != "" {
		h, m, err := parseTimeOfDay(at)
		if err != nil {
			return nil, err
		}
		atHour = h
		atMinute = m
	}

	// Check for weekday names
	switch every {
	case "monday":
		return &schedule{Type: scheduleWeekly, Weekdays: []time.Weekday{time.Monday}, AtHour: atHour, AtMinute: atMinute}, nil
	case "tuesday":
		return &schedule{Type: scheduleWeekly, Weekdays: []time.Weekday{time.Tuesday}, AtHour: atHour, AtMinute: atMinute}, nil
	case "wednesday":
		return &schedule{Type: scheduleWeekly, Weekdays: []time.Weekday{time.Wednesday}, AtHour: atHour, AtMinute: atMinute}, nil
	case "thursday":
		return &schedule{Type: scheduleWeekly, Weekdays: []time.Weekday{time.Thursday}, AtHour: atHour, AtMinute: atMinute}, nil
	case "friday":
		return &schedule{Type: scheduleWeekly, Weekdays: []time.Weekday{time.Friday}, AtHour: atHour, AtMinute: atMinute}, nil
	case "saturday":
		return &schedule{Type: scheduleWeekly, Weekdays: []time.Weekday{time.Saturday}, AtHour: atHour, AtMinute: atMinute}, nil
	case "sunday":
		return &schedule{Type: scheduleWeekly, Weekdays: []time.Weekday{time.Sunday}, AtHour: atHour, AtMinute: atMinute}, nil
	case "weekday":
		return &schedule{Type: scheduleWeekly, Weekdays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}, AtHour: atHour, AtMinute: atMinute}, nil
	case "weekend":
		return &schedule{Type: scheduleWeekly, Weekdays: []time.Weekday{time.Saturday, time.Sunday}, AtHour: atHour, AtMinute: atMinute}, nil
	}

	// Check for month
	if matches := monthRegex.FindStringSubmatch(every); matches != nil {
		return &schedule{Type: scheduleMonthly, AtHour: atHour, AtMinute: atMinute}, nil
	}

	// Check for interval patterns
	if matches := intervalRegex.FindStringSubmatch(every); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		unit := matches[2]

		switch unit {
		case "s", "sec", "secs", "second", "seconds":
			return &schedule{Type: scheduleInterval, Interval: time.Duration(n) * time.Second}, nil
		case "m", "min", "mins", "minute", "minutes":
			return &schedule{Type: scheduleInterval, Interval: time.Duration(n) * time.Minute}, nil
		case "h", "hr", "hrs", "hour", "hours":
			return &schedule{Type: scheduleInterval, Interval: time.Duration(n) * time.Hour}, nil
		case "d", "day", "days":
			return &schedule{Type: scheduleDaily, AtHour: atHour, AtMinute: atMinute}, nil
		}
	}

	return nil, fmt.Errorf("invalid schedule expression: %s", every)
}

func parseTimeOfDay(at string) (int, int, error) {
	parts := strings.Split(at, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format: %s (expected HH:MM)", at)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil || h < 0 || h > 23 {
		return 0, 0, fmt.Errorf("invalid hour: %s", parts[0])
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 0 || m > 59 {
		return 0, 0, fmt.Errorf("invalid minute: %s", parts[1])
	}
	return h, m, nil
}
