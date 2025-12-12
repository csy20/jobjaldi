import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/job_provider.dart';
import '../widgets/job_card.dart';
import '../constants/app_constants.dart';
import 'job_detail_screen.dart';
import 'bookmarks_screen.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  final TextEditingController _searchController = TextEditingController();
  bool _showSearch = false;

  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      ref.read(jobProvider.notifier).loadJobs('MAANG');
    });
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _refreshJobs() async {
    final state = ref.read(jobProvider);
    await ref.read(jobProvider.notifier).loadJobs(
      state.currentCategory,
      forceRefresh: true,
    );
  }

  void _changeCategory(String category) {
    ref.read(jobProvider.notifier).loadJobs(category);
  }

  @override
  Widget build(BuildContext context) {
    final jobState = ref.watch(jobProvider);
    final filteredJobs = jobState.filteredJobs;

    return Scaffold(
      appBar: AppBar(
        title: _showSearch
            ? TextField(
                controller: _searchController,
                autofocus: true,
                decoration: const InputDecoration(
                  hintText: 'Search jobs...',
                  border: InputBorder.none,
                ),
                onChanged: (value) {
                  ref.read(jobProvider.notifier).setSearchQuery(value);
                },
              )
            : const Text('JobJaldi'),
        actions: [
          IconButton(
            icon: Icon(_showSearch ? Icons.close : Icons.search),
            onPressed: () {
              setState(() {
                _showSearch = !_showSearch;
                if (!_showSearch) {
                  _searchController.clear();
                  ref.read(jobProvider.notifier).setSearchQuery('');
                }
              });
            },
          ),
          IconButton(
            icon: const Icon(Icons.bookmark),
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (context) => const BookmarksScreen(),
                ),
              );
            },
          ),
          PopupMenuButton<String>(
            icon: const Icon(Icons.more_vert),
            onSelected: (value) {
              if (value == 'clear_filters') {
                ref.read(jobProvider.notifier).clearFilters();
                _searchController.clear();
              }
            },
            itemBuilder: (context) => [
              const PopupMenuItem(
                value: 'clear_filters',
                child: Text('Clear Filters'),
              ),
            ],
          ),
        ],
      ),
      body: Column(
        children: [
          if (jobState.searchQuery.isNotEmpty || 
              jobState.levelFilter != null || 
              jobState.locationFilter != null ||
              jobState.remoteOnly)
            _buildFilterChips(),
          Expanded(
            child: RefreshIndicator(
              onRefresh: _refreshJobs,
              child: jobState.isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : jobState.error != null
                      ? _buildErrorState(jobState.error!)
                          : filteredJobs.isEmpty
                          ? _buildEmptyState(jobState.currentCategory)
                          : ListView.builder(
                              padding: const EdgeInsets.all(AppConstants.listPadding),
                              cacheExtent: AppConstants.cacheExtent,
                              itemCount: filteredJobs.length,
                              itemBuilder: (context, index) {
                                final job = filteredJobs[index];
                                return JobCard(
                                  job: job,
                                  onTap: () {
                                    Navigator.push(
                                      context,
                                      MaterialPageRoute(
                                        builder: (context) => JobDetailScreen(job: job),
                                      ),
                                    );
                                  },
                                );
                              },
                            ),
            ),
          ),
        ],
      ),
      floatingActionButton: _buildCategoryFab(),
    );
  }

  Widget _buildFilterChips() {
    final jobState = ref.watch(jobProvider);
    
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: [
            if (jobState.searchQuery.isNotEmpty)
              Chip(
                label: Text('Search: ${jobState.searchQuery}'),
                onDeleted: () {
                  _searchController.clear();
                  ref.read(jobProvider.notifier).setSearchQuery('');
                },
              ),
            if (jobState.levelFilter != null) ...[
              const SizedBox(width: 8),
              Chip(
                label: Text('Level: ${jobState.levelFilter}'),
                onDeleted: () {
                  ref.read(jobProvider.notifier).setLevelFilter(null);
                },
              ),
            ],
            if (jobState.locationFilter != null) ...[
              const SizedBox(width: 8),
              Chip(
                label: Text('Location: ${jobState.locationFilter}'),
                onDeleted: () {
                  ref.read(jobProvider.notifier).setLocationFilter(null);
                },
              ),
            ],
            if (jobState.remoteOnly) ...[
              const SizedBox(width: 8),
              Chip(
                label: const Text('Remote Only'),
                onDeleted: () {
                  ref.read(jobProvider.notifier).setRemoteOnly(false);
                },
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildErrorState(String error) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.error_outline, size: 64, color: Colors.red.shade300),
          const SizedBox(height: 16),
          Text(
            'Failed to fetch jobs',
            style: TextStyle(
              fontSize: 18,
              color: Colors.grey.shade600,
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(height: 8),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 32),
            child: Text(
              error,
              style: TextStyle(color: Colors.grey.shade500),
              textAlign: TextAlign.center,
            ),
          ),
          const SizedBox(height: 16),
          ElevatedButton.icon(
            onPressed: _refreshJobs,
            icon: const Icon(Icons.refresh),
            label: const Text('Retry'),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState(String category) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.work_off_outlined, size: 64, color: Colors.grey.shade400),
          const SizedBox(height: 16),
          Text(
            'No $category jobs found',
            style: TextStyle(
              fontSize: 18,
              color: Colors.grey.shade600,
              fontWeight: FontWeight.w500,
            ),
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

  Widget _buildCategoryFab() {
    return FloatingActionButton.extended(
      onPressed: () {
        _showCategoryBottomSheet();
      },
      icon: const Icon(Icons.filter_list),
      label: const Text('Select Category'),
    );
  }

  void _showCategoryBottomSheet() {
    showModalBottomSheet(
      context: context,
      builder: (context) {
        final currentCategory = ref.read(jobProvider).currentCategory;
        
        return Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const SizedBox(height: 16),
            const Text(
              'Select Job Category',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const Divider(),
            ListTile(
              leading: const Icon(Icons.business),
              title: const Text('MAANG'),
              selected: currentCategory == 'MAANG',
              onTap: () {
                _changeCategory('MAANG');
                Navigator.pop(context);
              },
            ),
            ListTile(
              leading: const Icon(Icons.computer),
              title: const Text('IT Company'),
              selected: currentCategory == 'IT Company',
              onTap: () {
                _changeCategory('IT Company');
                Navigator.pop(context);
              },
            ),
            ListTile(
              leading: const Icon(Icons.account_balance),
              title: const Text('Gov Job'),
              selected: currentCategory == 'Gov Job',
              onTap: () {
                _changeCategory('Gov Job');
                Navigator.pop(context);
              },
            ),
            ListTile(
              leading: const Icon(Icons.all_inclusive),
              title: const Text('All'),
              selected: currentCategory == 'All',
              onTap: () {
                _changeCategory('All');
                Navigator.pop(context);
              },
            ),
            const SizedBox(height: 16),
          ],
        );
      },
    );
  }
}
