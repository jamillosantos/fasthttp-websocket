VERSION ?= $(shell git describe --tags 2>/dev/null | cut -c 2-)
TEST_FLAGS ?=
REPO_OWNER ?= $(shell cd .. && basename "$$(pwd)")


test-short:
	make test-with-flags --ignore-errors TEST_FLAGS='-short'

test:
	@-rm -r .coverage
	@mkdir .coverage
	make test-with-flags TEST_FLAGS='-v -race -covermode atomic -coverprofile .coverage/_$$(RAND).txt -bench=. -benchmem'
	@echo 'mode: atomic' > .coverage/combined.txt
	@cat .coverage/*.txt | grep -v 'mode: atomic' >> .coverage/combined.txt


test-with-flags:
	go test $(TEST_FLAGS) .

html-coverage:
	go tool cover -html=.coverage/combined.txt

deps:
	-go get -v -t ./...
	-go test -i ./...

list-external-deps:
	$(call external_deps,'.')

# example: make release V=0.0.0
release:
	git tag v$(V)
	@read -p "Press enter to confirm and push to origin ..." && git push origin v$(V)


define external_deps
	@echo '-- $(1)';  go list -f '{{join .Deps "\n"}}' $(1) | grep -v github.com/$(REPO_OWNER)/migrate | xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}'

endef


.PHONY: build-cli clean test-short test test-with-flags deps html-coverage \
        list-external-deps release

SHELL = /bin/bash
RAND = $(shell echo $$RANDOM)
