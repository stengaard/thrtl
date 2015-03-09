package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func term(c net.Conn, setRate chan int64, thisRate func() int64) {
	var (
		err  error
		line string
		i    int64
	)
	defer func() {
		var s string
		if err != nil {
			s = fmt.Sprintf("Err: %s", err)
		} else {
			s = "kthxbye!"
		}

		fmt.Fprint(c, s+"\n")
		c.Close()
	}()

	r := bufio.NewReader(c)

	for {
		fmt.Fprint(c, "\n >")
		line, err = r.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "Err %q reading from %s\n", err, c.RemoteAddr())
			return
		}
		if err == io.EOF {
			err = nil
			return
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		cmd, args := fields[0], fields[1:]
		switch strings.ToUpper(cmd) {
		case "GET_RATE":
			fmt.Fprint(c, thisRate())
		case "SET_RATE":
			if len(args) == 0 {
				fmt.Fprintf(c, "argument required\n")
				printUsage(c)
				continue
			}
			i, err = strconv.ParseInt(args[0], 0, 64)
			if err != nil {
				fmt.Fprintf(c, "bad value - %s", err)
				continue
			}
			setRate <- i
			fmt.Fprintf(c, "value set to %dKBps", i)
		case "QUIT":
			return
		default:
			fmt.Fprintf(c, "unknown command %q\n", cmd)
			printUsage(c)
		}
	}
}

func printUsage(c io.Writer) {
	fmt.Fprintf(c, "Good command: \n")
	fmt.Fprintf(c, "GET_RATE     - return the current flow rate.\n")
	fmt.Fprintf(c, "SET_RATE <n> - set flow rate to n KBps.\n")
}
