# Makefile
.EXPORT_ALL_VARIABLES:	

GO111MODULE=on
GOPROXY=direct
GOSUMDB=off
GOPRIVATE=github.com/jeffotoni/benchmark-gocache

update:
	@echo "########## make test benchmark ... "
	@rm -f go.*
	go mod init benchmark-gocache
	go mod tidy
	@echo "done"
