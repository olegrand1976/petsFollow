import 'dart:async';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';

enum HeartRatePhase { ready, running, review }

class HeartRateFlowScreen extends StatefulWidget {
  const HeartRateFlowScreen({super.key, required this.petId});
  final String petId;

  @override
  State<HeartRateFlowScreen> createState() => _HeartRateFlowScreenState();
}

class _HeartRateFlowScreenState extends State<HeartRateFlowScreen> {
  HeartRatePhase phase = HeartRatePhase.ready;
  int secondsLeft = 60;
  int taps = 0;
  Timer? timer;
  String? sessionId;
  Map<String, dynamic>? result;
  DateTime? lastTap;

  Future<void> start() async {
    final sess = await ApiClient.instance.startHeartRate(widget.petId);
    setState(() {
      sessionId = sess['id'] as String?;
      phase = HeartRatePhase.running;
      secondsLeft = 60;
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
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Relevé envoyé au véto')));
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
    return Scaffold(
      appBar: AppBar(title: const Text('Relevé cardiaque')),
      body: Padding(
        padding: const EdgeInsets.all(24),
        child: switch (phase) {
          HeartRatePhase.ready => Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Tapotez à chaque battement pendant 60 secondes.'),
                const SizedBox(height: 24),
                FilledButton(onPressed: start, child: const Text('Démarrer')),
              ],
            ),
          HeartRatePhase.running => GestureDetector(
              onTap: onTap,
              behavior: HitTestBehavior.opaque,
              child: Center(
                child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
                  Text('$secondsLeft s', style: Theme.of(context).textTheme.displayLarge),
                  const SizedBox(height: 16),
                  Text('$taps battements', style: Theme.of(context).textTheme.headlineSmall),
                  const SizedBox(height: 24),
                  const Icon(Icons.favorite, size: 96),
                  const Text('Tapez ici à chaque battement'),
                ]),
              ),
            ),
          HeartRatePhase.review => Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('BPM: ${result?['bpm'] ?? '?'}', style: Theme.of(context).textTheme.headlineMedium),
                Text('Battements: $taps'),
                if (result?['isAlert'] == true) const Text('Alerte seuil', style: TextStyle(color: Colors.orangeAccent)),
                const SizedBox(height: 24),
                FilledButton(onPressed: validate, child: const Text('Valider et envoyer au véto')),
                TextButton(onPressed: restart, child: const Text('Recommencer')),
              ],
            ),
        },
      ),
    );
  }
}
