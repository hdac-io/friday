# Makefile for the "localnode" docker image.

build-docker-image:
	rm $(CURDIR)/casperlabs_linux_builder/build_casperlabs_ee_linux.sh
	cp $(CURDIR)/../install_casperlabs_ee.sh $(CURDIR)/casperlabs_linux_builder/build_casperlabs_ee_linux.sh
	docker build --tag friday/casperlabs_linux_builder casperlabs_linux_builder

build-ee-linux:
	docker run --rm -v $(CURDIR)/../../:/friday:Z friday/casperlabs_linux_builder
	mkdir -p $(CURDIR)/../../build/CasperLabs/
	cp -r $(CURDIR)/../../CasperLabs/execution-engine/target/ ../../build/CasperLabs/

