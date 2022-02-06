.PHONY: help
help: ## Lists the available commands.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: compile provision dashboard ## Compiles binaries, provisions all infrastructure and starts Grafana Dashbaord.

.PHONY: clean
clean: ## Remove build artifacts.
	docker-compose down;
	cd terraform; terraform apply -auto-approve -destroy; rm tfplan; cd ..;
	rm -rf bin;

.PHONY: rebuild
rebuild: clean all ## Tear down and bring everything back up.

.PHONY: compile
compile: ## Compiles the binaries.
	GOOS=linux GOARCH=amd64 go build -o bin/benchd cmd/benchd/main.go # Compile for server
	go build -o bin/benchctl cmd/benchctl/*.go # Compile for local
	go build -o bin/promdl cmd/promdl/*.go # Compile for local

.PHONY: provision
provision: compile ## Provisions all infrastructure in Google Cloud
	cd terraform; terraform init; terraform plan -out tfplan; terraform apply -auto-approve tfplan

.PHONY: dashboard
dashboard: ## Deploy a local Grafana instance for easier monitoring
	docker-compose up -d grafana

.PHONY: local
local: ## Spins up a local development environment.
	docker-compose up --build
