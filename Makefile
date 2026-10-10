# Makefile
.EXPORT_ALL_VARIABLES:

GO111MODULE=on
GOPROXY=direct
GOSUMDB=off
GOPRIVATE=github.com/jeffotoni/benchmark-gocache

.PHONY: test benchmark-current benchmark-current-save benchmark-current-count5 tidy

test:
	go test ./...

benchmark-current:
	go test -run '^$$' -bench '^BenchmarkCurrent' -benchmem -benchtime=3s -count=1

benchmark-current-save:
	go test -run '^$$' -bench '^BenchmarkCurrent' -benchmem -benchtime=3s -count=1 | tee benchmark-current.txt

benchmark-current-count5:
	go test -run '^$$' -bench '^BenchmarkCurrent' -benchmem -benchtime=3s -count=5

tidy:
	go mod tidy
