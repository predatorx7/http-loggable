.PHONY: build read write delete clean

read:
	go run bin/read/read.go

write:
	go run bin/write/write.go

build:
	make clean
	mkdir -p build/bin
	go build -o build/bin/write bin/write/write.go
	go build -o build/bin/read bin/read/read.go
	cp -r public build/public

delete:
	rm -rf logs/

clean:
	rm -rf build/
