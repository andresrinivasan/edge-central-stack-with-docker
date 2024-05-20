# Edge Central Stack with Docker

The purpose of this repo is to demonstrate how to manage an Edge Central stack using Docker rather than using the Edge Central CLI. You might consider this strategy if you are managing Edge Central containers alongside additional project containers and want a single operational interface.

## Prerequisites

- [IOTech Edge Central Software](https://www.iotechsys.com/products/edge-central/edge-central-installer-download/) Installed
- A valid (or [eval](https://www.iotechsys.com/resources/evaluating-the-software/edge-central-evaluation/)) license
- [docker and docker compose](https://www.docker.com/) daemon and commands (or [equivalent](https://github.com/abiosoft/colima))
- [httpie](https://httpie.io/) command (aka better curl)
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
| edgexpert up app-service --path=... | docker compose -f app-service.yaml... up | start-app-service |
| edgexpert down app-service --path=... | docker compose -f app-service.yaml... down | stop-app-service |

## Additional Information

Between passing arguments to Docker Compose knowing the minimum microservices needed for Edge Central, there's a bit more complexity. It is assumed you have installed Edge Central and have the Docker Compose files that come with the distribution.

## Set Your Environment Variables

You will need to set the following environment variables:

```sh
- LICENSE_FILE
- COMPOSE_FILE
- EDGEXPERT_IMAGE_REPO
- EDGEXPERT_IMAGE_VERSION
```

You can export them to your environment or add them to a `.env` style file and point to the `.env` file using the ENV environment variable. For example, if you put the following in make.env

```sh
LICENSE_FILE=/Users/andre/iotech-edge-central/2.3.2/EdgeXpert_Andre_Evaluation.lic 
COMPOSE_FILE=/Users/andre/iotech-edge-central/2.3.2/etc/edgexpert/docker-compose.yml:/Users/andre/iotech-edge-central/2.3.2/etc/edgexpert/docker-compose-port-mapping.yml:/Users/andre/repos/edge-central-stack-with-docker/core-data-override.yaml
EDGEXPERT_IMAGE_REPO=edgexpert
EDGEXPERT_IMAGE_VERSION=2.3
```

invoking `ENV=make.env make ...` will set up your environment for the make targets. Since `make` will use these environment variables, you can also export them from your shell:

```sh
export $(xargs <make.env)
make
```

> NOTE you will need to use absolute paths in the _env_ file as there is no shell expansion.
>
> FURTHER NOTE the assumption path assumption that Edge Central was installed to
> _/Users/andre/iotech-edge-central/2.3.2/_ and the license key file is located in
> _/Users/andre/iotech-edge-central/2.3.2/EdgeXpert_Andre_Evaluation.lic_. I suspect you'll need to tweak this for 
> your environment.

### Edge Central Service Environment Variables

Edge Central services can be configured with [environment variables](https://docs.iotechsys.com/edge-xpert23/core-services/core-config.html). For example, add `core-data-override.yaml` to the COMPOSE_FILE environment variable to disable data persistence.

I'm going to assume you're following this pattern.

### Adding the License Key

You must have a valid license key to start Edge Central. All IOTech microservices look for the license key in a Docker volume named _license-data_ with the path _/lic/LICENSE_FILE.lic_. An evaluation license can be requested via <https://www.iotechsys.com/resources/evaluating-the-software/edge-central-evaluation/>.

#### Example

```sh
make add-license
```

### Starting Edge Central

There is a common set of microservices that are more or less needed for Edge Central (see `EDGE_CENTRAL_COMMON_SERVICES` in the _Makefile_). For example, if events are not being stored in Redis, there is no need to start Redis. Those common services are added to the underlying Docker Compose command when using the _edgexpert_ command.

There are additional services that may be needed, for example Portainer. In the _Makefile_ these are added to `EDGE_CENTRAL_ADDITIONAL_SERVICES`. Please see the [IOTech Services](https://docs.iotechsys.com/edge-xpert23/cli/cli-services.html) documentation for an enumeration of the services included with Edge Central.

> Please remember this is where COMPOSE_FILE comes into play. This is not a search path; `Docker Compose` will merge Compose files enumerated **either** in the COMPOSE_FILE environment variable **or** passed via `-f`.

#### Example

Assuming the `COMPOSE_FILE` environment variable has been set as above

```sh
make start-edge-central
```

### Starting an App Service

This is a little more complex as overrides of the Edge Central `app-service.yml` are needed along with a couple of support files.

#### Override `app-service.yml`

The app-service.yml that comes with Edge Central requires a couple of overrides. See [the Edge Central HTTP Export Example repo](https://github.com/andresrinivasan/edge-central-http-export-example) for the source. First I felt that a Docker Volume was more Docker'ish rather than a [bind mount](https://docs.docker.com/storage/bind-mounts/) for the app service configuration file.

App services expect to find their configuration in res/APP-SERVICE-NAME/configuration.toml where APP-SERVICE-NAME is...wait for it...the name of the app service. The override Docker Compose file (`app-service-override.yaml`) includes an external volume dependency on `edgecentral-app-service-config` which is mounted as `/res`. This overrides the bind mount in `app-service.yml` which mounts `~/.edgexpert/...` as `/res`. When the app service is started (see `make start-app-service`), the app service configuration file is copied to `edgecentral-app-service-config:/res/javascript-http-export/configuration.toml`.

A ports key has also been added to the service to make it easier to trigger the app service directly from the host.

See the [the Edge Central HTTP Export Example repo](https://github.com/andresrinivasan/edge-central-http-export-example) for more details about the sample app service specifically.

#### Example

```sh
COMPOSE_FILE=${COMPOSE_FILE}:~/iotech-edge-central/2.3.2/etc/edgexpert/app-service.yml make start-app-service
```

### Creating Portainer Stacks

A Portainer Stack is ["...a collection of services, usually related to one application or usage"](https://docs.portainer.io/user/docker/stacks). Among other ways to define a Portainer Stack is to use a Docker Compose project. In the _Makefile_ the project is inferred from the Docker Compose File for Edge Central (i.e. edgexpert) and declared by setting a project for Alpine.

## v.Next

This demontration is complete though there are some remaining cleanup tasks:

- App service network assumes edgexpert_edgex-network but network is created based on PROJECT?
- Preset Portainer admin password
- Alpine down rather than kill
- Cleanup WARN[0000] mount of type `volume` should not define `bind` option
