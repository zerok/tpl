all: tpl

tpl: $(shell find . -name '*.go')
	cd cmd/tpl && go build -o ../../tpl

install:
	cd cmd/tpl && go install

test:
	go test ./... -v

clean:
	rm -f tpl

.PHONY: clean
.PHONY: install
.PHONY: all
