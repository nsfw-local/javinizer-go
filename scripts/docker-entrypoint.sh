#!/bin/sh
set -e

# Copy default config if it doesn't exist
if [ ! -f "/javinizer/config.yaml" ]; then
    echo "No config file found, copying default configuration..."
    cp /app/config/config.yaml.default /javinizer/config.yaml
    echo "✓ Default configuration created at /javinizer/config.yaml"
fi

# Execute the main command
exec "$@"
