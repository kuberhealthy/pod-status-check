IMAGE := "kuberhealthy/pod-status-check"
TAG := "latest"

# Build the pod status check container locally.
build:
	podman build -f Containerfile -t {{IMAGE}}:{{TAG}} .

# Run the unit tests for the pod status check.
test:
	go test ./...

# Build the pod status check binary locally.
binary:
	go build -o bin/pod-status-check ./cmd/pod-status-check
