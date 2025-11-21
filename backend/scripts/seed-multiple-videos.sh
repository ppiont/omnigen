#!/bin/bash

# Seed multiple sample videos with different statuses
# This creates a realistic video library for local development

# Add Homebrew to PATH
export PATH="/opt/homebrew/bin:$PATH"

ENDPOINT="http://localhost:8000"
REGION="us-east-1"
TABLE_NAME="omnigen-jobs-local"
USER_ID="dev-user-123"

# Set dummy credentials for DynamoDB Local
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy

echo "🎬 Seeding multiple sample videos..."
echo ""

# Helper function to create a video
create_video() {
    local job_id=$1
    local title=$2
    local status=$3
    local stage=$4
    local progress=$5
    local prompt=$6
    local duration=$7
    local age_hours=$8

    # Calculate timestamps based on age
    local created_at=$(($(date +%s) - ($age_hours * 3600)))
    local updated_at=$created_at
    local completed_at_field=""

    if [ "$status" = "completed" ]; then
        completed_at_field=', "completed_at": {"N": "'"$created_at"'"}'
    fi

    local video_url_field=""
    local thumbnail_field=""

    if [ "$status" = "completed" ]; then
        video_url_field=', "video_key": {"S": "jobs/'"$job_id"'/video.mp4"}'
        thumbnail_field=', "thumbnail_url": {"S": "https://via.placeholder.com/1920x1080/4a5568/ffffff?text='"${title// /+}"'"}'
    elif [ "$status" = "processing" ]; then
        thumbnail_field=', "thumbnail_url": {"S": "https://via.placeholder.com/1920x1080/f59e0b/ffffff?text=Processing"}'
    fi

    aws dynamodb put-item \
        --table-name $TABLE_NAME \
        --item '{
            "job_id": {"S": "'"$job_id"'"},
            "user_id": {"S": "'"$USER_ID"'"},
            "status": {"S": "'"$status"'"},
            "stage": {"S": "'"$stage"'"},
            "prompt": {"S": "'"$prompt"'"},
            "title": {"S": "'"$title"'"},
            "duration": {"N": "'"$duration"'"},
            "aspect_ratio": {"S": "16:9"},
            "progress_percent": {"N": "'"$progress"'"},
            "scenes_completed": {"N": "3"},
            "created_at": {"N": "'"$created_at"'"},
            "updated_at": {"N": "'"$updated_at"'"}
            '"$completed_at_field"'
            '"$video_url_field"'
            '"$thumbnail_field"'
        }' \
        --endpoint-url $ENDPOINT \
        --region $REGION \
        --no-cli-pager > /dev/null 2>&1

    if [ $? -eq 0 ]; then
        echo "   ✓ $title ($status)"
    else
        echo "   ✗ Failed to create $title"
    fi
}

# Create completed videos
echo "📦 Creating completed videos..."
create_video "video-$(date +%s)-1" "Mountain Sunrise" "completed" "completed" 100 "A serene mountain landscape at sunrise with birds flying across the sky" 30 2
create_video "video-$(date +%s)-2" "Coffee Commercial" "completed" "completed" 100 "Modern coffee shop with barista pouring latte art, warm lighting" 15 5
create_video "video-$(date +%s)-3" "Tech Product Launch" "completed" "completed" 100 "Sleek smartphone rotating on a futuristic pedestal with dynamic lighting" 20 12
create_video "video-$(date +%s)-4" "Fitness Motivation" "completed" "completed" 100 "Athletic person running at sunset, inspirational and energetic" 30 24

echo ""
echo "🔄 Creating processing videos..."
create_video "video-$(date +%s)-5" "Travel Adventure" "processing" "scene_generation" 45 "Aerial view of tropical beach with crystal clear water" 30 0
create_video "video-$(date +%s)-6" "Food Recipe" "processing" "audio_generation" 65 "Step by step pasta cooking tutorial with ingredients" 25 0

echo ""
echo "⏳ Creating pending videos..."
create_video "video-$(date +%s)-7" "Fashion Lookbook" "pending" "script_generation" 15 "Modern fashion collection showcase with runway vibes" 20 0

echo ""
echo "❌ Creating failed video..."
create_video "video-$(date +%s)-8" "Test Video" "failed" "script_generation" 10 "This is a test video that failed" 15 1

echo ""
echo "✅ Sample videos created successfully!"
echo ""
echo "🌐 View in app: http://localhost:5173/library"
echo ""

# Show count
echo "📊 Total videos in database:"
aws dynamodb scan \
    --table-name $TABLE_NAME \
    --select "COUNT" \
    --endpoint-url $ENDPOINT \
    --region $REGION \
    --no-cli-pager 2>&1 | grep -o '"Count": [0-9]*' || echo "   Count check unavailable"
