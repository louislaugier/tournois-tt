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