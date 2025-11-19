class AppConstants {
  static const int cacheTimeoutMinutes = 2;
  static const double cardElevation = 2.0;
  static const double borderRadius = 16.0;
  static const double cardMargin = 12.0;
  static const double listPadding = 16.0;
  static const double iconSize = 16.0;
  static const double companyLogoSize = 48.0;
  static const double listItemExtent = 120.0;
  static const double cacheExtent = 500.0;
  
  static const String cacheKeyPrefix = 'jobs_cache_';
  static const String cacheTimeKeyPrefix = 'jobs_cache_time_';
  static const String bookmarksKey = 'bookmarked_jobs';
  static const String searchHistoryKey = 'search_history';
  static const String themeKey = 'theme_mode';
  
  static const int maxSearchHistory = 10;
  static const int jobsPerPage = 20;
}
