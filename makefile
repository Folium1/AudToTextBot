.SILENT:

build:
	@if [ -f "go.mod" ]; then \
		rm go.mod; \
	fi
	@if [ -f "go.sum" ]; then \
		rm go.sum; \
	fi
	go mod init tgbot
	go mod tidy
	go build


run:
ifneq ("$(wildcard $(tgbot))","")
	rm tgbot
endif
	docker build -t tgbot .
	docker-compose up --build