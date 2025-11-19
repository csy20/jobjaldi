# JobJaldi - Feature Guide

## Quick Reference for New Features

### For Users

#### 🔍 Search Jobs
1. Tap the search icon in the app bar
2. Type your search query (searches title, company, location)
3. Results filter in real-time
4. Tap X to close search

#### 📌 Bookmark Jobs
1. Tap the bookmark icon on any job card
2. View all saved jobs by tapping the bookmark icon in app bar
3. Tap bookmark again to remove from saved jobs

#### 📄 View Job Details
1. Tap any job card
2. View full details including:
   - Company information
   - Location
   - Job level
   - Source
3. Tap "Apply Now" to open job application
4. Use share button to share job with others

#### 🔄 Refresh Jobs
1. Pull down on the job list to refresh
2. Or select a new category from the bottom sheet

#### 🎨 Dark Mode
- System automatically detects your device theme preference
- Can be extended to add manual toggle in settings

#### 🏷️ Filters
- Active filters appear as chips at the top
- Tap X on any chip to remove that filter
- Use menu (⋮) → "Clear Filters" to remove all

### For Developers

#### 📁 Project Structure

```
lib/
├── constants/
│   └── app_constants.dart       # All app constants
├── models/
│   └── job.dart                 # Job data model
├── providers/
│   └── job_provider.dart        # Riverpod state management
├── screens/
│   ├── home_screen.dart         # Main job listing
│   ├── job_detail_screen.dart   # Job details
│   └── bookmarks_screen.dart    # Saved jobs
├── services/
│   ├── job_agent_bridge.dart    # Native bridge
│   ├── job_service.dart         # Business logic
│   └── cache_service.dart       # Local storage
├── widgets/
│   └── job_card.dart            # Reusable job card
└── main.dart                     # App entry point
```

#### 🔧 Key Classes

##### JobNotifier (State Management)
```dart
// Load jobs
ref.read(jobProvider.notifier).loadJobs('FAANG', forceRefresh: true);

// Set search query
ref.read(jobProvider.notifier).setSearchQuery('engineer');

// Access state
final state = ref.watch(jobProvider);
final jobs = state.filteredJobs;
```

##### BookmarkNotifier
```dart
// Toggle bookmark
ref.read(bookmarkProvider.notifier).toggleBookmark(job);

// Check if bookmarked
final isBookmarked = ref.read(bookmarkProvider.notifier).isBookmarked(job);

// Get all bookmarks
final bookmarks = ref.watch(bookmarkProvider);
```

##### CacheService
```dart
// Cache jobs
await CacheService.cacheJobs('FAANG', jobs);

// Get cached jobs
final cached = await CacheService.getCachedJobs('FAANG');

// Bookmark operations
await CacheService.saveBookmark(job);
await CacheService.removeBookmark(job);
final bookmarks = await CacheService.getBookmarks();
```

##### JobService
```dart
// Fetch jobs
final jobs = await JobService.fetchJobs('FAANG', useCache: true);

// Filter jobs
final filtered = JobService.filterJobs(
  jobs,
  searchQuery: 'engineer',
  level: 'senior',
  isRemote: true,
);
```

#### 🎯 Adding New Features

##### Add a New Screen
1. Create file in `lib/screens/`
2. Extend `ConsumerWidget` or `ConsumerStatefulWidget`
3. Access providers with `ref.watch()` or `ref.read()`
4. Navigate: `Navigator.push(context, MaterialPageRoute(...))`

##### Add a New Filter
1. Add state field to `JobState` in `job_provider.dart`
2. Add setter method in `JobNotifier`
3. Update `filteredJobs` getter to use new filter
4. Add filter logic to `JobService.filterJobs()`
5. Add UI for filter in `home_screen.dart`

##### Add a New Cache Type
1. Add constant key in `app_constants.dart`
2. Add methods in `cache_service.dart`
3. Call from provider or service

#### 🧪 Testing Checklist

- [ ] Search functionality works
- [ ] Bookmarks persist after app restart
- [ ] Pull-to-refresh updates jobs
- [ ] Job details screen displays correctly
- [ ] Share functionality works
- [ ] Dark mode switches correctly
- [ ] Filters apply correctly
- [ ] Cache works (test within 2 minutes)
- [ ] Category selection works
- [ ] Error states display properly

#### 📝 Code Conventions

1. **Constants**: Always use `AppConstants` for magic numbers/strings
2. **Widgets**: Extract to separate files if >100 lines
3. **State**: Use Riverpod providers, avoid setState for complex state
4. **Async**: Always handle errors in async operations
5. **Navigation**: Use MaterialPageRoute for simple navigation
6. **Naming**: 
   - Screens: `*Screen`
   - Providers: `*Provider`
   - Services: `*Service`
   - Widgets: Descriptive names (e.g., `JobCard`)

#### 🔌 Extending the App

##### Add Backend API Integration
```dart
// In job_service.dart
static Future<List<Job>> fetchJobsFromAPI(String category) async {
  final response = await http.get(Uri.parse('$apiUrl/jobs/$category'));
  // Parse and return jobs
}
```

##### Add Pagination
```dart
// In JobState
final int currentPage;
final bool hasMore;

// In JobNotifier
Future<void> loadMoreJobs() async {
  if (state.hasMore && !state.isLoading) {
    // Fetch next page
  }
}
```

##### Add Notifications
```dart
// Use flutter_local_notifications
// Create notification service
// Schedule notifications for new jobs
```

#### 🐛 Common Issues

**Issue**: Jobs not loading
- Check internet connection
- Verify JobAgentBridge is initialized
- Check platform channel communication

**Issue**: Cache not working
- Ensure CacheService.init() is called in main()
- Check SharedPreferences permissions
- Verify cache timeout hasn't expired

**Issue**: Bookmarks not persisting
- Check CacheService initialization
- Verify JSON serialization in Job model
- Check app permissions for local storage

#### 📚 Resources

- [Riverpod Documentation](https://riverpod.dev)
- [Flutter Documentation](https://flutter.dev/docs)
- [Material Design 3](https://m3.material.io)

#### 🚀 Performance Tips

1. Use `const` constructors wherever possible
2. Keep `itemExtent` in ListView for fixed-height items
3. Cache network images (already have cached_network_image)
4. Limit rebuilds with proper Riverpod selectors
5. Profile app with Flutter DevTools

#### 🔐 Security Notes

- Never store sensitive data in SharedPreferences (use flutter_secure_storage)
- Validate all user inputs
- Sanitize URLs before launching
- Use HTTPS for API calls

---

**Need Help?** Check IMPLEMENTATION_SUMMARY.md for complete feature list and architecture details.
