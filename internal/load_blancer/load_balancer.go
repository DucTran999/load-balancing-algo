package loadbalancer

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/DucTran999/load-balancing-algo/pkg/backend"
	"github.com/rs/zerolog/log"
)

type Algorithm int

func (a Algorithm) String() string {
	switch a {
	case RoundRobin:
		return "Round Robin"
	case WeightedRoundRobin:
		return "Weighted Round Robin"
	case SourceIPHash:
		return "Source IP Hash"
	case LeastConnection:
		return "Least Connection"
	case LowestLatency:
		return "Lowest Response Time"
	case ResourceBase:
		return "Resource Base"
	default:
		return ""
	}
}

const (
	RoundRobin Algorithm = iota
	WeightedRoundRobin
	SourceIPHash
	LeastConnection
	LowestLatency
	ResourceBase
)

type LoadBalancer interface {
	Start() error
}

type loadBalancer struct {
	port   int
	host   string
	server *http.Server
}

func NewLoadBalancer(
	host string, port int, targets []*backend.SimpleHTTPServer, alg Algorithm,
) (*loadBalancer, error) {

	hdl, err := NewLoadBalancerHandler(alg, targets)
	if err != nil {
		return nil, err
	}

	lb := &loadBalancer{
		host: host,
		port: port,
		server: &http.Server{
			Addr:         net.JoinHostPort(host, strconv.Itoa(port)),
			Handler:      hdl,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	return lb, nil
}

func (lb *loadBalancer) Start() error {
	// Start HTTP server in a goroutine
	go func() {
		if err := lb.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("failed to start load balancer")
		}
	}()

	// Wait for server to become available
	address := fmt.Sprintf("localhost:%d", lb.port)
	for {
		// Use a 3s default dial timeout, overrideable via config
		dialer := net.Dialer{
			Timeout: 3 * time.Second,
		}
		if conn, err := dialer.Dial("tcp", address); err == nil {
			if errCloseConn := conn.Close(); errCloseConn != nil {
				log.Warn().Err(errCloseConn).Msg("failed to close tcp conn")
			}
			break
		}
		time.Sleep(100 * time.Millisecond) // prevent tight loop
	}

	log.Info().Msgf("load balancer running on %v", address)
	return nil
}
