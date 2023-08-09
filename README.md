# Simple Edge Central Stack

The purpose of this repo is to demonstrate how to manage an Edge Central stack without using the Edge Central CLI. This demonstration is just that; you may or may not want this particular set of services in your own environment.

## Demonstration

The Edge Central CLI is a wrapper around Docker and Docker Compose and exists to simplify starting/stopping the Edge Central microservices. I use simple targets in a Makefile to demonstrate what Docker commands are being used.

| Edge Central Command | Docker Command | Make Target |
| --- | --- | --- |
| edgexpert license install | N/A | add-license |
| edgexpert up | docker compose up ... | start-edge-central |
| edgexpert down | docker compose down ... | stop-edge-central |
| edgexpert gen | docker compose config ... | render-canonical-docker-compose |

## Additional Information

Between passing arguments to Docker Compose knowing the minimum microservices needed for Edge Central, there's a bit more complexity.

### Adding the License Key

You must have a valid license key to start Edge Central. All IOTech microservices look for the license key in a Docker volume named _license-data_ with the path _/lic/LICENSE_FILE.lic_. An evaluation license can be requested via <https://www.iotechsys.com/resources/evaluating-the-software/edge-central-evaluation/>.

### Starting Edge Central

There is a common set of microservices that are more or less needed for Edge Central (see `EDGE_CENTRAL_COMMON_SERVICES` in the _Makefile_). For example, if events are not being stored in Redis, there is no need to start Redis. Those common services are added to the underlying Docker Compose command when using _edgexpert_.

There are additional services that may be needed, for example Portainer. In the _Makefile_ these are added to `EDGE_CENTRAL_ADDITIONAL_SERVICES`. Please see the [IOTech Services](https://docs.iotechsys.com/edge-xpert23/cli/cli-services.html) documentation for an enumeration of the services included with Edge Central.

### Creating Portainer Stacks

A Portainer Stack is ["...a collection of services, usually related to one application or usage"](https://docs.portainer.io/user/docker/stacks). Among other ways to define a Portainer Stack is to use a Docker Compose project. In the _Makefile_ the project is inferred from the Docker Compose File for Edge Central (i.e. edgexpert) and declared by setting a project for Alpine.

## v.Next

This demontration is complete though there are some remaining cleanup tasks:

- Should app-configurable exit if there is no configuraiton file? See EDGEX_PROFILE?
- App service network assumes edgexpert_edgex-network but network is created based on PROJECT?
- Preset Portainer admin password
- Alpine down rather than kill
- Cleanup WARN[0000] mount of type `volume` should not define `bind` option
