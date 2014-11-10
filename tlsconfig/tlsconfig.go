package tlsconfig

import (
	"crypto/x509"
	"github.com/bbbacsa/deploy.io/vendor/crypto/tls"
	"io/ioutil"
	"os"
)

func GetTLSConfig(clientCertPEMData, clientKeyPEMData []byte) (*tls.Config, error) {
	certPool := x509.NewCertPool()

	certChainPath := os.Getenv("DEPLOY_HOST_CA")
	if certChainPath != "" {
		certChainData, err := ioutil.ReadFile(certChainPath)
		if err != nil {
			return nil, err
		}
		certPool.AppendCertsFromPEM(certChainData)
	} else {
		certPool.AppendCertsFromPEM([]byte(deployCerts))
	}

	clientCert, err := tls.X509KeyPair(clientCertPEMData, clientKeyPEMData)
	if err != nil {
		return nil, err
	}

	config := new(tls.Config)
	config.RootCAs = certPool
	config.Certificates = []tls.Certificate{clientCert}
	config.BuildNameToCertificate()

	return config, nil
}

var deployCerts string = `-----BEGIN CERTIFICATE-----
MIIEKTCCAxGgAwIBAgIJAP81C5xoXHunMA0GCSqGSIb3DQEBCwUAMIGqMQswCQYD
VQQGEwJQSDEPMA0GA1UECAwGTGFndW5hMRAwDgYDVQQHDAdDYWxhbWJhMS4wLAYD
VQQKDCVCcnljaGVUZWNoIEludGVybmV0IFNvbHV0aW9ucyBDb21wYW55MQwwCgYD
VQQLDANEZXYxGDAWBgNVBAMMDzEwNC4xMzEuMTU4LjEyNDEgMB4GCSqGSIb3DQEJ
ARYRYmJiYWNzYUBnbWFpbC5jb20wHhcNMTQxMTEwMDUyMzU5WhcNMTUxMTEwMDUy
MzU5WjCBqjELMAkGA1UEBhMCUEgxDzANBgNVBAgMBkxhZ3VuYTEQMA4GA1UEBwwH
Q2FsYW1iYTEuMCwGA1UECgwlQnJ5Y2hlVGVjaCBJbnRlcm5ldCBTb2x1dGlvbnMg
Q29tcGFueTEMMAoGA1UECwwDRGV2MRgwFgYDVQQDDA8xMDQuMTMxLjE1OC4xMjQx
IDAeBgkqhkiG9w0BCQEWEWJiYmFjc2FAZ21haWwuY29tMIIBIjANBgkqhkiG9w0B
AQEFAAOCAQ8AMIIBCgKCAQEAzWBb0yQS5ca0dQWhsrPdFcsfUGazhJ8EXM+2Np5s
bj7wiT06TSunB+ME1Aj61KKxb9gI1QSW8LJy9Xp/1R1r7SWJ5VAAb8oSXP92w0Dk
ph8MXPl1x8K3B22hJk0jiIADdS09AG30cp7osW6uqz7ARgsQh4khh2DohaB0zM1t
9uLrDgUdP9BAVlFVRSYpfKMBPZ5PfmKmYod9GYOA9/Nxs44N/PhmvFMI42cVoL88
YFZ3x/U7Iu495Hri9fJ1roHAh7Z7nGL3sD/iGd3bGXOeDztXWiqp59qSFUyvOuyC
K/paZnz5izBlaz6Zir+M+zMcjAh4qvccfWyRlNZWFtAONwIDAQABo1AwTjAdBgNV
HQ4EFgQUSBE1/wLJOO3R2TFs8katsJy9B+0wHwYDVR0jBBgwFoAUSBE1/wLJOO3R
2TFs8katsJy9B+0wDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEATZbL
PdPIH8ahQpGdtTb0shcxOYcuLcrn67kxEzeOkXcOsmHhw8RdWkQiglMPBwdyi0Xj
8yY9ri1Jv1jC/swAAmtsB6qd+oxJaiVn2G+okVX2xXaLCQROwfIcNEnwVxUXyNwG
hMZEiNT1kymy8MI5FwQqZ4hvbbUqcMSrB2O1z5C8zwDL2eXm8LjrmRkRpb+pP9fX
kwYPbQO4v0v4PKge2ezhWc4u0WFN3Zg68XS2YB5anKQzK1heFiB79mbHyRKF+t9c
F4Un6peNMm7WBxup68KTBCQb6lK6jhtTUvVirMCjwaXUQOHOrRT9QzoBrfgJm0OJ
gCAxbCdK3lcDQKxC8Q==
-----END CERTIFICATE-----`
