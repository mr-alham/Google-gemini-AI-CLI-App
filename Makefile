exec:
	./gemini

text:
	./gemini --text

image:
	./gemini --image

help:
	./gemini --help

run:
	go run main.go

build:
	go build -a -v -o gemini main.go
