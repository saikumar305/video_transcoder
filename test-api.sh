#!/bin/bash

echo "🎬 Video Transcoder API Test Script"
echo "=================================="

# Base URL
BASE_URL="http://localhost:8080"

# Check if server is running
echo "📡 Checking if server is running..."
if ! curl -s "$BASE_URL/" > /dev/null; then
    echo "❌ Server is not running. Please start it with: go run main.go"
    exit 1
fi
echo "✅ Server is running!"

# Function to find a video file
find_video_file() {
    echo "🔍 Looking for video files..."
    
    # Common video file locations and extensions
    SEARCH_PATHS=(
        "$HOME/Desktop"
        "$HOME/Downloads"
        "$HOME/Movies"
        "."
    )
    
    VIDEO_EXTENSIONS=("*.mp4" "*.mov" "*.avi" "*.mkv" "*.wmv" "*.flv" "*.webm")
    
    for path in "${SEARCH_PATHS[@]}"; do
        if [ -d "$path" ]; then
            for ext in "${VIDEO_EXTENSIONS[@]}"; do
                video_file=$(find "$path" -maxdepth 1 -name "$ext" -type f | head -1)
                if [ -n "$video_file" ]; then
                    echo "✅ Found video file: $video_file"
                    echo "$video_file"
                    return 0
                fi
            done
        fi
    done
    
    echo "❌ No video file found in common locations"
    echo "Please provide a video file path as argument: $0 /path/to/video.mp4"
    return 1
}

# Get video file
if [ $# -eq 0 ]; then
    VIDEO_FILE=$(find_video_file)
    if [ $? -ne 0 ]; then
        exit 1
    fi
else
    VIDEO_FILE="$1"
    if [ ! -f "$VIDEO_FILE" ]; then
        echo "❌ File not found: $VIDEO_FILE"
        exit 1
    fi
fi

echo "📹 Using video file: $VIDEO_FILE"

# Test 1: Upload video
echo ""
echo "📤 Test 1: Uploading video..."
UPLOAD_RESPONSE=$(curl -s -X POST -F "video=@$VIDEO_FILE" "$BASE_URL/api/v1/upload")
echo "Response: $UPLOAD_RESPONSE"

# Extract video ID from response
VIDEO_ID=$(echo $UPLOAD_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -z "$VIDEO_ID" ]; then
    echo "❌ Failed to extract video ID from upload response"
    exit 1
fi

echo "✅ Video uploaded successfully! ID: $VIDEO_ID"

# Test 2: Check status
echo ""
echo "📊 Test 2: Checking transcoding status..."
for i in {1..5}; do
    STATUS_RESPONSE=$(curl -s "$BASE_URL/api/v1/videos/$VIDEO_ID/status")
    echo "Status check $i: $STATUS_RESPONSE"
    
    # Check if transcoding is complete
    if echo "$STATUS_RESPONSE" | grep -q '"status":"completed"'; then
        echo "✅ Transcoding completed!"
        break
    elif echo "$STATUS_RESPONSE" | grep -q '"status":"failed"'; then
        echo "❌ Transcoding failed!"
        break
    fi
    
    echo "⏳ Transcoding in progress... waiting 10 seconds"
    sleep 10
done

# Test 3: List videos
echo ""
echo "📋 Test 3: Listing all videos..."
LIST_RESPONSE=$(curl -s "$BASE_URL/api/v1/videos")
echo "Response: $LIST_RESPONSE"

# Test 4: Download links
echo ""
echo "📥 Test 4: Download links..."
echo "Original: $BASE_URL/api/v1/videos/$VIDEO_ID/download"
echo "480p: $BASE_URL/api/v1/videos/$VIDEO_ID/download?quality=480p"
echo "720p: $BASE_URL/api/v1/videos/$VIDEO_ID/download?quality=720p"
echo "1080p: $BASE_URL/api/v1/videos/$VIDEO_ID/download?quality=1080p"

echo ""
echo "🎉 All tests completed!"
echo "Video ID: $VIDEO_ID"