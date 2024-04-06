test:
	go test -v ./...

gen:
	@bash pkg/weather/nws/testdata/package.sh