#!/bin/bash
set -e

echo "=== Nextcloud Test Server Initialization ==="

# Wait for Nextcloud to be fully initialized
while [ ! -f /var/www/html/config/config.php ]; do
    echo "Waiting for Nextcloud to be initialized..."
    sleep 5
done

echo "Nextcloud is initialized. Installing Tasks app..."

# Install and enable the Tasks app
php /var/www/html/occ app:install tasks || echo "Tasks app already installed"
php /var/www/html/occ app:enable tasks

# Install and enable Calendar app (CalDAV dependency)
php /var/www/html/occ app:install calendar || echo "Calendar app already installed"
php /var/www/html/occ app:enable calendar

# Create a test calendar for integration tests
echo "Creating test calendar..."
php /var/www/html/occ dav:create-calendar admin testcalendar "Test Calendar" || echo "Test calendar already exists"

echo "=== Initialization Complete ==="
echo ""
echo "Nextcloud Test Server is ready!"
echo "Access at: http://localhost:8080"
echo "Username: \${NEXTCLOUD_ADMIN_USER}"
echo ""
echo "CalDAV URL: http://localhost:8080/remote.php/dav"
echo ""
# WARNING: This is a test/development server only. Do NOT expose to external networks.
