import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/billing_addon.dart';
import 'package:petsfollow_mobile/core/api/open_url.dart';
import 'package:petsfollow_mobile/core/discovery/discovery_controller.dart';
import 'package:petsfollow_mobile/core/models/discovery_card.dart';
import 'package:petsfollow_mobile/core/models/discovery_progress.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/features/discovery/presentation/discovery_card_widget.dart';
import 'package:petsfollow_mobile/features/heartrate/presentation/heart_rate_flow_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_detail_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_form_screen.dart';
import 'package:petsfollow_mobile/features/shell/presentation/main_shell_screen.dart';
import 'package:petsfollow_mobile/features/vets/presentation/my_vets_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class HomeTab extends StatefulWidget {
  const HomeTab({super.key, this.onNavigateToPets});

  final VoidCallback? onNavigateToPets;

  @override
  State<HomeTab> createState() => _HomeTabState();
}

class _HomeTabState extends State<HomeTab> {
  List<Pet> pets = [];
  String? userName;
  bool loading = true;
  bool? hasVets;
  DiscoveryProgress? discoveryProgress;

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
    try {
      final progress = await DiscoveryController.instance.load();
      discoveryProgress = progress;
    } catch (_) {}
    try {
      final vets = await ApiClient.instance.getMyVets();
      hasVets = vets.isNotEmpty;
    } catch (_) {
      // Keep previous / null — do not hide the first-vet CTA on network errors.
    }
    try {
      final data = await ApiClient.instance.getPets();
      if (mounted) {
        setState(() {
          pets = data.map((p) => Pet.fromJson(Map<String, dynamic>.from(p as Map))).toList();
          loading = false;
        });
      }
    } catch (_) {
      if (mounted) setState(() => loading = false);
    }
  }

  List<DiscoveryCard> _discoveryCards(AppLocalizations l10n, DiscoveryProgress progress) {
    final base = [
      DiscoveryCard(dayIndex: 0, title: l10n.discoveryDay0Title, body: l10n.discoveryDay0Body),
      DiscoveryCard(dayIndex: 2, title: l10n.discoveryDay2Title, body: l10n.discoveryDay2Body),
      DiscoveryCard(dayIndex: 4, title: l10n.discoveryDay4Title, body: l10n.discoveryDay4Body),
      DiscoveryCard(dayIndex: 6, title: l10n.discoveryDay6Title, body: l10n.discoveryDay6Body),
    ];
    return DiscoveryController.instance.cardsWithProgress(base, progress);
  }

  Future<void> _completeMission(DiscoveryCard card) async {
    final progress = await DiscoveryController.instance.completeCard(card.cardKey);
    if (mounted) setState(() => discoveryProgress = progress);
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

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final greeting = userName != null && userName!.isNotEmpty
        ? l10n.greeting(userName!.split(' ').first)
        : l10n.myPets;
    final activePet = pets.where((p) => p.isActive).firstOrNull;
    final progress = discoveryProgress ?? DiscoveryProgress(userId: '', startedAt: DateTime.now());
    final cards = _discoveryCards(l10n, progress);
    final mission = DiscoveryController.instance.missionCardForToday(
      cards.where((c) => !c.completed && !c.locked).toList(),
      progress,
    );

    return PetsTabScaffold(
      title: const PetsAppBarLogo(),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: load,
              child: ListView(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
                children: [
                  Text(greeting, style: Theme.of(context).textTheme.headlineSmall),
                  const SizedBox(height: 4),
                  Text(l10n.appTagline, style: TextStyle(color: AppColors.textMuted)),
                  const SizedBox(height: 20),
                  if (hasVets == false) ...[
                    _AddFirstVetCard(
                      onAdd: () async {
                        await Navigator.push(
                          context,
                          MaterialPageRoute(builder: (_) => const MyVetsScreen()),
                        );
                        load();
                      },
                    ),
                    const SizedBox(height: 16),
                  ],
                  if (mission != null) ...[
                    DiscoveryCardWidget(
                      card: mission,
                      mission: true,
                      onComplete: () => _completeMission(mission),
                    ),
                  ],
                  if (pets.isEmpty)
                    _EmptyPetsState(
                      onAdd: () async {
                        await Navigator.push(
                          context,
                          MaterialPageRoute(builder: (_) => const PetFormScreen()),
                        );
                        load();
                      },
                    )
                  else ...[
                    if (activePet != null) ...[
                      Text(l10n.startMeasurement, style: Theme.of(context).textTheme.titleMedium),
                      const SizedBox(height: 12),
                      _ProminentFcCta(
                        pet: activePet,
                        onMeasure: () => _startMeasurement(activePet),
                      ),
                      const SizedBox(height: 12),
                      _UpsellBanner(
                        l10n: l10n,
                        hasHorse: pets.any((p) => p.species == 'horse'),
                        petCount: pets.length,
                      ),
                      const SizedBox(height: 24),
                    ],
                    Text(l10n.myPets, style: Theme.of(context).textTheme.titleMedium),
                    const SizedBox(height: 12),
                    ...pets.map(
                      (pet) => _PetHeroCard(
                        pet: pet,
                        speciesLabel: _speciesLabel(l10n, pet.species),
                        l10n: l10n,
                        onTap: () => _openPetDetail(pet),
                        onMeasure: pet.isActive ? () => _startMeasurement(pet) : null,
                        onResumePayment: pet.paymentStatus == 'pending_payment'
                            ? () => _resumePayment(pet)
                            : null,
                      ),
                    ),
                  ],
                  const SizedBox(height: 24),
                  Text(l10n.discoveryTitle, style: Theme.of(context).textTheme.titleMedium),
                  const SizedBox(height: 4),
                  Text(l10n.discoveryMission, style: TextStyle(color: AppColors.gold)),
                  const SizedBox(height: 12),
                  ...cards.map(
                    (card) => DiscoveryCardWidget(
                      card: card,
                      onComplete: card.locked || card.completed ? null : () => _completeMission(card),
                    ),
                  ),
                ],
              ),
            ),
      floatingActionButton: pets.isNotEmpty
          ? null
          : FloatingActionButton.extended(
              onPressed: () async {
                await Navigator.push(
                  context,
                  MaterialPageRoute(builder: (_) => const PetFormScreen()),
                );
                load();
              },
              icon: const Icon(Icons.add),
              label: Text(l10n.newPet),
            ),
    );
  }

  Future<void> _startMeasurement(Pet pet) async {
    await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => HeartRateFlowScreen(
          petId: pet.id,
          durationsSec: pet.heartrateDurationsSec,
        ),
      ),
    );
    load();
  }

  Future<void> _resumePayment(Pet pet) async {
    final url = await ApiClient.instance.resumeCheckout(pet.id);
    await openExternalUrl(url);
    load();
  }

  void _openPetDetail(Pet pet) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => PetDetailScreen(pet: pet, onUpdated: load)),
    );
  }
}

