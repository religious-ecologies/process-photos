BINARY_NAME=process-photos
BUILD_FLAGS:="-X main.version=$(shell date +"%Y-%m-%dT%H:%M:%S%z")"

build : $(BINARY_NAME)

$(BINARY_NAME): $(wildcard *.go)
	go mod tidy
	CGO_CFLAGS_ALLOW="-Xpreprocessor" go build -ldflags $(BUILD_FLAGS) -o $(BINARY_NAME)

install : $(wildcard *.go)
	go mod tidy
	CGO_CFLAGS_ALLOW="-Xpreprocessor" go install -ldflags $(BUILD_FLAGS) -o $(BINARY_NAME)

clean : 
	rm -f test-output/*

clobber :
	rm -f $(BINARY_NAME)
	
test : clean
	@mkdir -p test-output
	./$(BINARY_NAME) -r ccw -h 0.2 -w 0.2 --background purple test/IMG_006*.JPG -o test-output
	./$(BINARY_NAME) -r cw --background black test/IMG_010*.JPG -o test-output/

.PHONY : test
