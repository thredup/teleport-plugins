/*
Copyright 2021 Gravitational, Inc.

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

package client

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/gravitational/teleport/api/constants"
	"github.com/gravitational/trace"
	"golang.org/x/net/http2"
)

// Credentials are used to authenticate to Auth.
type Credentials interface {
	// Dialer is used to dial a connection to Auth.
	Dialer() (ContextDialer, error)
	// Config returns TLS configuration used to connect to Auth.
	Config() (*tls.Config, error)
}

// LoadTLS is used to load credentials directly from another *tls.Config.
func LoadTLS(tlsConfig *tls.Config) *TLSConfigCreds {
	return &TLSConfigCreds{
		tlsConfig: tlsConfig,
	}
}

// TLSConfigCreds are used to authenticate the client
// with a predefined *tls.Config.
type TLSConfigCreds struct {
	tlsConfig *tls.Config
}

// Dialer is used to dial a connection to Auth.
func (c *TLSConfigCreds) Dialer() (ContextDialer, error) {
	return nil, trace.NotImplemented("no dialer")
}

// Config returns TLS configuration used to connect to Auth.
func (c *TLSConfigCreds) Config() (*tls.Config, error) {
	return configure(c.tlsConfig), nil
}

// LoadKeyPair is used to load credentials from files on disk.
func LoadKeyPair(certFile string, keyFile string, caFile string) *KeyPairCreds {
	return &KeyPairCreds{
		certFile: certFile,
		keyFile:  keyFile,
		caFile:   caFile,
	}
}

// KeyPairCreds are used to authenticate the client
// with certificates generated in the given file paths.
type KeyPairCreds struct {
	certFile string
	keyFile  string
	caFile   string
}

// Dialer is used to dial a connection to Auth.
func (c *KeyPairCreds) Dialer() (ContextDialer, error) {
	return nil, trace.NotImplemented("no dialer")
}

// Config returns TLS configuration used to connect to Auth.
func (c *KeyPairCreds) Config() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(c.certFile, c.keyFile)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	cas, err := ioutil.ReadFile(c.caFile)
	if err != nil {
		return nil, trace.ConvertSystemError(err)
	}

	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(cas); !ok {
		return nil, trace.BadParameter("invalid TLS CA cert PEM")
	}

	return configure(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}), nil
}

// LoadIdentityFile is used to load credentials from an identity file on disk.
func LoadIdentityFile(path string) *IdentityCreds {
	return &IdentityCreds{
		path: path,
	}
}

// IdentityCreds are used to authenticate the client
// with an identity file generated in the given file path.
type IdentityCreds struct {
	path string
}

// Dialer is used to dial a connection to Auth.
func (c *IdentityCreds) Dialer() (ContextDialer, error) {
	return nil, trace.NotImplemented("no dialer")
}

// Config returns TLS configuration used to connect to Auth.
func (c *IdentityCreds) Config() (*tls.Config, error) {
	identityFile, err := ReadIdentityFile(c.path)
	if err != nil {
		return nil, trace.BadParameter("identity file could not be decoded: %v", err)
	}

	tlsConfig, err := identityFile.TLS()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return configure(tlsConfig), nil
}

func configure(c *tls.Config) *tls.Config {
	tlsConfig := c.Clone()

	tlsConfig.NextProtos = []string{http2.NextProtoTLS}

	if tlsConfig.ServerName == "" {
		tlsConfig.ServerName = constants.APIDomain
	}

	// This logic still appears to be necessary to force client to always send
	// a certificate regardless of the server setting. Otherwise the client may pick
	// not to send the client certificate by looking at certificate request.
	if len(tlsConfig.Certificates) > 0 {
		cert := tlsConfig.Certificates[0]
		tlsConfig.Certificates = nil
		tlsConfig.GetClientCertificate = func(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return &cert, nil
		}
	}

	return tlsConfig
}
