import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';
import 'package:geolocator/geolocator.dart';
import 'package:path_provider/path_provider.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/core/notifications/push_navigation.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/theme/appearance_settings_tile.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
import 'package:petsfollow_mobile/features/invite/presentation/app_invite_qr_screen.dart';
import 'package:petsfollow_mobile/features/messaging/presentation/messaging_screen.dart';
import 'package:petsfollow_mobile/features/profile/presentation/profile_screen.dart';
import 'package:petsfollow_mobile/features/settings/presentation/switch_profile_screen.dart';
import 'package:petsfollow_mobile/features/shell/presentation/pro_light_pet_screen.dart';
import 'package:petsfollow_mobile/features/support/presentation/support_report_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:record/record.dart';
import 'package:url_launcher/url_launcher.dart';

bool _isConsultationHasReportError(Object e) {
  if (e is! DioException) return false;
  final data = e.response?.data;
  if (data is! Map) return false;
  final err = data['error'];
  if (err is! Map) return false;
  final msgKey = (err['msgKey'] ?? err['messageKey'])?.toString();
  if (msgKey == 'consultation_has_report') return true;
  // Filet si msgKey absent : cancel walk-in → code conflict + 409.
  return e.response?.statusCode == 409 && err['code']?.toString() == 'conflict';
}

/// Terrain shell for care_pro (vet_light, farrier, …) and cabinet `vet`.
class ProLightShellScreen extends StatefulWidget {
  const ProLightShellScreen({super.key, required this.onLogout});

  final VoidCallback onLogout;

  @override
  State<ProLightShellScreen> createState() => _ProLightShellScreenState();
}

String proLightSpecialtyLabel(AppLocalizations l10n, String specialty) {
  switch (specialty) {
    case 'farrier':
      return l10n.proLightSpecialtyFarrier;
    case 'physio':
      return l10n.proLightSpecialtyPhysio;
    case 'behaviorist':
      return l10n.proLightSpecialtyBehaviorist;
    case 'groomer':
      return l10n.proLightSpecialtyGroomer;
    case 'breeder':
      return l10n.proLightSpecialtyBreeder;
    case 'vet_light':
      return l10n.proLightSpecialtyVetLight;
    default:
      return specialty;
  }
}

class _ProLightShellScreenState extends State<ProLightShellScreen> {
  int _index = 0;
  bool _loading = true;
  String? _error;
  List<dynamic> _visits = [];
  List<dynamic> _clients = [];
  List<dynamic> _pets = [];
  int _loadGen = 0;
  int _messagesUnread = 0;

  /// Agenda / Clients / Pets / Messages / Settings (identical for care_pro + cabinet staff).
  static const _messagesIndex = 3;
  static const _settingsIndex = 4;

  @override
  void initState() {
    super.initState();
    _bindPushNavigation();
    _load();
  }

  @override
  void dispose() {
    final nav = PushNavigation.instance;
    if (identical(nav.onSelectTab, _onPushSelectTab)) {
      nav.onSelectTab = null;
    }
    super.dispose();
  }

  void _onPushSelectTab(int i) {
    if (!mounted) return;
    // Client shell uses tabMessages=3; map to Pro Light messages index.
    final target = i == PushNavigation.tabMessages ? _messagesIndex : i;
    if (target < 0 || target > _settingsIndex) return;
    setState(() => _index = target);
  }

  void _bindPushNavigation() {
    PushNavigation.instance.onSelectTab = _onPushSelectTab;
  }

  List<Map<String, dynamic>> get _staffClientMaps => _clients
      .whereType<Map>()
      .map((e) => Map<String, dynamic>.from(e))
      .toList();

  List<Map<String, dynamic>> get _staffPetMaps => _pets
      .whereType<Map>()
      .map((e) => Map<String, dynamic>.from(e))
      .toList();

  bool _canWriteNotes(Map<String, dynamic> row) {
    // Secretary defaults: no pets.write_clinical — hide write UI to avoid 403.
    final role = ApiClient.instance.userRole;
    if (role == 'secretary') return false;
    if (role == 'vet' || role == 'vet_assistant') return true;
    final p = row['permission'] as String? ?? 'read';
    return p == 'write_notes' || p == 'full';
  }

  Future<void> _load({bool silent = false}) async {
    final gen = ++_loadGen;
    if (!silent) {
      setState(() {
        _loading = true;
        _error = null;
      });
    }
    try {
      final lists = await ApiClient.instance.loadProTerrainLists();
      if (!mounted || gen != _loadGen) return;
      setState(() {
        _visits = lists.visits;
        _clients = lists.clients;
        _pets = lists.pets;
        _loading = false;
        _error = null;
      });
    } catch (_) {
      if (!mounted || gen != _loadGen) return;
      if (silent) {
        // Keep current lists after a successful mutation; toast only.
        _toast(AppLocalizations.of(context)!.proLightLoadError);
        return;
      }
      setState(() {
        _error = AppLocalizations.of(context)!.proLightLoadError;
        _loading = false;
      });
    }
  }

