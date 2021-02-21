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
        # how many scripts are allowed to run at once? 0 for no limit:
        max-concurrency: 64
```

Here's an example of how to test your configuration from a shell:

```bash
# 1.1.1.1 should not be on any blocklists:
echo '{"ip": "1.1.1.1"}' | ./oragono-dnsbl ./config.yaml
# expected output:
# {"result":1,"banMessage":"","cacheNet":"","cacheSeconds":0,"error":""}

# a Tor exit node that should be blocked by our example config file;
# see https://www.dan.me.uk/torlist/?exit for an up-to-date list of exit nodes
echo '{"ip": "103.253.41.98"}' | ./oragono-dnsbl ./config.yaml
# expected output:
# {"result":3,"banMessage":"You need to enable SASL to access this network while using TOR","cacheNet":"","cacheSeconds":0,"error":""}
```
