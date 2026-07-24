import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/api/billing_addon.dart';
import 'package:petsfollow_mobile/core/api/open_url.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/core/models/vet_link.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/features/heartrate/presentation/heart_rate_flow_screen.dart';
import 'package:petsfollow_mobile/features/messaging/presentation/messaging_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/horse_health_panel.dart';
import 'package:petsfollow_mobile/features/pets/presentation/book_visit_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_timeline_screen.dart';
import 'package:petsfollow_mobile/features/vets/presentation/my_vets_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class PetDetailScreen extends StatefulWidget {
  const PetDetailScreen({super.key, required this.pet, this.onUpdated});

  final Pet pet;
  final VoidCallback? onUpdated;

  @override
  State<PetDetailScreen> createState() => _PetDetailScreenState();
}

class _PetDetailScreenState extends State<PetDetailScreen> with WidgetsBindingObserver {
  late Pet pet;
  List<VetLink> vets = [];
  bool loadingVets = true;
  String? vetsLoadError;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    pet = widget.pet;
    _loadVets();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      _reloadPet();
    }
  }

  Future<void> _reloadPet() async {
    try {
      final updated = await ApiClient.instance.getPet(pet.id);
      if (!mounted) return;
      setState(() => pet = Pet.fromJson(updated));
      widget.onUpdated?.call();
    } catch (_) {}
  }

  Future<void> _loadVets() async {
    final l10n = AppLocalizations.of(context)!;
    if (mounted) {
      setState(() {
        loadingVets = true;
        vetsLoadError = null;
      });
    }
    try {
      final data = await ApiClient.instance.getMyVets(primaryPracticeId: pet.practiceId);
      if (mounted) {
        setState(() {
          vets = data;
          loadingVets = false;
          vetsLoadError = null;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          loadingVets = false;
          vetsLoadError = mapApiError(e, l10n);
        });
      }
    }
  }

  String _speciesLabel(AppLocalizations l10n, String species) {
    switch (species) {
      case 'dog':
        return l10n.speciesDog;
      case 'cat':
        return l10n.speciesCat;
      case 'horse':
        return l10n.speciesHorse;
      default:
        return l10n.speciesOther;
    }
  }

  Future<void> _changePhoto() async {
    final l10n = AppLocalizations.of(context)!;
    final picker = ImagePicker();
    final file = await picker.pickImage(source: ImageSource.gallery, maxWidth: 1024, imageQuality: 85);
    if (file == null) return;
    try {
      await ApiClient.instance.uploadPetPhoto(pet.id, file.path);
      final updated = await ApiClient.instance.getPet(pet.id);
      if (mounted) {
        setState(() => pet = Pet.fromJson(updated));
        widget.onUpdated?.call();
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(l10n.photoUpdated)));
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorGeneric('photo'))),
        );
      }
    }
  }

  Future<void> _setPrimaryVet(VetLink vet) async {
    final l10n = AppLocalizations.of(context)!;
    try {
      await ApiClient.instance.setPetPrimaryPractice(pet.id, vet.practiceId);
      final updated = await ApiClient.instance.getPet(pet.id);
      if (mounted) {
        setState(() => pet = Pet.fromJson(updated));
        await _loadVets();
        widget.onUpdated?.call();
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(l10n.primaryVetSet)));
        }
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorGeneric('primary vet'))),
        );
      }
    }
  }

  Future<void> _pickPrimaryVet() async {
    final l10n = AppLocalizations.of(context)!;
    if (vets.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(l10n.noVets)));
      return;
    }
    await showModalBottomSheet<void>(
      context: context,
      builder: (ctx) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Text(l10n.setPrimaryVet, style: Theme.of(ctx).textTheme.titleMedium),
            ),
            ...vets.map(
              (v) => ListTile(
                leading: Icon(
                  v.isPrimary ? Icons.star : Icons.star_outline,
                  color: AppColors.gold,
                ),
                title: Text(v.practiceName),
                subtitle: Text(v.vetFullName),
                onTap: () {
                  Navigator.pop(ctx);
                  _setPrimaryVet(v);
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _requestVisit() async {
    final l10n = AppLocalizations.of(context)!;
    var practiceId = pet.practiceId;
    if (practiceId == null || practiceId.isEmpty) {
      final primary = vets.where((v) => v.isPrimary).firstOrNull ?? vets.firstOrNull;
      practiceId = primary?.practiceId;
    }
    if (practiceId == null || practiceId.isEmpty) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(l10n.noVets)));
      await Navigator.push(
        context,
        MaterialPageRoute(builder: (_) => const MyVetsScreen()),
      );
      if (mounted) await _loadVets();
      return;
    }
    final booked = await Navigator.push<bool>(
      context,
      MaterialPageRoute(
        builder: (_) => BookVisitScreen(
          petId: pet.id,
          petName: pet.name,
          practiceId: practiceId!,
        ),
      ),
    );
    if (booked == true && mounted) {
      final visits = await ApiClient.instance.getVisits(pet.id);
      if (!mounted) return;
      await NotificationService.instance.scheduleVisits(
        visits,
        visitLabel: l10n.upcomingVisit,
        petName: pet.name,
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final species = _speciesLabel(l10n, pet.species);
    final initial = pet.name.isNotEmpty ? pet.name.substring(0, 1).toUpperCase() : '?';
    final primaryVet = vets.where((v) => v.isPrimary).firstOrNull;

    return Scaffold(
      appBar: AppBar(title: Text(pet.name)),
      body: ListView(
        padding: scrollPaddingWithSystemBottom(context, all: 20),
        children: [
          Center(
            child: Column(
              children: [
                CircleAvatar(
                  radius: 56,
                  backgroundColor: AppColors.surfaceElevated,
                  backgroundImage: pet.photoUrl?.isNotEmpty == true ? NetworkImage(pet.photoUrl!) : null,
                  child: pet.photoUrl?.isNotEmpty == true
                      ? null
                      : Text(initial, style: const TextStyle(fontSize: 36, fontWeight: FontWeight.bold)),
                ),
                const SizedBox(height: 8),
                TextButton.icon(
                  onPressed: _changePhoto,
                  icon: const Icon(Icons.photo_camera_outlined),
                  label: Text(l10n.changePhoto),
                ),
                const SizedBox(height: 4),
                Text(species, style: Theme.of(context).textTheme.titleMedium?.copyWith(color: AppColors.gold)),
                Text(pet.breed, style: TextStyle(color: AppColors.textMuted)),
              ],
            ),
          ),
          if (pet.species == 'horse') ...[
            const SizedBox(height: 24),
            HorseHealthPanel(petId: pet.id, petName: pet.name),
          ],
          if (pet.isActive) ...[
            const SizedBox(height: 16),
            _UpsellHint(l10n: l10n, petId: pet.id, onPurchased: _reloadPet),
          ],
          const SizedBox(height: 24),
          if (pet.isActive)
            FilledButton.icon(
              onPressed: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => HeartRateFlowScreen(
                      petId: pet.id,
                      durationsSec: pet.heartrateDurationsSec,
                    ),
                  ),
                );
              },
              icon: const Icon(Icons.favorite),
              label: Text(l10n.startMeasurement),
            ),
          if (pet.needsResumePayment) ...[
            FilledButton.icon(
              onPressed: () async {
                final url = await ApiClient.instance.resumeCheckout(pet.id);
                await openExternalUrl(url);
                await _reloadPet();
              },
              icon: const Icon(Icons.payment),
              label: Text(l10n.paymentResume),
            ),
            const SizedBox(height: 8),
          ],
          const SizedBox(height: 8),
          _ActionTile(
            icon: Icons.history,
            label: l10n.visitHistory,
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => PetTimelineScreen(petId: pet.id, petName: pet.name)),
            ),
          ),
          _ActionTile(
            icon: Icons.event_available,
            label: l10n.requestVisit,
            onTap: _requestVisit,
          ),
          _ActionTile(
            icon: Icons.chat,
            label: l10n.vetMessaging,
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const MessagingScreen()),
            ),
          ),
          const Divider(height: 32),
          Row(
            children: [
              Expanded(child: Text(l10n.myVets, style: Theme.of(context).textTheme.titleSmall)),
              TextButton(
                onPressed: () async {
                  await Navigator.push(context, MaterialPageRoute(builder: (_) => const MyVetsScreen()));
                  _loadVets();
                },
                child: Text(l10n.addVetByEmail),
              ),
            ],
          ),
          if (loadingVets)
            const Center(child: Padding(padding: EdgeInsets.all(16), child: CircularProgressIndicator()))
          else if (vetsLoadError != null)
            ListTile(
              leading: const Icon(Icons.cloud_off_outlined, color: AppColors.textMuted),
              title: Text(vetsLoadError!),
              trailing: TextButton(
                onPressed: _loadVets,
                child: Text(l10n.retryAction),
              ),
            )
          else if (vets.isEmpty)
            ListTile(
              leading: const Icon(Icons.local_hospital_outlined),
              title: Text(l10n.noVets),
              subtitle: Text(l10n.addVetByEmail),
            )
          else
            ...vets.map(
              (v) => ListTile(
                leading: Icon(
                  v.isPrimary ? Icons.star : Icons.star_outline,
                  color: AppColors.gold,
                ),
                title: Text(v.practiceName),
                subtitle: Text('${v.vetFullName} · ${v.vetEmail}'),
                trailing: v.isPrimary
                    ? null
                    : IconButton(
                        icon: const Icon(Icons.star_outline),
                        tooltip: l10n.setPrimaryVet,
                        onPressed: () => _setPrimaryVet(v),
                      ),
              ),
            ),
          if (vets.length > 1)
            OutlinedButton.icon(
              onPressed: _pickPrimaryVet,
              icon: const Icon(Icons.swap_horiz),
              label: Text(primaryVet != null ? l10n.primaryVet : l10n.setPrimaryVet),
            ),
          if (pet.isActive && pet.entitlement?.isSubscription == true)
            Padding(
              padding: const EdgeInsets.only(top: 16),
              child: OutlinedButton.icon(
                onPressed: () async {
                  final url = await ApiClient.instance.billingPortal(pet.id);
                  await openExternalUrl(url);
                },
                icon: const Icon(Icons.settings),
                label: Text(l10n.manageSubscription),
              ),
            ),
        ],
      ),
    );
  }
}

