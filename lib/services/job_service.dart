import '../models/job.dart';
import 'job_agent_bridge.dart';
import 'cache_service.dart';

class JobService {
  static const List<Target> _faangTargets = [
    Target(provider: 'greenhouse', company: 'stripe'),
    Target(provider: 'greenhouse', company: 'airbnb'),
    Target(provider: 'greenhouse', company: 'roblox'),
    Target(provider: 'greenhouse', company: 'databricks'),
  ];

  static const List<Target> _itTargets = [
    Target(provider: 'greenhouse', company: 'asana'),
    Target(provider: 'greenhouse', company: 'scaleai'),
    Target(provider: 'lever', company: 'openai'),
  ];

  static const List<Target> _govTargets = [];

  static List<Target> getTargetsForCategory(String category) {
    switch (category) {
      case 'FAANG':
        return _faangTargets;
      case 'IT Company':
        return _itTargets;
      case 'Gov Job':
        return _govTargets;
      default:
        return [..._faangTargets, ..._itTargets];
    }
  }

  static Future<List<Job>> fetchJobs(String category, {bool useCache = true}) async {
    if (useCache) {
      final cached = await CacheService.getCachedJobs(category);
      if (cached != null && cached.isNotEmpty) {
        return cached;
      }
    }

    final targets = getTargetsForCategory(category);
    if (targets.isEmpty) {
      return [];
    }

    final jobs = await JobAgentBridge.scrapeMany(targets);
    
    if (jobs.isNotEmpty) {
      await CacheService.cacheJobs(category, jobs);
    }
    
    return jobs;
  }

  static List<Job> filterJobs(
    List<Job> jobs, {
    String? searchQuery,
    String? level,
    String? location,
    bool? isRemote,
  }) {
    var filtered = jobs;

    if (searchQuery != null && searchQuery.isNotEmpty) {
      final query = searchQuery.toLowerCase();
      filtered = filtered.where((job) {
        return job.title.toLowerCase().contains(query) ||
               job.company.toLowerCase().contains(query) ||
               (job.location?.toLowerCase().contains(query) ?? false);
      }).toList();
    }

    if (level != null && level.isNotEmpty) {
      filtered = filtered.where((job) {
        return job.level?.toLowerCase().contains(level.toLowerCase()) ?? false;
      }).toList();
    }

    if (location != null && location.isNotEmpty) {
      final loc = location.toLowerCase();
      filtered = filtered.where((job) {
        return job.location?.toLowerCase().contains(loc) ?? false;
      }).toList();
    }

    if (isRemote == true) {
      filtered = filtered.where((job) {
        final loc = job.location?.toLowerCase() ?? '';
        return loc.contains('remote') || loc.contains('work from home');
      }).toList();
    }

    return filtered;
  }
}
