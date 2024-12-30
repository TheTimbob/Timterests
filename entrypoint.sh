#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    source .env
fi

CERT_RENEWAL_INTERVAL=90 # Days

# Request or renew certificates
if [ ! -d "$CERT_DIR" ]; then
	echo "Certificates not found. Requesting new certificates..."
	echo "Certifcate details: Domain: ${DOMAIN} Email: ${EMAIL}"
	certbot certonly --dns-route53 -d "${DOMAIN}" -d "*.${DOMAIN}" --non-interactive --agree-tos -m "${EMAIL}"
else
	echo "Renewing certificates..."
	certbot renew --non-interactive
fi

echo "Starting the Go application..."
exec /app/main
