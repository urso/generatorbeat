BEATNAME=generatorbeat
BEAT_DIR=github.com/urso/
SYSTEM_TESTS=false
TEST_ENVIRONMENT=false
ES_BEATS=./vendor/github.com/elastic/beats
GOPACKAGES=$(shell glide novendor)
BEAT_DIR=github.com/urso/
PREFIX?=.

# Path to the libbeat Makefile
include $(ES_BEATS)/libbeat/scripts/Makefile


.PHONY: generate
generate:
	python scripts/generate_template.py    etc/fields.yml    etc/generatorbeat.template.json
	python scripts/generate_field_docs.py  etc/fields.yml    etc/generatorbeat.asciidoc

.PHONY: install-cfg
install-cfg:
	mkdir -p $(PREFIX)
	cp etc/generatorbeat.template.json     $(PREFIX)/generatorbeat.template.json
	cp etc/generatorbeat.yml               $(PREFIX)/generatorbeat.yml
	cp etc/generatorbeat.yml               $(PREFIX)/generatorbeat-linux.yml
	cp etc/generatorbeat.yml               $(PREFIX)/generatorbeat-binary.yml
	cp etc/generatorbeat.yml               $(PREFIX)/generatorbeat-darwin.yml
	cp etc/generatorbeat.yml               $(PREFIX)/generatorbeat-win.yml

.PHONY: update-deps
update-deps:
	glide update  --no-recursive
