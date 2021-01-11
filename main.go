package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/pedropombeiro/qnapexporter/lib/exporter/prometheus"
)

func main() {
	// Setup our Ctrl+C handler
	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	e := prometheus.NewExporter()

	// handle route using handler function
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")

		err := e.WriteMetrics(w)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	// listen to port
	server := &http.Server{Addr: ":9094"}
	go func() {
		for {
			select {
			case <-exitCh:
				fmt.Fprintln(os.Stderr, "Program aborted, exiting...")
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
				err := server.Shutdown(ctx)
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
				}
				cancel()
				return
			case <-time.After(1 * time.Second):
				break
			}
		}
	}()

	err := server.ListenAndServe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	os.Exit(1)
}