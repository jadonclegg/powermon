all = main.go commands/client.go commands/powermon.go commands/server.go

powermon: $(all)
	go build

powermon-arm: $(all)
	env GOOS=linux GOARCH=arm GOARM=5 go build -o powermon-arm

powermon-arm64: $(all)
	env GOOS=linux GOARCH=arm64 GOARM=7 go build -o powermon-arm64

.PHONY: arm
arm: powermon-arm

.PHONY: arm64
arm64: powermon-arm64

.PHONY: all
all: arm arm64 powermon