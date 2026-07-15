import 'dart:async';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

enum HeartRatePhase { ready, running, review }

class HeartRateFlowScreen extends StatefulWidget {
  const HeartRateFlowScreen({
    super.key,
    required this.petId,
    required this.durationsSec,
  });

  final String petId;
  final List<int> durationsSec;

  @override
  State<HeartRateFlowScreen> createState() => _HeartRateFlowScreenState();
}

class _HeartRateFlowScreenState extends State<HeartRateFlowScreen> {
  HeartRatePhase phase = HeartRatePhase.ready;
  int selectedDuration = 60;
  int secondsLeft = 60;
  int taps = 0;
  Timer? timer;
  String? sessionId;
  Map<String, dynamic>? result;
  DateTime? lastTap;

  @override
  void initState() {
    super.initState();
    final durations = widget.durationsSec.isEmpty ? [60] : widget.durationsSec;
    selectedDuration = durations.first;
  }

  Future<void> start() async {
    final sess = await ApiClient.instance.startHeartRate(
      widget.petId,
      durationSec: selectedDuration,
    );
    setState(() {
      sessionId = sess['id'] as String?;
      phase = HeartRatePhase.running;
      secondsLeft = selectedDuration;
      taps = 0;
    });
    timer?.cancel();
    timer = Timer.periodic(const Duration(seconds: 1), (t) async {
      if (secondsLeft <= 1) {
        t.cancel();
        await finish();
      } else {
        setState(() => secondsLeft--);
      }
    });
  }

  void onTap() {
    final now = DateTime.now();
    if (lastTap != null && now.difference(lastTap!).inMilliseconds < 150) return;
    lastTap = now;
    setState(() => taps++);
  }

  Future<void> finish() async {
    if (sessionId == null) return;
    final data = await ApiClient.instance.completeHeartRate(sessionId!, taps);
    setState(() {
      phase = HeartRatePhase.review;
      result = data;
    });
  }

  Future<void> validate() async {
    if (sessionId == null) return;
    await ApiClient.instance.validateHeartRate(sessionId!);
    if (mounted) {
      final l10n = AppLocalizations.of(context)!;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(l10n.sentToVet)));
      Navigator.pop(context);
    }
  }

  Future<void> restart() async {
    if (sessionId != null) {
      await ApiClient.instance.cancelHeartRate(sessionId!);
    }
    setState(() {
      phase = HeartRatePhase.ready;
      sessionId = null;
      result = null;
    });
  }

  @override
  void dispose() {
    timer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final durations = widget.durationsSec.isEmpty ? [60] : widget.durationsSec;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.heartRate)),
      body: Padding(
        padding: const EdgeInsets.all(24),
        child: switch (phase) {
          HeartRatePhase.ready => Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (durations.length > 1) ...[
                  Text(l10n.chooseDuration),
                  const SizedBox(height: 8),
                  Wrap(
                    spacing: 8,
                    children: durations.map((d) {
                      return ChoiceChip(
                        label: Text(l10n.durationSeconds(d)),
                        selected: selectedDuration == d,
                        onSelected: (_) => setState(() => selectedDuration = d),
                      );
                    }).toList(),
                  ),
                  const SizedBox(height: 16),
                ],
                Text(l10n.heartRateInstructionsDuration(selectedDuration)),
                const SizedBox(height: 24),
                FilledButton(onPressed: start, child: Text(l10n.start)),
              ],
            ),
          HeartRatePhase.running => GestureDetector(
              onTap: onTap,
              behavior: HitTestBehavior.opaque,
              child: Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(l10n.secondsLeft(secondsLeft), style: Theme.of(context).textTheme.displayLarge),
                    const SizedBox(height: 16),
                    Text(l10n.beatsCount(taps), style: Theme.of(context).textTheme.headlineSmall),
                    const SizedBox(height: 24),
                    const Icon(Icons.favorite, size: 96),
                    Text(l10n.tapHere),
                  ],
                ),
              ),
            ),
          HeartRatePhase.review => Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  l10n.bpmLabel('${result?['bpm'] ?? '?'}'),
                  style: Theme.of(context).textTheme.headlineMedium,
                ),
                Text(l10n.beatsLabel(taps)),
                if (result?['isAlert'] == true)
                  Text(l10n.thresholdAlert, style: const TextStyle(color: Colors.orangeAccent)),
                const SizedBox(height: 24),
                FilledButton(onPressed: validate, child: Text(l10n.validateAndSend)),
                TextButton(onPressed: restart, child: Text(l10n.restart)),
              ],
            ),
        },
      ),
    );
  }
}
