import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/features/heartrate/presentation/heart_rate_chart.dart';

class PetTimelineScreen extends StatefulWidget {
  const PetTimelineScreen({super.key, required this.petId});
  final String petId;

  @override
  State<PetTimelineScreen> createState() => _PetTimelineScreenState();
}

class _PetTimelineScreenState extends State<PetTimelineScreen> {
  List<dynamic> items = [];
  List<({DateTime date, int bpm, bool isAlert})> chartPoints = [];

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    final data = await ApiClient.instance.getTimeline(widget.petId);
    final sessions = await ApiClient.instance.getHeartRateSessions(widget.petId);
    setState(() {
      items = data;
      chartPoints = sessions
          .where((s) => s['bpm'] != null)
          .map((s) => (
                date: DateTime.parse(s['startedAt'] as String),
                bpm: s['bpm'] as int,
                isAlert: s['isAlert'] as bool? ?? false,
              ))
          .toList()
          .reversed
          .toList();
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.history)),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (chartPoints.isNotEmpty) ...[
            Text(l10n.latestValues, style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 8),
            HeartRateChart(points: chartPoints, height: 200),
            const SizedBox(height: 24),
          ],
          ...items.map((item) {
            final m = item as Map<String, dynamic>;
            return ListTile(
              title: Text(m['title'] as String? ?? ''),
              subtitle: Text(m['body'] as String? ?? ''),
            );
          }),
        ],
      ),
    );
  }
}
