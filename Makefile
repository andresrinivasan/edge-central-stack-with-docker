add-license:
	if [ x${LICENSE_FILE} = x ]; \
	then \
		echo "Please set the environment variable LICENSE_FILE"; \
	else \
		docker run --rm -v $(dir ${LICENSE_FILE}):/source -v license-data:/lic -w /source alpine cp $(notdir ${LICENSE_FILE}) /lic/LICENSE_FILE.lic; \
	fi

EDGE_CENTRAL_COMMON_SERVICES = core-data core-keeper core-metadata core-command redis mqtt-broker
EDGE_CENTRAL_ADDITIONAL_SERVICES = xpert-manager sys-mgmt device-virtual portainer

export COMPOSE_FILE?=/etc/edgexpert/docker-compose.yml

export EDGEXPERT_IMAGE_REPO=edgexpert
export EDGEXPERT_IMAGE_VERSION=2.3

start-edge-central:
	docker compose up -d $(EDGE_CENTRAL_COMMON_SERVICES) $(EDGE_CENTRAL_ADDITIONAL_SERVICES)

stop-edge-central:
	docker compose down

APP_SERVICE_NAME=javascript-http-export
export EDGEX_PROFILE=$(APP_SERVICE_NAME)

start-app-service: javascript-http-export.toml edgecentral-app-service-config app-service-override.yaml
	compose_files="$(shell echo ${COMPOSE_FILE} | sed 's/^/-f /' | sed 's/:/ -f /g')"; \
	docker compose $${compose_files} -f ./app-service-override.yaml -p $(APP_SERVICE_NAME) up app-service --detach --no-deps

http-trigger.json javascript-http-export.toml app-service-override.yaml:
	wget https://raw.githubusercontent.com/andresrinivasan/edge-central-http-export-example/main/$@

edgecentral-app-service-config:
	docker run --rm -v ${PWD}:/source -v edgecentral-app-service-config:/res -w /source alpine sh -c "mkdir -p /res/${APP_SERVICE_NAME}; cp ${APP_SERVICE_NAME}.toml /res/${APP_SERVICE_NAME}/configuration.toml"

test-app-service: http-trigger.json
	curl -d @http-trigger.json -H Content-Type:application/json -X POST http://localhost:59704/api/v2/trigger

stop-app-service:
	compose_files="$(shell echo ${COMPOSE_FILE} | sed 's/^/-f /' | sed 's/:/ -f /g')"; \
	docker compose $${compose_files} -f ./app-service-override.yaml -p $(APP_SERVICE_NAME) down app-service

start-alpine:
	docker compose -p this-is-the-alpine-project -f ./alpine.yaml up -d

stop-alpine:
	docker compose -f ./alpine.yaml kill

render-canonical-docker-compose:
	docker compose config $(EDGE_CENTRAL_COMMON_SERVICES) $(EDGE_CENTRAL_ADDITIONAL_SERVICES)

