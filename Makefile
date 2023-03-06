portage-exporter: $(shell find -name '*.go')
	go build ./cmd/portage-exporter
