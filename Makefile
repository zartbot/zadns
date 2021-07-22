GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
LDFLAGS = -s -w
BINARY_NAME=zadns
BINARY_PATH=./build/zadns
CONFIG_PATH=./config
MODEL_PATH=./model

all: build
build:  linux mac win
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -rf $(BINARY_PATH)
	mkdir -p $(BINARY_PATH)
	cp -r $(CONFIG_PATH) $(BINARY_PATH)
	cp -r $(MODEL_PATH) $(BINARY_PATH)

deps:
	$(GOGET) -u -v  github.com/zartbot/zadns

linux: $(info >> Starting build linux package...)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)"  -o $(BINARY_PATH)/$(BINARY_NAME)_linux -v

mac:  $(info >> Starting build mac package...)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_PATH)/$(BINARY_NAME)_mac -v

win:  $(info >> Starting build windows package...)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_PATH)/$(BINARY_NAME)_win.exe -v