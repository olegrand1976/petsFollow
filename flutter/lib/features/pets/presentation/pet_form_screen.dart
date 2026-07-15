import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
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
  List<Map<String, dynamic>> plans = [];

  @override
  void initState() {
    super.initState();
    _loadPlans();
  }

  Future<void> _loadPlans() async {
    try {
      final data = await ApiClient.instance.getBillingPlans();
      setState(() {
        plans = data.map((p) => Map<String, dynamic>.from(p as Map)).toList();
      });
    } catch (_) {
      /* fallback labels from API locale via Accept-Language */
    }
  }

  String _summary(AppLocalizations l10n) {
    final plan = plans.firstWhere(
      (p) => p['code'] == selectedPlan,
      orElse: () => {'label': selectedPlan},
    );
    final label = plan['label'] as String? ?? selectedPlan;
    if (autoRenew) {
      switch (selectedPlan) {
        case 'annual':
          return l10n.planAnnualSub(label);
        case 'triennial':
          return l10n.planTriennialSub;
        case 'quinquennial':
          return l10n.planQuinquennialSub;
      }
    }
    return l10n.planOneTime(label);
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
        final l10n = AppLocalizations.of(context)!;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorGeneric(e.toString()))),
        );
      }
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> _waitForPayment(String petId) async {
    final l10n = AppLocalizations.of(context)!;
    for (var i = 0; i < 20; i++) {
      await Future.delayed(const Duration(seconds: 2));
      final pet = await ApiClient.instance.getPet(petId);
      if (pet['paymentStatus'] == 'active') {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(l10n.paymentConfirmed)),
          );
        }
        return;
      }
    }
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.paymentPending)),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final displayPlans = plans.isNotEmpty
        ? plans
        : [
            {'code': 'annual', 'label': '25 € / an'},
            {'code': 'triennial', 'label': '60 € / 3 ans', 'recommended': true},
            {'code': 'quinquennial', 'label': '75 € / 5 ans'},
          ];

    return Scaffold(
      appBar: AppBar(title: Text(l10n.newPet)),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            TextField(controller: name, decoration: InputDecoration(labelText: l10n.petName)),
            TextField(controller: species, decoration: InputDecoration(labelText: l10n.species)),
            TextField(controller: breed, decoration: InputDecoration(labelText: l10n.breed)),
            const SizedBox(height: 24),
            Text(l10n.choosePlan, style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 8),
            ...displayPlans.map((plan) {
              final code = plan['code'] as String;
              final recommended = plan['recommended'] == true;
              return Card(
                color: selectedPlan == code ? Theme.of(context).colorScheme.primaryContainer : null,
                child: RadioListTile<String>(
                  value: code,
                  groupValue: selectedPlan,
                  onChanged: (v) => setState(() => selectedPlan = v!),
                  title: Row(
                    children: [
                      Text(plan['label'] as String? ?? code),
                      if (recommended) ...[
                        const SizedBox(width: 8),
                        Chip(
                          label: Text(l10n.recommended, style: const TextStyle(fontSize: 11)),
                          visualDensity: VisualDensity.compact,
                          backgroundColor: Theme.of(context).colorScheme.secondaryContainer,
                        ),
                      ],
                    ],
                  ),
                ),
              );
            }),
            SwitchListTile(
              title: Text(l10n.autoRenewTitle),
              subtitle: Text(l10n.autoRenewSubtitle),
              value: autoRenew,
              onChanged: (v) => setState(() => autoRenew = v),
            ),
            Text(_summary(l10n), style: Theme.of(context).textTheme.bodyMedium),
            const SizedBox(height: 24),
            FilledButton(
              onPressed: loading ? null : save,
              child: loading
                  ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                  : Text(l10n.continueToPayment),
            ),
          ],
        ),
      ),
    );
  }
}
