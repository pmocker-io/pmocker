.PHONY: build-cli run-gva run-gva-web test clean

build-cli:
	cd cli && go build -o pmocker.exe .

run-gva:
	cd gva/server && go run main.go

run-gva-web:
	cd gva/web && npm run dev

test:
	go work sync
	cd cli && go test ./...
	cd pkg && go test ./...

clean:
	rm -f cli/pmocker cli/pmocker.exe
	rm -f gva/server/gva-server gva/server/gva-server.exe
	find . -name "*.db" -delete
