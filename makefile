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

clean:
	if [ -f powermon-arm ]; then rm powermon-arm; fi
	if [ -f powermon-arm64 ]; then rm powermon-arm64; fi
	if [ -f powermon ]; then rm powermon; fi

install: powermon
	cp powermon /usr/bin

uninstall:
	if [ -f /usr/bin/powermon ]; then rm /usr/bin/powermon; fi

powermon.exe: $(all)
	env GOOS=windows go build