all: tpl

tpl: $(shell find . -name '*.go')
	cd cmd/tpl && go build -o ../../tpl

clean:
	rm -f tpl
