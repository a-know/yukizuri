.PHONY: all

serve:
	./yukizuri.bin -addr=":8080" -logging=true

build:
	go build -o yukizuri.bin
	