# logear

[![Build status](https://api.travis-ci.org/DLag/logear.png)](https://travis-ci.org/DLag/logear)

(log ear, log gear)

Logging system designed to be fast, reliable and flexible but in same time simple.

## Purpose

Logear can grab structurised messages from multiple inputs and deliver it to many destinations.
It can replace huge and slow systems like Logstash and Fluentd.
Logear writen in Go and don't needs any specific environment.

## Inputs

### (filetail) File input with json, messagepack or custom formats
Reads line by line from file, parses it with json lib or with custom regexp.

## Outputs

### (fluentd_forwarder) Fluentd_forwarder network protocol
Delivers messages to Fluent with native protocol. Supports messagepack and json output encoding.
