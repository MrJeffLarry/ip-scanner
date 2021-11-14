# include env from local
LOCAL_ENV_FILE=local.env
ifneq ("$(wildcard $(LOCAL_ENV_FILE))","")
	include $(LOCAL_ENV_FILE)
	export $(shell sed 's/=.*//' $(LOCAL_ENV_FILE))
endif

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool
GOGET=$(GOCMD) get
BINARY_FOLDER=build/

BINARY_NAME=ipscan

BINARY_UNIX=$(BINARY_FOLDER)$(BINARY_NAME)
BINARY_WIN32=$(BINARY_FOLDER)$(BINARY_NAME)_32.exe
BINARY_WIN64=$(BINARY_FOLDER)$(BINARY_NAME)_64.exe

SOURCE_CMD=cmd/client

all: test build
build: 
	@echo "Compiling source"
	$(GOBUILD) -o $(BINARY_UNIX) $(SOURCE_CMD)/main.go
build-run: 
	@echo "Compiling source"
	$(GOBUILD) -o $(BINARY_UNIX) $(SOURCE_CMD)/main.go
	$(BINARY_UNIX) 192.168.1.x 80,443,22
test:
	$(GOTEST) -v ./... -cover
testIntegration:
	docker start mongodb-test || docker run --name mongodb-test -p 28017:27017 -d mongo
	$(GOTEST) -v ./... -cover -tags=integration
	docker stop mongodb-test
	docker rm mongodb-test
cover:
	$(GOTEST) -v -coverprofile=./build/c.out ./...
	$(GOTOOL) cover -html=./build/c.out -o ./build/coverage.html
coverAll:
	docker start mongodb-test || docker run --name mongodb-test -p 28017:27017 -d mongo
	$(GOTEST) -v -tags=integration -coverprofile=./build/c.out ./...
	docker stop mongodb-test
	docker rm mongodb-test
	$(GOTOOL) cover -html=./build/c.out -o ./build/coverage.html
clean:
	$(GOCLEAN)
	rm -r build
run:
	$(GOCMD) run $(SOURCE_CMD)/main.go 192.168.1.x

run-db:
	docker start mongodb || docker run --name mongodb -v ~/Docker/mongodb:/data/db -p 27017:27017 -d mongo
stop-db:
	docker stop mongodb && docker rm mongodb	

# Cross compilation
#build-ubuntu:
#	GOOS=linux GARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) $(SOURCE_CMD)/main.go
#build-arm:
#	GOOS=linux GOARCH=arm GOARM=5 $(GOBUILD) -o $(BINARY_UNIX) $(SOURCE_CMD)/main.go
build-win32:
	GOOS=windows GOARCH=386 $(GOBUILD) -o $(BINARY_WIN32) $(SOURCE_CMD)/main.go
build-win64:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_WIN64) $(SOURCE_CMD)/main.go
build-mac:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) $(SOURCE_CMD)/main.go


install:
	@echo "Compiling source"
	$(GOBUILD) -o $(BINARY_UNIX) $(SOURCE_CMD)/main.go
	@echo "Copy to go bin"
	cp $(BINARY_UNIX) ~/go/bin/