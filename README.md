oragono-dnsbl
=============

This is a DNSBL client for use in Oragono as an IP-checking script.

To build the plugin, [install Go 1.14 or higher](https://golang.org/dl), then run `make`; this will build an `oragono-dnsbl` binary.

See `config.yaml` for an example of how to configure the plugin.

To configure oragono to use this plugin, add a section like this to your `server` block:

```yaml
    ip-check-script:
        enabled: true
        command: "/path/to/oragono-dnsbl"
        # constant list of args to pass to the command; the actual query
        # and result are transmitted over stdin/stdout:
        args: ['/path/to/config.yaml']
        # timeout for process execution, after which we send a SIGTERM:
        timeout: 4s
        # how long after the SIGTERM before we follow up with a SIGKILL:
        kill-timeout: 1s
```
