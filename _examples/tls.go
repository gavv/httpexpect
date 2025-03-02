package examples

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"path"
	"strconv"
)

/*
openssl ecparam -genkey -name secp384r1 -noout -out root.key

	openssl req -x509 -new -nodes -key root.key -sha256 -days 99999 -out root.pem \
		-subj "/C=US/ST=Default/L=Default/O=Default/OU=Root CA/CN=Root CA"

cat root.pem
*/
const rootPEM = `-----BEGIN CERTIFICATE-----
MIICYjCCAeigAwIBAgIUI0z85tUs0+AkSxegGbR8FIhDlLwwCgYIKoZIzj0EAwIw
ZzELMAkGA1UEBhMCVVMxEDAOBgNVBAgMB0RlZmF1bHQxEDAOBgNVBAcMB0RlZmF1
bHQxEDAOBgNVBAoMB0RlZmF1bHQxEDAOBgNVBAsMB1Jvb3QgQ0ExEDAOBgNVBAMM
B1Jvb3QgQ0EwIBcNMjUwMzAyMTYyNzU4WhgPMjI5ODEyMTUxNjI3NThaMGcxCzAJ
BgNVBAYTAlVTMRAwDgYDVQQIDAdEZWZhdWx0MRAwDgYDVQQHDAdEZWZhdWx0MRAw
DgYDVQQKDAdEZWZhdWx0MRAwDgYDVQQLDAdSb290IENBMRAwDgYDVQQDDAdSb290
IENBMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAErFzNSZjpYSDSltw59xWyJ4qWmhof
e0idFvpU5IR/ESQCH2XLCB7pCPipcaMle6uXkBgLmDnlfx9uEjDyPCoH8/kzO9jU
LxmP5qs6COOvp/te3NNtJ5d61YVsFvJtOV63o1MwUTAdBgNVHQ4EFgQU8QzxNU1Y
Un8xwId7x+kw4+XSpMswHwYDVR0jBBgwFoAU8QzxNU1YUn8xwId7x+kw4+XSpMsw
DwYDVR0TAQH/BAUwAwEB/zAKBggqhkjOPQQDAgNoADBlAjEA8G1boeaHX38AHHTo
1HY40uz7xMw4iO6fyQ4oVgcpAc/gTFy+dLK1ZTcf6cSeabi8AjBxIpXixwrljDz+
PpT7MUaOkjIOB/oL4saF83iwjFOp3iHQ0JKrK8/agyHcQVac4vU=
-----END CERTIFICATE-----`

