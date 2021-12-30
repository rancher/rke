/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cert

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"net"
)

// MakeCSR generates a PEM-encoded CSR using the supplied private key, subject, and SANs.
// All key types that are implemented via crypto.Signer are supported (This includes
// *rsa.PrivateKey, *ecdsa.PrivateKey and ed25519.PrivateKey).
func MakeCSR(privateKey interface{}, subject *pkix.Name, dnsSANs []string, ipSANs []net.IP) (csr []byte, err error) {
	template := &x509.CertificateRequest{
		Subject:     *subject,
		DNSNames:    dnsSANs,
		IPAddresses: ipSANs,
	}

	return MakeCSRFromTemplate(privateKey, template)
}

// MakeCSRFromTemplate generates a PEM-encoded CSR using the supplied private
// key and certificate request as a template. All key types that are
// implemented via crypto.Signer are supported (This includes *rsa.PrivateKey,
// *ecdsa.PrivateKey and ed25519.PrivateKey)
func MakeCSRFromTemplate(privateKey interface{}, template *x509.CertificateRequest) ([]byte, error) {
	t := *template
	t.SignatureAlgorithm = sigType(privateKey)

	csrDER, err := x509.CreateCertificateRequest(cryptorand.Reader, &t, privateKey)
	if err != nil {
		return nil, err
	}

	csrPemBlock := &pem.Block{
		Type:  CertificateRequestBlockType,
		Bytes: csrDER,
	}

	return pem.EncodeToMemory(csrPemBlock), nil
}

func sigType(privateKey interface{}) x509.SignatureAlgorithm {
	if key, ok := privateKey.(*rsa.PrivateKey); ok {
		// Customize the signature for RSA keys, depending on the key size
		keySize := key.N.BitLen()
		switch {
		case keySize >= 4096:
			return x509.SHA512WithRSA
		case keySize >= 3072:
			return x509.SHA384WithRSA
		default:
			return x509.SHA256WithRSA
		}
	} else if key, ok := privateKey.(*ecdsa.PrivateKey); ok {
		// Customize the signature for ECDSA keys, depending on the curve used
		switch key.Params().Name {
		case "P-512":
			return x509.ECDSAWithSHA512
		case "P-384":
			return x509.ECDSAWithSHA384
		case "P-256":
			return x509.ECDSAWithSHA256
		case "P-224":
			return x509.ECDSAWithSHA1
		}
	} else if _, ok := privateKey.(ed25519.PrivateKey); ok {
		return x509.PureEd25519
	}
	return x509.UnknownSignatureAlgorithm
}
