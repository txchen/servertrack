# servertrack

To build:

```bash
$ go get
$ go build
```

To run without build:

```bash
$ go run main.go stat.go template.go
```

To run tests:

```bash
$ go get github.com/stretchr/testify/assert
$ go test -v
$ go test -v -race
```

To run benchmark test:

```bash
$ go test -bench=.
```

The server would listen on port 30000, it has two apis:
* GET: http://localhost:30000/load?servername=xxxxx
* POST: http://localhost:30000/load?servername=yyyyy&cpuload=xx.xx&ramload=yy.yy

To load some data into the service, there is one script:

```bash
$ go get github.com/jawher/mow.cli
$ go run scripts/loadtest.go -h
# for example
$ go run scripts/loadtest.go -s host1 -d 20150404 -c=20000
# or, you can run multiple instances together
$ go run scripts/loadtest.go -s host2 -d 20150307 -c=10000 -t=3 & \
  go run scripts/loadtest.go -s host2 -d 20150303 -c=10000 -t=3 & \
  go run scripts/loadtest.go -s host3 -d 20150302 -c=10000 -t=3
```
