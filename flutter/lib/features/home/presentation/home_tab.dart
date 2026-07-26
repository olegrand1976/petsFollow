import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/api/open_url.dart';
import 'package:petsfollow_mobile/core/discovery/discovery_controller.dart';
import 'package:petsfollow_mobile/core/models/discovery_card.dart';
import 'package:petsfollow_mobile/core/models/discovery_progress.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/core/ui/load_error_view.dart';
import 'package:petsfollow_mobile/features/discovery/presentation/discovery_card_widget.dart';
import 'package:petsfollow_mobile/features/heartrate/presentation/heart_rate_flow_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_create_flow.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_detail_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_quick_actions.dart';
import 'package:petsfollow_mobile/features/settings/presentation/feature_modules_controller.dart';
import 'package:petsfollow_mobile/features/shell/presentation/main_shell_screen.dart';
import 'package:petsfollow_mobile/features/vets/presentation/widgets/add_vet_panel.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class HomeTab extends StatefulWidget {
  const HomeTab({super.key, this.onNavigateToPets});

  final VoidCallback? onNavigateToPets;

  @override
  State<HomeTab> createState() => _HomeTabState();
}

class _HomeTabState extends State<HomeTab> with WidgetsBindingObserver {
  List<Pet> pets = [];
  String? userName;
  bool loading = true;
  String? loadError;
  bool _hasLoadedOnce = false;
  bool? hasVets;
  DiscoveryProgress? discoveryProgress;
  int householdEpoch = 0;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    load();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      load();
    }
  }

  Future<void> load() async {
    final l10n = AppLocalizations.of(context)!;
    final keepStale = _hasLoadedOnce;
    if (!keepStale && mounted) {
      setState(() {
        loading = true;
        loadError = null;
      });
    }
    try {
      final me = await ApiClient.instance.getMe();
      userName = me['fullName'] as String?;
    } catch (_) {}
    await FeatureModulesController.instance.load();
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
          loadError = null;
          _hasLoadedOnce = true;
          householdEpoch++;
        });
      }
    } catch (e) {
      if (!mounted) return;
      final msg = mapApiError(e, l10n);
      if (keepStale) {
        setState(() => loading = false);
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
      } else {
        setState(() {
          loading = false;
          loadError = msg;
          householdEpoch++;
        });
      }
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

  Future<void> _openPetForm() => openPetFormAndFollowUp(
        context,
        onReload: load,
        hasLinkedVets: hasVets,
      );

  Future<void> _openKennelEncode() => openKennelEncodeAndFollowUp(
        context,
        onReload: load,
        hasLinkedVets: hasVets,
      );

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
    final p = PetsPalette.of(context);
    final greeting = userName != null && userName!.isNotEmpty
        ? l10n.greeting(userName!.split(' ').first)
        : l10n.myPets;
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
          : loadError != null
              ? LoadErrorView(message: loadError!, onRetry: load)
              : RefreshIndicator(
              onRefresh: load,
              child: ListView(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
                children: [
                  Text(greeting, style: Theme.of(context).textTheme.headlineSmall),
                  const SizedBox(height: 4),
                  Text(l10n.appTagline, style: TextStyle(color: p.textMuted)),
                  const SizedBox(height: 20),
                  if (hasVets == false) ...[
                    _AddFirstVetCard(
                      onLinked: load,
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
                      onAdd: _openPetForm,
                    )
                  else ...[
                    if (FeatureModulesController.instance.kennel) ...[
                      _KennelEncodeButton(l10n: l10n, onPressed: _openKennelEncode),
                      const SizedBox(height: 12),
                    ],
                    if (FeatureModulesController.instance.family)
                      _FamilyHouseholdCard(
                        key: ValueKey(householdEpoch),
                        l10n: l10n,
                      ),
                    if (FeatureModulesController.instance.family)
                      const SizedBox(height: 24),
                    Text(l10n.myPets, style: Theme.of(context).textTheme.titleMedium),
                    const SizedBox(height: 12),
                    ...pets.map(
                      (pet) => _PetHeroCard(
                        pet: pet,
                        speciesLabel: _speciesLabel(l10n, pet.species),
                        l10n: l10n,
                        onTap: () => _openPetDetail(pet),
                        onMeasure: pet.isOwner && pet.isActive && !pet.needsVetLink
                            ? () => _startMeasurement(pet)
                            : null,
                        onWeightRecorded:
                            pet.isOwner && pet.isActive && !pet.needsVetLink ? load : null,
                        onResumePayment:
                            pet.needsResumePayment ? () => _resumePayment(pet) : null,
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
              onPressed: _openPetForm,
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
  const _AddFirstVetCard({required this.onLinked});

  final VoidCallback onLinked;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final p = PetsPalette.of(context);
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: p.surfaceElevated,
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
          Text(l10n.homeAddFirstVetBody, style: TextStyle(color: p.textMuted, height: 1.35)),
          const SizedBox(height: 14),
          AddVetPanel(compact: true, onLinked: onLinked),
        ],
      ),
    );
  }
}

class _FamilyHouseholdCard extends StatefulWidget {
  const _FamilyHouseholdCard({super.key, required this.l10n});

  final AppLocalizations l10n;

  @override
  State<_FamilyHouseholdCard> createState() => _FamilyHouseholdCardState();
}

class _FamilyHouseholdCardState extends State<_FamilyHouseholdCard> {
  Map<String, dynamic>? _data;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final data = await ApiClient.instance.getHousehold();
      if (!mounted) return;
      setState(() => _data = data);
    } catch (_) {
      if (mounted) setState(() => _data = null);
    }
  }

  @override
  Widget build(BuildContext context) {
    final data = _data;
    if (data == null) return const SizedBox.shrink();
    final count = (data['petCount'] as num?)?.toInt() ?? 0;
    final pack = '${data['pack'] ?? 'family'}';
    final upcoming = (data['upcomingReminders'] as List?) ?? const [];
    final title = pack == 'kennel'
        ? widget.l10n.kennelHouseholdTitle(count)
        : widget.l10n.familyHouseholdTitle(count);
    final dateFmt = DateFormat.yMMMd(Localizations.localeOf(context).toString());
    final now = DateTime.now();
    final p = PetsPalette.of(context);
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: p.surfaceElevated,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppColors.primary.withValues(alpha: 0.25)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
          ),
          if (upcoming.isNotEmpty) ...[
            const SizedBox(height: 8),
            Text(widget.l10n.familyHouseholdNext, style: TextStyle(color: p.textMuted, fontSize: 12)),
            const SizedBox(height: 6),
            ...upcoming.take(3).map((raw) {
              final item = Map<String, dynamic>.from(raw as Map);
              final petName = '${item['petName'] ?? ''}';
              final reminderTitle = '${item['title'] ?? item['type'] ?? ''}';
              final dueAt = DateTime.tryParse('${item['dueAt'] ?? ''}')?.toLocal();
              final isOverdue = item['isOverdue'] == true ||
                  (dueAt != null && dueAt.isBefore(now));
              final dateLabel = dueAt == null
                  ? null
                  : (isOverdue
                      ? '${widget.l10n.careOverdue} · ${dateFmt.format(dueAt)}'
                      : dateFmt.format(dueAt));
              return Padding(
                padding: const EdgeInsets.only(bottom: 4),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: Text(
                        '• $petName — $reminderTitle',
                        style: const TextStyle(fontSize: 12),
                        overflow: TextOverflow.ellipsis,
                        maxLines: 2,
                      ),
                    ),
                    if (dateLabel != null) ...[
                      const SizedBox(width: 8),
                      Text(
                        dateLabel,
                        style: TextStyle(
                          fontSize: 11,
                          color: isOverdue ? AppColors.alert : p.textMuted,
                        ),
                      ),
                    ],
                  ],
                ),
              );
            }),
          ],
        ],
      ),
    );
  }
}

