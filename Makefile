.PHONY: help
help: ## Lists the available commands.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: compile provision ## Compiles binaries, generates configurations and provisions all infrastructure.

.PHONY: compile
compile: ## Compiles the binaries.
	GOOS=linux GOARCH=amd64 go build -o benchd cmd/benchd/main.go # Compile for server
	go build -o benchd cmd/benchctl/main.go # Compile for local

.PHONY: generate
generate: ## Generates local config files for different benchmark runs.
	echo "todo"

.PHONY: provision
provision: ## Provisions all infrastructure in Google Cloud
	cd terraform; terraform init; terraform plan -out planfile; terraform apply -auto-approve planfile

.PHONY: deploy
deploy: generate ##

.PHONY: local
local: ## Spins up a local development environment.
	docker-compose --build

.PHONY: clean
clean: ## Remove build artifacts.
	cd terraform; terraform apply -auto-approve -destroy; rm terraform.tfstate terraform.tfstate.backup; cd ..;
	rm benchd benchctl terraform/planfile;
