# Testing S3 Auto-Cleanup

This document explains how to test the automatic S3 bucket cleanup functionality.

## 🧪 Testing Methods

### Method 1: Manual Test Endpoints (Recommended)

The server includes two test endpoints for easy testing:

#### 1. List Objects in Bucket
```bash
# List all objects currently in the S3 bucket
curl http://localhost:8080/list
```

**Response:**
```json
{
  "count": 5,
  "objects": [
    "raw/video1.mp4",
    "processed/video1/master.m3u8",
    ...
  ]
}
```

#### 2. Manually Trigger Cleanup
```bash
# Manually trigger cleanup (deletes all objects)
curl -X POST http://localhost:8080/cleanup
```

**Response:**
```json
{
  "status": "success",
  "deleted": 5,
  "remaining": 0,
  "message": "Deleted 5 objects"
}
```

### Method 2: Using the Test Script

A bash script is provided for automated testing:

```bash
# Make script executable
chmod +x test_cleanup.sh

# Run the test
./test_cleanup.sh
```

**Note:** Requires `jq` for JSON formatting. Install with:
- macOS: `brew install jq`
- Linux: `sudo apt-get install jq` or `sudo yum install jq`

### Method 3: Test with Shorter Interval

For testing the automatic cleanup (without waiting 2 hours), you can temporarily modify the ticker interval in `main.go`:

```go
// Change from 2 hours to 30 seconds for testing
ticker := time.NewTicker(30 * time.Second)
```

**⚠️ Remember to change it back to 2 hours after testing!**

## 📝 Step-by-Step Testing Guide

### Test 1: Verify List Endpoint
1. Start your server: `go run cmd/server/main.go`
2. Upload a video file (this will add objects to S3)
3. Check objects: `curl http://localhost:8080/list`
4. Verify you see the uploaded files

### Test 2: Manual Cleanup
1. List objects: `curl http://localhost:8080/list` (note the count)
2. Trigger cleanup: `curl -X POST http://localhost:8080/cleanup`
3. List again: `curl http://localhost:8080/list` (should show 0 objects)
4. Verify the response shows correct deletion count

### Test 3: Automatic Cleanup (with modified interval)
1. Modify the ticker to 30 seconds in `main.go`
2. Start the server
3. Upload a video file
4. Wait 30 seconds
5. Check server logs - you should see cleanup messages
6. List objects - should be empty

### Test 4: Error Handling
1. Test with invalid S3 credentials (should log errors but not crash)
2. Test with empty bucket (should handle gracefully)
3. Test with very large bucket (should handle pagination)

## 🔍 What to Check

### Server Logs
Look for these log messages:
```
S3 auto-cleanup goroutine started. Will delete all files every 2 hours.
Starting scheduled S3 bucket cleanup...
Deleted X objects from bucket transcoder.project
Successfully deleted all X objects from bucket transcoder.project
Scheduled cleanup completed successfully
```

### API Responses
- `/list` should return accurate object count
- `/cleanup` should return success status and deletion count
- Both endpoints should handle errors gracefully

### S3 Bucket
- Verify objects are actually deleted in your S3 console
- Check that the bucket is empty after cleanup
- Verify no partial deletions (all or nothing)

## 🐛 Troubleshooting

### Issue: "No objects found in bucket"
- This is normal if the bucket is empty
- Upload a test file first

### Issue: Cleanup endpoint returns error
- Check S3 credentials are configured correctly
- Verify bucket name is correct
- Check AWS region matches your bucket

### Issue: Objects not deleting
- Check S3 permissions (need DeleteObject permission)
- Verify bucket name in code matches actual bucket
- Check server logs for specific error messages

## 🎯 Expected Behavior

1. **On Server Start:**
   - Goroutine starts immediately
   - Log message: "S3 auto-cleanup goroutine started..."

2. **Every 2 Hours:**
   - Cleanup runs automatically
   - All objects are deleted
   - Success message logged

3. **Manual Cleanup:**
   - Immediate execution
   - Returns JSON response with deletion count
   - Objects removed from bucket

4. **Empty Bucket:**
   - Handles gracefully
   - Returns "No objects found" message
   - No errors thrown

## 📊 Testing Checklist

- [ ] List endpoint works correctly
- [ ] Manual cleanup deletes all objects
- [ ] Empty bucket handled gracefully
- [ ] Server logs show cleanup messages
- [ ] Automatic cleanup runs on schedule (with test interval)
- [ ] Multiple objects deleted correctly
- [ ] Error handling works (invalid credentials, etc.)
- [ ] Pagination works for large buckets (>1000 objects)

