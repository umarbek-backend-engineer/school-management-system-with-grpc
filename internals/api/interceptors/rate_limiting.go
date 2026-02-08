package interceptors

import (
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type rateLimiter struct {
	mux       sync.Mutex
	visitor   map[string]int
	limit     int
	resetTime time.Duration
}

func NewRateLimiter(limit int, resetTime time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitor:   make(map[string]int),
		limit:     limit,
		resetTime: resetTime,
	}

	go rl.resetVisitorCount()
	return rl
}

func (rl *rateLimiter) resetVisitorCount() {
	for {
		time.Sleep(rl.resetTime)
		rl.mux.Lock()
		rl.visitor = make(map[string]int)
		rl.mux.Unlock()
	}
}

func (rl *rateLimiter) RateLimitIntercepter(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	rl.mux.Lock()
	defer rl.mux.Unlock()

	pr, ok := peer.FromContext(ctx) // trying to get the user IP
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unable to get the client IP")
	}

	// getting  and initializing visitor ip
	visitorIP := pr.Addr.String()		
	rl.visitor[visitorIP]++

	log.Printf("++++++++++++++++++ Visitor count from IP: %s: %d\n", visitorIP, rl.visitor[visitorIP])

	//  if the visito made more request then the limit the interceptor will block the request from this user for a pacified time
	if rl.visitor[visitorIP] > rl.limit {
		return nil, status.Error(codes.ResourceExhausted, "Too many Requests")
	}
	return handler(ctx, req)
}
