#!/bin/bash

# Seed a sample completed video into local DynamoDB
# This creates a mock video entry that appears in the Video Library

# Add Homebrew to PATH
export PATH="/opt/homebrew/bin:$PATH"

ENDPOINT="http://localhost:8000"
REGION="us-east-1"
TABLE_NAME="omnigen-jobs-local"

# Set dummy credentials for DynamoDB Local
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy

# Generate timestamps
CREATED_AT=$(date +%s)
UPDATED_AT=$CREATED_AT
COMPLETED_AT=$CREATED_AT

# Sample job ID and user ID (matches dev mode user)
JOB_ID="sample-video-$(date +%s)"
USER_ID="dev-user-123"

echo "🎬 Creating sample video in local DynamoDB..."
echo "   Job ID: $JOB_ID"
echo "   User ID: $USER_ID"
echo ""

# Create a completed video job with sample data
aws dynamodb put-item \
    --table-name $TABLE_NAME \
    --item '{
        "job_id": {"S": "'"$JOB_ID"'"},
        "user_id": {"S": "'"$USER_ID"'"},
        "status": {"S": "completed"},
        "stage": {"S": "completed"},
        "prompt": {"S": "A serene mountain landscape at sunrise with birds flying across the sky"},
        "title": {"S": "Mountain Sunrise"},
        "duration": {"N": "30"},
        "aspect_ratio": {"S": "16:9"},
        "style": {"S": "Cinematic"},
        "tone": {"S": "Inspiring"},
        "tempo": {"S": "Slow"},
        "progress_percent": {"N": "100"},
        "scenes_completed": {"N": "5"},
        "video_key": {"S": "jobs/'"$JOB_ID"'/video.mp4"},
        "thumbnail_url": {"S": "https://via.placeholder.com/1920x1080/4a5568/ffffff?text=Sample+Video"},
        "audio_url": {"S": "https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3"},
        "created_at": {"N": "'"$CREATED_AT"'"},
        "updated_at": {"N": "'"$UPDATED_AT"'"},
        "completed_at": {"N": "'"$COMPLETED_AT"'"}
    }' \
    --endpoint-url $ENDPOINT \
    --region $REGION \
    --no-cli-pager

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Sample video created successfully!"
    echo ""
    echo "📋 Video Details:"
    echo "   Title: Mountain Sunrise"
    echo "   Status: completed"
    echo "   Duration: 30 seconds"
    echo "   Created: $(date -r $CREATED_AT)"
    echo ""
    echo "🌐 View in app: http://localhost:5173/library"
    echo ""

    # Verify the item was created
    echo "🔍 Verifying item in database..."
    aws dynamodb get-item \
        --table-name $TABLE_NAME \
        --key '{"job_id": {"S": "'"$JOB_ID"'"}}' \
        --endpoint-url $ENDPOINT \
        --region $REGION \
        --no-cli-pager \
        --output json | jq -r '.Item.title.S // "Not found"' | while read title; do
            if [ "$title" != "Not found" ]; then
                echo "   ✓ Confirmed: $title"
            else
                echo "   ✗ Error: Item not found"
            fi
        done
else
    echo ""
    echo "❌ Failed to create sample video"
    exit 1
fi
