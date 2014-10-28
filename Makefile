all: index.html Robert_Cambridge.pdf

index.html: README.md
	go run .

Robert_Cambridge.pdf: README.md
	go run .
