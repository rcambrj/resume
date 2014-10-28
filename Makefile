all: README.html README.pdf

README.html: README.md
	go run .

README.pdf: README.md
	go run .
