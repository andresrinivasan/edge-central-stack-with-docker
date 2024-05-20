
## From https://unix.stackexchange.com/questions/235223/makefile-include-env-file
ifdef ENV
include $(ENV)
export $(shell sed 's/=.*//' $$ENV)
endif

#export $(shell [ ! -n "$(ENV)" ] || cat $(ENV) | grep -v --perl-regexp '^('$$(env | sed 's/=.*//'g | tr '\n' '|')')\=')

# test:
# 	echo $$LICENSE_FILE
# 	echo $$COMPOSE_FILE


add-license:
ifndef LICENSE_FILE
	$(error LICENSE_FILE environment variable must point to a valid Edge Central license file)
else
	docker run --rm -v $(dir $(LICENSE_FILE)):/source -v license-data:/lic -w /source alpine cp $(notdir $(LICENSE_FILE)) /lic/LICENSE_FILE.lic
endif

require-compose-file:
ifndef COMPOSE_FILE
	$(error COMPOSE_FILE environment variable must point to at least one Docker Compose file)
endif

EDGE_CENTRAL_COMMON_SERVICES ?= core-data core-keeper core-metadata core-command redis mqtt-broker
EDGE_CENTRAL_ADDITIONAL_SERVICES ?= xpert-manager sys-mgmt device-virtual portainer

EDGEXPERT_IMAGE_REPO ?= edgexpert
EDGEXPERT_IMAGE_VERSION ?= 2.3

export EDGEXPERT_IMAGE_REPO
export EDGEXPERT_IMAGE_VERSION

start-edge-central: require-compose-file
	docker compose up --detach $(EDGE_CENTRAL_COMMON_SERVICES) $(EDGE_CENTRAL_ADDITIONAL_SERVICES);

stop-edge-central: require-compose-file
	docker compose down

start-app-services: http-export-app-service mqtt-export-service

http-export-app-service: require-compose-file javascript-http-export.toml app-service-override.yaml http-export-app-service-config
	export APP_SERVICE_NAME=javascript-http-export; \
	export EDGEX_PROFILE=$${APP_SERVICE_NAME}; \
	export COMPOSE_FILE=$${COMPOSE_FILE}:app-service-override.yaml; \
	docker compose -p $${APP_SERVICE_NAME} up app-service --detach --no-deps

http-trigger.json javascript-http-export.toml app-service-override.yaml:
	http -q -d https://raw.githubusercontent.com/andresrinivasan/edge-central-http-export-example/main/$@

http-export-app-service-config:
	export APP_SERVICE_NAME=javascript-http-export; \
	export EDGEX_PROFILE=$${APP_SERVICE_NAME}; \
	export COMPOSE_FILE=$${COMPOSE_FILE}:app-service-override.yaml; \
	docker run --rm -v $${PWD}:/source -v edgecentral-app-service-config:/res -w /source alpine sh -c "mkdir -p /res/$${APP_SERVICE_NAME}; cp $${APP_SERVICE_NAME}.toml /res/$${APP_SERVICE_NAME}/configuration.toml"

test-http-export-app-service: http-trigger.json
	http POST http://localhost:59704/api/v2/trigger <http-trigger.json

mqtt-export-service:

stop-http-export-app-service: require-compose-file http-export-vars
	export APP_SERVICE_NAME=javascript-http-export; \
	export EDGEX_PROFILE=$${APP_SERVICE_NAME}; \
	export COMPOSE_FILE=$${COMPOSE_FILE}:app-service-override.yaml; \
	docker compose -p $${APP_SERVICE_NAME} down app-service

start-alpine:
	docker compose -p this-is-the-alpine-project -f ./alpine.yaml up -d

stop-alpine:
	docker compose -f ./alpine.yaml kill

render-canonical-docker-compose: require-compose-file
	docker compose config $(EDGE_CENTRAL_COMMON_SERVICES) $(EDGE_CENTRAL_ADDITIONAL_SERVICES)

