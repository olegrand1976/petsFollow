import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:url_launcher/url_launcher.dart';

class PetFormScreen extends StatefulWidget {
  const PetFormScreen({super.key});

  @override
  State<PetFormScreen> createState() => _PetFormScreenState();
}

class _PetFormScreenState extends State<PetFormScreen> {
  final name = TextEditingController();
  final species = TextEditingController(text: 'dog');
  final breed = TextEditingController();
  String selectedPlan = 'triennial';
  bool autoRenew = true;
  bool loading = false;

  static const plans = [
    {'code': 'annual', 'label': '25 € / an'},
    {'code': 'triennial', 'label': '60 € / 3 ans', 'recommended': true},
    {'code': 'quinquennial', 'label': '75 € / 5 ans'},
  ];

  String get summary {
    final plan = plans.firstWhere((p) => p['code'] == selectedPlan);
    final label = plan['label'] as String;
    if (autoRenew) {
      if (selectedPlan == 'annual') return '$label, renouvelé automatiquement';
      if (selectedPlan == 'triennial') return '60 € tous les 3 ans, renouvelé automatiquement';
      return '75 € tous les 5 ans, renouvelé automatiquement';
    }
    return '$label, paiement unique';
  }

  Future<void> save() async {
    setState(() => loading = true);
    try {
      final res = await ApiClient.instance.createPet({
        'name': name.text,
        'species': species.text,
        'breed': breed.text,
        'plan': selectedPlan,
        'billingMode': autoRenew ? 'subscription' : 'one_time',
      });
      final checkoutUrl = res['checkoutUrl'] as String?;
      final pet = res['pet'] as Map<String, dynamic>? ?? res;
      final petId = pet['id'] as String?;
      if (checkoutUrl != null && await canLaunchUrl(Uri.parse(checkoutUrl))) {
        await launchUrl(Uri.parse(checkoutUrl), mode: LaunchMode.externalApplication);
      }
      if (!mounted || petId == null) return;
      await _waitForPayment(petId);
      if (mounted) Navigator.pop(context);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Erreur: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> _waitForPayment(String petId) async {
    for (var i = 0; i < 20; i++) {
      await Future.delayed(const Duration(seconds: 2));
      final pet = await ApiClient.instance.getPet(petId);
      if (pet['paymentStatus'] == 'active') {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Paiement confirmé — animal actif')),
          );
        }
        return;
      }
    }
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Paiement en attente — vous pourrez reprendre plus tard')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Nouvel animal')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
          TextField(controller: name, decoration: const InputDecoration(labelText: 'Nom')),
          TextField(controller: species, decoration: const InputDecoration(labelText: 'Espèce')),
          TextField(controller: breed, decoration: const InputDecoration(labelText: 'Race')),
          const SizedBox(height: 24),
          Text('Choisissez votre formule', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          ...plans.map((plan) {
            final code = plan['code'] as String;
            final recommended = plan['recommended'] == true;
            return Card(
              color: selectedPlan == code ? Theme.of(context).colorScheme.primaryContainer : null,
              child: RadioListTile<String>(
                value: code,
                groupValue: selectedPlan,
                onChanged: (v) => setState(() => selectedPlan = v!),
                title: Row(children: [
                  Text(plan['label'] as String),
                  if (recommended) ...[
                    const SizedBox(width: 8),
                    Chip(
                      label: const Text('Recommandé', style: TextStyle(fontSize: 11)),
                      visualDensity: VisualDensity.compact,
                      backgroundColor: Theme.of(context).colorScheme.secondaryContainer,
                    ),
                  ],
                ]),
              ),
            );
          }),
          SwitchListTile(
            title: const Text('Renouveler automatiquement'),
            subtitle: const Text('Prélèvement à chaque échéance'),
            value: autoRenew,
            onChanged: (v) => setState(() => autoRenew = v),
          ),
          Text(summary, style: Theme.of(context).textTheme.bodyMedium),
          const SizedBox(height: 24),
          FilledButton(
            onPressed: loading ? null : save,
            child: loading
                ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                : const Text('Continuer vers le paiement'),
          ),
        ]),
      ),
    );
  }
}
