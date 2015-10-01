package fluentd_forwarder

import (
	basiclogger "../../basiclogger"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"gopkg.in/vmihailenco/msgpack.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"regexp"
	"time"
)

const module = "fluentd_forwarder"

type Fluentd_forwarder struct {
	tag                           string
	c                             chan *basiclogger.Message
	conn                          net.Conn
	SSLCertificate, SSLKey, SSLCA string
	tlsConfig                     tls.Config
	tlsEnabled                    bool
	host                          string
	hosts                         []string
	timeout                       time.Duration
}

var hostname string
var hostport_re, _ = regexp.Compile("^(.+):([0-9]+)$")

func init() {
	hostname, _ = os.Hostname()
	rand.Seed(time.Now().UnixNano())
}

func Init(conf map[string]interface{}) *Fluentd_forwarder {
	if hosts_raw, ok := conf["hosts"]; !ok || len(hosts_raw.([]interface{})) == 0 {
		log.Fatal("[", module, "] You must specify hosts")
	} else {
		var hosts []string
		for _, hostport_raw := range hosts_raw.([]interface{}) {
			hostport := hostport_raw.(string)
			submatch := hostport_re.FindSubmatch([]byte(hostport))
			if submatch == nil {
				log.Printf("[%s] Invalid host:port given: %s", module, hostport)
			} else {
				hosts = append(hosts, hostport)
			}
		}
		if len(hosts) == 0 {
			log.Fatal("[", module, "] There is no valid hosts")
		} else {
			if timeout, ok := conf["timeout"]; !ok || timeout.(int64) <= 0 {
				log.Fatal("[", module, "] You must specify right timeout")
			} else {
				var SSLCertificate, SSLKey, SSLCA string
				if cert, ok := conf["ssl_cert"]; ok && len(cert.(string)) > 0 {
					SSLCertificate = cert.(string)
					if key, ok := conf["ssl_key"]; ok && len(key.(string)) > 0 {
						SSLKey = key.(string)
					}
				}
				if ca, ok := conf["ssl_ca"]; ok && len(ca.(string)) > 0 {
					SSLCA = ca.(string)
				}
				tag, _ := conf["tag"].(string)
				res := Fluentd_forwarder{
					tag:            tag,
					c:              make(chan *basiclogger.Message),
					conn:           nil,
					hosts:          hosts,
					SSLCertificate: SSLCertificate,
					SSLKey:         SSLKey,
					SSLCA:          SSLCA,
					timeout:        time.Second * time.Duration(timeout.(int64))}
				res.loadCerts()
				return &res
			}
		}
	}
	return nil
}

func (v Fluentd_forwarder) Tag() string {
	return v.tag
}

//TODO: SSL support
func (v Fluentd_forwarder) Send(message *basiclogger.Message) error {
	var err error
	now := time.Now().UnixNano()
	for {
		if v.conn == nil {
			if !v.connect() {
				if v.conn != nil {
					v.conn.Close()
				}
				time.Sleep(time.Second)
				continue
			}
		}
		message.Data["host"] = hostname
		val := []interface{}{v.tag, now, message.Data}
		m, err := msgpack.Marshal(val)
		if err != nil {
			fmt.Printf("[%s] Bogus message: %v", v.tag, message.Data)
			break
		}
		v.conn.SetDeadline(time.Now().Add(v.timeout))
		_, err = v.conn.Write(m)
		if err != nil {
			log.Printf("[%s] Socket write error %v", v.tag, err)
			v.conn.Close()
			time.Sleep(time.Second)
			continue
		}
		break
	}
	return err
}

func (v Fluentd_forwarder) connect() bool {
	hostport := v.hosts[rand.Int()%len(v.hosts)]
	submatch := hostport_re.FindSubmatch([]byte(hostport))
	host := string(submatch[1])
	port := string(submatch[2])
	ips, err := net.LookupIP(host)

	if err != nil {
		log.Printf("[%s] DNS lookup failure \"%s\": %s", v.tag, host, err)
		return false
	}

	ip := ips[rand.Int()%len(ips)]
	var addressport string

	address := ip.String()
	if len(ip) == net.IPv4len {
		addressport = fmt.Sprintf("%s:%s", address, port)
	} else if len(ip) == net.IPv6len {
		addressport = fmt.Sprintf("[%s]:%s", address, port)
	}

	log.Printf("[%s] Connecting to %s (%s) \n", v.tag, addressport, host)

	conn, err := net.DialTimeout("tcp", addressport, v.timeout)
	if err != nil {
		log.Printf("[%s] Failure connecting to %s: %s\n", v.tag, address, err)
		return false
	} else {
		v.conn = conn
		v.host = host
	}
	if v.tlsEnabled {
		v.tlsConfig.ServerName = host

		connTLS := tls.Client(v.conn, &v.tlsConfig)
		connTLS.SetDeadline(time.Now().Add(v.timeout))
		err = connTLS.Handshake()
		if err != nil {
			log.Printf("[%s] Failed to tls handshake with %s %s\n", v.tag, address, err)
			connTLS.Close()
			return false
		} else {
			v.conn = connTLS
		}
	}
	log.Printf("[%s] Connected to %s\n", v.tag, address)
	return true
}

func (v Fluentd_forwarder) loadCerts() {
	v.tlsEnabled = false
	v.tlsConfig.MinVersion = tls.VersionTLS10
	if len(v.SSLCertificate) > 0 && len(v.SSLKey) > 0 {
		log.Printf("[%s] Loading client ssl certificate and key from \"%s\" and \"%s\"\n", v.tag,
			v.SSLCertificate, v.SSLKey)
		cert, err := tls.LoadX509KeyPair(v.SSLCertificate, v.SSLKey)
		if err != nil {
			log.Fatalf("[%s] Failed loading client ssl certificate: %s\n", v.tag, err)
		}
		v.tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if len(v.SSLCA) > 0 {
		log.Printf("[%s] Loading CA certificate from file: %s\n", v.tag, v.SSLCA)
		v.tlsConfig.RootCAs = x509.NewCertPool()

		pemdata, err := ioutil.ReadFile(v.SSLCA)
		if err != nil {
			log.Fatalf("[%s] Failure reading CA certificate: %s\n", v.tag, err)
		}

		block, _ := pem.Decode(pemdata)
		if block == nil {
			log.Fatalf("[%s] Failed to decode PEM data of CA certificate from \"%s\"\n", v.tag, v.SSLCA)
		}
		if block.Type != "CERTIFICATE" {
			log.Fatalf("[%s] This is not a certificate file: %s\n", v.tag, v.SSLCA)
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			log.Fatalf("[%s] Failed to parse CA certificate: %s\n", v.tag, v.SSLCA)
		}
		v.tlsConfig.RootCAs.AddCert(cert)
	}
	v.tlsEnabled = true
}