  Future<void> _openMaps(Map<String, dynamic> visit) async {
    final lat = visit['lat'];
    final lng = visit['lng'];
    final address = (visit['addressText'] as String?)?.trim() ?? '';
    Uri? uri;
    if (lat is num && lng is num) {
      uri = Uri.parse('https://www.google.com/maps/dir/?api=1&destination=$lat,$lng');
    } else if (address.isNotEmpty) {
      uri = Uri.parse(
        'https://www.google.com/maps/dir/?api=1&destination=${Uri.encodeComponent(address)}',
      );
    }
    if (uri != null) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    }
  }

  void _openPet(String petId, {String? petName, String? permission}) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => ProLightPetScreen(
          petId: petId,
          petName: petName,
          permission: permission,
          onNewConsultation: _startNewConsultation,
        ),
      ),
    );
  }

  /// Étape 1 consultation terrain : visite confirmée → CR sheet.
  bool _consultBusy = false;

  Future<void> _startNewConsultation(Map<String, dynamic> pet) async {
    final petId = pet['id'] as String? ?? '';
    if (petId.isEmpty || _consultBusy) return;
    _consultBusy = true;
    final l10n = AppLocalizations.of(context)!;
    try {
      final visit = await ApiClient.instance.createVisit(
        petId,
        scheduledAt: DateTime.now(),
        confirmDirect: true,
        silentConfirm: true,
        consultationSession: true,
        durationMinutes: 30,
      );
      if (!mounted) return;
      await _load(silent: true);
      if (!mounted) return;
      await _openReport({
        'id': visit.id,
        'petId': visit.petId,
        'permission': pet['permission'] ?? 'write_notes',
        'clientUserId': pet['ownerUserId'],
        'ownerUserId': pet['ownerUserId'],
      }, showNextSteps: true);
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(mapApiError(e, l10n))),
      );
    } finally {
      _consultBusy = false;
    }
  }

  void _openClientPets(Map<String, dynamic> client) {
    final clientId = client['userId'] as String? ?? '';
    final name = client['fullName'] as String? ?? '';
    final pets = _pets
        .map((p) => Map<String, dynamic>.from(p as Map))
        .where((p) => (p['ownerUserId'] as String?) == clientId)
        .toList();
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (ctx) {
          final l10n = AppLocalizations.of(ctx)!;
          return Scaffold(
            appBar: AppBar(title: Text(name.isEmpty ? l10n.proLightClients : name)),
            body: pets.isEmpty
                ? Center(child: Text(l10n.proLightNoPets))
                : ListView.separated(
                    padding: const EdgeInsets.all(16),
                    itemCount: pets.length,
                    separatorBuilder: (_, __) => const Divider(height: 1),
                    itemBuilder: (_, i) {
                      final p = pets[i];
                      final petId = p['id'] as String? ?? '';
                      final canConsult = _canWriteNotes(p);
                      return ListTile(
                        title: Text('${p['name'] ?? ''}'),
                        subtitle: Text('${p['species'] ?? ''}'),
                        trailing: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            if (canConsult && petId.isNotEmpty)
                              IconButton(
                                key: Key('pro_light_client_pet_consultation_$petId'),
                                tooltip: l10n.proLightNewConsultation,
                                icon: const Icon(Icons.medical_services_outlined),
                                onPressed: () => _startNewConsultation(p),
                              ),
                            const Icon(Icons.chevron_right),
                          ],
                        ),
                        onTap: petId.isEmpty
                            ? null
                            : () => _openPet(
                                  petId,
                                  petName: p['name'] as String?,
                                  permission: p['permission'] as String?,
                                ),
                      );
                    },
                  ),
          );
        },
      ),
    );
  }

  Future<void> _openReport(
    Map<String, dynamic> visit, {
    bool showNextSteps = false,
  }) async {
    final visitId = visit['id'] as String?;
    if (visitId == null) return;
    final canWrite = _canWriteNotes(visit);
    final l10n = AppLocalizations.of(context)!;
    String initialText = '';
    String transcript = '';
    String improved = '';
    var status = 'none';
    try {
      final report = await ApiClient.instance.getVisitReport(visitId);
      transcript = (report['transcriptText'] as String?) ?? '';
      improved = (report['improvedText'] as String?) ?? '';
      initialText = (report['bodyText'] as String?) ??
          (transcript.isNotEmpty ? transcript : '');
      status = report['status'] as String? ?? 'draft';
    } catch (_) {
      if (showNextSteps) {
        try {
          await ApiClient.instance.updateVisit(visitId, 'cancelled');
        } catch (_) {
          // best-effort cleanup
        }
        if (mounted) await _load(silent: true);
      }
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.proLightActionFailed)),
      );
      return;
    }
    if (!mounted) return;
    final saved = await showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      builder: (ctx) => _VisitReportSheet(
        visitId: visitId,
        initialText: initialText,
        initialTranscript: transcript,
        initialImproved: improved,
        initialStatus: status,
        canWrite: canWrite,
        clientUserId: (visit['clientUserId'] ?? visit['ownerUserId']) as String?,
        petId: visit['petId'] as String?,
        showNextStepsOnSave: showNextSteps,
      ),
    );
    // Walk-in consultation: cancel orphan visit if CR sheet closed without save.
    if (showNextSteps && saved != true) {
      try {
        await ApiClient.instance.updateVisit(visitId, 'cancelled');
      } on DioException catch (e) {
        if (!_isConsultationHasReportError(e)) {
          // best-effort cleanup
        }
        // 409 consultation_has_report → CR already persisted; keep visit.
      } catch (_) {
        // best-effort cleanup
      }
      if (mounted) await _load(silent: true);
    }
  }

  void _toast(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(message)));
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final specialty = ApiClient.instance.userSpecialty ?? '';
    return Container(
      decoration: BoxDecoration(gradient: AppTheme.loginGradientOf(context)),
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          title: const PetsLogo(variant: PetsLogoVariant.horizontal, height: 28),
          actions: [
            IconButton(
              key: const Key('pro_light_support_btn'),
              tooltip: l10n.supportMenu,
              onPressed: () => openSupportReport(context, source: 'flutter_pro_light'),
              icon: const Icon(Icons.support_agent_outlined),
            ),
            IconButton(
              tooltip: l10n.appInviteTitle,
              onPressed: () {
                Navigator.of(context).push(
                  MaterialPageRoute<void>(
                    builder: (_) => const AppInviteQrScreen(),
                  ),
                );
              },
              icon: const Icon(Icons.qr_code_2),
            ),
            IconButton(
              onPressed: _load,
              icon: const Icon(Icons.refresh),
            ),
          ],
        ),
        body: _loading
            ? const Center(child: CircularProgressIndicator())
            : _error != null
                ? Center(child: Text(_error!, style: const TextStyle(color: AppColors.alert)))
                : IndexedStack(
                    index: _index,
                    children: [
                      _VisitsTab(
                        visits: _visits,
                        emptyAllLabel: specialty == 'farrier'
                            ? l10n.proLightEmptyFarrier
                            : l10n.proLightNoVisits,
                        emptyTodayLabel: l10n.proLightNoTourToday,
                        emptyWeekLabel: l10n.proLightNoTourWeek,
                        tourTodayLabel: l10n.proLightTourToday,
                        tourWeekLabel: l10n.proLightTourWeek,
                        tourAllLabel: l10n.proLightTourAll,
                        onMaps: _openMaps,
                        onReport: _openReport,
                        canWriteNotes: _canWriteNotes,
                        onOpenPet: (visit) {
                          final petId = visit['petId'] as String?;
                          if (petId == null || petId.isEmpty) return;
                          _openPet(petId, petName: visit['petName'] as String?);
                        },
                        onSaveLocation: (visit, address, {lat, lng, clearCoords = false}) async {
                          final id = visit['id'] as String?;
                          if (id == null) return;
                          try {
                            await ApiClient.instance.updateVisitLocation(
                              id,
                              address,
                              lat: lat,
                              lng: lng,
                              clearCoords: clearCoords,
                            );
                            await _load(silent: true);
                          } catch (_) {
                            _toast(l10n.proLightActionFailed);
                          }
                        },
                        onMarkDone: (visit) async {
                          final id = visit['id'] as String?;
                          if (id == null) return;
                          try {
                            await ApiClient.instance.updateVisit(id, 'done');
                            await _load(silent: true);
                          } catch (_) {
                            _toast(l10n.proLightActionFailed);
                          }
                        },
                        addressLabel: l10n.proLightAddress,
                        mapsLabel: l10n.proLightOpenMaps,
                        reportLabel: l10n.proLightReportTitle,
                        petLabel: l10n.proLightPets,
                        readOnlyLabel: l10n.proLightReadOnly,
                        gpsLabel: l10n.proLightUseGps,
                        gpsDeniedLabel: l10n.proLightGpsDenied,
                        doneLabel: l10n.careDone,
                      ),
                      _ListTab(
                        empty: l10n.proLightNoClients,
                        items: _clients
                            .map((c) => Map<String, dynamic>.from(c as Map))
                            .toList(),
                        titleKey: 'fullName',
                        subtitleKey: 'email',
                        onTap: _openClientPets,
                      ),
                      _ListTab(
                        empty: l10n.proLightNoPets,
                        items: _pets
                            .map((p) => Map<String, dynamic>.from(p as Map))
                            .toList(),
                        titleKey: 'name',
                        subtitleKey: 'species',
                        onTap: (row) {
                          final id = row['id'] as String?;
                          if (id == null || id.isEmpty) return;
                          _openPet(
                            id,
                            petName: row['name'] as String?,
                            permission: row['permission'] as String?,
                          );
                        },
                      ),
                      MessagingScreen(
                        key: const Key('pro_light_messaging'),
                        embedded: true,
                        active: _index == _messagesIndex,
                        staffMode: true,
                        staffClients: _staffClientMaps,
                        staffPets: _staffPetMaps,
                        onUnreadTotalChanged: (n) {
                          if (!mounted || n == _messagesUnread) return;
                          setState(() => _messagesUnread = n);
                        },
                      ),
                      _SettingsTab(onLogout: widget.onLogout),
                    ],
                  ),
        bottomNavigationBar: NavigationBar(
          key: const Key('pro_light_nav'),
          selectedIndex: _index.clamp(0, _settingsIndex),
          onDestinationSelected: (i) => setState(() => _index = i),
          destinations: [
            NavigationDestination(icon: const Icon(Icons.event), label: l10n.proLightAgenda),
            NavigationDestination(icon: const Icon(Icons.people), label: l10n.proLightClients),
            NavigationDestination(icon: const Icon(Icons.pets), label: l10n.proLightPets),
            NavigationDestination(
              icon: Badge(
                isLabelVisible: _messagesUnread > 0,
                label: Text(_messagesUnread > 99 ? '99+' : '$_messagesUnread'),
                child: const Icon(Icons.chat_bubble_outline),
              ),
              label: l10n.vetMessaging,
            ),
            NavigationDestination(
              icon: const Icon(Icons.settings_outlined),
              label: l10n.proLightSettings,
            ),
          ],
        ),
      ),
    );
  }
}

