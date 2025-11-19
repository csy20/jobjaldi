# Troubleshooting Guide - JobJaldi Loading Issues

## Issue: App Stuck Loading / No Jobs Showing

If the app is stuck loading for a long time without showing jobs, try these solutions:

### Quick Fixes

1. **Reset Circuit Breakers** (if they're blocking requests)
   - The circuit breakers might be open after initial failures
   - Restart the app to reset circuit breakers
   - Or rebuild the AAR to reset state

2. **Check Network Connection**
   - Ensure device has internet connectivity
   - Try switching between WiFi and mobile data

3. **Clear App Cache**
   - Restart the app
   - Or clear app data in Android settings

### Debugging Steps

1. **Check Logs**
   - Look for errors in Android logcat:
     ```bash
     adb logcat | grep -i "JobagentChannel\|jobagent\|scrape"
     ```
   - Look for Flutter console output:
     ```bash
     flutter logs
     ```

2. **Common Error Patterns**

   **Circuit Breaker Open:**
   ```
   Error: circuit breaker is open
   ```
   - Solution: Restart app or wait 15 seconds for auto-recovery

   **Rate Limiting:**
   - If rate limiter is blocking, requests will wait
   - Should not block indefinitely (max wait is based on rate)

   **Network Errors:**
   ```
   Error: greenhouse request failed: ...
   ```
   - Check internet connection
   - Verify API endpoints are accessible

   **Timeout Errors:**
   ```
   Error: context cancelled: context deadline exceeded
   ```
   - Increase timeout in `jobagent.go` if needed
   - Check if requests are taking too long

3. **Test Individual Components**

   **Test Scraping Directly:**
   ```bash
   cd scrapers
   go run test_scraper.go
   ```

   **Check if AAR is Built:**
   ```bash
   ls -lh android/app/libs/jobagent.aar
   ```

### Code-Level Fixes Applied

1. **Better Error Handling**
   - Errors now propagate to Flutter instead of being swallowed
   - Added logging in Android layer
   - Better error messages in UI

2. **Circuit Breaker Improvements**
   - Less aggressive (10 failures instead of 5)
   - Faster recovery (15s instead of 30s)
   - Added reset functions

3. **Rate Limiting Fixes**
   - Non-blocking check before waiting
   - Prevents indefinite blocking

4. **Partial Success Handling**
   - Returns jobs even if some targets fail
   - Better error reporting

### Manual Reset Options

If circuit breakers are stuck:

1. **Restart App** - Simplest solution
2. **Rebuild AAR** - Resets all state:
   ```bash
   cd scrapers
   ./scripts/build_aar.sh
   ```
3. **Clear Cache** - In Go code:
   ```go
   jobagent.ClearCache()
   jobagent.ResetCircuitBreakers()
   ```

### Expected Behavior

- **First Load**: 2-5 seconds (depending on network)
- **Cached Load**: <100ms
- **With Retries**: Up to 10 seconds (3 retries × ~3s each)
- **Circuit Breaker Open**: Returns error immediately, recovers in 15s

### If Still Not Working

1. **Verify AAR is Up-to-Date**
   ```bash
   cd scrapers
   ./scripts/build_aar.sh
   cd ../android
   ./gradlew clean
   ```

2. **Check Go Module Dependencies**
   ```bash
   cd scrapers
   go mod tidy
   go mod verify
   ```

3. **Test Scraping Independently**
   ```bash
   cd scrapers
   go run test_scraper.go
   ```

4. **Check Flutter Dependencies**
   ```bash
   flutter pub get
   flutter clean
   flutter pub get
   ```

5. **Rebuild Everything**
   ```bash
   # Rebuild AAR
   cd scrapers && ./scripts/build_aar.sh
   
   # Rebuild Android
   cd ../android && ./gradlew clean build
   
   # Rebuild Flutter
   cd .. && flutter clean && flutter pub get
   ```

### Performance Expectations

- **Normal Operation**: 2-5 seconds for first load
- **Cached**: <100ms
- **With Failures**: Up to 10 seconds with retries
- **Circuit Breaker**: Immediate failure, 15s recovery

If loading takes more than 30 seconds, there's likely an issue that needs investigation.


