BINARY_NAME=process-photos
BUILD_FLAGS:="-X main.version=$(shell date +"%Y-%m-%dT%H:%M:%S%z")"

build : $(BINARY_NAME)

$(BINARY_NAME): $(wildcard *.go)
	go mod tidy
	CGO_CFLAGS_ALLOW="-Xpreprocessor" go build -ldflags $(BUILD_FLAGS) -o $(BINARY_NAME)

clean : 
	rm -f $(BINARY_NAME)
	rm -f out/*
	