class _SettingsTab extends StatelessWidget {
  const _SettingsTab({required this.onLogout});

  final VoidCallback onLogout;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final specialty = ApiClient.instance.userSpecialty ?? '';
    return ListView(
      children: [
        ListTile(
          leading: const Icon(Icons.badge_outlined),
          title: Text(l10n.proLightSpecialty),
          subtitle: Text(specialty.isEmpty ? '—' : proLightSpecialtyLabel(l10n, specialty)),
        ),
        ListTile(
          leading: const Icon(Icons.person_outline),
          title: Text(l10n.myData),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const ProfileScreen()),
          ),
        ),
        ListTile(
          leading: const Icon(Icons.swap_horiz),
          title: Text(l10n.switchProfile),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const SwitchProfileScreen()),
          ),
        ),
        ListTile(
          leading: const Icon(Icons.qr_code_2),
          title: Text(l10n.appInviteTitle),
          subtitle: Text(l10n.appInviteHintShort),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const AppInviteQrScreen()),
          ),
        ),
        ListenableBuilder(
          listenable: LocaleController.instance,
          builder: (context, _) {
            final code = LocaleController.instance.languageCode;
            return ListTile(
              leading: const Icon(Icons.language),
              title: Text(l10n.language),
              trailing: DropdownButton<String>(
                value: code,
                underline: const SizedBox.shrink(),
                items: [
                  DropdownMenuItem(value: 'fr', child: Text(l10n.languageFr)),
                  DropdownMenuItem(value: 'nl', child: Text(l10n.languageNl)),
                  DropdownMenuItem(value: 'en', child: Text(l10n.languageEn)),
                  DropdownMenuItem(value: 'es', child: Text(l10n.languageEs)),
                  DropdownMenuItem(value: 'et', child: Text(l10n.languageEt)),
                  DropdownMenuItem(value: 'it', child: Text(l10n.languageIt)),
                ],
                onChanged: (next) async {
                  if (next == null || next == code) return;
                  try {
                    if (ApiClient.instance.token != null) {
                      await ApiClient.instance.updateLocale(next);
                    } else {
                      await LocaleController.instance.setLocale(next);
                    }
                  } catch (_) {
                    await LocaleController.instance.setLocale(next);
                  }
                },
              ),
            );
          },
        ),
        const AppearanceSettingsTile(),
        ListTile(
          key: const Key('pro_light_settings_support'),
          leading: const Icon(Icons.support_agent_outlined),
          title: Text(l10n.supportMenu),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => openSupportReport(context, source: 'flutter_pro_light'),
        ),
        ListTile(
          leading: const Icon(Icons.logout),
          title: Text(l10n.logout),
          onTap: () async {
            await ApiClient.instance.logout();
            onLogout();
          },
        ),
      ],
    );
  }
}

