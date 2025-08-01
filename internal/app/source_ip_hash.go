package app

import (
	loadbalancer "github.com/DucTran999/load-balancing-algo/internal/load_blancer"
	"github.com/DucTran999/load-balancing-algo/internal/tools"
	"github.com/DucTran999/load-balancing-algo/pkg/backend"
	"github.com/rs/zerolog"
)

func RunSourceIPhashApp(logger zerolog.Logger) {
	// Initialize the backend builder and configure number of backend servers
	backendBuilder := backend.NewBackendBuilder(logger)
	backendBuilder.SetNumberOfBackends(5)

	// Build the backend servers
	backends, err := backendBuilder.Build()
	if err != nil {
		logger.Fatal().Msgf("failed when build backends: %v", err)
	}

	// Create a new load balancer on localhost:8080 using the backends and source ip algorithm
	lb, err := loadbalancer.NewLoadBalancer("localhost", 8080, backends, loadbalancer.SourceIPHash)
	if err != nil {
		logger.Fatal().Msgf("failed to init loadbalancer: %v", err)
	}

	// Start the load balancer asynchronously
	if err := lb.Start(); err != nil {
		logger.Fatal().Msgf("failed to start load balancer: %v", err)
	}

	// Initialize a request sender component and start sending requests asynchronously
	rs := tools.NewRequestSender(20)
	go rs.SendNow()

	// Wait for a graceful shutdown signal and stop the first backend cleanly
	GracefulShutdown(logger, backendBuilder.ShutdownAllBackends)
}
