run:
	go run .

build:
	mkdir -p builds
	go build -o builds/farmbotsimulator .

install:
	@echo "Still working on this"
	cp builds/farmbotsimulator /usr/bin/
	farmbotsimulator init