class _UpsellHint extends StatefulWidget {
  const _UpsellHint({required this.l10n, required this.petId, this.onPurchased});

  final AppLocalizations l10n;
  final String petId;
  final Future<void> Function()? onPurchased;

  @override
  State<_UpsellHint> createState() => _UpsellHintState();
}

class _UpsellHintState extends State<_UpsellHint> {
  BillingAddon? _carePlus;
  // Fail-closed: hide upsell until entitlements confirm Care+ is missing.
  bool _hasCarePlus = true;
  bool _ready = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final catalog = await BillingAddon.fetchCatalog();
      final ents = await AddonEntitlements.load();
      if (!mounted) return;
      setState(() {
        _carePlus = BillingAddon.byCode(catalog, 'care_plus');
        _hasCarePlus = ents == null ? true : ents.hasCarePlus;
        _ready = true;
      });
    } catch (_) {
      if (mounted) setState(() => _ready = true);
    }
  }

  Future<void> _buy(BuildContext context, String code) async {
    try {
      final url = await ApiClient.instance.startAddonCheckout(addonCode: code);
      await openExternalUrl(url);
      await _load();
      await widget.onPurchased?.call();
    } catch (_) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(widget.l10n.paymentResume)));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    if (!_ready || _hasCarePlus || _carePlus == null) return const SizedBox.shrink();
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: AppColors.surfaceElevated,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.gold.withValues(alpha: 0.25)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(widget.l10n.carePlusUpsell, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600)),
          const SizedBox(height: 8),
          TextButton(
            onPressed: () => _buy(context, _carePlus!.code),
            child: Text(
              _carePlus!.label.isNotEmpty
                  ? _carePlus!.label
                  : widget.l10n.carePlusUpsell.split('—').first.trim(),
            ),
          ),
        ],
      ),
    );
  }
}

class _ActionTile extends StatelessWidget {
  const _ActionTile({required this.icon, required this.label, required this.onTap});

  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        leading: Icon(icon, color: AppColors.primary),
        title: Text(label),
        trailing: const Icon(Icons.chevron_right),
        onTap: onTap,
      ),
    );
  }
}
