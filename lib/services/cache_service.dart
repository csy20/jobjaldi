import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import '../models/job.dart';
import '../constants/app_constants.dart';

class CacheService {
  static SharedPreferences? _prefs;

  static Future<void> init() async {
    _prefs ??= await SharedPreferences.getInstance();
  }

  static Future<void> cacheJobs(String category, List<Job> jobs) async {
    await init();
    final key = AppConstants.cacheKeyPrefix + category;
    final timeKey = AppConstants.cacheTimeKeyPrefix + category;
    
    final jobsJson = jobs.map((job) => job.toJson()).toList();
    await _prefs!.setString(key, jsonEncode(jobsJson));
    await _prefs!.setString(timeKey, DateTime.now().toIso8601String());
  }

  static Future<List<Job>?> getCachedJobs(String category) async {
    await init();
    final key = AppConstants.cacheKeyPrefix + category;
    final timeKey = AppConstants.cacheTimeKeyPrefix + category;
    
    final cached = _prefs!.getString(key);
    final timeStr = _prefs!.getString(timeKey);
    
    if (cached == null || timeStr == null) return null;
    
    final cacheTime = DateTime.parse(timeStr);
    if (DateTime.now().difference(cacheTime).inMinutes > AppConstants.cacheTimeoutMinutes) {
      return null;
    }
    
    final decoded = jsonDecode(cached) as List;
    return decoded.map((json) => Job.fromJson(json)).toList();
  }

  static Future<void> saveBookmark(Job job) async {
    await init();
    final bookmarks = await getBookmarks();
    if (!bookmarks.any((j) => j.dedupeKey() == job.dedupeKey())) {
      bookmarks.add(job);
      final jobsJson = bookmarks.map((j) => j.toJson()).toList();
      await _prefs!.setString(AppConstants.bookmarksKey, jsonEncode(jobsJson));
    }
  }

  static Future<void> removeBookmark(Job job) async {
    await init();
    final bookmarks = await getBookmarks();
    bookmarks.removeWhere((j) => j.dedupeKey() == job.dedupeKey());
    final jobsJson = bookmarks.map((j) => j.toJson()).toList();
    await _prefs!.setString(AppConstants.bookmarksKey, jsonEncode(jobsJson));
  }

  static Future<List<Job>> getBookmarks() async {
    await init();
    final cached = _prefs!.getString(AppConstants.bookmarksKey);
    if (cached == null) return [];
    
    final decoded = jsonDecode(cached) as List;
    return decoded.map((json) => Job.fromJson(json)).toList();
  }

  static Future<bool> isBookmarked(Job job) async {
    final bookmarks = await getBookmarks();
    return bookmarks.any((j) => j.dedupeKey() == job.dedupeKey());
  }

  static Future<void> addSearchQuery(String query) async {
    await init();
    final history = await getSearchHistory();
    history.remove(query);
    history.insert(0, query);
    
    if (history.length > AppConstants.maxSearchHistory) {
      history.removeRange(AppConstants.maxSearchHistory, history.length);
    }
    
    await _prefs!.setStringList(AppConstants.searchHistoryKey, history);
  }

  static Future<List<String>> getSearchHistory() async {
    await init();
    return _prefs!.getStringList(AppConstants.searchHistoryKey) ?? [];
  }

  static Future<void> clearSearchHistory() async {
    await init();
    await _prefs!.remove(AppConstants.searchHistoryKey);
  }

  static Future<void> saveThemeMode(String mode) async {
    await init();
    await _prefs!.setString(AppConstants.themeKey, mode);
  }

  static Future<String> getThemeMode() async {
    await init();
    return _prefs!.getString(AppConstants.themeKey) ?? 'system';
  }
}
