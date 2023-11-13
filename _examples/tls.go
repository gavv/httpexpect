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

func NewRootCertPool() *x509.CertPool {
	const rootPEM = `
-----BEGIN CERTIFICATE-----
MIIBUzCB+6ADAgECAgEBMAoGCCqGSM49BAMCMBIxEDAOBgNVBAoTB1Rlc3QgQ0Ew
HhcNMjMxMTEzMTIyNTEzWhcNMjQwNTExMTIyNTEzWjASMRAwDgYDVQQKEwdUZXN0
IENBMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEHBm1CiEs4CKw0ynUlzaTz9Pi
ROnBwosfX3xYIEz5l1rN119FEJLWQFx8xBASkpZDz+Eehw9QdPaqwapDKGVgbaNC
MEAwDgYDVR0PAQH/BAQDAgIEMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFExX
luCP8Bp73F1L7UuFCM/NFgPkMAoGCCqGSM49BAMCA0cAMEQCIFLFQRgIUbjzA0c1
Pennq6gP/WiJpppZPQq5IYR4V7BfAiBVoGh+32UOJ13YYO8HsL/6P7KIwZKXkJpJ
LoibTriVMg==
-----END CERTIFICATE-----
`
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		panic("failed to parse root certificate")
	}
	return roots
}

func ExampleTlsServer() *httptest.Server {
	certPem := []byte(`-----BEGIN CERTIFICATE-----
MIIBbDCCARKgAwIBAgIBAjAKBggqhkjOPQQDAjASMRAwDgYDVQQKEwdUZXN0IENB
MB4XDTIzMTExMzEyMjUxN1oXDTI0MDUxMTEyMjUxN1owEjEQMA4GA1UEChMHQWNt
ZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABEvlkPnSh5jYMD4MSkjJH7HW
iDR/UnqIJrI3nV0FTotWly0z3nMy0FCM1VxyGJc8HcKi2KPIaQmVF2sYCLwu8xuj
WTBXMA4GA1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDATAfBgNVHSME
GDAWgBRMV5bgj/Aae9xdS+1LhQjPzRYD5DAPBgNVHREECDAGhwR/AAABMAoGCCqG
SM49BAMCA0gAMEUCIQDHQVvWrOvagkYT9/qeSZ7xUwTTWiRfvWmlCgLf5NXu7AIg
ea/Q6OcG41k25PXVn3VRLRBEfSFIsuJzTyTNXCHx8vY=
-----END CERTIFICATE-----`)

	keyPem := []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIHHIE/n9wJI/dm1vnwhd8Jm/Wi04R+m8wYfUnkCFu4QnoAoGCCqGSM49
AwEHoUQDQgAES+WQ+dKHmNgwPgxKSMkfsdaINH9SeogmsjedXQVOi1aXLTPeczLQ
UIzVXHIYlzwdwqLYo8hpCZUXaxgIvC7zGw==
-----END EC PRIVATE KEY-----`)

	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		log.Fatal(err)
	}

	server := httptest.NewUnstartedServer(TlsHandler())
	server.TLS = &tls.Config{Certificates: []tls.Certificate{cert}}
	return server
}

func TlsHandler() http.Handler {
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
			if _, ok := items[name]; ok {
				delete(items, name)
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
