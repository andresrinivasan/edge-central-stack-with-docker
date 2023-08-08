add-license:
	if [ x${LICENSE_FILE} = x ]; \
	then \
		echo "Please set the environment variable LICENSE_FILE"; \
	else \
		docker run --rm -v $(dir ${LICENSE_FILE}):/source -v license-data:/lic -w /source alpine cp $(notdir ${LICENSE_FILE}) /lic/LICENSE_FILE.lic; \
	fi

EDGE_CENTRAL_COMMON_SERVICES = core-data core-keeper core-metadata core-command redis mqtt-broker
EDGE_CENTRAL_ADDITIONAL_SERVICES = xpert-manager sys-mgmt device-virtual portainer

export COMPOSE_FILE ?= /etc/edgexpert/docker-compose.yml:/etc/edgexpert/app-service.yml

## export COMPOSE_PROJECT_NAME=edgexpert
export EDGEXPERT_IMAGE_REPO=edgexpert
export EDGEXPERT_IMAGE_VERSION=2.3

start-edge-central:
	docker compose up -d $(EDGE_CENTRAL_COMMON_SERVICES) $(EDGE_CENTRAL_ADDITIONAL_SERVICES)

stop-edge-central:
	docker compose down

start-app-service:
	docker compose up -d app-service --path ${APP-SERVICE-CONFIG}

stop-app-service:
	docker compose down app-service --path ${APP-SERVICE-CONFIG}

start-alpine:
	docker compose -p this-is-alpine-project -f ./alpine.yaml up -d

stop-alpine:
	docker compose -f ./alpine.yaml kill

render-canonical-docker-compose:
	docker compose config $(EDGE_CENTRAL_COMMON_SERVICES) $(EDGE_CENTRAL_ADDITIONAL_SERVICES)

