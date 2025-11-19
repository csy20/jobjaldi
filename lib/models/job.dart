import 'dart:convert';

class Job {
  const Job({
    required this.title,
    required this.company,
    required this.url,
    required this.source,
    this.location,
    this.level,
  });

  final String title;
  final String company;
  final String url;
  final String source;
  final String? location;
  final String? level;

  factory Job.fromJson(Map<String, dynamic> json) {
    return Job(
      title: json['title'] as String? ?? '',
      company: json['company'] as String? ?? '',
      url: json['url'] as String? ?? '',
      source: json['source'] as String? ?? '',
      location: json['location'] as String?,
      level: json['level'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'title': title,
      'company': company,
      'url': url,
      'source': source,
      if (location != null) 'location': location,
      if (level != null) 'level': level,
    };
  }

  String dedupeKey() {
    final parts = <String>[
      source,
      company,
      title,
      location ?? '',
      url,
    ];
    return parts.map((value) => value.toLowerCase()).join('|');
  }

  @override
  String toString() => jsonEncode(toJson());
}
