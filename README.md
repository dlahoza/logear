# logear

[![Build status](https://api.travis-ci.org/DLag/logear.png)](https://travis-ci.org/DLag/logear)

*(log ear, log gear)*

Logging system designed to be fast, reliable and flexible but in same time simple.

## Purpose

**Logear** can grab structured messages from multiple inputs and deliver it to many destinations.
It can replace huge and slow systems like Logstash and Fluentd.
Logear written in Go and doesn't require any specific environment.

## Build and Install

```
# go get -u github.com/DLag/logear
# mv $GOPATH/bin/logear /usr/local/sbin/logear
# logear -v
```

## Logear forwarder protocol

The protocol implements serialisation with MessagePack, zlib compression and SSL certificate
checks and encoding. This protocol can be more efficient than Lumberjack.

## Filters
Logear support JSON and custom regexp filters. Regexp implemented with [Google RE2 library](https://github.com/google/re2/).

## Inputs
- **filetail** - File input with json, MessagePack or custom filters
Reads line by line from a file, parses it with json, msgpack or with custom regexp.
- **out_logear_forwarder** - Fluentd_forwarder network protocol
Receive messages from another instance of Logear.

## Outputs

- **fluentd_forwarder** - Fluentd_forwarder network protocol
Delivers messages to Fluent with the native protocol. Supports messagepack and json output encoding.
- **out_logear_forwarder** - Fluentd_forwarder network protocol
Delivers messages to another instance of Logear.

## Examples
You can find example config in [/example](https://github.com/DLag/logear/tree/master/example) directory.