class _AddFirstVetCard extends StatelessWidget {
  const _AddFirstVetCard({required this.onAdd});

  final VoidCallback onAdd;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.surfaceElevated,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: AppColors.primary.withValues(alpha: 0.25)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.local_hospital_outlined, color: AppColors.primary),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  l10n.homeAddFirstVetTitle,
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w600),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(l10n.homeAddFirstVetBody, style: TextStyle(color: AppColors.textMuted, height: 1.35)),
          const SizedBox(height: 14),
          FilledButton.icon(
            onPressed: onAdd,
            icon: const Icon(Icons.person_add_alt_1),
            label: Text(l10n.homeAddFirstVetCta),
          ),
        ],
      ),
    );
  }
}

class _UpsellBanner extends StatefulWidget {
  const _UpsellBanner({
    required this.l10n,
    required this.hasHorse,
    required this.petCount,
  });

  final AppLocalizations l10n;
  final bool hasHorse;
  final int petCount;

  @override
  State<_UpsellBanner> createState() => _UpsellBannerState();
}

class _UpsellBannerState extends State<_UpsellBanner> {
  BillingAddon? _carePlus;
  BillingAddon? _horse;
  BillingAddon? _family;
  bool _hasCarePlus = false;
  bool _hasHorsePack = false;
  bool _hasFamily = false;

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
        _horse = BillingAddon.byCode(catalog, 'horse');
        _family = BillingAddon.byCode(catalog, 'family');
        _hasCarePlus = ents.hasCarePlus;
        _hasHorsePack = ents.hasHorse;
        _hasFamily = ents.hasFamily;
      });
    } catch (_) {}
  }

  Future<void> _buy(BuildContext context, String code) async {
    try {
      final url = await ApiClient.instance.startAddonCheckout(addonCode: code);
      await openExternalUrl(url);
      await _load();
    } catch (_) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(widget.l10n.paymentResume)),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final showCare = !_hasCarePlus && _carePlus != null;
    final showHorse = widget.hasHorse && !_hasHorsePack && _horse != null;
    final showFamily = widget.petCount >= 2 && !_hasFamily && _family != null;
    if (!showCare && !showHorse && !showFamily) return const SizedBox.shrink();

    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: AppColors.surfaceElevated,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppColors.gold.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (showCare) ...[
            Text(widget.l10n.carePlusUpsell, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13)),
            const SizedBox(height: 8),
            OutlinedButton(
              onPressed: () => _buy(context, 'care_plus'),
              child: Text(_carePlus!.label.isNotEmpty ? _carePlus!.label : widget.l10n.activateAddon),
            ),
          ],
          if (showCare && (showHorse || showFamily)) const SizedBox(height: 12),
          if (showFamily) ...[
            Text(widget.l10n.familyPackHint, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13)),
            const SizedBox(height: 8),
            OutlinedButton(
              onPressed: () => _buy(context, 'family'),
              child: Text(_family!.label.isNotEmpty ? _family!.label : widget.l10n.activateAddon),
            ),
          ],
          if (showFamily && showHorse) const SizedBox(height: 12),
          if (showHorse) ...[
            Text(widget.l10n.horsePackUpsell, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13)),
            const SizedBox(height: 8),
            OutlinedButton(
              onPressed: () => _buy(context, 'horse'),
              child: Text(_horse!.label.isNotEmpty ? _horse!.label : widget.l10n.activateAddon),
            ),
          ],
        ],
      ),
    );
  }
}

