.PHONY: install
install:
	@cd cmd/spotify-downloader/ && go build && go install
	@echo "Installed binary into $$GOPATH/bin"