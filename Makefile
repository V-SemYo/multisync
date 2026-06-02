build:
	go build -o multisync .

test:
	go test ./... -v

vet:
	go vet ./...

run:
	go run . --config config.yaml --once

clean:
	rm -f multisync