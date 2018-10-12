package publish

import (
	"os/exec"
	"sync"
	"time"
)

type command struct {
	name      string
	args      []string
	cmd       *exec.Cmd
	mu        *sync.Mutex
	stdout    chan string
	stderr    chan string
	started   bool      // cmd.Start called, no error
	stopped   bool      // Stop called
	done      bool      // run() done
	final     bool      // status finalized in Status
	startTime time.Time // if started true
	status    *Status
	doneChan  chan struct{} // closed when done running
}

type Status struct {
	Cmd      string
	PID      int
	Complete bool     // false if stopped or signaled
	Exit     int      // exit code of process
	Error    error    // Go error
	StartTs  int64    // Unix ts (nanoseconds), zero if Cmd not started
	StopTs   int64    // Unix ts (nanoseconds), zero if Cmd not started or running
	Runtime  float64  // seconds, zero if Cmd not started
	Stdout   []string // buffered STDOUT; see Cmd.Status for more info
	Stderr   []string // buffered STDERR; see Cmd.Status for more info
}

type Cmd map[string]*command

func NewCmd() Cmd {
	return make(map[string]*command)
}

func (r *Cmd)AddCmd(name, bin string, args ...string) *Cmd {
	(*r)[name] = newCommand(bin, args...)

	return r
}

func newCommand(name string, args ...string) *command {
	return &command{
		name: name,
		args: args,
		mu:   &sync.Mutex{},
		status: &Status{
			Cmd:      name,
			PID:      0,
			Complete: false,
			Exit:     -1,
			Error:    nil,
			Runtime:  0,
		},
		doneChan: make(chan struct{}),
	}
}

func Start() <-chan Status {
	
}