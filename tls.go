package DarkYoung

import (
    "crypto/tls"
    "crypto/x509"
    "github.com/pkg/errors"
    "google.golang.org/grpc/credentials"
    "io/ioutil"
)

func loadCredentials(caPath, certPath, keyPath string) (credentials.TransportCredentials, error) {
    certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
    if err != nil {
        return nil, err
    }
    // 加载ca认证
    certPool := x509.NewCertPool()
    ca, err := ioutil.ReadFile(caPath)
    if err != nil {
        return nil, err
    }
    if ok := certPool.AppendCertsFromPEM(ca); !ok { // 认证ca失败
        return nil, errors.New("failed to append client certs")
    }

    // Create the TLS credentials
    creds := credentials.NewTLS(&tls.Config{
        ClientAuth:   tls.RequireAndVerifyClientCert,
        Certificates: []tls.Certificate{certificate},
        ClientCAs:    certPool,
    })
    return creds, nil
}

func loadClientCredentials(caPath, certPath, keyPath string, ServerAddr string) (credentials.TransportCredentials, error) {
    certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
    if err != nil {
        return nil, err
    }
    // 加载ca认证
    certPool := x509.NewCertPool()
    ca, err := ioutil.ReadFile(caPath)
    if err != nil {
        return nil, err
    }
    if ok := certPool.AppendCertsFromPEM(ca); !ok { // 认证ca失败
        return nil, errors.New("failed to append client certs")
    }

    // Create the TLS credentials
    creds := credentials.NewTLS(&tls.Config{
        ServerName:ServerAddr,
        Certificates: []tls.Certificate{certificate},
        RootCAs:    certPool,
    })
    return creds, nil
}