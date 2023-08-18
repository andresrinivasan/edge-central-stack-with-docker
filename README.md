# Simple Edge Central Stack

The purpose of this repo is to demonstrate how to manage an Edge Central stack without using the Edge Central CLI. This demonstration is just that; you may or may not want this particular set of services in your own environment.

## Prerequisites

- [IOTech Edge Central Software](https://www.iotechsys.com/products/edge-central/edge-central-installer-download/) Installed
- A valid (or [eval](https://www.iotechsys.com/resources/evaluating-the-software/edge-central-evaluation/)) license
- [docker and docker compose](https://www.docker.com/) daemon and commands (or [equivalent](https://github.com/abiosoft/colima))
- [curl](https://curl.se/) command
- [jq](https://jqlang.github.io/jq/) command
- [wget](https://www.gnu.org/software/wget/) command

## Demonstration

The Edge Central CLI is a wrapper around Docker and Docker Compose and exists to simplify starting/stopping the Edge Central microservices. I use simple targets in a Makefile to demonstrate what Docker commands are being used.

| Edge Central Command | Docker Command | Make Target |
| --- | --- | --- |
| edgexpert license install | N/A | add-license |
| edgexpert up | docker compose up ... | start-edge-central |
| edgexpert down | docker compose down ... | stop-edge-central |
| edgexpert gen | docker compose config ... | render-canonical-docker-compose |
| edgexpert up app-service --path=... | docker compose -f app-service.yaml... | start-app-service |

## Additional Information

Between passing arguments to Docker Compose knowing the minimum microservices needed for Edge Central, there's a bit more complexity.

### Adding the License Key

You must have a valid license key to start Edge Central. All IOTech microservices look for the license key in a Docker volume named _license-data_ with the path _/lic/LICENSE_FILE.lic_. An evaluation license can be requested via <https://www.iotechsys.com/resources/evaluating-the-software/edge-central-evaluation/>.

### Starting Edge Central

There is a common set of microservices that are more or less needed for Edge Central (see `EDGE_CENTRAL_COMMON_SERVICES` in the _Makefile_). For example, if events are not being stored in Redis, there is no need to start Redis. Those common services are added to the underlying Docker Compose command when using _edgexpert_.

There are additional services that may be needed, for example Portainer. In the _Makefile_ these are added to `EDGE_CENTRAL_ADDITIONAL_SERVICES`. Please see the [IOTech Services](https://docs.iotechsys.com/edge-xpert23/cli/cli-services.html) documentation for an enumeration of the services included with Edge Central.

### Starting an App Service

This is a little more complex as a custom version of the Edge Central `app-service.yaml` is required along with a couple of support files.

#### Custom Version of `app-service.yml`

The app-service.yml that comes with Edge Central requires a couple of modifications. See [the Edge Central HTTP Export Example repo](https://github.com/andresrinivasan/edge-central-http-export-example) for the source and additional story telling. First and most important is to use a Docker Volume rather than a [bind mount](https://docs.docker.com/storage/bind-mounts/) for the app service configuration file. 

App services expect to find their configuration in res/APP-SERVICE-NAME/configuration.toml where APP-SERVICE-NAME is...wait for it...the name of the app service. The modified Docker Compose file (`custom-app-service.yaml`) now includes an external volume dependency on `edgecentral-app-service-config` which is mounted as `/res`. When the app service is started via `make`, the app service configuration file is copied to `edgecentral-app-service-config:/res/javascript-http-export/configuration.toml`.

A ports key has also been added to the service to make it easier to trigger the app service directly from the host.

See the [the Edge Central HTTP Export Example repo](https://github.com/andresrinivasan/edge-central-http-export-example) for more details about the sample app service specifically.

### Creating Portainer Stacks

A Portainer Stack is ["...a collection of services, usually related to one application or usage"](https://docs.portainer.io/user/docker/stacks). Among other ways to define a Portainer Stack is to use a Docker Compose project. In the _Makefile_ the project is inferred from the Docker Compose File for Edge Central (i.e. edgexpert) and declared by setting a project for Alpine.

## v.Next

This demontration is complete though there are some remaining cleanup tasks:

- App service network assumes edgexpert_edgex-network but network is created based on PROJECT?
- Preset Portainer admin password
- Alpine down rather than kill
- Cleanup WARN[0000] mount of type `volume` should not define `bind` option
