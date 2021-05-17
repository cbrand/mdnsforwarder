# MDNSForwarder

This is an app which allows to forward MDNS messages between different networks. The primary intend is to forward MDNS queries
through unicast addresses (like a VPN network).

It has been succesfully tested to run on an OpenWRT router using the mips processor connecting to a docker container network in Kubernetes.

## Usage

```bash
mdnsforwarder --log-level DEBUG --config config.json
```

If no config path is specified the default (`/etc/config/mdnsforwarder`) is used.

Refer to the config.json for an example. The configuration options are:

- `mdnsInterfaces`: list of interfaces which should route mdnstraffic between each other
- `listeners`: Host/Port combination where unicast messages including mdns messages via UDP is expected
- `targets`: The list of targets where MDNS traffic from the local interfaces should be forwarded to.

## Availability

- The program is pulished as a self running go-executable for various platforms in the
  [releases tab](https://github.com/cbrand/mdnsforwarder/releases).
- To run it in a docker container, the image is published in Docker Hub with the image
  [`cbrand/mdnsforwarder`](https://hub.docker.com/r/cbrand/mdnsforwarder).

## Caveats

The remote MDNS traffic forwarding doesn't have any security. It doesn't encrypt any traffic and relies on protection from a lower layer (for example by utilizing a VPN). DO NOT USE THIS DIRECTLY OVER THE OPEN INTERNET.

## License

The application is published via the [MIT license](LICENSE).
