import '../models/job.dart';
import 'job_agent_bridge.dart';
import 'cache_service.dart';

class JobService {
  // MAANG-tier global tech companies with India presence
  static const List<Target> _maangTargets = [
    Target(provider: 'greenhouse', company: 'doordash'),        // DoorDash - Pune, Hyderabad
    Target(provider: 'greenhouse', company: 'coinbase'),        // Coinbase - Hyderabad
    Target(provider: 'greenhouse', company: 'uberfreight'),     // Uber Freight - Hyderabad
    Target(provider: 'greenhouse', company: 'stripe'),          // Stripe - India
    Target(provider: 'greenhouse', company: 'databricks'),      // Databricks - India
    Target(provider: 'greenhouse', company: 'reddit'),          // Reddit - India
    Target(provider: 'greenhouse', company: 'pinterest'),       // Pinterest - India
    Target(provider: 'greenhouse', company: 'airbnb'),          // Airbnb - India
    Target(provider: 'greenhouse', company: 'lyft'),            // Lyft - India
  ];

  // Indian tech companies and startups
  static const List<Target> _itTargets = [
    Target(provider: 'greenhouse', company: 'policybazaar'),    // PolicyBazaar - India
    Target(provider: 'greenhouse', company: 'razorpay'),        // Razorpay - India
    Target(provider: 'greenhouse', company: 'swiggy'),          // Swiggy - India
    Target(provider: 'greenhouse', company: 'zomato'),          // Zomato - India
    Target(provider: 'lever', company: 'paytm'),                // Paytm - India
    Target(provider: 'greenhouse', company: 'meesho'),          // Meesho - India
    Target(provider: 'greenhouse', company: 'phonepe'),         // PhonePe - India
    Target(provider: 'greenhouse', company: 'cred'),            // CRED - India
    Target(provider: 'greenhouse', company: 'groww'),           // Groww - India
  ];

  static const List<Target> _govTargets = [];

  static List<Target> getTargetsForCategory(String category) {
    switch (category) {
      case 'MAANG':
        return _maangTargets;
      case 'IT Company':
        return _itTargets;
      case 'Gov Job':
        return _govTargets;
      default:
        return [..._maangTargets, ..._itTargets];
    }
  }

  static Future<List<Job>> fetchJobs(String category, {bool useCache = true}) async {
    if (useCache) {
      final cached = await CacheService.getCachedJobs(category);
      if (cached != null && cached.isNotEmpty) {
        // Shuffle cached jobs to show variety
        final shuffled = List<Job>.from(cached)..shuffle();
        return shuffled;
      }
    }

    final targets = getTargetsForCategory(category);
    if (targets.isEmpty) {
      return [];
    }

    final jobs = await JobAgentBridge.scrapeMany(targets);
    
    if (jobs.isNotEmpty) {
      // Shuffle jobs to mix companies together
      jobs.shuffle();
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
