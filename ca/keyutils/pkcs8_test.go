package keyutils

import (
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/require"
)

const ecKeyPEM = `-----BEGIN EC PRIVATE KEY-----
MGgCAQEEHAJNi85auxiK2BUfAPoXgYfWt7k57TkK1ng9SCSgBwYFK4EEACGhPAM6
AARNGXPlfrrZHwxfe1yhNNmITamxlJwxA2jpf/qAf4vnE87UZNylsC3xT5Quf/hL
kGCr5YiADxsjiA==
-----END EC PRIVATE KEY-----
`

const encryptedPKCS8InvalidPadding = `-----BEGIN ENCRYPTED PRIVATE KEY-----
MIHOMEkGCSqGSIb3DQEFDTA8MBsGCSqGSIb3DQEFDDAOBAjxk6v6kjceLAICCAAw
HQYJYIZIAWUDBAEqBBBVCqGMzL53rwf6Bv4OEPeJBIGAEuEUhjZd/d1BEbntAoZU
3cCB6ewYMqj97p6MncR1EFq+a26R/ehoCZg7O2L5AJrZK8K6UuZG8HxpZkraS5Mh
L5dg6PPGclig3Xn1sCPUmHi13x+DPISBuUdkQEep5lEpqxLSRQerllbXmhaTznAk
aqc20eq8ndE9DjZ7gDPnslY=
-----END ENCRYPTED PRIVATE KEY-----`

func TestPKCS8RejectsInvalidPadding(t *testing.T) {
	block, _ := pem.Decode([]byte(encryptedPKCS8InvalidPadding))
	require.NotNil(t, block)
	for _, formatter := range []Formatter{Default, FIPS} {
		_, err := formatter.DecryptPEMBlock(block, []byte("poonies"))
		require.Error(t, err)
		_, err = formatter.ParsePrivateKeyPEMWithPassword([]byte(encryptedPKCS8InvalidPadding), []byte("poonies"))
		require.Error(t, err)
		_, err = formatter.DecryptPEMBlock(decryptedPKCS8Block, []byte("ponies"))
		require.Error(t, err)
	}
}

func TestPKCS8KeyConversion(t *testing.T) {
	_, err := ECToPKCS8PEM([]byte("garbage pem"))
	require.Error(t, err)
	_, err = ECToPKCS8PEM(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: []byte("garbage key")}))
	require.Error(t, err)
	out, err := ECToPKCS8PEM([]byte(ecKeyPEM))
	require.NoError(t, err)
	require.Equal(t, []byte(decryptedPKCS8), out)

	_, err = PKCS8ToECPEM([]byte("garbage pem"))
	require.Error(t, err)
	_, err = PKCS8ToECPEM(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("garbage key")}))
	require.Error(t, err)
	out, err = PKCS8ToECPEM([]byte(decryptedPKCS8))
	require.NoError(t, err)
	require.Equal(t, []byte(ecKeyPEM), out)
}