class _EmptyPetsState extends StatelessWidget {
  const _EmptyPetsState({required this.onAdd});

  final VoidCallback onAdd;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          children: [
            Icon(Icons.pets, size: 48, color: AppColors.gold.withValues(alpha: 0.8)),
            const SizedBox(height: 16),
            Text(l10n.emptyPetsTitle, style: Theme.of(context).textTheme.titleLarge),
            const SizedBox(height: 8),
            Text(
              l10n.emptyPetsBody,
              textAlign: TextAlign.center,
              style: TextStyle(color: AppColors.textMuted, height: 1.4),
            ),
            const SizedBox(height: 20),
            FilledButton.icon(
              onPressed: onAdd,
              icon: const Icon(Icons.add),
              label: Text(l10n.newPet),
            ),
          ],
        ),
      ),
    );
  }
}

class _ProminentFcCta extends StatelessWidget {
  const _ProminentFcCta({required this.pet, required this.onMeasure});

  final Pet pet;
  final VoidCallback onMeasure;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(28),
        gradient: LinearGradient(
          colors: [AppColors.primary, AppColors.primary.withValues(alpha: 0.7)],
        ),
        boxShadow: [
          BoxShadow(
            color: AppColors.primary.withValues(alpha: 0.3),
            blurRadius: 16,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onMeasure,
          borderRadius: BorderRadius.circular(28),
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: Row(
              children: [
                const Icon(Icons.favorite, color: AppColors.bg, size: 36),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        l10n.heartRate,
                        style: Theme.of(context).textTheme.titleMedium?.copyWith(
                              color: AppColors.bg,
                              fontWeight: FontWeight.bold,
                            ),
                      ),
                      Text(
                        pet.name,
                        style: TextStyle(color: AppColors.bg.withValues(alpha: 0.85)),
                      ),
                    ],
                  ),
                ),
                Icon(Icons.arrow_forward, color: AppColors.bg.withValues(alpha: 0.9)),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _PetHeroCard extends StatelessWidget {
  const _PetHeroCard({
    required this.pet,
    required this.speciesLabel,
    required this.l10n,
    required this.onTap,
    this.onMeasure,
    this.onResumePayment,
  });

  final Pet pet;
  final String speciesLabel;
  final AppLocalizations l10n;
  final VoidCallback onTap;
  final VoidCallback? onMeasure;
  final VoidCallback? onResumePayment;

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(28),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  CircleAvatar(
                    radius: 28,
                    backgroundColor: AppColors.surfaceElevated,
                    backgroundImage: pet.photoUrl?.isNotEmpty == true
                        ? NetworkImage(pet.photoUrl!)
                        : null,
                    child: pet.photoUrl?.isNotEmpty == true
                        ? null
                        : Text(
                            pet.name.isNotEmpty ? pet.name.substring(0, 1).toUpperCase() : '?',
                            style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
                          ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(pet.name, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                        Text('$speciesLabel · ${pet.breed}', style: TextStyle(color: AppColors.textMuted)),
                      ],
                    ),
                  ),
                  _PaymentBadge(pet: pet, l10n: l10n),
                ],
              ),
              if (onMeasure != null) ...[
                const SizedBox(height: 12),
                SizedBox(
                  width: double.infinity,
                  child: FilledButton(
                    onPressed: onMeasure,
                    child: Text(l10n.startMeasurement),
                  ),
                ),
              ],
              if (onResumePayment != null) ...[
                const SizedBox(height: 12),
                SizedBox(
                  width: double.infinity,
                  child: OutlinedButton(
                    onPressed: onResumePayment,
                    child: Text(l10n.paymentResume),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _PaymentBadge extends StatelessWidget {
  const _PaymentBadge({required this.pet, required this.l10n});

  final Pet pet;
  final AppLocalizations l10n;

  @override
  Widget build(BuildContext context) {
    final ent = pet.entitlement;
    if (pet.paymentStatus == 'pending_payment') {
      return _BadgeChip(label: l10n.badgePendingPayment, color: AppColors.alert);
    }
    if (pet.isActive) {
      if (ent?.isSubscription == true) {
        return _BadgeChip(label: l10n.badgeAutoRenew, color: AppColors.gold);
      }
      return _BadgeChip(label: l10n.badgeActive, color: AppColors.primary);
    }
    return const SizedBox.shrink();
  }
}

class _BadgeChip extends StatelessWidget {
  const _BadgeChip({required this.label, required this.color});

  final String label;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.5)),
      ),
      child: Text(label, style: TextStyle(color: color, fontSize: 11, fontWeight: FontWeight.w600)),
    );
  }
}