/*
openssl ecparam -genkey -name secp384r1 -noout -out server.key
cat server.key
*/
const keyPEM = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDDTT/qNszAfIQFRv9y34x1RgM3hFVAp5U3a/btjYqEgqxYk8kvUGlFr
+qEprddwNqCgBwYFK4EEACKhZANiAARiOW9fjG7w3oscwVgIV09b4j8OeHZU0Zm7
tZETBGwIzFBiYfYkJZsdqd7xItm2NI9pIwUN1IOUaMWj04pt4QPimRF9595dsQRR
QBi1vJGmGpbzVQMrdPX76841f7ijjMk=
-----END EC PRIVATE KEY-----`

/*
cat >san.cnf <<-EOF
[req]
req_extensions = req_ext
distinguished_name = dn

[dn]
C = US
ST = Default
L = Default
O = Default
OU = Server
CN = localhost

[req_ext]
subjectAltName = @alt_names

[alt_names]
IP.1 = 127.0.0.1
EOF

	openssl req -new -key server.key -out server.csr -config san.cnf \
		-subj "/C=US/ST=Default/L=Default/O=Default/OU=Root CA/CN=Root CA"

	openssl x509 -req -in server.csr -CA root.pem -CAkey root.key -CAcreateserial \
		-out server.crt -days 99999 -sha256 -extensions req_ext -extfile san.cnf

cat server.crt
*/
const certPEP = `-----BEGIN CERTIFICATE-----
MIICYzCCAeigAwIBAgIUWwtxD72tzDuOZhszd51AVmuZpswwCgYIKoZIzj0EAwIw
ZzELMAkGA1UEBhMCVVMxEDAOBgNVBAgMB0RlZmF1bHQxEDAOBgNVBAcMB0RlZmF1
bHQxEDAOBgNVBAoMB0RlZmF1bHQxEDAOBgNVBAsMB1Jvb3QgQ0ExEDAOBgNVBAMM
B1Jvb3QgQ0EwIBcNMjUwMzAyMTY0NTQzWhgPMjI5ODEyMTUxNjQ1NDNaMGcxCzAJ
BgNVBAYTAlVTMRAwDgYDVQQIDAdEZWZhdWx0MRAwDgYDVQQHDAdEZWZhdWx0MRAw
DgYDVQQKDAdEZWZhdWx0MRAwDgYDVQQLDAdSb290IENBMRAwDgYDVQQDDAdSb290
IENBMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEYjlvX4xu8N6LHMFYCFdPW+I/Dnh2
VNGZu7WREwRsCMxQYmH2JCWbHane8SLZtjSPaSMFDdSDlGjFo9OKbeED4pkRfefe
XbEEUUAYtbyRphqW81UDK3T1++vONX+4o4zJo1MwUTAPBgNVHREECDAGhwR/AAAB
MB0GA1UdDgQWBBRmxPMxdBiiNpML14s8SWKQCOuLZjAfBgNVHSMEGDAWgBTxDPE1
TVhSfzHAh3vH6TDj5dKkyzAKBggqhkjOPQQDAgNpADBmAjEAy0Bq3IU8jXkfz6be
QwmYr+tqdBUnWpSwvIgySTU7nF1qT8CUF1Nq/xKbl1FQfFy9AjEA4ni6pxNc4v7a
yaCpOxVFyMz6wFOdTdWBBR4MFNi/HsAcSGMvSIPM+PMYdFc0FmN3
-----END CERTIFICATE-----`

// NewRootCertPool creates a new custom root-certificate set.
//
// In this example, it's used so that the server's certificates are trusted.
// In real world use it's better to omit this in order to use the
// default root set of the current operating system.
func NewRootCertPool() *x509.CertPool {
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		panic("failed to parse root certificate")
	}
	return roots
}

// ExampleTLSServer creates a httptest.Server with hardcoded key pair.
func ExampleTLSServer() *httptest.Server {
	cert, err := tls.X509KeyPair([]byte(certPEP), []byte(keyPEM))
	if err != nil {
		log.Fatal(err)
	}

	server := httptest.NewUnstartedServer(TLSHandler())
	server.TLS = &tls.Config{Certificates: []tls.Certificate{cert}}
	return server
}

// TLSHandler creates http.Handler for tls server
//
// Routes:
//
//	GET /fruits           get item map
//	GET /fruits/{name}    get item amount
//	PUT /fruits/{name}    add or update fruit (amount in body)
func TLSHandler() http.Handler {
	items := map[string]int{}

	mux := http.NewServeMux()

	mux.HandleFunc("/tls/", func(writer http.ResponseWriter, request *http.Request) {
		_, name := path.Split(request.URL.Path)

		switch request.Method {
		case "PUT":
			var data int
			if err := json.NewDecoder(request.Body).Decode(&data); err != nil {
				panic(err)
			}
			items[name] += data
			writer.WriteHeader(http.StatusNoContent)

		case "DELETE":
			var data int
			if err := json.NewDecoder(request.Body).Decode(&data); err != nil {
				panic(err)
			}
			if _, ok := items[name]; ok {
				items[name] -= data
				writer.WriteHeader(http.StatusNoContent)
			} else {
				writer.WriteHeader(http.StatusNotFound)
			}

		case "GET":
			if amount, ok := items[name]; ok {
				_, err := writer.Write([]byte(strconv.Itoa(amount)))
				if err != nil {
					writer.WriteHeader(http.StatusServiceUnavailable)
				}
			} else {
				writer.WriteHeader(http.StatusNotFound)
			}

		default:
			writer.WriteHeader(http.StatusBadRequest)
		}
	})

	mux.HandleFunc("/tls", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case "GET":

			all, err := json.Marshal(&items)
			if err != nil {
				panic(err)
			}
			writer.Header().Set("Content-Type", "application/json")
			writer.Write(all)

		default:
			writer.WriteHeader(http.StatusBadRequest)
		}
	})

	return mux
}
