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
- `skip_own_ip`: Defaults to `true`. If set to `false`, this will handle all messages which are also send by the own list and should only be used in special environments.

### Installing on OpenWRT

The installation process on OpenWRT requires you to know the processor architecture of the device and pick out the correct executable.
You can test this by downloading from the [releases](https://github.com/cbrand/mdnsforwarder/releases) tab the executable and see if it runs. You require ssh access to the device for it to work.

To for example test if your device is arm64 compatible do the following:

```bash
wget -O mdnsforwarder https://github.com/cbrand/mdnsforwarder/releases/download/1.1.0/mdnsforwarder-arm64
chmod +x /root/mdnsforwarder
/root/mdnsforwarder --help
```

If you do not see an error message you have the right architecture.

Other more likely candidates for your router are

- mips
```bash
wget -O mdnsforwarder https://github.com/cbrand/mdnsforwarder/releases/download/1.1.0/mdnsforwarder-mips
chmod +x /root/mdnsforwarder
/root/mdnsforwarder --help
```

- mipsle
```bash
wget -O /root/mdnsforwarder https://github.com/cbrand/mdnsforwarder/releases/download/1.1.0/mdnsforwarder-mipsle
chmod +x /root/mdnsforwarder
/root/mdnsforwarder --help
```
Once you figure this out make a note where you have downloaded the executable and move it - if desired - to another location.

Afterwards you need to install it as a service. Add a new file in `/etc/init.d/mdnsforwarder` and add the following content:

```bash
#!/bin/sh /etc/rc.common
USE_PROCD=1
START=95
STOP=01
start_service() {
    procd_open_instance
    # The /root/mdnsforwarder must be adjusted to the folder where you downloaded it in.
    procd_set_param command /root/mdnsforwarder
    procd_close_instance
}
```

Additionally, make the script executable via `chmod +x /etc/init.d/mdnsforwarder` in your command line.

Make sure that you have the configuration of the mdnsforwarder added to `/etc/config/mdnsforwarder` for example with:

```json
{
    "mdnsInterfaces": [
        "eth0",
        "eth1"
    ],
    "listeners": [
        "0.0.0.0:53531"
    ],
    "targets": [
        "remote:53531"
    ]
}
```

Afterwards enable and start the service:

```bash
/etc/init.d/mdnsforwarder enable
/etc/init.d/mdnsforwarder start
```

On restart the mdnsforwarder will now start automatically. You can also verify the process to run without a reboot by checking if it is there:

```bash
ps | grep mdnsforwarder
```

## Availability

- The program is pulished as a self running go-executable for various platforms in the
  [releases tab](https://github.com/cbrand/mdnsforwarder/releases).
- To run it in a docker container, the image is published in Docker Hub with the image
  [`cbrand/mdnsforwarder`](https://hub.docker.com/r/cbrand/mdnsforwarder).

## Caveats

The remote MDNS traffic forwarding doesn't have any security. It doesn't encrypt any traffic and relies on protection from a lower layer (for example by utilizing a VPN). DO NOT USE THIS DIRECTLY OVER THE OPEN INTERNET.

## License

The application is published via the [MIT license](LICENSE).
