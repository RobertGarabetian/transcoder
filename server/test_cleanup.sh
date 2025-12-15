#!/bin/bash

# Test script for S3 cleanup functionality
# Usage: ./test_cleanup.sh

SERVER_URL="http://localhost:8080"

echo "🧪 Testing S3 Cleanup Functionality"
echo "===================================="
echo ""

# Step 1: List objects before cleanup
echo "📋 Step 1: Listing objects in S3 bucket..."
echo "GET $SERVER_URL/list"
echo ""
BEFORE=$(curl -s "$SERVER_URL/list")
echo "$BEFORE" | jq '.'
BEFORE_COUNT=$(echo "$BEFORE" | jq '.count')
echo ""
echo "Objects found: $BEFORE_COUNT"
echo ""

# Step 2: Trigger cleanup
echo "🧹 Step 2: Triggering cleanup..."
echo "POST $SERVER_URL/cleanup"
echo ""
CLEANUP_RESULT=$(curl -s -X POST "$SERVER_URL/cleanup")
echo "$CLEANUP_RESULT" | jq '.'
echo ""

# Step 3: List objects after cleanup
echo "📋 Step 3: Listing objects after cleanup..."
echo "GET $SERVER_URL/list"
echo ""
AFTER=$(curl -s "$SERVER_URL/list")
echo "$AFTER" | jq '.'
AFTER_COUNT=$(echo "$AFTER" | jq '.count')
echo ""
echo "Objects remaining: $AFTER_COUNT"
echo ""

# Summary
echo "===================================="
echo "✅ Test Summary:"
echo "   Before: $BEFORE_COUNT objects"
echo "   After:  $AFTER_COUNT objects"
echo "   Deleted: $((BEFORE_COUNT - AFTER_COUNT)) objects"
echo ""

