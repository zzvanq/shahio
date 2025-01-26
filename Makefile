BINARY_FILENAME=main

build:
	echo "Building..."
	go build -o $(BINARY_FILENAME) main.go

run:
	echo "Running..."
	go build -o $(BINARY_FILENAME) main.go
	./$(BINARY_FILENAME)
	
