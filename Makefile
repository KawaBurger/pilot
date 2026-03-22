.PHONY: build build-web build-server deploy stop

build: build-web build-server

build-web:
	cd web && npm run build

build-server:
	go build -o pilot ./cmd/pilot

deploy: build stop
	nohup ./pilot serve > ~/.pilot/pilot.log 2>&1 &
	@echo "Pilot running on :8090 (log: ~/.pilot/pilot.log)"

stop:
	-@pkill -f "./pilot serve" 2>/dev/null; true
