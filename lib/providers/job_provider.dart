import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/job.dart';
import '../services/job_service.dart';
import '../services/cache_service.dart';

class JobState {
  final List<Job> jobs;
  final bool isLoading;
  final String? error;
  final String currentCategory;
  final String searchQuery;
  final String? levelFilter;
  final String? locationFilter;
  final bool remoteOnly;

  const JobState({
    this.jobs = const [],
    this.isLoading = false,
    this.error,
    this.currentCategory = 'MAANG',
    this.searchQuery = '',
    this.levelFilter,
    this.locationFilter,
    this.remoteOnly = false,
  });

  JobState copyWith({
    List<Job>? jobs,
    bool? isLoading,
    String? error,
    String? currentCategory,
    String? searchQuery,
    String? levelFilter,
    String? locationFilter,
    bool? remoteOnly,
  }) {
    return JobState(
      jobs: jobs ?? this.jobs,
      isLoading: isLoading ?? this.isLoading,
      error: error,
      currentCategory: currentCategory ?? this.currentCategory,
      searchQuery: searchQuery ?? this.searchQuery,
      levelFilter: levelFilter ?? this.levelFilter,
      locationFilter: locationFilter ?? this.locationFilter,
      remoteOnly: remoteOnly ?? this.remoteOnly,
    );
  }

  List<Job> get filteredJobs {
    return JobService.filterJobs(
      jobs,
      searchQuery: searchQuery.isEmpty ? null : searchQuery,
      level: levelFilter,
      location: locationFilter,
      isRemote: remoteOnly ? true : null,
    );
  }
}

class JobNotifier extends StateNotifier<JobState> {
  JobNotifier() : super(const JobState());

  Future<void> loadJobs(String category, {bool forceRefresh = false}) async {
    if (state.isLoading) return;

    state = state.copyWith(
      isLoading: true,
      currentCategory: category,
      error: null,
    );

    try {
      final jobs = await JobService.fetchJobs(category, useCache: !forceRefresh);
      state = state.copyWith(
        jobs: jobs,
        isLoading: false,
        error: null,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  void setSearchQuery(String query) {
    state = state.copyWith(searchQuery: query);
    if (query.isNotEmpty) {
      CacheService.addSearchQuery(query);
    }
  }

  void setLevelFilter(String? level) {
    state = state.copyWith(levelFilter: level);
  }

  void setLocationFilter(String? location) {
    state = state.copyWith(locationFilter: location);
  }

  void setRemoteOnly(bool remoteOnly) {
    state = state.copyWith(remoteOnly: remoteOnly);
  }

  void clearFilters() {
    state = state.copyWith(
      searchQuery: '',
      levelFilter: null,
      locationFilter: null,
      remoteOnly: false,
    );
  }
}

final jobProvider = StateNotifierProvider<JobNotifier, JobState>((ref) {
  return JobNotifier();
});

class BookmarkNotifier extends StateNotifier<List<Job>> {
  BookmarkNotifier() : super([]) {
    _loadBookmarks();
  }

  Future<void> _loadBookmarks() async {
    state = await CacheService.getBookmarks();
  }

  Future<void> toggleBookmark(Job job) async {
    final isBookmarked = state.any((j) => j.dedupeKey() == job.dedupeKey());
    
    if (isBookmarked) {
      await CacheService.removeBookmark(job);
      state = state.where((j) => j.dedupeKey() != job.dedupeKey()).toList();
    } else {
      await CacheService.saveBookmark(job);
      state = [...state, job];
    }
  }

  bool isBookmarked(Job job) {
    return state.any((j) => j.dedupeKey() == job.dedupeKey());
  }
}

final bookmarkProvider = StateNotifierProvider<BookmarkNotifier, List<Job>>((ref) {
  return BookmarkNotifier();
});

final themeModeProvider = StateNotifierProvider<ThemeModeNotifier, String>((ref) {
  return ThemeModeNotifier();
});

class ThemeModeNotifier extends StateNotifier<String> {
  ThemeModeNotifier() : super('system') {
    _loadThemeMode();
  }

  Future<void> _loadThemeMode() async {
    state = await CacheService.getThemeMode();
  }

  Future<void> setThemeMode(String mode) async {
    state = mode;
    await CacheService.saveThemeMode(mode);
  }
}
