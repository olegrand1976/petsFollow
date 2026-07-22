import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/media_url.dart';
import 'package:petsfollow_mobile/core/api/open_url.dart';
import 'package:petsfollow_mobile/features/heartrate/presentation/heart_rate_chart.dart';
import 'package:petsfollow_mobile/features/heartrate/presentation/heart_rate_flow_screen.dart';
import 'package:petsfollow_mobile/features/messaging/presentation/messaging_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_form_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_timeline_screen.dart';
import 'package:petsfollow_mobile/features/settings/presentation/settings_menu_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key, required this.onLogout});

  final VoidCallback onLogout;

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  List<dynamic> pets = [];
  String? userName;
  Map<String, List<({DateTime date, int bpm, bool isAlert})>> chartData = {};

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    try {
      final me = await ApiClient.instance.getMe();
      userName = me['fullName'] as String?;
    } catch (_) {}
    final data = await ApiClient.instance.getPets();
    final charts = <String, List<({DateTime date, int bpm, bool isAlert})>>{};
    for (final p in data) {
      final pet = p as Map<String, dynamic>;
      final petId = pet['id'] as String;
      try {
        final sessions = await ApiClient.instance.getHeartRateSessions(petId);
        charts[petId] = sessions
            .where((s) => s['bpm'] != null)
            .map((s) => (
                  date: DateTime.parse(s['startedAt'] as String),
                  bpm: s['bpm'] as int,
                  isAlert: s['isAlert'] as bool? ?? false,
                ))
            .take(7)
            .toList()
            .reversed
            .toList();
      } catch (_) {
        charts[petId] = [];
      }
    }
    setState(() {
      pets = data;
      chartData = charts;
    });
  }

  List<int> _durationsForPet(Map<String, dynamic> pet) {
    final raw = pet['heartrateDurationsSec'] as List<dynamic>?;
    if (raw == null || raw.isEmpty) return [60];
    return raw.map((e) => e as int).toList();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final greeting = userName != null && userName!.isNotEmpty
        ? l10n.greeting(userName!.split(' ').first)
        : l10n.myPets;

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.myPets),
        actions: [
          IconButton(
            icon: const Icon(Icons.settings),
            tooltip: l10n.settings,
            onPressed: () => Navigator.push(
              context,
              MaterialPageRoute(
                builder: (_) => SettingsMenuScreen(onLogout: widget.onLogout),
              ),
            ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () async {
          await Navigator.push(context, MaterialPageRoute(builder: (_) => const PetFormScreen()));
          load();
        },
        child: const Icon(Icons.add),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text(greeting, style: Theme.of(context).textTheme.headlineSmall),
          const SizedBox(height: 16),
          ...pets.map((p) {
            final pet = p as Map<String, dynamic>;
            final petId = pet['id'] as String;
            final status = pet['paymentStatus'] as String? ?? 'pending_payment';
            final ent = pet['entitlement'] as Map<String, dynamic>?;
            final points = chartData[petId] ?? [];
            return Card(
              margin: const EdgeInsets.only(bottom: 12),
              child: InkWell(
                onTap: () => _openPetMenu(context, petId, pet['name'] as String, status, ent, _durationsForPet(pet)),
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          CircleAvatar(
                            backgroundImage: (pet['photoUrl'] as String?)?.isNotEmpty == true
                                ? NetworkImage(resolveMediaUrl(pet['photoUrl'] as String)!)
                                : null,
                            child: (pet['photoUrl'] as String?)?.isNotEmpty == true
                                ? null
                                : Text((pet['name'] as String? ?? '?').substring(0, 1).toUpperCase()),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(pet['name'] as String? ?? '', style: const TextStyle(fontWeight: FontWeight.bold)),
                                Text('${pet['species']} · ${pet['breed']}'),
                              ],
                            ),
                          ),
                        ],
                      ),
                      if (status == 'active' && points.isNotEmpty) ...[
                        const SizedBox(height: 12),
                        Text(l10n.latestValues, style: Theme.of(context).textTheme.labelLarge),
                        const SizedBox(height: 8),
                        HeartRateChart(points: points),
                      ],
                      if (status == 'active')
                        Padding(
                          padding: const EdgeInsets.only(top: 12),
                          child: FilledButton(
                            onPressed: () {
                              Navigator.push(
                                context,
                                MaterialPageRoute(
                                  builder: (_) => HeartRateFlowScreen(
                                    petId: petId,
                                    durationsSec: _durationsForPet(pet),
                                  ),
                                ),
                              ).then((_) => load());
                            },
                            child: Text(l10n.startMeasurement),
                          ),
                        ),
                    ],
                  ),
                ),
              ),
            );
          }),
        ],
      ),
    );
  }

  void _openPetMenu(
    BuildContext context,
    String petId,
    String name,
    String status,
    Map<String, dynamic>? ent,
    List<int> durations,
  ) {
    final l10n = AppLocalizations.of(context)!;
    showModalBottomSheet(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(title: Text(name, style: const TextStyle(fontWeight: FontWeight.bold))),
            if (status == 'pending_payment')
              ListTile(
                leading: const Icon(Icons.payment),
                title: Text(l10n.paymentResume),
                onTap: () async {
                  Navigator.pop(context);
                  final url = await ApiClient.instance.resumeCheckout(petId);
                  await openExternalUrl(url);
                  load();
                },
              ),
            if (status == 'active' && ent?['billingMode'] == 'subscription')
              ListTile(
                leading: const Icon(Icons.settings),
                title: Text(l10n.manageSubscription),
                onTap: () async {
                  Navigator.pop(context);
                  final url = await ApiClient.instance.billingPortal(petId);
                  await openExternalUrl(url);
                },
              ),
            ListTile(
              leading: const Icon(Icons.photo_camera_outlined),
              title: Text(l10n.changePhoto),
              onTap: () async {
                Navigator.pop(context);
                final picker = ImagePicker();
                final file = await picker.pickImage(
                  source: ImageSource.gallery,
                  maxWidth: 1024,
                  imageQuality: 85,
                );
                if (file == null) return;
                try {
                  await ApiClient.instance.uploadPetPhoto(petId, file.path);
                  if (context.mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text(l10n.photoUpdated)),
                    );
                  }
                  load();
                } catch (_) {
                  if (context.mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text(l10n.errorGeneric('photo'))),
                    );
                  }
                }
              },
            ),
            ListTile(
              leading: const Icon(Icons.favorite),
              title: Text(l10n.heartRate),
              onTap: () {
                Navigator.pop(context);
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => HeartRateFlowScreen(petId: petId, durationsSec: durations),
                  ),
                ).then((_) => load());
              },
            ),
            ListTile(
              leading: const Icon(Icons.history),
              title: Text(l10n.history),
              onTap: () {
                Navigator.pop(context);
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (_) => PetTimelineScreen(petId: petId)),
                );
              },
            ),
            ListTile(
              leading: const Icon(Icons.chat),
              title: Text(l10n.vetMessaging),
              onTap: () {
                Navigator.pop(context);
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (_) => const MessagingScreen()),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}
