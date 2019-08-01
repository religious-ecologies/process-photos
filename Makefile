BINARY_NAME=process-photos
BUILD_FLAGS:="-X main.version=$(shell date +"%Y-%m-%dT%H:%M:%S%z")"

build : $(BINARY_NAME)

$(BINARY_NAME): $(wildcard *.go)
	go mod tidy
	CGO_CFLAGS_ALLOW="-Xpreprocessor" go build -ldflags $(BUILD_FLAGS) -o $(BINARY_NAME)

clean : 
	rm -f test-output/*

clobber :
	rm -f $(BINARY_NAME)
	
test : clean
	@mkdir -p test-output
	./$(BINARY_NAME) -h 0.2 -w 0.2 --background "srgb(146, 147, 199)" test/purple-IMG_006*.JPG -o test-output/
	./$(BINARY_NAME) -h 0.2 -w 0.075 -p 50 --background purple test/purple-IMG_5*.JPG -o test-output/
	./$(BINARY_NAME) -r cw --background black test/black-*.JPG -o test-output/
	./$(BINARY_NAME) --background gray -h 0.16 -w 0.1 test/gray-*.JPG -o test-output/

.PHONY : test
