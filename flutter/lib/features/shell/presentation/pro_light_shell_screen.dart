import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';
import 'package:geolocator/geolocator.dart';
import 'package:path_provider/path_provider.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
import 'package:petsfollow_mobile/features/invite/presentation/app_invite_qr_screen.dart';
import 'package:petsfollow_mobile/features/profile/presentation/profile_screen.dart';
import 'package:petsfollow_mobile/features/shell/presentation/pro_light_pet_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:record/record.dart';
import 'package:url_launcher/url_launcher.dart';

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

  @override
  void initState() {
    super.initState();
    _load();
  }

  bool _canWriteNotes(Map<String, dynamic> row) {
    if (ApiClient.instance.userRole == 'vet') return true;
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

  void _openPet(String petId, {String? petName}) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => ProLightPetScreen(petId: petId, petName: petName),
      ),
    );
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
                      return ListTile(
                        title: Text('${p['name'] ?? ''}'),
                        subtitle: Text('${p['species'] ?? ''}'),
                        trailing: const Icon(Icons.chevron_right),
                        onTap: petId.isEmpty
                            ? null
                            : () => Navigator.of(ctx).push(
                                  MaterialPageRoute(
                                    builder: (_) => ProLightPetScreen(
                                      petId: petId,
                                      petName: p['name'] as String?,
                                    ),
                                  ),
                                ),
                      );
                    },
                  ),
          );
        },
      ),
    );
  }

  Future<void> _openReport(Map<String, dynamic> visit) async {
    final visitId = visit['id'] as String?;
    if (visitId == null) return;
    final canWrite = _canWriteNotes(visit);
    final l10n = AppLocalizations.of(context)!;
    String initialText = '';
    var status = 'none';
    try {
      final report = await ApiClient.instance.getVisitReport(visitId);
      initialText = (report['bodyText'] as String?) ??
          (report['transcriptText'] as String?) ??
          '';
      status = report['status'] as String? ?? 'draft';
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.proLightActionFailed)),
      );
      return;
    }
    if (!mounted) return;
    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (ctx) => _VisitReportSheet(
        visitId: visitId,
        initialText: initialText,
        initialStatus: status,
        canWrite: canWrite,
      ),
    );
  }

  void _toast(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(message)));
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final specialty = ApiClient.instance.userSpecialty ?? '';
    final specialtyLabel = proLightSpecialtyLabel(l10n, specialty);
    final isVet = ApiClient.instance.userRole == 'vet';
    final title = isVet
        ? l10n.proLightVetTitle
        : (specialtyLabel.isEmpty
            ? l10n.proLightTitle
            : '${l10n.proLightTitle} · $specialtyLabel');
    return Container(
      decoration: const BoxDecoration(gradient: AppTheme.loginGradient),
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          title: Row(
            children: [
              const PetsLogo(height: 28),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  title,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
          actions: [
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
                          _openPet(id, petName: row['name'] as String?);
                        },
                      ),
                      _SettingsTab(onLogout: widget.onLogout),
                    ],
                  ),
        bottomNavigationBar: NavigationBar(
          key: const Key('pro_light_nav'),
          selectedIndex: _index,
          onDestinationSelected: (i) => setState(() => _index = i),
          destinations: [
            NavigationDestination(icon: const Icon(Icons.event), label: l10n.proLightAgenda),
            NavigationDestination(icon: const Icon(Icons.people), label: l10n.proLightClients),
            NavigationDestination(icon: const Icon(Icons.pets), label: l10n.proLightPets),
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
    required this.initialStatus,
    required this.canWrite,
  });

  final String visitId;
  final String initialText;
  final String initialStatus;
  final bool canWrite;

  @override
  State<_VisitReportSheet> createState() => _VisitReportSheetState();
}

