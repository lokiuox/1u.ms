# 1u.ms

## What is this

This is a small set of zero-configuration DNS utilities for assisting in detection and exploitation of SSRF-related vulnerabilities. It provides easy to use DNS rebinding utility, as well as a way to get resolvable resource records with any given contents.

The tool does not employ any novel techniques and is not unique in any sense. All features are trivial to implement and can be easily found in other similar tools.

The service is currently run at the [1u.ms](http://1u.ms/) domain (and its' subdomains).

#### Source code & self-host

The source is available on [github](https://github.com/neex/1u.ms). However, the code is shitty as hell.

If you want to self-host it, follow this steps:

1. Get a domain name of your choice and put NS records so that your server serves that domain.
2. Perform `go get github.com/neex/1u.ms`.
3. Download and modify [1u.ms.yaml](https://github.com/neex/1u.ms/blob/master/1u.ms.yaml).
4. Run like this: `1u.ms 1u.ms.yaml`.

## Usage

#### A-record

If you want to get a record that resolves to an IP, use the following subdomain:

`make-<IP>-rr.1u.ms`

For example, domain `make-1.2.3.4-rr.1u.ms` resolves to `1.2.3.4`:

```shell
$ host -t A make-1.2.3.4-rr.1u.ms
make-1.2.3.4-rr.1u.ms has address 1.2.3.4
```

You can use dashes instead of dots as long as the IP is valid:

```shell
$ host -t A make-1-2-3-4-rr.1u.ms
make-1-2-3-4-rr.1u.ms has address 1.2.3.4
```

You can place some unique prefix/suffix before `make` or after `rr` (dots are allowed):

```shell
$ host -t A a.prefix-make-1-2-3-4-rr-and.a-suffix.1u.ms
a.prefix-make-1-2-3-4-rr-and.a-suffix.1u.ms has address 1.2.3.4
```

Multiple records can be separated by `-and-`:

```shell
$ host -t A make-1-2-3-4-and-5-6-7-8-rr.1u.ms
make-1-2-3-4-and-5-6-7-8-rr.1u.ms has address 1.2.3.4
make-1-2-3-4-and-5-6-7-8-rr.1u.ms has address 5.6.7.8
```

#### DNS rebinding

In the context of SSRF bugs, DNS rebinding is a well-known technique targeting TOCTOU type of vulnerabilities during IP blacklisting or whitelisting. It is performed using a domain that resolves in a legit IP during the first request (check) and to the forbidden one during the second request (use).

To generate a domain name with this behavior, use the following syntax:

`make-<IP1>-rebind-<IP2>-rr.1u.ms`

For example, the domain name `make-1.2.3.4-rebind-169.254-169.254-rr.1u.ms` will first resolve to `1.2.3.4` and then to `169.254.169.254`:

```shell
$ host -t A make-1.2.3.4-rebind-169.254-169.254-rr.1u.ms
make-1.2.3.4-rebind-169.254-169.254-rr.1u.ms has address 1.2.3.4
$ host -t A make-1.2.3.4-rebind-169.254-169.254-rr.1u.ms
make-1.2.3.4-rebind-169.254-169.254-rr.1u.ms has address 169.254.169.254
```

The logic behind the feature is as follows:
* if there were no requests to this domain during last 5 seconds, it's resolved to the first IP;
* otherwise, it's resolved to the second one.

You can use prefixes before `make-` and suffix after `-rr` in order to uniqualize the domain name (e.g. `prefix-make-1.2.3.4-rebind-169.254-169.254-rr-suffix.1u.ms`). The timeouts are separate for each domain name.

If you need to change the default 5 seconds timeout, use the following syntax:

`make-<IP1>-rebindfor<interval>-<IP2>-rr.1u.ms`

where `<interval>` is something like `10s` (10 seconds) or `5m` (5 minutes).

If you need that "whitelisted" IP (which is IP1 in our examples) be returned multiple times before rebinding, use the following syntax:

`make-<IP1>-rebindfor<interval>after<num>times-<IP2>-rr.1u.ms`

For example, `make-1.2.3.4-rebindfor30safter2times-127.0.0.1-rr.1u.ms` will resolve in `1.2.3.4` first two times, and then will resolve in `127.0.0.1` for next 30 seconds.

#### AAAA-record

To make up a domain that resolves only to an IPv6 address, use the following syntax:

`make-ip-v6-<IP>-rr.1u.ms`

Colons must be replaced with letter `l`. As always, random prefix and suffix can be used:

```shell
$ host -t AAAA prefix-make-ip-v6-1l2ll3-rr-suffix.1u.ms
prefix-make-ip-v6-1l2ll3-rr-suffix.1u.ms has IPv6 address 1:2::3
```

#### CNAMEs

By default, unparsable addresses are considered as CNAMEs:

```shell
$ host make-example.com-rr.1u.ms
make-example.com-rr.1u.ms is an alias for example.com.
...
```

To force a domain to be a CNAME, add `cname-` prefix:

```shell
$ host -t A make-cname-example.com-rr.1u.ms
make-cname-example.com-rr.1u.ms is an alias for example.com.
...
```

#### Other record types

If the thing between `make-` and `-rr` is a parsable record, it is returned for any type of request.

```shell
$ host -t TXT make-blahblah-rr.1u.ms
make-blahblah-rr.1u.ms descriptive text "blahblah"
```

#### Hex encoding

You can encode the contents of a record in hex and add a `hex-` prefix after `make-`:

```shell
$ host -t A make-hex-312e322e332e34-rr.1u.ms
make-hex-312e322e332e34-rr.1u.ms has address 1.2.3.4
```

#### Note on DNS TTLs

Some servers don't want to handle zero TTL replies. Default TTL is 1 for "service" domains and 0 for others.

If you want to change TTL, add `set-<number>-ttl` anywhere in the domain name.

#### Log viewing

The log of all DNS requests is public. There are the following endpoints:

* [http://1u.ms/last](http://1u.ms/last) — gives last 100 requests
* http://1u.ms/log — an infinite loading page with current log records, like `tail -f`. Intended usage is running `curl http://1u.ms/log` in a terminal while doing your experiments.
* [http://1u.ms/log?grep=`<regexp>`](http://1u.ms/log?grep=<regexp>) — same as above, but show only matching lines

## Contacts & FAQ

If you have any questions or suggestions in mind, feel free to contact me via [@neexemil](https://t.me/neexemil) on Telegram or [@emil_lerner](https://twitter.com/emil_lerner) on Twitter.

#### Is this tool free for any type of usage?

Sure.

#### But what if I use it during some illegal adventure or DDoS it with a huge amount of traffic?

These are awful ideas which I don't like.

#### I've read the code and have concluded that you're a noob. It is the shittiest program ever.

Sorry.
