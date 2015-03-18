## thrtl
A small tool to limit IOPS through a pipe. Reads from `stdin` - writes to `stdout`. 
Gives you a TCP address where you can set throughput.

Feedback is very welcome.

## Install 
    go get github.com/stengaard/thrtl

## Usage 


    server1$ thrtl < somefile | nc server2 2222
    server2$ nc -l 2222 > some-other-file