class _VisitReportSheetState extends State<_VisitReportSheet> {
  late final TextEditingController _controller;
  late String _status;
  bool _busy = false;
  bool _recording = false;
  final AudioRecorder _recorder = AudioRecorder();

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(text: widget.initialText);
    _status = widget.initialStatus;
  }

  @override
  void dispose() {
    _controller.dispose();
    _recorder.dispose();
    super.dispose();
  }

  String get _specialty => ApiClient.instance.userSpecialty ?? '';

  Future<bool> _confirmAudioConsent() async {
    final l10n = AppLocalizations.of(context)!;
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.proLightAudioConsentTitle),
        content: Text(l10n.proLightAudioConsentBody),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text(MaterialLocalizations.of(ctx).cancelButtonLabel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: Text(l10n.proLightAudioConsentAccept),
          ),
        ],
      ),
    );
    return ok == true;
  }

  Future<void> _applyTranscript(Map<String, dynamic> transcribed) async {
    final text = (transcribed['bodyText'] as String?) ??
        (transcribed['transcriptText'] as String?) ??
        '';
    if (text.isNotEmpty) {
      _controller.text = text;
    }
  }

  Future<void> _toggleDictation() async {
    if (_recording) {
      final path = await _recorder.stop();
      setState(() => _recording = false);
      if (path == null || path.isEmpty) return;
      final transcribed = await ApiClient.instance.transcribeVisitReport(
        widget.visitId,
        path,
        filename: 'dictation.m4a',
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
    setState(() => _recording = true);
  }

  Future<void> _run(Future<void> Function() action, {bool popOnOk = false}) async {
    setState(() => _busy = true);
    try {
      await action();
      if (!mounted) return;
      if (popOnOk) Navigator.pop(context);
    } catch (e) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      final msg = e is StateError && e.message == 'mic_denied'
          ? l10n.proLightMicDenied
          : l10n.proLightActionFailed;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final isFinal = _status == 'final';
    final readOnly = !widget.canWrite || isFinal || _busy || _status == 'none' && !widget.canWrite;
    final hint =
        _specialty == 'farrier' ? l10n.proLightReportHintFarrier : l10n.proLightReportHint;
    return Padding(
      padding: EdgeInsets.only(
        left: 16,
        right: 16,
        top: 16,
        bottom: MediaQuery.of(context).viewInsets.bottom + 16,
      ),
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
          const SizedBox(height: 12),
          if (widget.canWrite && !isFinal)
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                OutlinedButton(
                  onPressed: _busy
                      ? null
                      : () => _run(() async {
                            await ApiClient.instance
                                .putVisitReport(widget.visitId, _controller.text);
                          }, popOnOk: true),
                  child: Text(l10n.save),
                ),
                FilledButton(
                  onPressed: _busy
                      ? null
                      : () => _run(() async {
                            await ApiClient.instance
                                .putVisitReport(widget.visitId, _controller.text);
                            final improved = await ApiClient.instance
                                .improveVisitReport(widget.visitId);
                            if (!mounted) return;
                            _controller.text =
                                (improved['bodyText'] as String?) ?? _controller.text;
                          }),
                  child: Text(l10n.proLightImproveAi),
                ),
                FilledButton.tonal(
                  onPressed: _busy
                      ? null
                      : () => _run(() async {
                            await ApiClient.instance
                                .putVisitReport(widget.visitId, _controller.text);
                            final finalized = await ApiClient.instance
                                .finalizeVisitReport(widget.visitId);
                            if (!mounted) return;
                            setState(() {
                              _status = finalized['status'] as String? ?? 'final';
                            });
                          }, popOnOk: true),
                  child: Text(l10n.proLightFinalizeReport),
                ),
                FilledButton.tonal(
                  onPressed: _busy
                      ? null
                      : () => _run(() async {
                            await _toggleDictation();
                          }),
                  child: Text(
                    _recording ? l10n.proLightDictationStop : l10n.proLightDictationStart,
                  ),
                ),
                OutlinedButton(
                  onPressed: _busy || _recording
                      ? null
                      : () => _run(() async {
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
                            );
                            if (!mounted) return;
                            await _applyTranscript(transcribed);
                          }),
                  child: Text(l10n.proLightTranscribeAudio),
                ),
              ],
            ),
        ],
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

  /// Aligné Nuxt `visitDisplayAt` / modèle `Visit.displayDate` (sans fallback createdAt).
  DateTime? _parseDisplayAt(Map<String, dynamic> v) {
    final proposed = v['proposedScheduledAt']?.toString();
    if (proposed != null && proposed.isNotEmpty) {
      final at = DateTime.tryParse(proposed)?.toLocal();
      if (at != null) return at;
    }
    final raw = v['scheduledAt']?.toString();
    if (raw == null || raw.isEmpty) return null;
    return DateTime.tryParse(raw)?.toLocal();
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
