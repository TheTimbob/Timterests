# Simple Makefile for a Go project

# Build the application
all: build test
templ-install:
	@if ! command -v templ > /dev/null; then \
		read -p "Go's 'templ' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/a-h/templ/cmd/templ@latest; \
			if [ ! -x "$$(command -v templ)" ]; then \
				echo "templ installation failed. Exiting..."; \
				exit 1; \
			fi; \
		else \
			echo "You chose not to install templ. Exiting..."; \
			exit 1; \
		fi; \
	fi

build: templ-install
	@echo "Building..."
	@templ generate

	@CGO_ENABLED=1 go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Coverage report
coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=tmp/coverage.out -covermode=atomic ./...
	@cat tmp/coverage.out | grep -v "_templ.go" > tmp/cover.out
	@go tool cover -func=tmp/cover.out

complexity:
	@echo "Calculating code complexity..."
	@if command -v gocyclo > /dev/null; then \
		gocyclo -over 15 .; \
	else \
		read -p "Go's 'gocyclo' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/fzipp/gocyclo@latest; \
			gocyclo -over 15 .; \
		else \
			echo "You chose not to install gocyclo. Exiting..."; \
			exit 1; \
		fi; \
	fi

# Clean the binary and temp files
clean:
	@echo "Cleaning..."
	@rm -f main
	@rm -rf tmp/*
	@go clean

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# Deploy to AWS
deploy:
	@echo "Deploying..."
	@if [ -f "utils/aws-deploy.sh" ]; then \
		./utils/aws-deploy.sh; \
	else \
		echo "Deployment script not found."; \
		exit 1; \
	fi


.PHONY: all build run test coverage clean watch templ-install
