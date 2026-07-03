#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CERT_DIR="${SCRIPT_DIR}/../registry/certs"
mkdir -p "$CERT_DIR"

openssl req -newkey rsa:2048 -nodes -sha256 \
  -keyout "${CERT_DIR}/auth.key" \
  -x509 -days 3650 \
  -out "${CERT_DIR}/auth.crt" \
  -subj "/CN=private-registry-auth"

chmod 600 "${CERT_DIR}/auth.key"
chmod 644 "${CERT_DIR}/auth.crt"

# Generate JWKS from the certificate so the registry can validate tokens
# without depending on the API server startup order.
python3 - "${CERT_DIR}/auth.crt" "${CERT_DIR}/jwks.json" <<'PY'
import sys, json, hashlib, base64, subprocess, re, tempfile, os

cert_path, out_path = sys.argv[1], sys.argv[2]
with tempfile.NamedTemporaryFile(mode='wb', suffix='.pem', delete=False) as pub_f:
    pub_pem = subprocess.check_output(['openssl', 'x509', '-in', cert_path, '-pubkey', '-noout'])
    pub_f.write(pub_pem)
    pub_path = pub_f.name
try:
    text = subprocess.check_output(['openssl', 'rsa', '-pubin', '-in', pub_path, '-noout', '-text'], text=True)
finally:
    os.unlink(pub_path)
hex_n = re.search(r'Modulus:([\s\S]*?)Exponent:', text).group(1)
hex_n = ''.join(c for c in hex_n if c in '0123456789abcdefABCDEF')
if hex_n.startswith('00') and len(hex_n) % 2 == 0:
    hex_n = hex_n[2:]
n = base64.urlsafe_b64encode(bytes.fromhex(hex_n)).rstrip(b'=').decode()
exp_match = re.search(r'Exponent: (\d+)', text)
e_int = int(exp_match.group(1))
e = base64.urlsafe_b64encode(e_int.to_bytes((e_int.bit_length() + 7) // 8, 'big')).rstrip(b'=').decode()
jwk_json = json.dumps({'e': e, 'kty': 'RSA', 'n': n}, sort_keys=True).encode()
kid = base64.urlsafe_b64encode(hashlib.sha256(jwk_json).digest()).rstrip(b'=').decode()
jwks = {'keys': [{'kty': 'RSA', 'kid': kid, 'use': 'sig', 'alg': 'RS256', 'n': n, 'e': e}]}
with open(out_path, 'w') as f:
    json.dump(jwks, f, indent=2)
PY

chmod 644 "${CERT_DIR}/jwks.json"

echo "Generated:"
echo "  ${CERT_DIR}/auth.key"
echo "  ${CERT_DIR}/auth.crt"
echo "  ${CERT_DIR}/jwks.json"
