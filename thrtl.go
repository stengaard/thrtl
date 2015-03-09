// thrtl is a live-throttling pipe-through. It reads from stdin and writes to stdout.
// By default it writes as fast a possible, but the maximum throughput can be adjusted
// in a socket. Most likely not useful to most people.
package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"
)

var (
	ctl = flag.String("ctl", ":0", "Control location - (eg: <addr>:<port>)")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
}

func main() {

	flag.Usage = usage
	flag.Parse()

	s := &throttledReader{
		int: os.Stdin,
		throttle: throttle{
			KBps:    0,
			newRate: make(chan int64, 1),
		},
	}
	var clean []func()
	done := make(chan struct{})

	kill := make(chan os.Signal)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGQUIT)

	list, err := net.Listen("tcp", *ctl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err listening on addr: %s\n", err)
		os.Exit(1)
	}
	if tcpAddr, ok := list.Addr().(*net.TCPAddr); ok {
		fmt.Fprintf(os.Stderr, "Connect to control plane using telnet %s %d\n", tcpAddr.IP, tcpAddr.Port)
	} else {
		fmt.Fprintf(os.Stderr, "Control plane ready at %v\n", list.Addr())
	}

	go func() {
		defer list.Close()
		for {
			c, err := list.Accept()
			if err != nil {
				continue
			}
			go term(c, s.newRate, func() int64 {
				return s.throttle.KBps
			})

			select {
			case <-done:
				return
			default:
			}
		}
	}()

	// this is the important bit
	clean = append(clean, func() {
		os.Stdout.Close()
	})
	go func() {
		io.Copy(os.Stdout, s)
		close(done)
	}()

	select {
	case <-kill:
	case <-done:
	}

	for _, c := range clean {
		c()
	}

}

type throttledReader struct {
	int io.Reader
	throttle
}

type throttle struct {
	KBps    int64
	newRate chan int64
}

func (r *throttledReader) Read(b []byte) (n int, err error) {
	f := func() int {
		n, err = r.int.Read(b)
		return n
	}
	r.throttle.Delay(f)
	return n, err
}

// Delay delays the execution of time as close a possible to (n/1024)/t.KBps
func (t *throttle) Delay(doRead func() (n int)) {
	select {
	case rate := <-t.newRate:
		t.KBps = rate
	default:
	}

	start := time.Now()
	n := doRead()
	readTime := time.Now().Sub(start)
	sleep := byteTime(t.KBps, n)
	time.Sleep(sleep - readTime)
}

// byteTime returns the time required for n bytes at kbps.
func byteTime(kbps int64, n int) time.Duration {
	if kbps == 0 {
		return 0
	}
	d := time.Duration(float64(time.Second) * (float64(n) / 1024) / float64(kbps))
	return d
}
