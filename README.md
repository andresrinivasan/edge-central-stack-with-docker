# simple-edge-central-stack

- Should app-configurable exit if there is no configuraiton file? See EDGEX_PROFILE?
- App service network assumes edgexpert_edgex-network but network is created based on PROJECT?
- Preset Portainer admin password
- Alpine down rather than kill
- Cleanup WARN[0000] mount of type `volume` should not define `bind` option

- pkg-config and 0mq are required (note Go binding is via pebbe/zmq4 which creates the dependency on libzmq)
  - brew install pkg-config zmq
