package in_logear_forwarder

import (
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"github.com/DLag/logear/basiclogger"
	"gopkg.in/vmihailenco/msgpack.v2"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"regexp"
	"time"
)

const module = "in_logear_forwarder"

type In_logear_forwarder struct {
	tag          string
	messageQueue chan *basiclogger.Message
	tlsConfig    tls.Config
	bind         string
	timeout      time.Duration
}

var hostport_re, _ = regexp.Compile("^(.+):([0-9]+)$")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Init(messageQueue chan *basiclogger.Message, conf map[string]interface{}) *In_logear_forwarder {
	var SSLCertificate, SSLKey, SSLCA string
	var tlsConfig tls.Config
	var timeout int64
	tag := conf["tag"].(string)
	bind := conf["bind"].(string)
	if timeout_raw, ok := conf["timeout"]; !ok || timeout_raw.(int64) <= 0 {
		log.Fatal("[", module, "] You must specify right timeout")
	} else {
		timeout = timeout_raw.(int64)
	}
	if cert, ok := conf["ssl_cert"]; ok && len(cert.(string)) > 0 {
		SSLCertificate = cert.(string)
		if key, ok := conf["ssl_key"]; ok && len(key.(string)) > 0 {
			SSLKey = key.(string)
			if len(SSLCertificate) > 0 && len(SSLKey) > 0 {
				tlsConfig.MinVersion = tls.VersionTLS12
				log.Printf("[%s] Loading server ssl certificate and key from \"%s\" and \"%s\"", tag,
					SSLCertificate, SSLKey)
				cert, err := tls.LoadX509KeyPair(SSLCertificate, SSLKey)
				if err != nil {
					log.Fatalf("[%s] Failed loading server ssl certificate: %s", tag, err)
				}
				tlsConfig.Certificates = []tls.Certificate{cert}
				if ca, ok := conf["ssl_ca"]; ok && len(ca.(string)) > 0 {
					SSLCA = ca.(string)
					if len(SSLCA) > 0 {
						log.Printf("[%s] Loading CA certificate from file: %s\n", tag, SSLCA)
						tlsConfig.ClientCAs = x509.NewCertPool()
						tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
						pemdata, err := ioutil.ReadFile(SSLCA)
						if err != nil {
							log.Fatalf("[%s] Failure reading CA certificate: %s\n", tag, err)
						}

						block, _ := pem.Decode(pemdata)
						if block == nil {
							log.Fatalf("[%s] Failed to decode PEM data of CA certificate from \"%s\"\n", tag, SSLCA)
						}
						if block.Type != "CERTIFICATE" {
							log.Fatalf("[%s] This is not a certificate file: %s\n", tag, SSLCA)
						}

						cacert, err := x509.ParseCertificate(block.Bytes)
						if err != nil {
							log.Fatalf("[%s] Failed to parse CA certificate: %s\n", tag, SSLCA)
						}
						tlsConfig.ClientCAs.AddCert(cacert)
					}
				}
				v := &In_logear_forwarder{tag: tag,
					messageQueue: messageQueue,
					tlsConfig:    tlsConfig,
					bind:         bind,
					timeout:      time.Second * time.Duration(timeout)}
				return v
			}
		}
	}

	return nil
}

func (v *In_logear_forwarder) Tag() string {
	return v.tag
}

func (v *In_logear_forwarder) Listener() {
	go v.listen()
}

func (v *In_logear_forwarder) listen() {
	listener, err := tls.Listen("tcp4", v.bind, &v.tlsConfig)
	if err != nil {
		log.Fatalf("[%s] Can't start listen \"%s\", error: %v", v.tag, v.bind, err)
	}
	defer listener.Close()
	log.Printf("[%s] Waiting for connections", v.tag)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[%s] Can't accept client %v", v.tag, err)
		}
		go v.worker(conn)
	}
}

func (v *In_logear_forwarder) worker(conn net.Conn) {

	log.Printf("[%s] Accepted connection from %s", v.tag, conn.RemoteAddr().String())
	for {
		conn.SetReadDeadline(time.Now().Add(v.timeout))
		csize, err := v.readInt64(conn)
		if err != nil {
			log.Printf("[%s] Can't read size of compressed payload, closing connection, error: %v", v.tag, err)
			conn.Close()
			return
		}
		size, err := v.readInt64(conn)
		if err != nil {
			log.Printf("[%s] Can't read size of uncompressed payload, closing connection, error: %v", v.tag, err)
			conn.Close()
			return
		}
		log.Printf("[%s] Waiting for %d bytes of payload", v.tag, csize)
		cpayload := make([]byte, int(csize))
		n, err := conn.Read(cpayload)
		if err != nil || int64(n) != csize {
			log.Printf("[%s] Can't read compressed payload, closing connection, error: %v", v.tag, err)
			conn.Close()
			return
		}
		bpayload := bytes.NewReader(cpayload)
		zpayload, err := zlib.NewReader(bpayload)
		if err != nil {
			log.Printf("[%s] Can't start zlib handler, closing connection, error: %v", v.tag, err)
			conn.Close()
			return
		}
		payload := make([]byte, size)
		n, err = zpayload.Read(payload)
		if err != nil || int64(n) != size {
			log.Printf("[%s] Can't uncompress payload, closing connection, error: %v", v.tag, err)
			conn.Close()
			return
		}
		var data map[string]interface{}
		err = msgpack.Unmarshal(payload, &data)
		if err != nil {
			log.Printf("[%s] Can't parse payload error: %v", v.tag, err)
			conn.Close()
			return
		}
		if _, ok := data["@timestamp"]; !ok {
			data["@timestamp"] = time.Now()
		}
		v.messageQueue <- &basiclogger.Message{Time: time.Now(), Data: data}
	}
}

func (v *In_logear_forwarder) readInt64(r io.Reader) (int64, error) {
	var data int64
	err := binary.Read(r, binary.BigEndian, &data)
	if err != nil {
		return 0, err
	}
	return data, nil
}
