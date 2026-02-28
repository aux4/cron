package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func buildURL(port, path string, params map[string]string) string {
	u := fmt.Sprintf("http://localhost:%s%s", port, path)
	v := url.Values{}
	for key, val := range params {
		if val != "" {
			v.Set(key, val)
		}
	}
	if len(v) > 0 {
		u += "?" + v.Encode()
	}
	return u
}

func stopServer(args []string) {
	port := getArg(args, 0, "8421")

	pidFile := pidFilePath(port)
	data, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scheduler not running on port %s\n", port)
		os.Exit(1)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid pid file: %s\n", pidFile)
		os.Remove(pidFile)
		os.Exit(1)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "process %d not found\n", pid)
		os.Remove(pidFile)
		os.Exit(1)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		fmt.Fprintf(os.Stderr, "failed to stop scheduler: %v\n", err)
		os.Remove(pidFile)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "{\"status\":\"STOPPED\",\"port\":\"%s\",\"pid\":%d}\n", port, pid)
}

func addEntry(args []string) {
	port := getArg(args, 0, "8421")
	name := getArg(args, 1, "")
	every := getArg(args, 2, "")
	at := getArg(args, 3, "")
	run := getArg(args, 4, "")

	if name == "" {
		fmt.Fprintln(os.Stderr, "task name is required")
		os.Exit(1)
	}
	if every == "" {
		fmt.Fprintln(os.Stderr, "schedule expression is required")
		os.Exit(1)
	}
	if run == "" {
		fmt.Fprintln(os.Stderr, "run command is required")
		os.Exit(1)
	}

	params := map[string]string{
		"name":  name,
		"every": every,
		"at":    at,
		"run":   run,
	}

	resp, err := http.Post(buildURL(port, "/add", params), "application/json", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "%s\n", body)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s", body)
}

func removeEntry(args []string) {
	port := getArg(args, 0, "8421")
	name := getArg(args, 1, "")

	if name == "" {
		fmt.Fprintln(os.Stderr, "task name is required")
		os.Exit(1)
	}

	params := map[string]string{"name": name}

	resp, err := http.Post(buildURL(port, "/remove", params), "application/json", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "%s\n", body)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s", body)
}

func pauseEntry(args []string) {
	port := getArg(args, 0, "8421")
	name := getArg(args, 1, "")

	if name == "" {
		fmt.Fprintln(os.Stderr, "task name is required")
		os.Exit(1)
	}

	params := map[string]string{"name": name}

	resp, err := http.Post(buildURL(port, "/pause", params), "application/json", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "%s\n", body)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s", body)
}

func resumeEntry(args []string) {
	port := getArg(args, 0, "8421")
	name := getArg(args, 1, "")

	if name == "" {
		fmt.Fprintln(os.Stderr, "task name is required")
		os.Exit(1)
	}

	params := map[string]string{"name": name}

	resp, err := http.Post(buildURL(port, "/resume", params), "application/json", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "%s\n", body)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s", body)
}

func listEntries(args []string) {
	port := getArg(args, 0, "8421")

	resp, err := http.Get(buildURL(port, "/list", nil))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "%s\n", body)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s", body)
}

func showHistory(args []string) {
	port := getArg(args, 0, "8421")
	name := getArg(args, 1, "")
	limit := getArg(args, 2, "10")

	if name == "" {
		fmt.Fprintln(os.Stderr, "task name is required")
		os.Exit(1)
	}

	params := map[string]string{
		"name":  name,
		"limit": limit,
	}

	resp, err := http.Get(buildURL(port, "/history", params))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "%s\n", body)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s", body)
}
