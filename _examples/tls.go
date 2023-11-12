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
MIIBVDCB+6ADAgECAgEBMAoGCCqGSM49BAMCMBIxEDAOBgNVBAoTB1Rlc3QgQ0Ew
HhcNMjMxMTEyMjI1MDUxWhcNMjQwNTEwMjI1MDUyWjASMRAwDgYDVQQKEwdUZXN0
IENBMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAExCfkUYqpDK4p8kqoiv8N3NJN
TLRPXO34/bHMK1LlLgKZl/pVNoRDkBezOuA1JY7P84yIbHQURrdSk1fxSPuSYaNC
MEAwDgYDVR0PAQH/BAQDAgIEMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFGzL
/3opNd1chvJe3FfNKD0/r500MAoGCCqGSM49BAMCA0gAMEUCIQCQlXyc7ZOG/Pzm
1EXeRIk+kfhTSjm2N9VU2kfK9sXZygIgSSv3lfL+sIr/HsWU0JXgadKgTQXpLdv3
tpQZpaV/Nxc=
-----END CERTIFICATE-----
`

func ExampleTlsServer() *httptest.Server {
	certPem := []byte(`-----BEGIN CERTIFICATE-----
MIIC5TCCAc2gAwIBAgIBATANBgkqhkiG9w0BAQsFADASMRAwDgYDVQQKEwdBY21l
IENvMB4XDTIzMTExMjIyNTgwNloXDTI0MDUxMDIyNTgwNlowEjEQMA4GA1UEChMH
QWNtZSBDbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKIVbmvhGdNJ
ni4e3Bwxolg7M6caoud0IJHQvzyK/XwuPk3Gwp3+fjtC3SuPf8cxPhur9hCskFuK
ngQ28T6cANUCYTjPsYY8PDOPLTTrkuuBFDdO/IxDZvZuBW3ZO3QqUAECSt8stsh8
wheuoBlzR+hKwJJK3pmkEaNKrLmZXopqXuecW1SSteqhzf763zkVi8ZLwJziKUuD
FjC4R5ChEq0pDzBKQdBWIv8Jptv7kIT1fN39Yiim4pZ2KJ9atjMx1ukrrBd0nKB/
7C5nu/zlbNsot6zc8Su8Z5WCQqXoCCrPIaYnlh/+Lq9VJhYD0Pjo60xKWzzbIi+G
EjIDn9XTUg8CAwEAAaNGMEQwDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsG
AQUFBwMBMAwGA1UdEwEB/wQCMAAwDwYDVR0RBAgwBocEfwAAATANBgkqhkiG9w0B
AQsFAAOCAQEAcbIeCvXJpjH8QjPE7agC5rstxuRvbYHtSn7hKixojwPAFKXIQqRZ
2l/GhJUTmymtc+vaAVSLpRAps+TuB/TXetdJqT8aadk6GFgd8FLMARCpMy5xrSSv
qVO99viHSIpaR/T1HDKCSAz6llEId1yi+XCYR0fOSg5tXqwF8py6N19Lcp1Zy1f6
0Ms5Pa7X2crF5f1bkoJiFM3XFTWOUy8MXuZuQoeOtGdLAePVWvQWWwZsO+2NuQkF
7dcimZJ8gpp1QcGEFx8+RIgGkbtp1CMEx8BEf5kkfMbje1/OdZvwsfZOTI1RaUWA
ABxPrh6fMq68uLduLGXiIoiV/yP3Lm/Xdw==
-----END CERTIFICATE-----

`)
	keyPem := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAohVua+EZ00meLh7cHDGiWDszpxqi53QgkdC/PIr9fC4+TcbC
nf5+O0LdK49/xzE+G6v2EKyQW4qeBDbxPpwA1QJhOM+xhjw8M48tNOuS64EUN078
jENm9m4Fbdk7dCpQAQJK3yy2yHzCF66gGXNH6ErAkkremaQRo0qsuZleimpe55xb
VJK16qHN/vrfORWLxkvAnOIpS4MWMLhHkKESrSkPMEpB0FYi/wmm2/uQhPV83f1i
KKbilnYon1q2MzHW6SusF3ScoH/sLme7/OVs2yi3rNzxK7xnlYJCpegIKs8hpieW
H/4ur1UmFgPQ+OjrTEpbPNsiL4YSMgOf1dNSDwIDAQABAoIBAAib7M6MGUwQt/cp
KnXQ6ReYpWi10HtMvsIf/Vhg5Y/oAOUurn2n29qX9ZlvuNDCu9LKcnp2QACsvzHo
HS4/KQgnZTSYS4yevG/cpgEOljIuG/3IEz/8AIcMVvt7s127NZ6oGYP7IwZJIiIR
420Wo3YiKlJa6bHtdgZfXAdLryrY3+PxkGwXCoETUlvHks5QBFXYmLLjfKwqGreX
RPFkDq0OBlYDbIrfddR8iM3YZao+2dDSCFKBU40RKzY8lAU7cz/IoS4ooQ+1ouuj
77bkQaE9C4a9DA+U4nXUAkqiWk1v0hP5YgydNPAxtfkWKJoGS/WS41ZcLWUgQ5O+
zdP7G7kCgYEA05WmK3YTG+IcW9t3fOPLOEGvKdE8l/iA7EyXcwunYzE2v47ebciF
vXmEye2qRlvfogD3o6YrbjDQPYhBQmmr1WnGWNgpLC+shy3oOrb1e4OBV7EHoXhW
MeWfUoAmWGgBddU/ZfPyIoarsJMdvNu3fVGoncwuDGGyX6pumb/gcw0CgYEAxBun
eFVhxgC9FdYBPr9HCvEqMOdulyjin28M3GnWmdrDqBlP/QuQgqXsnAn0XlZZb4Qf
CDudDyyZncSG8AODeFMPEB5vdWqroq7zTsMufRwI+4qJSJgxVHLitWnh+egnWsmq
fBFVG5K8RgSeB77Yl4oyejRVcOZ9Fi2vh8GZwosCgYBa5dCUnU5KTVJ3mAp2SfqV
OYrCAVTxyN3CJolt8FTCBXOKyhr+uQXTx6/nfEYJohCqLZY15P6FgU0FElNO78zV
i3Kd2oedpwGMtYkuKEm//VgEz1YC5YrKNubCb7GJi20NLUbmSu38LTT3T8yXxSDI
Itu4pu4lfZc/CB4pyUfoxQKBgQC5nGcEuONisd5FpZjmF8qY66uAP/vnLDZaqpPk
pnQMiQc4ukR//4sWbQ8mnTFifJ4Hs2hftXSxIQiAT7tbvieYIh0Wp4fc/UpYHviA
qrH8jiVeV0Aaqpm+EULMa9wLWZSuFEO9S/Zes6JpLwOX1yVPQOkHyzK3OiBYdoM1
naL3gwKBgFIXQ8mgtpsA9AnzrfdigKP4C78HxvEq2iLR6HLGurmtdb88rGX9qZnT
zcXCOJ4ZS+ha08YPor/NOoI4olz8GTa0qTORcxGKsAe6F0tln0p7AeKd37jaebdU
j0IkvdefQeTVjMZKhO1px6FMl8Jg2kWaPchU5JQb9LnMNLtv/FlG
-----END RSA PRIVATE KEY-----
`)

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
