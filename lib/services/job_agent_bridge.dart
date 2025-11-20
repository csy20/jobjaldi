import 'dart:convert';

import 'package:flutter/services.dart';

import '../models/job.dart';

class Target {
  const Target({
    required this.provider,
    required this.company,
  });

  final String provider;
  final String company;

  Map<String, String> toJson() => <String, String>{
        'provider': provider,
        'company': company,
      };
}

class JobAgentBridge {
  const JobAgentBridge._();

  static const MethodChannel _channel = MethodChannel('jobagent');

  static Future<List<Job>> scrapeMany(List<Target> targets) async {
    final payload = <String, dynamic>{
      'targets': targets.map((target) => target.toJson()).toList(),
    };

    try {
      final cfg = jsonEncode(payload);
      final response = await _channel.invokeMethod<String>('scrapeMany', cfg);
      return _parseJobs(response);
    } on PlatformException catch (e) {
      // Log error details for debugging
      print('PlatformException in scrapeMany: ${e.code} - ${e.message}');
      print('Details: ${e.details}');
      // Return empty list but don't hide the error
      return <Job>[];
    } catch (e) {
      print('Error in scrapeMany: $e');
      return <Job>[];
    }
  }

  static List<Job> _parseJobs(String? response) {
    if (response == null || response.isEmpty) {
      return <Job>[];
    }

    final decoded = jsonDecode(response);
    if (decoded is! List) {
      return <Job>[];
    }

    final deduped = <String, Job>{};
    for (final entry in decoded) {
      if (entry is Map) {
        final json = Map<String, dynamic>.from(entry);
        final job = Job.fromJson(json);
        deduped[job.dedupeKey()] = job;
      }
    }
    return deduped.values.toList();
  }
}
