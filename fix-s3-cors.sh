#!/bin/bash
# Script to update S3 bucket CORS configuration for direct uploads

# Get the bucket name from environment or use default pattern
BUCKET_NAME="${S3_ASSETS_BUCKET:-omnigen-assets-971422717446}"

echo "Updating CORS configuration for bucket: $BUCKET_NAME"

# Create CORS configuration JSON
cat > /tmp/cors-config.json << 'EOF'
{
  "CORSRules": [
    {
      "AllowedHeaders": ["*"],
      "AllowedMethods": ["GET", "HEAD", "PUT", "POST", "DELETE"],
      "AllowedOrigins": ["*"],
      "ExposeHeaders": ["ETag", "x-amz-server-side-encryption", "x-amz-request-id", "x-amz-id-2"],
      "MaxAgeSeconds": 3000
    }
  ]
}
EOF

# Apply CORS configuration
aws s3api put-bucket-cors \
  --bucket "$BUCKET_NAME" \
  --cors-configuration file:///tmp/cors-config.json

if [ $? -eq 0 ]; then
  echo "✅ CORS configuration updated successfully!"
  echo "You can now upload files directly to S3 from the frontend."
else
  echo "❌ Failed to update CORS configuration."
  echo "Make sure you have AWS credentials configured and the bucket name is correct."
  exit 1
fi

# Clean up
rm -f /tmp/cors-config.json

