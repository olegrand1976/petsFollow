import 'dart:async';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/billing_addon.dart';
import 'package:petsfollow_mobile/core/api/open_url.dart';
import 'package:petsfollow_mobile/core/review/in_app_review_helper.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

enum HeartRatePhase { ready, running, review }

class HeartRateFlowScreen extends StatefulWidget {
  const HeartRateFlowScreen({
    super.key,
    required this.petId,
    required this.durationsSec,
  });

  final String petId;

  /// Durations enabled by the pet's primary practice (vet settings).
  final List<int> durationsSec;

  @override
  State<HeartRateFlowScreen> createState() => _HeartRateFlowScreenState();
}

class _HeartRateFlowScreenState extends State<HeartRateFlowScreen> {
  HeartRatePhase phase = HeartRatePhase.ready;
  late int selectedDuration;
  int secondsLeft = 0;
  int taps = 0;
  Timer? timer;
  String? sessionId;
  Map<String, dynamic>? result;
  DateTime? lastTap;

  /// Practice-configured durations, ascending. Never invent options the vet did not enable.
  List<int> get _practiceDurations {
    final raw = widget.durationsSec
        .where((d) => d == 15 || d == 30 || d == 60)
        .toSet()
        .toList()
      ..sort();
    return raw;
  }

  @override
  void initState() {
    super.initState();
    final durations = _practiceDurations;
    // Prefer the longest duration the vet enabled (clinical default).
    selectedDuration = durations.isEmpty ? 60 : durations.last;
    secondsLeft = selectedDuration;
  }

  Future<void> start() async {
    final durations = _practiceDurations;
    if (durations.isEmpty || !durations.contains(selectedDuration)) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.heartRateNoDurationConfigured)),
      );
      return;
    }
    try {
      final sess = await ApiClient.instance.startHeartRate(
        widget.petId,
        durationSec: selectedDuration,
      );
      if (!mounted) return;
      setState(() {
        sessionId = sess['id'] as String?;
        phase = HeartRatePhase.running;
        secondsLeft = selectedDuration;
        taps = 0;
      });
      timer?.cancel();
      timer = Timer.periodic(const Duration(seconds: 1), (t) async {
        if (!mounted) {
          t.cancel();
          return;
        }
        if (secondsLeft <= 1) {
          t.cancel();
          await finish();
        } else {
          setState(() => secondsLeft--);
        }
      });
    } catch (_) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.errorGeneric('heartrate'))),
      );
    }
  }

  void onTap() {
    final now = DateTime.now();
    if (lastTap != null && now.difference(lastTap!).inMilliseconds < 150)
      return;
    lastTap = now;
    setState(() => taps++);
  }

  Future<void> finish() async {
    if (sessionId == null) return;
    try {
      final data = await ApiClient.instance.completeHeartRate(sessionId!, taps);
      if (!mounted) return;
      setState(() {
        phase = HeartRatePhase.review;
        result = data;
      });
    } catch (_) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.errorGeneric('heartrate'))),
      );
    }
  }

  Future<void> validate() async {
    if (sessionId == null) return;
    try {
      await ApiClient.instance.validateHeartRate(sessionId!);
    } catch (_) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.errorGeneric('heartrate'))),
      );
      return;
    }
    if (!mounted) return;
    final l10n = AppLocalizations.of(context)!;
    var showCarePlus = false;
    try {
      final ents = await AddonEntitlements.load();
      showCarePlus = !ents.hasCarePlus;
    } catch (_) {}
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(showCarePlus
            ? '${l10n.sentToVet}\n${l10n.carePlusUpsell}'
            : l10n.sentToVet),
        action: showCarePlus
            ? SnackBarAction(
                label: l10n.activateAddon,
                onPressed: () async {
                  try {
                    final url = await ApiClient.instance
                        .startAddonCheckout(addonCode: 'care_plus');
                    await openExternalUrl(url);
                  } catch (_) {}
                },
              )
            : null,
        duration: Duration(seconds: showCarePlus ? 6 : 3),
      ),
    );
    await _maybeAskReview(l10n);
    if (mounted) Navigator.pop(context);
  }

  Future<void> _maybeAskReview(AppLocalizations l10n) async {
    if (!await InAppReviewHelper.shouldShowDialog()) return;
    if (!mounted) return;
    final yes = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.reviewAskTitle),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: Text(l10n.reviewAskNo)),
          FilledButton(
              onPressed: () => Navigator.pop(ctx, true),
              child: Text(l10n.reviewAskYes)),
        ],
      ),
    );
    await InAppReviewHelper.recordAsked();
    if (yes == true) {
      await InAppReviewHelper.openStoreReview();
    }
  }

  Future<void> restart() async {
    timer?.cancel();
    final id = sessionId;
    if (id != null) {
      try {
        await ApiClient.instance.cancelHeartRate(id);
      } catch (_) {}
    }
    if (!mounted) return;
    setState(() {
      phase = HeartRatePhase.ready;
      sessionId = null;
      result = null;
    });
  }

  @override
  void dispose() {
    timer?.cancel();
    final id = sessionId;
    if (id != null) {
      unawaited(() async {
        try {
          await ApiClient.instance.cancelHeartRate(id);
        } catch (_) {}
      }());
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final durations = _practiceDurations;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.heartRate)),
      body: SafeArea(
        top: false,
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: switch (phase) {
            HeartRatePhase.ready => Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(l10n.chooseDuration),
                  const SizedBox(height: 8),
                  if (durations.isEmpty)
                    Text(l10n.heartRateNoDurationConfigured)
                  else
                    Wrap(
                      spacing: 8,
                      children: durations.map((d) {
                        return ChoiceChip(
                          label: Text(l10n.durationSeconds(d)),
                          selected: selectedDuration == d,
                          onSelected: durations.length == 1
                              ? null
                              : (_) => setState(() => selectedDuration = d),
                        );
                      }).toList(),
                    ),
                  const SizedBox(height: 16),
                  Text(l10n.heartRateInstructionsDuration(selectedDuration)),
                  const SizedBox(height: 24),
                  FilledButton(
                    onPressed: durations.isEmpty ? null : start,
                    child: Text(l10n.start),
                  ),
                ],
              ),
            HeartRatePhase.running => GestureDetector(
                onTap: onTap,
                behavior: HitTestBehavior.opaque,
                child: Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(l10n.secondsLeft(secondsLeft),
                          style: Theme.of(context).textTheme.displayLarge),
                      const SizedBox(height: 16),
                      Text(l10n.beatsCount(taps),
                          style: Theme.of(context).textTheme.headlineSmall),
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
                    Text(l10n.thresholdAlert,
                        style: const TextStyle(color: Colors.orangeAccent)),
                  const SizedBox(height: 24),
                  FilledButton(
                      onPressed: validate, child: Text(l10n.validateAndSend)),
                  TextButton(onPressed: restart, child: Text(l10n.restart)),
                ],
              ),
          },
        ),
      ),
    );
  }
}
