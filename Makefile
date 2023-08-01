add-license:
	if [ x${LICENSE_FILE} = x ]; \
	then \
		echo "Please set the environment variable LICENSE_FILE"; \
	else \
		docker run --rm -v $(dir ${LICENSE_FILE}):/source -v license-data:/lic -w /source alpine cp $(notdir ${LICENSE_FILE}) /lic/LICENSE_FILE.lic; \
	fi

EDGE_CENTRAL_CORE_SERVICES = core-data core-keeper core-metadata core-command redis mqtt-broker

export COMPOSE_FILE ?= /etc/edgexpert/docker-compose.yml:/etc/edgexpert/app-service.yml


export COMPOSE_PROJECT_NAME=edgexpert
export EDGEXPERT_IMAGE_REPO=edgexpert
export EDGEXPERT_IMAGE_VERSION=2.3

start-edge-central:
	docker compose up -d xpert-manager sys-mgmt device-virtual portainer $(EDGE_CENTRAL_CORE_SERVICES)

start-app-service:
	docker compose up -d app-service --path ${APP-SERVICE-CONFIG}

stop-edge-central:
	docker compose down

start-alpine:
	docker compose -p this-is-alpine-project -f ./alpine.yaml up -d

stop-alpine:
	docker compose -f ./alpine.yaml kill
