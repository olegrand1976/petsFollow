import 'dart:io';

import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:path_provider/path_provider.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/api/open_url.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/core/models/pet_species.dart';
import 'package:petsfollow_mobile/core/models/vet_link.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/features/heartrate/presentation/heart_rate_chart.dart';
import 'package:petsfollow_mobile/features/heartrate/presentation/heart_rate_flow_screen.dart';
import 'package:petsfollow_mobile/features/heartrate/supports_heart_rate.dart';
import 'package:petsfollow_mobile/features/messaging/presentation/messaging_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/horse_health_panel.dart';
import 'package:petsfollow_mobile/features/pets/presentation/book_visit_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_edit_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/blood_pressure_chart.dart';
import 'package:petsfollow_mobile/features/pets/presentation/lab_panels_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_quick_actions.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_timeline_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/weight_chart.dart';
import 'package:petsfollow_mobile/features/settings/presentation/feature_modules_controller.dart';
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
  List<({DateTime date, int bpm, bool isAlert})> hrPoints = [];
  List<({DateTime date, double kg})> weightPoints = [];
  List<({DateTime date, int sys, int dia})> bpPoints = [];
  int labPanelCount = 0;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    pet = widget.pet;
    _loadVets();
    _loadCharts();
    FeatureModulesController.instance.load().then((_) {
      if (mounted) setState(() {});
    });
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
      await _loadCharts();
    } catch (_) {}
  }

  Future<void> _loadCharts() async {
    List<({DateTime date, int bpm, bool isAlert})>? nextHr;
    List<({DateTime date, double kg})>? nextWeight;
    List<({DateTime date, int sys, int dia})>? nextBp;
    int? nextLabCount;

    try {
      final sessions = await ApiClient.instance.getHeartRateSessions(pet.id);
      nextHr = sessions
          .whereType<Map>()
          .map((e) {
            final bpm = e['bpm'];
            final started =
                DateTime.tryParse('${e['startedAt'] ?? e['endedAt'] ?? ''}');
            if (bpm is! num || started == null) return null;
            return (
              date: started,
              bpm: bpm.round(),
              isAlert: e['isAlert'] == true,
            );
          })
          .whereType<({DateTime date, int bpm, bool isAlert})>()
          .toList();
    } catch (_) {}

    try {
      final weights = await ApiClient.instance.getWeightReadings(pet.id);
      nextWeight = weights
          .whereType<Map>()
          .map((e) {
            final kg = e['weightKg'];
            final at = DateTime.tryParse('${e['recordedAt'] ?? ''}');
            if (kg is! num || at == null) return null;
            return (date: at, kg: kg.toDouble());
          })
          .whereType<({DateTime date, double kg})>()
          .toList();
    } catch (_) {}

    try {
      final bps = await ApiClient.instance.getBloodPressureReadings(pet.id);
      nextBp = bps
          .whereType<Map>()
          .map((e) {
            final sys = e['systolicMmHg'];
            final dia = e['diastolicMmHg'];
            final at = DateTime.tryParse('${e['recordedAt'] ?? ''}');
            if (sys is! num || dia is! num || at == null) return null;
            return (date: at, sys: sys.round(), dia: dia.round());
          })
          .whereType<({DateTime date, int sys, int dia})>()
          .toList();
    } catch (_) {}

    try {
      final labs = await ApiClient.instance.getLabPanels(pet.id);
      nextLabCount = labs.length;
    } catch (_) {}

    if (!mounted) return;
    setState(() {
      if (nextHr != null) hrPoints = nextHr;
      if (nextWeight != null) weightPoints = nextWeight;
      if (nextBp != null) bpPoints = nextBp;
      if (nextLabCount != null) labPanelCount = nextLabCount;
    });
  }

  Future<void> _loadVets() async {
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
        final l10n = AppLocalizations.of(context)!;
        setState(() {
          loadingVets = false;
          vetsLoadError = mapApiError(e, l10n);
        });
      }
    }
  }

  String _speciesLabel(AppLocalizations l10n, String species) =>
      speciesLabel(l10n, species);

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

  Future<void> _openHealthBookPdf() async {
    final l10n = AppLocalizations.of(context)!;
    try {
      final bytes = await ApiClient.instance.downloadPetHealthBook(pet.id);
      if (bytes.isEmpty) throw StateError('empty pdf');
      final dir = await getTemporaryDirectory();
      final file = File('${dir.path}/health-book-${pet.id}.pdf');
      await file.writeAsBytes(bytes, flush: true);
      final opened = await openExternalUrl(file.uri.toString());
      if (!opened && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorCouldNotOpenLink)),
        );
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorGeneric('pdf'))),
        );
      }
    }
  }

  Future<void> _setPrimaryVet(VetLink vet) async {
    final l10n = AppLocalizations.of(context)!;
    try {
      await ApiClient.instance.setPetPrimaryPractice(pet.id, vet.practiceId);
      await ApiClient.instance.ensureFreshSession();
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
    if (vets.isEmpty) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(l10n.noVets)));
      await Navigator.push(
        context,
        MaterialPageRoute(builder: (_) => const MyVetsScreen()),
      );
      if (mounted) await _loadVets();
      return;
    }
    final filter = pet.practiceId?.trim();
    final filtered = (filter != null && filter.isNotEmpty)
        ? vets.where((v) => v.practiceId == filter).toList()
        : vets;
    if (filtered.isEmpty) {
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
          practiceIdFilter: filter,
          initialVets: filtered,
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
    final p = PetsPalette.of(context);
    final species = _speciesLabel(l10n, pet.species);
    final initial = pet.name.isNotEmpty ? pet.name.substring(0, 1).toUpperCase() : '?';
    final primaryVet = vets.where((v) => v.isPrimary).firstOrNull;

    return Scaffold(
      appBar: AppBar(
        title: Text(pet.name),
        actions: [
          if (pet.isOwner)
            IconButton(
              icon: const Icon(Icons.edit_outlined),
              tooltip: l10n.editPet,
              onPressed: () async {
                final changed = await Navigator.push<bool>(
                  context,
                  MaterialPageRoute(builder: (_) => PetEditScreen(pet: pet)),
                );
                if (changed == true) {
                  await _reloadPet();
                  widget.onUpdated?.call();
                }
              },
            ),
        ],
      ),
      body: ListView(
        padding: scrollPaddingWithSystemBottom(context, all: 20),
        children: [
          Center(
            child: Column(
              children: [
                CircleAvatar(
                  radius: 56,
                  backgroundColor: p.surfaceElevated,
                  backgroundImage: pet.photoUrl?.isNotEmpty == true ? NetworkImage(pet.photoUrl!) : null,
                  child: pet.photoUrl?.isNotEmpty == true
                      ? null
                      : Text(initial, style: const TextStyle(fontSize: 36, fontWeight: FontWeight.bold)),
                ),
                const SizedBox(height: 8),
                if (pet.isOwner)
                  TextButton.icon(
                    onPressed: _changePhoto,
                    icon: const Icon(Icons.photo_camera_outlined),
                    label: Text(l10n.changePhoto),
                  )
                else
                  Padding(
                    padding: const EdgeInsets.only(bottom: 4),
                    child: Text(
                      pet.sharedAccessLabel(l10n),
                      style: TextStyle(color: p.textMuted, fontSize: 13),
                    ),
                  ),
                const SizedBox(height: 4),
                Text(species, style: Theme.of(context).textTheme.titleMedium?.copyWith(color: AppColors.gold)),
                Text(pet.breed, style: TextStyle(color: p.textMuted)),
                if (pet.microchipNumber != null) ...[
                  const SizedBox(height: 4),
                  Text(
                    l10n.petMicrochipLabel(pet.microchipNumber!),
                    style: TextStyle(color: p.textMuted, fontSize: 13),
                  ),
                ],
                if (pet.healthBookNumber != null) ...[
                  const SizedBox(height: 4),
                  Text(
                    l10n.petHealthBookNumberLabel(pet.healthBookNumber!),
                    style: TextStyle(color: p.textMuted, fontSize: 13),
                  ),
                ],
                if (pet.healthBookPdfAttached) ...[
                  TextButton.icon(
                    key: Key('pet_health_book_open_${pet.id}'),
                    onPressed: _openHealthBookPdf,
                    icon: const Icon(Icons.picture_as_pdf_outlined),
                    label: Text(l10n.petHealthBookOpenPdf),
                  ),
                ],
                if (pet.weightKg != null) ...[
                  const SizedBox(height: 4),
                  Text(
                    l10n.weightLastLabel(
                      pet.weightKg!.toStringAsFixed(
                        pet.weightKg! == pet.weightKg!.roundToDouble() ? 0 : 2,
                      ),
                    ),
                    style: TextStyle(color: p.textMuted, fontSize: 13),
                  ),
                ],
              ],
            ),
          ),
          if (pet.isOwner && pet.needsVetLink) ...[
            const SizedBox(height: 24),
            Card(
              key: Key('pet_link_vet_banner_${pet.id}'),
              color: p.surfaceElevated,
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(
                      l10n.linkVetAfterSaveTitle,
                      style: Theme.of(context).textTheme.titleSmall,
                    ),
                    const SizedBox(height: 4),
                    Text(
                      l10n.linkVetAfterSaveBody,
                      style: TextStyle(color: p.textMuted, fontSize: 13),
                    ),
                    const SizedBox(height: 12),
                    FilledButton.icon(
                      key: Key('pet_link_vet_cta_${pet.id}'),
                      onPressed: () async {
                        await Navigator.push(
                          context,
                          MaterialPageRoute(builder: (_) => const MyVetsScreen()),
                        );
                        if (mounted) {
                          await _loadVets();
                          await _reloadPet();
                        }
                      },
                      icon: const Icon(Icons.local_hospital_outlined),
                      label: Text(l10n.addVetSearchLabel),
                    ),
                  ],
                ),
              ),
            ),
          ],
          if (pet.isOwner &&
              pet.isActive &&
              !pet.needsVetLink &&
              pet.species == 'horse' &&
              FeatureModulesController.instance.horse) ...[
            const SizedBox(height: 24),
            HorseHealthPanel(
              petId: pet.id,
              petName: pet.name,
              foodChainStatus: pet.foodChainStatus,
              domicileLocation: pet.domicileLocation,
            ),
          ],
          if (hrPoints.isNotEmpty || weightPoints.isNotEmpty || bpPoints.isNotEmpty) ...[
            const SizedBox(height: 24),
            if (hrPoints.isNotEmpty) ...[
              Text(l10n.heartRateShort, style: Theme.of(context).textTheme.titleMedium),
              const SizedBox(height: 8),
              HeartRateChart(points: hrPoints, height: 180),
              const SizedBox(height: 16),
            ],
            if (weightPoints.isNotEmpty) ...[
              Text(l10n.weightShort, style: Theme.of(context).textTheme.titleMedium),
              const SizedBox(height: 8),
              WeightChart(points: weightPoints, height: 180),
              const SizedBox(height: 16),
            ],
            if (bpPoints.isNotEmpty) ...[
              Text(l10n.bloodPressureShort, style: Theme.of(context).textTheme.titleMedium),
              const SizedBox(height: 8),
              BloodPressureChart(points: bpPoints, height: 180),
            ],
          ],
          if (labPanelCount > 0) ...[
            const SizedBox(height: 16),
            OutlinedButton.icon(
              key: Key('pet_labs_open_${pet.id}'),
              onPressed: () async {
                await Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => LabPanelsScreen(petId: pet.id),
                  ),
                );
              },
              icon: const Icon(Icons.science_outlined, size: 18),
              label: Text(l10n.labsOpen),
            ),
          ],
          const SizedBox(height: 24),
          if (pet.isOwner && pet.isActive) ...[
            PetQuickActions(
              petId: pet.id,
              onHeartRate: supportsHeartRateControl(pet.species)
                  ? () async {
                      await Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (_) => HeartRateFlowScreen(
                            petId: pet.id,
                            durationsSec: pet.heartrateDurationsSec,
                            species: pet.species,
                          ),
                        ),
                      );
                      if (mounted) await _loadCharts();
                    }
                  : null,
              onWeightRecorded: () async {
                await _reloadPet();
              },
              onBloodPressureRecorded: () async {
                await _loadCharts();
              },
            ),
            const SizedBox(height: 8),
          ],
          if (pet.needsResumePayment) ...[
            Text(
              l10n.paymentFeaturesLocked,
              style: TextStyle(color: p.textMuted, fontSize: 13),
            ),
            const SizedBox(height: 8),
            FilledButton.icon(
              key: Key('pet_resume_payment_${pet.id}'),
              onPressed: () async {
                final messenger = ScaffoldMessenger.of(context);
                final l10n = AppLocalizations.of(context)!;
                try {
                  final url = await ApiClient.instance.resumeCheckout(pet.id);
                  final opened = await openExternalUrl(url);
                  if (!opened && mounted) {
                    messenger.showSnackBar(
                      SnackBar(content: Text(l10n.errorCouldNotOpenLink)),
                    );
                  }
                  // Do not reload here — wait for payment deep link / resume.
                } catch (e) {
                  if (!mounted) return;
                  messenger.showSnackBar(
                    SnackBar(content: Text(mapApiError(e, l10n))),
                  );
                }
              },
              icon: const Icon(Icons.payment),
              label: Text(l10n.paymentResume),
            ),
            const SizedBox(height: 8),
          ],
          const SizedBox(height: 8),
          if (pet.isActive) ...[
            _ActionTile(
              key: Key('pet_consultations_${pet.id}'),
              icon: Icons.description_outlined,
              label: l10n.consultationsHistory,
              onTap: () => Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) => PetTimelineScreen(
                    petId: pet.id,
                    petName: pet.name,
                    canWriteNotes: pet.canWriteNotes,
                  ),
                ),
              ),
            ),
            _ActionTile(
              icon: Icons.history,
              label: l10n.visitHistory,
              onTap: () => Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) => PetTimelineScreen(
                    petId: pet.id,
                    petName: pet.name,
                    canWriteNotes: pet.canWriteNotes,
                  ),
                ),
              ),
            ),
          ],
          if (pet.isOwner && pet.isActive && !pet.needsVetLink) ...[
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
                MaterialPageRoute(
                  builder: (_) => MessagingScreen(initialPetId: pet.id),
                ),
              ),
            ),
          ],
          if (pet.isOwner && pet.isActive)
            _ActionTile(
              key: Key('pet_send_dossier_${pet.id}'),
              icon: Icons.send_outlined,
              label: l10n.sendDossierToPro,
              onTap: _sendDossierToPro,
            ),
          if (pet.isOwner) ...[
            if (pet.isActive && !pet.needsVetLink) const Divider(height: 32),
            Row(
              children: [
                Expanded(child: Text(l10n.myVets, style: Theme.of(context).textTheme.titleSmall)),
                // Flexible : sur un écran de 360 dp le libellé du bouton réclame
                // plus que la largeur totale, écrasant le titre à zéro.
                Flexible(
                  child: TextButton(
                    onPressed: () async {
                      await Navigator.push(context, MaterialPageRoute(builder: (_) => const MyVetsScreen()));
                      if (mounted) {
                        await _loadVets();
                        await _reloadPet();
                      }
                    },
                    child: Text(l10n.addVetSearchLabel, textAlign: TextAlign.end),
                  ),
                ),
              ],
            ),
            if (loadingVets)
              const Center(child: Padding(padding: EdgeInsets.all(16), child: CircularProgressIndicator()))
            else if (vetsLoadError != null)
              ListTile(
                leading: Icon(Icons.cloud_off_outlined, color: p.textMuted),
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
                subtitle: Text(l10n.addVetSearchHint),
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
                  key: Key('pet_manage_subscription_${pet.id}'),
                  onPressed: _openBillingPortal,
                  icon: const Icon(Icons.settings),
                  label: Text(l10n.manageSubscription),
                ),
              ),
          ],
        ],
      ),
    );
  }

  Future<void> _openBillingPortal() async {
    final l10n = AppLocalizations.of(context)!;
    try {
      final url = await ApiClient.instance.billingPortal(pet.id);
      final opened = await openExternalUrl(url);
      if (!opened && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorCouldNotOpenLink)),
        );
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(mapApiError(e, l10n))),
      );
    }
  }

  Future<void> _sendDossierToPro() async {
    final l10n = AppLocalizations.of(context)!;
    final email = await showDialog<String>(
      context: context,
      builder: (ctx) => _SendDossierEmailDialog(l10n: l10n),
    );
    if (email == null || email.isEmpty || !mounted) return;
    if (!email.contains('@')) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.sendDossierInvalidEmail)),
      );
      return;
    }
    final messenger = ScaffoldMessenger.of(context);
    try {
      await ApiClient.instance.sendPetDossierShare(pet.id, email);
      if (!mounted) return;
      messenger.showSnackBar(SnackBar(content: Text(l10n.sendDossierSuccess)));
    } catch (e) {
      if (!mounted) return;
      messenger.showSnackBar(SnackBar(content: Text(mapApiError(e, l10n))));
    }
  }
}

