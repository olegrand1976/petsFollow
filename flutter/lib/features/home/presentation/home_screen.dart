import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/features/heartrate/presentation/heart_rate_flow_screen.dart';
import 'package:petsfollow_mobile/features/messaging/presentation/messaging_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_form_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_timeline_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  List<dynamic> pets = [];

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    final data = await ApiClient.instance.getPets();
    setState(() => pets = data);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Mes animaux')),
      floatingActionButton: FloatingActionButton(
        onPressed: () async {
          await Navigator.push(context, MaterialPageRoute(builder: (_) => const PetFormScreen()));
          load();
        },
        child: const Icon(Icons.add),
      ),
      body: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: pets.length,
        itemBuilder: (_, i) {
          final p = pets[i] as Map<String, dynamic>;
          final status = p['paymentStatus'] as String? ?? 'pending_payment';
          final ent = p['entitlement'] as Map<String, dynamic>?;
          String badge;
          switch (status) {
            case 'active':
              badge = ent?['billingMode'] == 'subscription' ? 'Renouvellement auto' : 'Actif';
              if (ent?['validUntil'] != null) badge += ' · expire ${ent!['validUntil'].toString().substring(0, 10)}';
              break;
            case 'pending_payment':
              badge = 'En attente de paiement';
              break;
            default:
              badge = status;
          }
          return Card(
            child: ListTile(
              title: Text(p['name'] as String? ?? ''),
              subtitle: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                Text('${p['species']} · ${p['breed']}'),
                Text(badge, style: TextStyle(color: status == 'active' ? Colors.green : Colors.orange, fontSize: 12)),
              ]),
              onTap: () => _openPetMenu(context, p['id'] as String, p['name'] as String, status, ent),
            ),
          );
        },
      ),
    );
  }

  void _openPetMenu(BuildContext context, String petId, String name, String status, Map<String, dynamic>? ent) {
    showModalBottomSheet(
      context: context,
      builder: (_) => SafeArea(
        child: Column(mainAxisSize: MainAxisSize.min, children: [
          ListTile(title: Text(name, style: const TextStyle(fontWeight: FontWeight.bold))),
          if (status == 'pending_payment')
            ListTile(
              leading: const Icon(Icons.payment),
              title: const Text('Reprendre le paiement'),
              onTap: () async {
                Navigator.pop(context);
                final url = await ApiClient.instance.resumeCheckout(petId);
                await launchUrl(Uri.parse(url), mode: LaunchMode.externalApplication);
                load();
              },
            ),
          if (status == 'active' && ent?['billingMode'] == 'subscription')
            ListTile(
              leading: const Icon(Icons.settings),
              title: const Text('Gérer mon abonnement'),
              onTap: () async {
                Navigator.pop(context);
                final url = await ApiClient.instance.billingPortal(petId);
                await launchUrl(Uri.parse(url), mode: LaunchMode.externalApplication);
              },
            ),
          ListTile(
            leading: const Icon(Icons.favorite),
            title: const Text('Relevé cardiaque'),
            onTap: () {
              Navigator.pop(context);
              Navigator.push(context, MaterialPageRoute(builder: (_) => HeartRateFlowScreen(petId: petId)));
            },
          ),
          ListTile(
            leading: const Icon(Icons.history),
            title: const Text('Historique'),
            onTap: () {
              Navigator.pop(context);
              Navigator.push(context, MaterialPageRoute(builder: (_) => PetTimelineScreen(petId: petId)));
            },
          ),
          ListTile(
            leading: const Icon(Icons.chat),
            title: const Text('Messagerie véto'),
            onTap: () {
              Navigator.pop(context);
              Navigator.push(context, MaterialPageRoute(builder: (_) => const MessagingScreen()));
            },
          ),
        ]),
      ),
    );
  }
}