class _KennelEncodeButton extends StatelessWidget {
  const _KennelEncodeButton({required this.l10n, required this.onPressed});

  final AppLocalizations l10n;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    return OutlinedButton.icon(
      onPressed: onPressed,
      icon: const Icon(Icons.pets_outlined, size: 18),
      label: Text(l10n.kennelQuickEncodeTitle),
    );
  }
}

class _EmptyPetsState extends StatelessWidget {
  const _EmptyPetsState({required this.onAdd});

  final VoidCallback onAdd;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final p = PetsPalette.of(context);
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
              style: TextStyle(color: p.textMuted, height: 1.4),
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

class _PetHeroCard extends StatelessWidget {
  const _PetHeroCard({
    required this.pet,
    required this.speciesLabel,
    required this.l10n,
    required this.onTap,
    this.onMeasure,
    this.onWeightRecorded,
    this.onResumePayment,
  });

  final Pet pet;
  final String speciesLabel;
  final AppLocalizations l10n;
  final VoidCallback onTap;
  final VoidCallback? onMeasure;
  final VoidCallback? onWeightRecorded;
  final VoidCallback? onResumePayment;

  @override
  Widget build(BuildContext context) {
    final p = PetsPalette.of(context);
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
                    backgroundColor: p.surfaceElevated,
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
                        Text('$speciesLabel · ${pet.breed}', style: TextStyle(color: p.textMuted)),
                      ],
                    ),
                  ),
                  _PaymentBadge(pet: pet, l10n: l10n),
                ],
              ),
              if (onMeasure != null) ...[
                const SizedBox(height: 12),
                PetQuickActions(
                  petId: pet.id,
                  onHeartRate: onMeasure!,
                  onWeightRecorded: onWeightRecorded,
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
    final p = PetsPalette.of(context);
    if (pet.isSharedAccess) {
      return _BadgeChip(label: pet.sharedAccessLabel(l10n), color: p.textMuted);
    }
    if (pet.needsResumePayment) {
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
