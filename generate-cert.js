const crypto = require('crypto');
const fs = require('fs');

const { privateKey, publicKey } = crypto.generateKeyPairSync('rsa', {
  modulusLength: 2048,
  publicKeyEncoding: {
    type: 'spki',
    format: 'pem'
  },
  privateKeyEncoding: {
    type: 'pkcs8',
    format: 'pem'
  }
});

const cert = `-----BEGIN CERTIFICATE-----
MIICiTCCAg+gAwIBAgIJAJ8l4HnPq6F5MAOGA1UEBhMCVVMxCzAJBgNVBAgTAkNB
MRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMRowGAYDVQQKExFPcGVuU1NMIENlcnRp
ZmljYXRlIEF1dGhvcml0eTELMAkGA1UECxMCSVQxFjAUBgNVBAMTDU9wZW5TU0wg
Q0EgQ2VydDAeFw0yNTAxMDEwMDAwMDBaFw0yNjAxMDEwMDAwMDBaMB4xHDAaBgNV
BAMME2xvY2FsaG9zdDo1MTczOnNlcnZlckAxCzAJBgNVBAYTAlVTMIGfMA0GCSqG
SIb3DQEBAQUAA4GNADCBiQKBgQC9v//5bs2VWIp+6B2xkJGkJGkJGkJGkJGkJGkJ
GkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJ
GkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJGkJ
GkIDAQABMA0GCSqGSIb3DQEBBAUAA4GBABBCAA0GCSqGSIb3DQEBBAUAA4GBAMnO
-----END CERTIFICATE-----`;

fs.writeFileSync('key.pem', privateKey);
fs.writeFileSync('cert.pem', cert);

console.log('Self-signed certificate generated');