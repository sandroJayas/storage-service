package middleware

import (
	"github.com/sandroJayas/storage-service/config"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Limits unauthenticated routes per IP address
// Allows:
// 1 request per second
// up to 3 requests burst
// Cleans up idle IPs every 5 mins

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	clients = make(map[string]*clientLimiter)
	mu      sync.Mutex

	// Cleanup old clients periodically
	cleanupInterval = time.Minute * 5
)

func init() {
	go cleanupClients()
}

func cleanupClients() {
	for {
		time.Sleep(cleanupInterval)
		mu.Lock()
		for ip, cl := range clients {
			if time.Since(cl.lastSeen) > time.Minute*10 {
				delete(clients, ip)
			}
		}
		mu.Unlock()
	}
}

func getLimiterForIP(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	cl, exists := clients[ip]
	if !exists {
		limiter := rate.NewLimiter(1, 3) // 1 request/sec with burst of 3
		clients[ip] = &clientLimiter{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	cl.lastSeen = time.Now()
	return cl.limiter
}

// RateLimitMiddleware applies a rate limit per IP
func RateLimitMiddleware() gin.HandlerFunc {
	if config.AppConfig.AppEnv == "testing" {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getLimiterForIP(ip)

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}
