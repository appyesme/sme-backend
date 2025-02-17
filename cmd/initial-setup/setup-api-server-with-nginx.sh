#!/bin/bash

# This is one time setup script to configure the API server with Nginx and SSL.
# Steps:
# 1. Go to digitalocean.com and open droplet in console.
# 2. Add this file in the root directory of the droplet.
# 3. Change the variables at the top of the file.
# 4. Run the script.
# 5. Restart the droplet.

# Variables (Replace these with your actual values)
API_URL='api.yes-me.com'
CONTACT_EMAIL='appyesme@gmail.com'
APP_WORKDIR='/root/apps/sme/'
APP_EXECUTABLE='/root/apps/sme/sme-backend'
APP_PORT=8080

# Update and Install Dependencies & necessary packages
echo "Updating system and installing required packages...";
sudo apt update;
sudo apt install nginx -y;

# Configure UFW (Firewall)
echo "Configuring UFW...";
sudo ufw enable;
sudo ufw allow 80;    # Allow HTTP
sudo ufw allow 443;   # Allow HTTPS
sudo ufw status;      # Check UFW status

# Create Systemd Service for the Go Application
echo "Creating systemd service for the Go app...";
sudo bash -c "cat > /etc/systemd/system/sme.service" <<EOF
[Unit]
Description=SME API Service
After=network.target

[Service]
Type=simple

Restart=on-failure

WorkingDirectory=$APP_WORKDIR
ExecStart= /bin/bash -lc $APP_EXECUTABLE

[Install]
WantedBy=multi-user.target
EOF

# Reload and Start the Service
sudo systemctl daemon-reload;
sudo systemctl start sme.service;
sudo systemctl enable sme.service;
sudo systemctl status sme.service;

# Verify the Go Application
echo "Testing the Go app locally...";
curl -s http://localhost:$APP_PORT/version || {
    echo "ERROR: The Go app did not respond. Check the service logs."
    exit 1
}

# Configure Nginx
echo "Configuring Nginx for $API_URL...";
sudo bash -c "cat > /etc/nginx/conf.d/sme_api_server.conf" <<EOF
server {
    listen 80;
    server_name $API_URL;
    client_max_body_size 100M;

    location / {
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header Host \$http_host;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_pass http://localhost:$APP_PORT;
    }
}
EOF

# Test and Restart Nginx
sudo nginx -t;
sudo systemctl restart nginx;

# Install and Configure Certbot
echo "Installing Certbot for SSL..."
sudo snap install --classic certbot;
sudo ln -s /snap/bin/certbot /usr/bin/certbot;

echo "Generating SSL certificate for $API_URL..."
sudo certbot --nginx --non-interactive --agree-tos --email $CONTACT_EMAIL --domains $API_URL;

# Test HTTPS
echo "Testing HTTPS...";
curl -k -s https://$API_URL/version || {
    echo "ERROR: HTTPS test failed. Check Nginx or Certbot configuration."
    exit 1
}

# Set Up Automatic Certificate Renewal
echo "Setting up automatic certificate renewal...";
sudo certbot renew --dry-run;

echo "Setup complete! Your app should be live at https://$API_URL/version"