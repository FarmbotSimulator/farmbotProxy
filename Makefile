run:
	go run .

build:
	mkdir -p builds
	go build -o builds/farmbotproxy .

install:
	@echo "Still working on this"
	cp builds/farmbotproxy/* /usr/bin/
	farmbotproxy init

debs:
	mkdir -p builds
	go build -o builds/farmbotproxy .
	mkdir -p installer/farmbotProxy_1.2-1_amd64/usr/local/bin
	cp builds/farmbotproxy installer/farmbotProxy_1.2-1_amd64/usr/local/bin/
	mkdir -p installer/farmbotProxy_1.2-1_amd64/DEBIAN
	cp deb/p* installer/farmbotProxy_1.2-1_amd64/DEBIAN/
	cp deb/control_amd64/* installer/farmbotProxy_1.2-1_amd64/DEBIAN/
	dpkg-deb --build --root-owner-group installer/farmbotProxy_1.2-1_amd64
	
	mkdir -p installer/farmbotProxy_1.2-1_i386/usr/local/bin
	cp builds/farmbotproxy installer/farmbotProxy_1.2-1_i386/usr/local/bin/
	mkdir -p installer/farmbotProxy_1.2-1_i386/DEBIAN
	cp deb/p* installer/farmbotProxy_1.2-1_i386/DEBIAN/
	cp deb/control_i386/* installer/farmbotProxy_1.2-1_i386/DEBIAN/
	dpkg-deb --build --root-owner-group installer/farmbotProxy_1.2-1_i386

	mkdir -p installer/farmbotProxy_1.2-1_arm64/usr/local/bin
	cp builds/farmbotproxy installer/farmbotProxy_1.2-1_arm64/usr/local/bin/
	mkdir -p installer/farmbotProxy_1.2-1_arm64/DEBIAN
	cp deb/p* installer/farmbotProxy_1.2-1_arm64/DEBIAN/
	cp deb/control_arm64/* installer/farmbotProxy_1.2-1_arm64/DEBIAN/
	dpkg-deb --build --root-owner-group installer/farmbotProxy_1.2-1_arm64