class _VisitReportSheet extends StatefulWidget {
  const _VisitReportSheet({
    required this.visitId,
    required this.initialText,
    required this.initialTranscript,
    required this.initialImproved,
    required this.initialStatus,
    required this.canWrite,
    this.clientUserId,
    this.petId,
    this.showNextStepsOnSave = false,
  });

  final String visitId;
  final String initialText;
  final String initialTranscript;
  final String initialImproved;
  final String initialStatus;
  final bool canWrite;
  final String? clientUserId;
  final String? petId;
  final bool showNextStepsOnSave;

  @override
  State<_VisitReportSheet> createState() => _VisitReportSheetState();
}

class _VisitReportSheetState extends State<_VisitReportSheet>
    with SingleTickerProviderStateMixin {
  late final TextEditingController _controller;
  late String _status;
  late String _transcript;
  late String _improved;
  late String _persistedBody;
  bool _busy = false;
  bool _recording = false;
  int _recordingSeconds = 0;
  Timer? _recordingTimer;
  final AudioRecorder _recorder = AudioRecorder();
  late final AnimationController _pulseController;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(text: widget.initialText);
    _status = widget.initialStatus;
    _transcript = widget.initialTranscript;
    _improved = widget.initialImproved;
    _persistedBody = widget.initialText;
    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 900),
    );
  }

  String get _historySaved {
    final body = _persistedBody.trim();
    if (body.isEmpty) return '';
    if (body == _transcript.trim() || body == _improved.trim()) return '';
    return _persistedBody;
  }

  void _applyPayload(Map<String, dynamic> data) {
    final transcript = (data['transcriptText'] as String?) ?? _transcript;
    final improved = (data['improvedText'] as String?) ?? _improved;
    final body = (data['bodyText'] as String?) ??
        (transcript.isNotEmpty ? transcript : _controller.text);
    final status = data['status'] as String? ?? _status;
    setState(() {
      _transcript = transcript;
      _improved = improved;
      _status = status;
      _persistedBody = body;
      _controller.text = body;
    });
  }

  @override
  void dispose() {
    _recordingTimer?.cancel();
    _pulseController.dispose();
    _controller.dispose();
    _recorder.dispose();
    super.dispose();
  }

  String get _specialty => ApiClient.instance.userSpecialty ?? '';

  String get _recordingClock {
    final m = (_recordingSeconds ~/ 60).toString().padLeft(2, '0');
    final s = (_recordingSeconds % 60).toString().padLeft(2, '0');
    return '$m:$s';
  }

  void _startRecordingTimer() {
    _recordingTimer?.cancel();
    _recordingSeconds = 0;
    _pulseController.repeat(reverse: true);
    _recordingTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (!mounted) return;
      setState(() => _recordingSeconds++);
    });
  }

  void _stopRecordingTimer() {
    _recordingTimer?.cancel();
    _recordingTimer = null;
    _recordingSeconds = 0;
    _pulseController
      ..stop()
      ..value = 1;
  }

  Future<bool> _confirmAudioConsent() async {
    final l10n = AppLocalizations.of(context)!;
    var checked = false;
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (ctx, setLocal) {
            return AlertDialog(
              title: Text(l10n.proLightAudioConsentTitle),
              content: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(l10n.proLightAudioConsentBody),
                  const SizedBox(height: 12),
                  CheckboxListTile(
                    key: const Key('pro_light_audio_consent_checkbox'),
                    contentPadding: EdgeInsets.zero,
                    controlAffinity: ListTileControlAffinity.leading,
                    value: checked,
                    onChanged: (v) => setLocal(() => checked = v ?? false),
                    title: Text(l10n.proLightAudioConsentClientCheck),
                  ),
                ],
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.pop(ctx, false),
                  child: Text(MaterialLocalizations.of(ctx).cancelButtonLabel),
                ),
                FilledButton(
                  key: const Key('pro_light_audio_consent_accept'),
                  onPressed: checked ? () => Navigator.pop(ctx, true) : null,
                  child: Text(l10n.proLightAudioConsentAccept),
                ),
              ],
            );
          },
        );
      },
    );
    return ok == true;
  }

  Future<void> _applyTranscript(Map<String, dynamic> transcribed) async {
    _applyPayload(transcribed);
  }

  Future<void> _toggleDictation() async {
    if (_recording) {
      final path = await _recorder.stop();
      _stopRecordingTimer();
      setState(() => _recording = false);
      if (path == null || path.isEmpty) return;
      final transcribed = await ApiClient.instance.transcribeVisitReport(
        widget.visitId,
        path,
        filename: 'dictation.m4a',
        clientAudioConsent: true,
      );
      if (!mounted) return;
      await _applyTranscript(transcribed);
      return;
    }
    if (!await _confirmAudioConsent()) return;
    if (!await _recorder.hasPermission()) {
      throw StateError('mic_denied');
    }
    final dir = await getTemporaryDirectory();
    final path =
        '${dir.path}/pf_visit_${widget.visitId}_${DateTime.now().millisecondsSinceEpoch}.m4a';
    await _recorder.start(
      const RecordConfig(encoder: AudioEncoder.aacLc),
      path: path,
    );
    if (!mounted) return;
    _startRecordingTimer();
    setState(() => _recording = true);
  }

  Future<void> _pickAudioFile() async {
    if (!await _confirmAudioConsent()) return;
    final picked = await FilePicker.pickFiles(
      type: FileType.custom,
      allowedExtensions: const ['mp3', 'm4a', 'wav', 'ogg', 'webm'],
    );
    if (picked == null || picked.files.isEmpty) return;
    final file = picked.files.first;
    final path = file.path;
    if (path == null || path.isEmpty) {
      throw StateError('no_path');
    }
    final transcribed = await ApiClient.instance.transcribeVisitReport(
      widget.visitId,
      path,
      filename: file.name,
      clientAudioConsent: true,
    );
    if (!mounted) return;
    await _applyTranscript(transcribed);
  }

  Future<void> _run(Future<void> Function() action, {bool popOnOk = false}) async {
    setState(() => _busy = true);
    try {
      await action();
      if (!mounted) return;
      if (popOnOk) {
        if (widget.showNextStepsOnSave) {
          await _showConsultationNextSteps();
        }
        if (mounted) Navigator.pop(context, true);
      }
    } catch (e) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      final msg = e is StateError && e.message == 'mic_denied'
          ? l10n.proLightMicDenied
          : mapApiError(e, l10n);
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  String get _proSiteBase {
    const defined = String.fromEnvironment('PRO_PUBLIC_SITE_URL');
    if (defined.isNotEmpty) return defined.replaceAll(RegExp(r'/$'), '');
    return 'http://localhost:3002';
  }

  Future<void> _openProPath(String pathAndQuery) async {
    final uri = Uri.parse('$_proSiteBase$pathAndQuery');
    await launchUrl(uri, mode: LaunchMode.externalApplication);
  }

  Future<void> _showConsultationNextSteps() async {
    final l10n = AppLocalizations.of(context)!;
    final clientId = widget.clientUserId ?? '';
    final petId = widget.petId ?? '';
    final visitId = widget.visitId;
    final role = ApiClient.instance.userRole;
    final showProBilling = role == 'vet' || role == 'vet_assistant';
    await showModalBottomSheet<void>(
      context: context,
      builder: (ctx) {
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(
                  l10n.proLightConsultationNextTitle,
                  style: Theme.of(ctx).textTheme.titleMedium,
                ),
                const SizedBox(height: 12),
                if (showProBilling) ...[
                  FilledButton(
                    key: const Key('pro_light_consultation_cta_daf'),
                    onPressed: () async {
                      final q = <String, String>{
                        if (clientId.isNotEmpty) 'clientUserId': clientId,
                        if (petId.isNotEmpty) 'petId': petId,
                        'visitId': visitId,
                      };
                      final qs = Uri(queryParameters: q).query;
                      await _openProPath('/daf/nouveau?$qs');
                      if (ctx.mounted) Navigator.pop(ctx);
                    },
                    child: Text(l10n.proLightConsultationCtaDaf),
                  ),
                  const SizedBox(height: 8),
                  OutlinedButton(
                    key: const Key('pro_light_consultation_cta_invoice'),
                    onPressed: () async {
                      final q = <String, String>{
                        if (clientId.isNotEmpty) 'clientUserId': clientId,
                        'visitId': visitId,
                        'mode': 'direct',
                      };
                      final qs = Uri(queryParameters: q).query;
                      await _openProPath('/invoicing?$qs');
                      if (ctx.mounted) Navigator.pop(ctx);
                    },
                    child: Text(l10n.proLightConsultationCtaInvoice),
                  ),
                  const SizedBox(height: 8),
                ],
                TextButton(
                  key: const Key('pro_light_consultation_cta_done'),
                  onPressed: () async {
                    try {
                      await ApiClient.instance.updateVisit(visitId, 'done');
                    } catch (_) {}
                    if (ctx.mounted) Navigator.pop(ctx);
                  },
                  child: Text(l10n.proLightConsultationCtaDone),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _recordingBanner(AppLocalizations l10n) {
    return Container(
      key: const Key('pro_light_cr_recording_banner'),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
      decoration: BoxDecoration(
        color: AppColors.alert.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.alert.withValues(alpha: 0.4)),
      ),
      child: Row(
        children: [
          FadeTransition(
            opacity: Tween<double>(begin: 0.35, end: 1).animate(_pulseController),
            child: const Icon(Icons.mic, color: AppColors.alert, size: 28),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  l10n.proLightRecordingInProgress,
                  style: Theme.of(context).textTheme.titleSmall?.copyWith(
                        color: AppColors.alert,
                        fontWeight: FontWeight.w600,
                      ),
                ),
                Text(
                  _recordingClock,
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
            ),
          ),
          FilledButton(
            key: const Key('pro_light_cr_dictation_stop'),
            style: FilledButton.styleFrom(backgroundColor: AppColors.alert),
            onPressed: _busy
                ? null
                : () => _run(() async {
                      await _toggleDictation();
                    }),
            child: Text(l10n.proLightDictationStop),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final isFinal = _status == 'final';
    final readOnly = !widget.canWrite || isFinal || _busy || _recording;
    final hint =
        _specialty == 'farrier' ? l10n.proLightReportHintFarrier : l10n.proLightReportHint;
    final bottomInset = MediaQuery.of(context).viewInsets.bottom;
    return PopScope(
      canPop: !_busy,
      child: SafeArea(
      child: Padding(
        padding: EdgeInsets.only(
          left: 16,
          right: 16,
          top: 8,
          bottom: bottomInset + 16,
        ),
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                isFinal
                    ? '${l10n.proLightReportTitle} · ${l10n.proLightReportFinal}'
                    : (!widget.canWrite
                        ? '${l10n.proLightReportTitle} · ${l10n.proLightReadOnly}'
                        : l10n.proLightReportTitle),
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _controller,
                maxLines: 8,
                readOnly: readOnly || !widget.canWrite,
                decoration: InputDecoration(hintText: hint),
              ),
              if (_transcript.isNotEmpty ||
                  _improved.isNotEmpty ||
                  _historySaved.isNotEmpty) ...[
                const SizedBox(height: 8),
                ExpansionTile(
                  tilePadding: EdgeInsets.zero,
                  title: Text(l10n.proLightReportHistoryTitle),
                  children: [
                    Align(
                      alignment: Alignment.centerLeft,
                      child: Text(
                        l10n.proLightReportHistoryTranscript,
                        style: Theme.of(context).textTheme.labelLarge,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Align(
                      alignment: Alignment.centerLeft,
                      child: Text(
                        _transcript.isEmpty
                            ? l10n.proLightReportHistoryEmpty
                            : _transcript,
                      ),
                    ),
                    if (_improved.isNotEmpty) ...[
                      const SizedBox(height: 12),
                      Align(
                        alignment: Alignment.centerLeft,
                        child: Text(
                          l10n.proLightReportHistoryImproved,
                          style: Theme.of(context).textTheme.labelLarge,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Align(
                        alignment: Alignment.centerLeft,
                        child: Text(_improved),
                      ),
                    ],
                    if (_historySaved.isNotEmpty) ...[
                      const SizedBox(height: 12),
                      Align(
                        alignment: Alignment.centerLeft,
                        child: Text(
                          l10n.proLightReportHistorySaved,
                          style: Theme.of(context).textTheme.labelLarge,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Align(
                        alignment: Alignment.centerLeft,
                        child: Text(_historySaved),
                      ),
                    ],
                  ],
                ),
              ],
              if (widget.canWrite && !isFinal) ...[
                const SizedBox(height: 16),
                if (_recording)
                  _recordingBanner(l10n)
                else ...[
                  FilledButton.tonalIcon(
                    key: const Key('pro_light_cr_dictation'),
                    onPressed: _busy
                        ? null
                        : () => _run(() async {
                              await _toggleDictation();
                            }),
                    icon: const Icon(Icons.mic_none),
                    label: Text(l10n.proLightDictationStart),
                  ),
                  Align(
                    alignment: Alignment.centerLeft,
                    child: TextButton.icon(
                      key: const Key('pro_light_cr_audio_file'),
                      onPressed: _busy
                          ? null
                          : () => _run(() async {
                                await _pickAudioFile();
                              }),
                      icon: const Icon(Icons.attach_file, size: 18),
                      label: Text(l10n.proLightTranscribeAudio),
                    ),
                  ),
                ],
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(
                      child: OutlinedButton(
                        key: const Key('pro_light_cr_save'),
                        onPressed: _busy || _recording
                            ? null
                            : () => _run(() async {
                                  final saved = await ApiClient.instance
                                      .putVisitReport(widget.visitId, _controller.text);
                                  if (!mounted) return;
                                  _applyPayload(saved);
                                }, popOnOk: true),
                        child: Text(l10n.save),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: FilledButton(
                        key: const Key('pro_light_cr_finalize'),
                        onPressed: _busy || _recording
                            ? null
                            : () => _run(() async {
                                  await ApiClient.instance
                                      .putVisitReport(widget.visitId, _controller.text);
                                  final finalized = await ApiClient.instance
                                      .finalizeVisitReport(widget.visitId);
                                  if (!mounted) return;
                                  _applyPayload(finalized);
                                }, popOnOk: true),
                        child: Text(l10n.proLightFinalizeReport),
                      ),
                    ),
                  ],
                ),
              ],
            ],
          ),
        ),
      ),
    ),
    );
  }
}

enum _TourFilter { today, week, all }

class _VisitsTab extends StatefulWidget {
  const _VisitsTab({
    required this.visits,
    required this.emptyAllLabel,
    required this.emptyTodayLabel,
    required this.emptyWeekLabel,
    required this.tourTodayLabel,
    required this.tourWeekLabel,
    required this.tourAllLabel,
    required this.onMaps,
    required this.onReport,
    required this.canWriteNotes,
    required this.onOpenPet,
    required this.onSaveLocation,
    required this.onMarkDone,
    required this.addressLabel,
    required this.mapsLabel,
    required this.reportLabel,
    required this.petLabel,
    required this.readOnlyLabel,
    required this.gpsLabel,
    required this.gpsDeniedLabel,
    required this.doneLabel,
  });

  final List<dynamic> visits;
  final String emptyAllLabel;
  final String emptyTodayLabel;
  final String emptyWeekLabel;
  final String tourTodayLabel;
  final String tourWeekLabel;
  final String tourAllLabel;
  final Future<void> Function(Map<String, dynamic>) onMaps;
  final Future<void> Function(Map<String, dynamic>) onReport;
  final bool Function(Map<String, dynamic>) canWriteNotes;
  final void Function(Map<String, dynamic>) onOpenPet;
  final Future<void> Function(
    Map<String, dynamic>,
    String, {
    double? lat,
    double? lng,
    bool clearCoords,
  }) onSaveLocation;
  final Future<void> Function(Map<String, dynamic>) onMarkDone;
  final String addressLabel;
  final String mapsLabel;
  final String reportLabel;
  final String petLabel;
  final String readOnlyLabel;
  final String gpsLabel;
  final String gpsDeniedLabel;
  final String doneLabel;

  @override
  State<_VisitsTab> createState() => _VisitsTabState();
}

class _VisitsTabState extends State<_VisitsTab> {
  _TourFilter _filter = _TourFilter.today;

  /// Aligné Nuxt `visitDisplayAt` / `Visit.displayDate`.
  DateTime? _parseDisplayAt(Map<String, dynamic> v) {
    for (final key in ['proposedScheduledAt', 'scheduledAt', 'createdAt']) {
      final raw = v[key]?.toString();
      if (raw == null || raw.isEmpty) continue;
      final at = DateTime.tryParse(raw)?.toLocal();
      if (at != null) return at;
    }
    return null;
  }

  bool _isOpenTourStatus(Map<String, dynamic> v) {
    final s = v['status']?.toString() ?? '';
    return s != 'done' && s != 'cancelled';
  }

  List<Map<String, dynamic>> _filteredVisits() {
    final now = DateTime.now();
    final startToday = DateTime(now.year, now.month, now.day);
    final endToday = startToday.add(const Duration(days: 1));
    final endWeek = startToday.add(const Duration(days: 7));

    final rows = widget.visits
        .map((e) => Map<String, dynamic>.from(e as Map))
        .toList();

    Iterable<Map<String, dynamic>> filtered = rows;
    if (_filter == _TourFilter.today) {
      filtered = rows.where((v) {
        if (!_isOpenTourStatus(v)) return false;
        final at = _parseDisplayAt(v);
        return at != null && !at.isBefore(startToday) && at.isBefore(endToday);
      });
    } else if (_filter == _TourFilter.week) {
      filtered = rows.where((v) {
        if (!_isOpenTourStatus(v)) return false;
        final at = _parseDisplayAt(v);
        return at != null && !at.isBefore(startToday) && at.isBefore(endWeek);
      });
    }

    final list = filtered.toList();
    if (_filter != _TourFilter.all) {
      list.sort((a, b) {
        final da = _parseDisplayAt(a) ?? DateTime.fromMillisecondsSinceEpoch(0);
        final db = _parseDisplayAt(b) ?? DateTime.fromMillisecondsSinceEpoch(0);
        return da.compareTo(db);
      });
    }
    return list;
  }

  String _emptyForFilter() {
    switch (_filter) {
      case _TourFilter.today:
        return widget.emptyTodayLabel;
      case _TourFilter.week:
        return widget.emptyWeekLabel;
      case _TourFilter.all:
        return widget.emptyAllLabel;
    }
  }

  String _formatWhen(Map<String, dynamic> v) {
    final at = _parseDisplayAt(v);
    if (at == null) {
      return v['status']?.toString() ?? '';
    }
    final hh = at.hour.toString().padLeft(2, '0');
    final mm = at.minute.toString().padLeft(2, '0');
    if (_filter == _TourFilter.today) {
      return '$hh:$mm';
    }
    final d = at.day.toString().padLeft(2, '0');
    final m = at.month.toString().padLeft(2, '0');
    return '$d/$m $hh:$mm';
  }

  @override
  Widget build(BuildContext context) {
    final filtered = _filteredVisits();
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
          child: SegmentedButton<_TourFilter>(
            showSelectedIcon: false,
            style: const ButtonStyle(
              visualDensity: VisualDensity.compact,
              tapTargetSize: MaterialTapTargetSize.shrinkWrap,
            ),
            segments: [
              ButtonSegment(
                value: _TourFilter.today,
                label: FittedBox(fit: BoxFit.scaleDown, child: Text(widget.tourTodayLabel)),
              ),
              ButtonSegment(
                value: _TourFilter.week,
                label: FittedBox(fit: BoxFit.scaleDown, child: Text(widget.tourWeekLabel)),
              ),
              ButtonSegment(
                value: _TourFilter.all,
                label: FittedBox(fit: BoxFit.scaleDown, child: Text(widget.tourAllLabel)),
              ),
            ],
            selected: {_filter},
            onSelectionChanged: (s) {
              if (s.isEmpty) return;
              setState(() => _filter = s.first);
            },
          ),
        ),
        Expanded(
          child: filtered.isEmpty
              ? Center(child: Text(_emptyForFilter()))
              : ListView.separated(
                  padding: const EdgeInsets.all(16),
                  itemCount: filtered.length,
                  separatorBuilder: (_, __) => const SizedBox(height: 8),
                  itemBuilder: (context, i) {
                    final v = filtered[i];
                    final writable = widget.canWriteNotes(v);
                    final title = [
                      v['petName'],
                      v['clientName'],
                    ].whereType<String>().where((s) => s.isNotEmpty).join(' · ');
                    final when = _formatWhen(v);
                    final address = (v['addressText'] as String?) ?? '';
                    return Card(
                      child: Padding(
                        padding: const EdgeInsets.all(12),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            InkWell(
                              onTap: () => widget.onOpenPet(v),
                              child: Text(
                                title.isEmpty ? '—' : title,
                                style: Theme.of(context).textTheme.titleMedium,
                              ),
                            ),
                            Text(when, style: Theme.of(context).textTheme.bodySmall),
                            if (address.isNotEmpty) Text(address),
                            if (!writable)
                              Padding(
                                padding: const EdgeInsets.only(top: 4),
                                child: Text(
                                  widget.readOnlyLabel,
                                  style: Theme.of(context).textTheme.labelSmall,
                                ),
                              ),
                            const SizedBox(height: 8),
                            Wrap(
                              spacing: 8,
                              children: [
                                TextButton.icon(
                                  onPressed: () => widget.onOpenPet(v),
                                  icon: const Icon(Icons.pets, size: 18),
                                  label: Text(widget.petLabel),
                                ),
                                TextButton.icon(
                                  onPressed: () => widget.onMaps(v),
                                  icon: const Icon(Icons.map_outlined, size: 18),
                                  label: Text(widget.mapsLabel),
                                ),
                                TextButton.icon(
                                  onPressed: () => widget.onReport(v),
                                  icon: const Icon(Icons.note_alt_outlined, size: 18),
                                  label: Text(widget.reportLabel),
                                ),
                                if (writable && (v['status']?.toString() == 'confirmed'))
                                  TextButton.icon(
                                    onPressed: () => widget.onMarkDone(v),
                                    icon: const Icon(Icons.check_circle_outline, size: 18),
                                    label: Text(widget.doneLabel),
                                  ),
                                if (writable)
                                  TextButton.icon(
                                    onPressed: () async {
                                      final messenger = ScaffoldMessenger.of(context);
                                      final originalAddress = address.trim();
                                      final ctrl = TextEditingController(text: address);
                                      double? lat =
                                          (v['lat'] is num) ? (v['lat'] as num).toDouble() : null;
                                      double? lng =
                                          (v['lng'] is num) ? (v['lng'] as num).toDouble() : null;
                                      var coordsClearedByEdit = false;
                                      var capturing = false;
                                      final ok = await showDialog<bool>(
                                        context: context,
                                        builder: (ctx) {
                                          return StatefulBuilder(
                                            builder: (ctx, setDlg) {
                                              return AlertDialog(
                                                title: Text(widget.addressLabel),
                                                content: Column(
                                                  mainAxisSize: MainAxisSize.min,
                                                  children: [
                                                    TextField(
                                                      controller: ctrl,
                                                      onChanged: (_) {
                                                        if (lat != null || lng != null) {
                                                          setDlg(() {
                                                            lat = null;
                                                            lng = null;
                                                            coordsClearedByEdit = true;
                                                          });
                                                        }
                                                      },
                                                    ),
                                                    const SizedBox(height: 8),
                                                    TextButton.icon(
                                                      onPressed: capturing
                                                          ? null
                                                          : () async {
                                                              setDlg(() => capturing = true);
                                                              final pos = await _captureGps();
                                                              if (!ctx.mounted) return;
                                                              setDlg(() => capturing = false);
                                                              if (pos == null) {
                                                                messenger.showSnackBar(
                                                                  SnackBar(
                                                                    content: Text(widget.gpsDeniedLabel),
                                                                  ),
                                                                );
                                                                return;
                                                              }
                                                              setDlg(() {
                                                                lat = pos.latitude;
                                                                lng = pos.longitude;
                                                                coordsClearedByEdit = false;
                                                                if (ctrl.text.trim().isEmpty) {
                                                                  ctrl.text =
                                                                      '${pos.latitude.toStringAsFixed(5)}, ${pos.longitude.toStringAsFixed(5)}';
                                                                }
                                                              });
                                                            },
                                                      icon: capturing
                                                          ? const SizedBox(
                                                              width: 18,
                                                              height: 18,
                                                              child: CircularProgressIndicator(strokeWidth: 2),
                                                            )
                                                          : const Icon(Icons.my_location, size: 18),
                                                      label: Text(
                                                        lat != null ? '${widget.gpsLabel} ✓' : widget.gpsLabel,
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                                actions: [
                                                  TextButton(
                                                    onPressed: () => Navigator.pop(ctx, false),
                                                    child: Text(AppLocalizations.of(ctx)!.cancel),
                                                  ),
                                                  FilledButton(
                                                    onPressed: () => Navigator.pop(ctx, true),
                                                    child: Text(AppLocalizations.of(ctx)!.save),
                                                  ),
                                                ],
                                              );
                                            },
                                          );
                                        },
                                      );
                                      final text = ctrl.text.trim();
                                      ctrl.dispose();
                                      if (ok == true) {
                                        final clearCoords = coordsClearedByEdit &&
                                            lat == null &&
                                            lng == null &&
                                            text != originalAddress;
                                        await widget.onSaveLocation(
                                          v,
                                          text,
                                          lat: lat,
                                          lng: lng,
                                          clearCoords: clearCoords,
                                        );
                                      }
                                    },
                                    icon: const Icon(Icons.edit_location_alt_outlined, size: 18),
                                    label: Text(widget.addressLabel),
                                  ),
                              ],
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
        ),
      ],
    );
  }
}

Future<Position?> _captureGps() async {
  try {
    var permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
    }
    if (permission == LocationPermission.denied ||
        permission == LocationPermission.deniedForever) {
      return null;
    }
    final enabled = await Geolocator.isLocationServiceEnabled();
    if (!enabled) return null;
    return Geolocator.getCurrentPosition(
      locationSettings: const LocationSettings(
        accuracy: LocationAccuracy.high,
        timeLimit: Duration(seconds: 12),
      ),
    );
  } catch (_) {
    return null;
  }
}

class _ListTab extends StatelessWidget {
  const _ListTab({
    required this.empty,
    required this.items,
    required this.titleKey,
    required this.subtitleKey,
    required this.onTap,
  });

  final String empty;
  final List<Map<String, dynamic>> items;
  final String titleKey;
  final String subtitleKey;
  final void Function(Map<String, dynamic>) onTap;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return Center(child: Text(empty));
    }
    return ListView.separated(
      padding: const EdgeInsets.all(16),
      itemCount: items.length,
      separatorBuilder: (_, __) => const Divider(height: 1),
      itemBuilder: (context, i) {
        final row = items[i];
        return ListTile(
          title: Text('${row[titleKey] ?? ''}'),
          subtitle: Text('${row[subtitleKey] ?? ''}'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => onTap(row),
        );
      },
    );
  }
}
