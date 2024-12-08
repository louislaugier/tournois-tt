# Certbot SSL Certificates

This directory contains the SSL certificate data and configuration for the application.

## Structure

- `conf/` - Contains Let's Encrypt configuration and certificates (not committed to git)
- `data/` - Contains Let's Encrypt webroot for domain validation (not committed to git)

## Notes

- These directories are kept empty in git using `.gitkeep` files
- The actual certificate data and configuration will be generated when running certbot in production
- Never commit SSL certificates or private keys to version control

## First-time Setup

When deploying to production:

1. Make sure your domain's DNS is pointing to your server
2. Set up your environment variables:
   ```bash
   # Copy the example environment file
   cp .env.example .env
   
   # Edit .env with your actual values
   DOMAIN=your-domain.com
   EMAIL=your-email@domain.com
   ```
3. The certbot container will automatically obtain and renew SSL certificates when you run `docker-compose up -d` 