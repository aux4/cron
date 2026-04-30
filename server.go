package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var defaultStderr io.Writer = os.Stderr

func startServer(args []string) {
	port := getArg(args, 0, "8421")
	dir := getArg(args, 1, ".")

	// If already running, exit successfully
	pidFile := pidFilePath(port)
	if pidData, err := os.ReadFile(pidFile); err == nil {
		pid, err := strconv.Atoi(strings.TrimSpace(string(pidData)))
		if err == nil {
			if proc, err := os.FindProcess(pid); err == nil {
				if proc.Signal(syscall.Signal(0)) == nil {
					fmt.Fprintf(os.Stdout, "{\"status\":\"ALREADY_RUNNING\",\"pid\":%d,\"port\":\"%s\"}\n", pid, port)
					return
				}
			}
		}
		os.Remove(pidFile)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid directory: %v\n", err)
		os.Exit(1)
	}

	store := NewCronStore(absDir)
	if err := store.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load cron entries: %v\n", err)
		os.Exit(1)
	}
	if err := store.LoadHistory(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load history: %v\n", err)
		os.Exit(1)
	}

	scheduler := NewScheduler(store)

	mux := http.NewServeMux()

	mux.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			httpError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		name := r.URL.Query().Get("name")
		every := r.URL.Query().Get("every")
		at := r.URL.Query().Get("at")
		in := r.URL.Query().Get("in")
		maxStr := r.URL.Query().Get("max")
		run := r.URL.Query().Get("run")

		if name == "" {
			httpError(w, http.StatusBadRequest, "name is required")
			return
		}
		if every == "" && in == "" && at == "" {
			httpError(w, http.StatusBadRequest, "every, in, or at is required")
			return
		}
		if run == "" {
			httpError(w, http.StatusBadRequest, "run is required")
			return
		}

		// Validate schedule expression
		if in != "" {
			if _, err := parseIn(in); err != nil {
				httpError(w, http.StatusBadRequest, err.Error())
				return
			}
		} else if every == "" && at != "" {
			if _, err := parseAt(at); err != nil {
				httpError(w, http.StatusBadRequest, err.Error())
				return
			}
		} else {
			if _, err := parseSchedule(every, at); err != nil {
				httpError(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		max := 0
		if maxStr != "" {
			n, err := strconv.Atoi(maxStr)
			if err != nil || n < 1 {
				httpError(w, http.StatusBadRequest, "max must be a positive integer")
				return
			}
			max = n
		}

		entry := CronEntry{
			Name:  name,
			Every: every,
			At:    at,
			In:    in,
			Max:   max,
			Run:   run,
			State: "active",
		}

		if err := store.Add(entry); err != nil {
			httpError(w, http.StatusConflict, err.Error())
			return
		}

		scheduler.Schedule(entry)
		httpJSON(w, http.StatusCreated, entry)
	})

	mux.HandleFunc("/remove", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			httpError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		name := r.URL.Query().Get("name")
		if name == "" {
			httpError(w, http.StatusBadRequest, "name is required")
			return
		}

		scheduler.Unschedule(name)

		if err := store.Remove(name); err != nil {
			httpError(w, http.StatusNotFound, err.Error())
			return
		}

		httpJSON(w, http.StatusOK, map[string]string{"name": name, "status": "REMOVED"})
	})

	mux.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			httpError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		name := r.URL.Query().Get("name")
		if name == "" {
			httpError(w, http.StatusBadRequest, "name is required")
			return
		}

		scheduler.Unschedule(name)

		entry, err := store.SetState(name, "paused")
		if err != nil {
			httpError(w, http.StatusNotFound, err.Error())
			return
		}

		httpJSON(w, http.StatusOK, map[string]string{"name": entry.Name, "state": entry.State})
	})

	mux.HandleFunc("/resume", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			httpError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		name := r.URL.Query().Get("name")
		if name == "" {
			httpError(w, http.StatusBadRequest, "name is required")
			return
		}

		entry, err := store.SetState(name, "active")
		if err != nil {
			httpError(w, http.StatusNotFound, err.Error())
			return
		}

		scheduler.Schedule(*entry)

		httpJSON(w, http.StatusOK, map[string]string{"name": entry.Name, "state": entry.State})
	})

	mux.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			httpError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		entries := store.List()
		httpJSON(w, http.StatusOK, entries)
	})

	mux.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			httpError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		name := r.URL.Query().Get("name")
		if name == "" {
			httpError(w, http.StatusBadRequest, "name is required")
			return
		}
		limitStr := r.URL.Query().Get("limit")
		limit := 10
		if limitStr != "" {
			if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
				limit = n
			}
		}
		history := store.GetHistory(name, limit)
		httpJSON(w, http.StatusOK, history)
	})

	pidFile = pidFilePath(port)
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write pid file: %v\n", err)
		os.Exit(1)
	}

	// Start scheduling all active entries
	scheduler.Start()

	addr := fmt.Sprintf(":%s", port)
	fmt.Fprintf(os.Stderr, "cron scheduler started on port %s\n", port)

	server := &http.Server{Addr: addr, Handler: mux}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		fmt.Fprintln(os.Stderr, "\nshutting down...")
		scheduler.Stop()
		os.Remove(pidFile)
		server.Close()
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		os.Remove(pidFile)
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func pidFilePath(port string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf(".aux4-cron-%s.pid", port))
}

func httpError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func httpJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
