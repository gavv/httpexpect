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

const rootPEM = `
-----BEGIN CERTIFICATE-----
MIIDXzCCAkegAwIBAgILBAAAAAABIVhTCKIwDQYJKoZIhvcNAQELBQAwTDEgMB4
GA1UECxMXR2xvYmFsU2lnbiBSb290IENBIC0gUjMxEzARBgNVBAoTCkdsb2JhbF
NpZ24xEzARBgNVBAMTCkdsb2JhbFNpZ24wHhcNMDkwMzE4MTAwMDAwWhcNMjkwM
zE4MTAwMDAwWjBMMSAwHgYDVQQLExdHbG9iYWxTaWduIFJvb3QgQ0EgLSBSMzET
MBEGA1UEChMKR2xvYmFsU2lnbjETMBEGA1UEAxMKR2xvYmFsU2lnbjCCASIwDQY
JKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMwldpB5BngiFvXAg7aEyiie/QV2Ec
WtiHL8RgJDx7KKnQRfJMsuS+FggkbhUqsMgUdwbN1k0ev1LKMPgj0MK66X17YUh
hB5uzsTgHeMCOFJ0mpiLx9e+pZo34knlTifBtc+ycsmWQ1z3rDI6SYOgxXG71uL
0gRgykmmKPZpO/bLyCiR5Z2KYVc3rHQU3HTgOu5yLy6c+9C7v/U9AOEGM+iCK65
TpjoWc4zdQQ4gOsC0p6Hpsk+QLjJg6VfLuQSSaGjlOCZgdbKfd/+RFO+uIEn8rU
AVSNECMWEZXriX7613t2Saer9fwRPvm2L7DWzgVGkWqQPabumDk3F2xmmFghcCA
wEAAaNCMEAwDgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0O
BBYEFI/wS3+oLkUkrk1Q+mOai97i3Ru8MA0GCSqGSIb3DQEBCwUAA4IBAQBLQNv
AUKr+yAzv95ZURUm7lgAJQayzE4aGKAczymvmdLm6AC2upArT9fHxD4q/c2dKg8
dEe3jgr25sbwMpjjM5RcOO5LlXbKr8EpbsU8Yt5CRsuZRj+9xTaGdWPoO4zzUhw
8lo/s7awlOqzJCK6fBdRoyV3XpYKBovHd7NADdBj+1EbddTKJd+82cEHhXXipa0
095MJ6RMG3NzdvQXmcIfeg7jLQitChws/zyrVQ4PkX4268NXSb7hLi18YIvDQVE
TI53O9zJrlAGomecsMx86OyXShkDOOyyGeMlhLxS67ttVb9+E7gUJTb0o2HLO02
JQZR7rkpeDMdmztcpHWD9f
-----END CERTIFICATE-----`

func ExampleTlsServer() *httptest.Server {
	certPem := []byte(`-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`)
	keyPem := []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49
AwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q
EKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==
-----END EC PRIVATE KEY-----`)

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		panic("failed to parse root certificate")
	}

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
		var data int
		if err := json.NewDecoder(request.Body).Decode(&data); err != nil {
			panic(err)
		}

		switch request.Method {
		case "PUT":
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
			if _, ok := items[name]; ok {
				writer.Write([]byte(strconv.Itoa(items[name])))
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
