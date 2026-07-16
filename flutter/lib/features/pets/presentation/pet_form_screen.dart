import 'dart:io';

import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:url_launcher/url_launcher.dart';

class PetFormScreen extends StatefulWidget {
  const PetFormScreen({super.key});

  @override
  State<PetFormScreen> createState() => _PetFormScreenState();
}

class _PetFormScreenState extends State<PetFormScreen> {
  final name = TextEditingController();
  String selectedSpecies = 'dog';
  final breed = TextEditingController();
  String selectedPlan = 'triennial';
  bool autoRenew = true;
  bool loading = false;
  List<Map<String, dynamic>> plans = [];
  XFile? photoFile;

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

  Future<void> _pickPhoto() async {
    final l10n = AppLocalizations.of(context)!;
    final source = await showModalBottomSheet<ImageSource>(
      context: context,
      builder: (ctx) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.photo_camera_outlined),
              title: Text(l10n.takePhoto),
              onTap: () => Navigator.pop(ctx, ImageSource.camera),
            ),
            ListTile(
              leading: const Icon(Icons.photo_library_outlined),
              title: Text(l10n.chooseFromGallery),
              onTap: () => Navigator.pop(ctx, ImageSource.gallery),
            ),
          ],
        ),
      ),
    );
    if (source == null) return;
    final picker = ImagePicker();
    final file = await picker.pickImage(
      source: source,
      maxWidth: 1024,
      maxHeight: 1024,
      imageQuality: 85,
      preferredCameraDevice: CameraDevice.rear,
    );
    if (file != null) setState(() => photoFile = file);
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
        'species': selectedSpecies,
        'breed': breed.text,
        'plan': selectedPlan,
        'billingMode': autoRenew ? 'subscription' : 'one_time',
      });
      final checkoutUrl = res['checkoutUrl'] as String?;
      final pet = res['pet'] as Map<String, dynamic>? ?? res;
      final petId = pet['id'] as String?;
      if (petId != null && photoFile != null) {
        try {
          await ApiClient.instance.uploadPetPhoto(petId, photoFile!.path);
        } catch (_) {
          /* payment flow continues even if photo fails */
        }
      }
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
    final initial = (name.text.isNotEmpty ? name.text : '?').substring(0, 1).toUpperCase();

    return Scaffold(
      appBar: AppBar(title: Text(l10n.newPet)),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Center(
              child: Column(
                children: [
                  GestureDetector(
                    onTap: _pickPhoto,
                    child: Container(
                      width: 140,
                      height: 140,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        border: Border.all(color: AppColors.primary, width: 3),
                        boxShadow: [
                          BoxShadow(
                            color: AppColors.primary.withValues(alpha: 0.18),
                            blurRadius: 12,
                            offset: const Offset(0, 4),
                          ),
                        ],
                      ),
                      child: ClipOval(
                        child: photoFile != null
                            ? Image.file(
                                File(photoFile!.path),
                                fit: BoxFit.cover,
                                width: 140,
                                height: 140,
                              )
                            : ColoredBox(
                                color: AppColors.surfaceElevated,
                                child: Center(
                                  child: Text(initial, style: const TextStyle(fontSize: 36, fontWeight: FontWeight.w600)),
                                ),
                              ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    l10n.photoFrameHint,
                    textAlign: TextAlign.center,
                    style: TextStyle(color: AppColors.textMuted, fontSize: 12),
                  ),
                  TextButton.icon(
                    onPressed: _pickPhoto,
                    icon: const Icon(Icons.photo_camera_outlined),
                    label: Text(photoFile == null ? l10n.addPhoto : l10n.changePhoto),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),
            TextField(controller: name, decoration: InputDecoration(labelText: l10n.petName), onChanged: (_) => setState(() {})),
            const SizedBox(height: 12),
            DropdownButtonFormField<String>(
              value: selectedSpecies,
              decoration: InputDecoration(labelText: l10n.species),
              items: [
                DropdownMenuItem(value: 'dog', child: Text(l10n.speciesDog)),
                DropdownMenuItem(value: 'cat', child: Text(l10n.speciesCat)),
                DropdownMenuItem(value: 'horse', child: Text(l10n.speciesHorse)),
                DropdownMenuItem(value: 'other', child: Text(l10n.speciesOther)),
              ],
              onChanged: (v) => setState(() => selectedSpecies = v ?? 'dog'),
            ),
            const SizedBox(height: 12),
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
