import 'dart:math' as math;

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';

import 'models/job.dart';
import 'services/job_agent_bridge.dart';

void main() {
  runApp(const JobApp());
}

class JobApp extends StatelessWidget {
  const JobApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'JobJaldi',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF00BFA5), // Teal Accent
          brightness: Brightness.light,
          surface: const Color(0xFFF5F7FA),
        ),
        useMaterial3: true,
        appBarTheme: const AppBarTheme(
          centerTitle: true,
          elevation: 0,
          backgroundColor: Colors.white,
          titleTextStyle: TextStyle(
            color: Colors.black87,
            fontSize: 22,
            fontWeight: FontWeight.bold,
            letterSpacing: -0.5,
          ),
        ),
        cardTheme: CardThemeData(
          elevation: 2,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
          ),
          color: Colors.white,
        ),
      ),
      home: const HomeScreen(),
    );
  }
}

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  List<Job> _jobs = <Job>[];
  bool _isFetching = false;
  String _currentCategory = 'All';
  DateTime? _lastFetchTime;

  // Define targets for different categories
  static const List<Target> _faangTargets = [
    Target(provider: 'greenhouse', company: 'stripe'),
    Target(provider: 'greenhouse', company: 'airbnb'),
    Target(provider: 'greenhouse', company: 'roblox'),
    Target(provider: 'greenhouse', company: 'databricks'),
  ];

  static const List<Target> _itTargets = [
    Target(provider: 'greenhouse', company: 'asana'),
    Target(provider: 'greenhouse', company: 'scaleai'),
    Target(provider: 'lever', company: 'openai'), // Assuming adapter exists/works
  ];

  static const List<Target> _govTargets = []; // Placeholder

  @override
  void initState() {
    super.initState();
    _refreshJobs('FAANG');
  }

  Future<void> _refreshJobs(String category) async {
    if (_isFetching) return;

    // Avoid redundant fetches within 2 minutes
    if (_currentCategory == category && 
        _lastFetchTime != null && 
        DateTime.now().difference(_lastFetchTime!).inMinutes < 2) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Using cached ${category} jobs'),
          duration: const Duration(seconds: 1),
          behavior: SnackBarBehavior.floating,
        ),
      );
      return;
    }

    setState(() {
      _isFetching = true;
      // Check if category changed before updating
      if (_currentCategory != category) {
        _jobs = [];
      }
      _currentCategory = category;
    });

    final messenger = ScaffoldMessenger.of(context);
    List<Target> targets;

    switch (category) {
      case 'FAANG':
        targets = _faangTargets;
        break;
      case 'IT Company':
        targets = _itTargets;
        break;
      case 'Gov Job':
        targets = _govTargets;
        break;
      default:
        targets = [..._faangTargets, ..._itTargets];
    }

    try {
      if (targets.isEmpty) {
        await Future.delayed(const Duration(milliseconds: 500));
        if (mounted) setState(() => _jobs = []);
      } else {
        final jobs = await JobAgentBridge.scrapeMany(targets);
        if (mounted) {
          setState(() {
            _jobs = jobs;
            _lastFetchTime = DateTime.now();
          });
        }
      }

      if (mounted) {
        messenger.hideCurrentSnackBar();
        messenger.showSnackBar(
          SnackBar(
            content: Text('Found ${_jobs.length} $category jobs'),
            behavior: SnackBarBehavior.floating,
            shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(10)),
            duration: const Duration(seconds: 2),
          ),
        );
      }
    } catch (error) {
      if (mounted) {
        // Log error for debugging
        print('Error fetching jobs: $error');
        messenger.hideCurrentSnackBar();
        messenger.showSnackBar(
          SnackBar(
            content: Text('Failed to fetch jobs: ${error.toString()}'),
            backgroundColor: Colors.red,
            duration: const Duration(seconds: 5),
            action: SnackBarAction(
              label: 'Retry',
              textColor: Colors.white,
              onPressed: () => _refreshJobs(category),
            ),
          ),
        );
        // Set fetching to false even on error
        setState(() {
          _isFetching = false;
        });
      }
    } finally {
      if (mounted) {
        setState(() => _isFetching = false);
      }
    }
  }

  Future<void> _openJob(Job job) async {
    final uri = Uri.tryParse(job.url);
    if (uri == null) return;
    await launchUrl(uri, mode: LaunchMode.externalApplication);
  }

  @override
  Widget build(BuildContext context) {
    final jobs = _jobs;

    return Scaffold(
      appBar: AppBar(
        title: const Text('JobJaldi'),
      ),
      body: _isFetching
          ? const Center(child: CircularProgressIndicator())
          : jobs.isEmpty
              ? _buildEmptyState()
              : ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: jobs.length,
                  itemBuilder: (context, index) {
                    final job = jobs[index];
                    return _buildJobCard(job);
                  },
                ),
      floatingActionButton: ExpandableFab(
        distance: 112.0,
        children: [
          ActionButton(
            onPressed: () => _refreshJobs('FAANG'),
            icon: const Icon(Icons.business),
            label: 'FAANG',
          ),
          ActionButton(
            onPressed: () => _refreshJobs('IT Company'),
            icon: const Icon(Icons.computer),
            label: 'IT Company',
          ),
          ActionButton(
            onPressed: () => _refreshJobs('Gov Job'),
            icon: const Icon(Icons.account_balance),
            label: 'Gov Job',
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.work_off_outlined, size: 64, color: Colors.grey.shade400),
          const SizedBox(height: 16),
          Text(
            'No ${_currentCategory} jobs found',
            style: TextStyle(fontSize: 18, color: Colors.grey.shade600, fontWeight: FontWeight.w500),
          ),
          const SizedBox(height: 8),
          Text(
            'Try adjusting filters or select another category',
            style: TextStyle(color: Colors.grey.shade500),
          ),
        ],
      ),
    );
  }

  Widget _buildJobCard(Job job) {
    final levelText = (job.level ?? '').isEmpty ? null : job.level;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: InkWell(
        onTap: () => _openJob(job),
        borderRadius: BorderRadius.circular(16),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    width: 48,
                    height: 48,
                    decoration: BoxDecoration(
                      color: Colors.teal.shade50,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Center(
                      child: Text(
                        job.company.isNotEmpty ? job.company[0].toUpperCase() : '?',
                        style: const TextStyle(
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                          color: Colors.teal,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          job.title,
                          style: const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            height: 1.2,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          job.company,
                          style: TextStyle(
                            fontSize: 14,
                            color: Colors.grey.shade700,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  if (job.location != null && job.location!.isNotEmpty) ...[
                    Icon(Icons.location_on_outlined, size: 16, color: Colors.grey.shade500),
                    const SizedBox(width: 4),
                    Expanded(
                      child: Text(
                        job.location!,
                        style: TextStyle(fontSize: 13, color: Colors.grey.shade600),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  ],
                  if (levelText != null) ...[
                    const SizedBox(width: 8),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                      decoration: BoxDecoration(
                        color: Colors.blue.shade50,
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Text(
                        levelText.toUpperCase(),
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.bold,
                          color: Colors.blue.shade700,
                        ),
                      ),
                    ),
                  ],
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// --- Animated FAB Implementation ---

@immutable
class ExpandableFab extends StatefulWidget {
  const ExpandableFab({
    super.key,
    this.initialOpen,
    required this.distance,
    required this.children,
  });

  final bool? initialOpen;
  final double distance;
  final List<Widget> children;

  @override
  State<ExpandableFab> createState() => _ExpandableFabState();
}

class _ExpandableFabState extends State<ExpandableFab>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _expandAnimation;
  bool _open = false;

  @override
  void initState() {
    super.initState();
    _open = widget.initialOpen ?? false;
    _controller = AnimationController(
      value: _open ? 1.0 : 0.0,
      duration: const Duration(milliseconds: 350), // Slightly slower for elegance
      vsync: this,
    );
    _expandAnimation = CurvedAnimation(
      curve: Curves.fastOutSlowIn,
      parent: _controller,
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _toggle() {
    setState(() {
      _open = !_open;
      if (_open) {
        _controller.forward();
      } else {
        _controller.reverse();
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Stack(
      alignment: Alignment.bottomRight,
      clipBehavior: Clip.none,
      fit: StackFit.expand,
      children: [
        _buildScrim(),
        Positioned(
          right: 16,
          bottom: 16,
          child: _buildTapToCloseFab(),
        ),
        ..._buildExpandingActionButtons(),
        Positioned(
          right: 16,
          bottom: 16,
          child: _buildTapToOpenFab(),
        ),
      ],
    );
  }

  Widget _buildScrim() {
    return AnimatedBuilder(
      animation: _expandAnimation,
      builder: (context, child) {
        return IgnorePointer(
          ignoring: !_open,
          child: GestureDetector(
            onTap: _toggle,
            child: Container(
              color: Colors.black.withOpacity(_expandAnimation.value * 0.4),
            ),
          ),
        );
      },
    );
  }

  Widget _buildTapToCloseFab() {
    return SizedBox(
      width: 56.0,
      height: 56.0,
      child: Center(
        child: Material(
          shape: const CircleBorder(),
          clipBehavior: Clip.antiAlias,
          elevation: 4.0,
          child: InkWell(
            onTap: _toggle,
            child: Padding(
              padding: const EdgeInsets.all(8.0),
              child: Icon(
                Icons.close,
                color: Theme.of(context).primaryColor,
              ),
            ),
          ),
        ),
      ),
    );
  }

  List<Widget> _buildExpandingActionButtons() {
    final children = <Widget>[];
    final count = widget.children.length;
    final step = 90.0 / (count - 1);
    for (var i = 0, angleInDegrees = 0.0;
        i < count;
        i++, angleInDegrees += step) {
      children.add(
        _ExpandingActionButton(
          // 90 degrees is Up (dy=1, dx=0 -> bottom increases, right static)
          // 0 degrees is Left (dy=0, dx=1 -> bottom static, right increases)
          directionInDegrees: 90.0 - angleInDegrees,
          maxDistance: widget.distance,
          progress: _expandAnimation,
          index: i,
          child: widget.children[i],
        ),
      );
    }
    return children;
  }

  Widget _buildTapToOpenFab() {
    return IgnorePointer(
      ignoring: _open,
      child: AnimatedContainer(
        transformAlignment: Alignment.center,
        transform: Matrix4.diagonal3Values(
          _open ? 0.7 : 1.0,
          _open ? 0.7 : 1.0,
          1.0,
        ),
        duration: const Duration(milliseconds: 250),
        curve: const Interval(0.0, 0.5, curve: Curves.easeOut),
        child: AnimatedOpacity(
          opacity: _open ? 0.0 : 1.0,
          curve: const Interval(0.25, 1.0, curve: Curves.easeInOut),
          duration: const Duration(milliseconds: 250),
          child: FloatingActionButton.extended(
            onPressed: _toggle,
            icon: const Icon(Icons.filter_list),
            label: const Text('Select Category'),
            elevation: 4,
            highlightElevation: 8,
          ),
        ),
      ),
    );
  }
}

@immutable
class _ExpandingActionButton extends StatelessWidget {
  const _ExpandingActionButton({
    required this.directionInDegrees,
    required this.maxDistance,
    required this.progress,
    required this.child,
    required this.index,
  });

  final double directionInDegrees;
  final double maxDistance;
  final Animation<double> progress;
  final Widget child;
  final int index;

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: progress,
      builder: (context, child) {
        final offset = Offset.fromDirection(
          directionInDegrees * (math.pi / 180.0),
          progress.value * maxDistance,
        );
        
        // Staggered fade and scale
        final delay = index * 0.1;
        final intervalStart = delay;
        final intervalEnd = (delay + 0.5).clamp(0.0, 1.0);
        
        final curvedValue = Curves.easeOutBack.transform(
          ((progress.value - intervalStart) / (intervalEnd - intervalStart))
              .clamp(0.0, 1.0),
        );

        return Positioned(
          right: 16.0 + offset.dx,
          bottom: 16.0 + offset.dy,
          child: Transform.scale(
            scale: curvedValue,
            child: Opacity(
              opacity: curvedValue.clamp(0.0, 1.0),
              child: child!,
            ),
          ),
        );
      },
      child: child,
    );
  }
}

@immutable
class ActionButton extends StatelessWidget {
  const ActionButton({
    super.key,
    this.onPressed,
    required this.icon,
    required this.label,
  });

  final VoidCallback? onPressed;
  final Widget icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withOpacity(0.15),
                blurRadius: 8,
                offset: const Offset(0, 4),
              ),
            ],
          ),
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          child: Text(
            label,
            style: const TextStyle(
              fontWeight: FontWeight.w600,
              fontSize: 14,
              color: Colors.black87,
            ),
          ),
        ),
        const SizedBox(width: 12),
        Material(
          shape: const CircleBorder(),
          clipBehavior: Clip.antiAlias,
          color: theme.colorScheme.primary,
          elevation: 6.0,
          shadowColor: theme.colorScheme.primary.withOpacity(0.4),
          child: IconButton(
            onPressed: onPressed,
            icon: icon,
            color: theme.colorScheme.onPrimary,
            padding: const EdgeInsets.all(12),
            iconSize: 24,
          ),
        ),
      ],
    );
  }
}
