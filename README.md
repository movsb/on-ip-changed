# on-ip-changed

On-IP-Changed is a small utility that periodically gets the IP of the system running this program and invokes handlers to notify what the change is.

Support both IPv4 and IPv6.

## Config File

Full configuration example:

```yaml
daemon:
  # tasks execution interval
  interval: 1m
  # if multiple getters, at most `concurrency`
  # randomly selected getters will be used.
  concurrency: 1
  # timeout out for getter and handler execution.
  timeout: 15s
  # notify the first IP address?
  # The first is what we get when we restart the daemon. if restarted 
  # frequently, handlers will be executed frequently too, with the same IP.
  initial: false

tasks:
  - name: test
    # if multiple getters, handlers will be executed if and only if when
    # a majority ( > a half ) of the getters return the same IP address.
    getters:
      - type: domain
        domain: example.com
    handlers:
      - shell: 
          command: echo --- $IP ---
    ipv6only: true
    ipv4only: true
```

### (Environment) Variables

- $IP
- $IPv4
- $IPv6

## Concepts

* **Tasks** Each contains *getters* and *handlers*.
* **Getter** A getter is an IP getter, which gets a kind of IP for localhost, website, router, domain, etc..
* **Handler** A handler is what should be doing when we get an IP.

You can write your own getters and handlers easily.

## Getters

### Domain

**Domain** getter gets the IP of a domain.

This can be useful, for example, when you use dynamic DNS and you want to [update the WireGuard peer's Endpoint][resolve-endpoint], since it [doesn't update][wg-dns] automatically.

[resolve-endpoint]: https://github.com/WireGuard/wireguard-tools/tree/master/contrib/reresolve-dns
[wg-dns]: https://lists.zx2c4.com/pipermail/wireguard/2017-November/002028.html

Example configuration:

```yaml
type: domain
domain: home.example.com
```

### Website

**Website** getter gets your network's outbound IP address reported by a website. This is usually used to get your public IP address, which can be used to update your DDNS record.

Example configuration:

```yaml
type: website
url: domain.to.get.my.ip.address.example.com
format: json
path: ip
```

#### Format

**Format** specifies what content type is returned from that website and how we should parse the content to get the IP address.

Can be one of:

* **text**

  The content is plain text IP address.

  `path` is not used.

* **json**

  The content is a JSON object containing a field specifying the IP.

  Use `path` to specify the path reaching to that field.

  For example, a JSON with this content:

  ```json
  {
    "data": {
      "ip": "1.1.1.1"
    }
  }
  ```

  Then, the `path` should be `data.ip`.

* **search**

  This enables searching for the first IP address in the content using a regexp matching a single IPv4 address.

  `path` is not used.
  
  **search** currently doesn't work for IPv6 addresses.
  
#### List

Some example websites which can give you your IP address:

```yaml
- type: website
  url: https://ifconfig.co/ip
  format: text
- type: website
  url: https://wtfismyip.com/text
  format: text
- type: website
  url: https://ip.cn/api/index?ip=&type=0
  format: json
  path: ip
- type: website
  url: https://myip.ipip.net/
  format: search
- type: website
  url: https://myip.com.tw/
  format: search
- type: website
  url: http://ip-api.com/line/?fields=query
  format: text
- type: website
  url: https://ip.sendev.cc
  format: text
- type: website
  url: https://api.ipify.org/
  format: text
```

### Ifconfig

**Ifconfig** gets the IP address of an interface by its name.

Example configuration:

```yaml
type: ifconfig
name: eth0
```

### Asus

**Asus** gets the WAN IP address of the Asus router family (not well tested).

Example configuration:

```yaml
type: asus
address: 192.168.1.1
username: asus
password: asus
```

Getting the outbound IPv4 address from routers can be useful when you use VPN.
Because `website` will report the IP address of your VPN server, which you mostly cannot control its port forwarding rules.

## Handlers

### Shell

**Shell** handler executes a shell command with IP passed by environment variable `IP`.

Example configuration:

```yaml
shell:
  command: echo $IP

  # or

  command: |
    #!/bin/bash

    set -eu

    echo IP: $IP

  # or

  command:
    - my_ddns_updater
    - -c
    - --long-option
    - $IP
```

Additional arguments can be set:

* **shell**

  Specify what shell will be used when `command` is a string instead of an array.

  ```yaml
  shell: fish
  ```

* **env**

  Additional environment variables.

  ```yaml
  env:
    key: value
    foo: bar
  ```

* **work_dir**

  ```yaml
  work_dir: ~/data/
  ```

  The working directory of the command.

  Home directory in style of `~` (no `~user`) in the `work_dir` will be expanded to their respective directories.

### HTTP

**HTTP** handler makes a request.

Example configuration:

```yaml
http:
  endpoint: http://example.com
  args:
    ip: $IP
  headers:
    token: ttt
  method: GET
  body: Your IP changed to $IP.
```

The result URL will be: <http://example.com/?ip=1.2.3.4> 。

For example, you can use [Chanify](https://github.com/chanify/chanify) to notify your latest IP address.

### DnsPod

**DnsPod** handler updates your DNS record of DnsPod.

Example configuration:

```yaml
dnspod:
  token: 123,abcdefgh
  email: your@example.com
  domain: example.com
  record: subdomain
```

There should have been more DNS provider handlers...
