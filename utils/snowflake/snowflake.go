package snowflake

import (
	"sync"
	"time"
)

// epoch is 2023-01-01T00:00:00Z in Unix seconds.
// Layout: (secondsSinceEpoch << 10) | (machineID << 5) | sequence
// Produces ~12-digit IDs in the current era.
const epoch = int64(1672531200)

// Generator produces monotonically increasing 12-digit snowflake IDs.
type Generator struct {
	mu        sync.Mutex
	machineID int64
	lastSec   int64
	seq       int64
}

var defaultGen = &Generator{machineID: 1}

// NewGenerator creates a Generator with the given machineID (0–31).
func NewGenerator(machineID int) *Generator {
	return &Generator{machineID: int64(machineID & 0x1F)}
}

// NextID returns a unique 12-digit ID.
func (g *Generator) NextID() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().Unix() - epoch
	if now == g.lastSec {
		g.seq = (g.seq + 1) & 0x1F
		if g.seq == 0 {
			// sequence exhausted — spin until the next second
			for time.Now().Unix()-epoch == g.lastSec {
				time.Sleep(time.Millisecond)
			}
			now = time.Now().Unix() - epoch
		}
	} else {
		g.lastSec = now
		g.seq = 0
	}

	return (now << 10) | (g.machineID << 5) | g.seq
}

// NextID returns a unique ID from the package-level default generator (machineID=1).
func NextID() int64 { return defaultGen.NextID() }