class _SendDossierEmailDialog extends StatefulWidget {
  const _SendDossierEmailDialog({required this.l10n});

  final AppLocalizations l10n;

  @override
  State<_SendDossierEmailDialog> createState() => _SendDossierEmailDialogState();
}

class _SendDossierEmailDialogState extends State<_SendDossierEmailDialog> {
  late final TextEditingController _controller;
  bool _consented = false;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = widget.l10n;
    return AlertDialog(
      key: const Key('pet_send_dossier_dialog'),
      // Depuis l'ajout de l'avertissement PHI, le dialogue dépasse la hauteur
      // laissée par le clavier sur un petit écran (le champ e-mail a autofocus).
      scrollable: true,
      title: Text(l10n.sendDossierToPro),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            l10n.sendDossierPhiWarning,
            style: Theme.of(context).textTheme.bodySmall,
          ),
          const SizedBox(height: 12),
          TextField(
            key: const Key('pet_send_dossier_email'),
            controller: _controller,
            keyboardType: TextInputType.emailAddress,
            autofocus: true,
            decoration: InputDecoration(
              labelText: l10n.sendDossierEmailLabel,
              hintText: l10n.sendDossierEmailHint,
            ),
          ),
          CheckboxListTile(
            key: const Key('pet_send_dossier_consent'),
            value: _consented,
            onChanged: (v) => setState(() => _consented = v ?? false),
            controlAffinity: ListTileControlAffinity.leading,
            contentPadding: EdgeInsets.zero,
            title: Text(
              l10n.sendDossierPhiConsent,
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(l10n.cancel),
        ),
        FilledButton(
          key: const Key('pet_send_dossier_confirm'),
          onPressed: _consented
              ? () => Navigator.pop(context, _controller.text.trim())
              : null,
          child: Text(l10n.sendDossierConfirm),
        ),
      ],
    );
  }
}

class _ActionTile extends StatelessWidget {
  const _ActionTile({
    super.key,
    required this.icon,
    required this.label,
    required this.onTap,
  });

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
