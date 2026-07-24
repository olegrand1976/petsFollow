import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';
import 'package:geolocator/geolocator.dart';
import 'package:path_provider/path_provider.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
import 'package:petsfollow_mobile/features/profile/presentation/profile_screen.dart';
import 'package:petsfollow_mobile/features/shell/presentation/pro_light_pet_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:record/record.dart';
import 'package:url_launcher/url_launcher.dart';

/// Terrain shell for care_pro (vet_light, farrier, …).
class ProLightShellScreen extends StatefulWidget {
  const ProLightShellScreen({super.key, required this.onLogout});

  final VoidCallback onLogout;

  @override
  State<ProLightShellScreen> createState() => _ProLightShellScreenState();
}

String _specialtyLabel(AppLocalizations l10n, String specialty) {
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
    final p = row['permission'] as String? ?? 'read';
    return p == 'write_notes' || p == 'full';
  }

  Future<void> _load() async {
    final gen = ++_loadGen;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final visits = await ApiClient.instance.listCareProVisits();
      final clients = await ApiClient.instance.listCareProClients();
      final pets = await ApiClient.instance.listCareProPets();
      if (!mounted || gen != _loadGen) return;
      setState(() {
        _visits = visits;
        _clients = clients;
        _pets = pets;
        _loading = false;
      });
    } catch (_) {
      if (!mounted || gen != _loadGen) return;
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
    final specialtyLabel = _specialtyLabel(l10n, specialty);
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
                  specialtyLabel.isEmpty
                      ? l10n.proLightTitle
                      : '${l10n.proLightTitle} · $specialtyLabel',
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
          actions: [
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
                        emptyLabel: specialty == 'farrier'
                            ? l10n.proLightEmptyFarrier
                            : l10n.proLightNoVisits,
                        onMaps: _openMaps,
                        onReport: _openReport,
                        canWriteNotes: _canWriteNotes,
                        onOpenPet: (visit) {
                          final petId = visit['petId'] as String?;
                          if (petId == null || petId.isEmpty) return;
                          _openPet(petId, petName: visit['petName'] as String?);
                        },
                        onSaveLocation: (visit, address, {lat, lng}) async {
                          final id = visit['id'] as String?;
                          if (id == null) return;
                          try {
                            await ApiClient.instance.updateVisitLocation(
                              id,
                              address,
                              lat: lat,
                              lng: lng,
                            );
                            await _load();
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
          subtitle: Text(specialty.isEmpty ? '—' : _specialtyLabel(l10n, specialty)),
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

class _VisitsTab extends StatelessWidget {
  const _VisitsTab({
    required this.visits,
    required this.emptyLabel,
    required this.onMaps,
    required this.onReport,
    required this.canWriteNotes,
    required this.onOpenPet,
    required this.onSaveLocation,
    required this.addressLabel,
    required this.mapsLabel,
    required this.reportLabel,
    required this.petLabel,
    required this.readOnlyLabel,
    required this.gpsLabel,
    required this.gpsDeniedLabel,
  });

  final List<dynamic> visits;
  final String emptyLabel;
  final Future<void> Function(Map<String, dynamic>) onMaps;
  final Future<void> Function(Map<String, dynamic>) onReport;
  final bool Function(Map<String, dynamic>) canWriteNotes;
  final void Function(Map<String, dynamic>) onOpenPet;
  final Future<void> Function(
    Map<String, dynamic>,
    String, {
    double? lat,
    double? lng,
  }) onSaveLocation;
  final String addressLabel;
  final String mapsLabel;
  final String reportLabel;
  final String petLabel;
  final String readOnlyLabel;
  final String gpsLabel;
  final String gpsDeniedLabel;

  @override
  Widget build(BuildContext context) {
    if (visits.isEmpty) {
      return Center(child: Text(emptyLabel));
    }
    return ListView.separated(
      padding: const EdgeInsets.all(16),
      itemCount: visits.length,
      separatorBuilder: (_, __) => const SizedBox(height: 8),
      itemBuilder: (context, i) {
        final v = Map<String, dynamic>.from(visits[i] as Map);
        final writable = canWriteNotes(v);
        final title = [
          v['petName'],
          v['clientName'],
        ].whereType<String>().where((s) => s.isNotEmpty).join(' · ');
        final when = v['scheduledAt']?.toString() ?? v['status']?.toString() ?? '';
        final address = (v['addressText'] as String?) ?? '';
        return Card(
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                InkWell(
                  onTap: () => onOpenPet(v),
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
                      readOnlyLabel,
                      style: Theme.of(context).textTheme.labelSmall,
                    ),
                  ),
                const SizedBox(height: 8),
                Wrap(
                  spacing: 8,
                  children: [
                    TextButton.icon(
                      onPressed: () => onOpenPet(v),
                      icon: const Icon(Icons.pets, size: 18),
                      label: Text(petLabel),
                    ),
                    TextButton.icon(
                      onPressed: () => onMaps(v),
                      icon: const Icon(Icons.map_outlined, size: 18),
                      label: Text(mapsLabel),
                    ),
                    TextButton.icon(
                      onPressed: () => onReport(v),
                      icon: const Icon(Icons.note_alt_outlined, size: 18),
                      label: Text(reportLabel),
                    ),
                    if (writable)
                      TextButton.icon(
                        onPressed: () async {
                          final ctrl = TextEditingController(text: address);
                          double? lat;
                          double? lng;
                          final ok = await showDialog<bool>(
                            context: context,
                            builder: (ctx) {
                              return StatefulBuilder(
                                builder: (ctx, setDlg) {
                                  return AlertDialog(
                                    title: Text(addressLabel),
                                    content: Column(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        TextField(controller: ctrl),
                                        const SizedBox(height: 8),
                                        TextButton.icon(
                                          onPressed: () async {
                                            final pos = await _captureGps(ctx, gpsDeniedLabel);
                                            if (pos == null) return;
                                            setDlg(() {
                                              lat = pos.latitude;
                                              lng = pos.longitude;
                                              if (ctrl.text.trim().isEmpty) {
                                                ctrl.text =
                                                    '${pos.latitude.toStringAsFixed(5)}, ${pos.longitude.toStringAsFixed(5)}';
                                              }
                                            });
                                          },
                                          icon: const Icon(Icons.my_location, size: 18),
                                          label: Text(
                                            lat != null ? '$gpsLabel ✓' : gpsLabel,
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
                            await onSaveLocation(v, text, lat: lat, lng: lng);
                          }
                        },
                        icon: const Icon(Icons.edit_location_alt_outlined, size: 18),
                        label: Text(addressLabel),
                      ),
                  ],
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

Future<Position?> _captureGps(BuildContext context, String deniedLabel) async {
  try {
    var permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
    }
    if (permission == LocationPermission.denied ||
        permission == LocationPermission.deniedForever) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(deniedLabel)));
      }
      return null;
    }
    final enabled = await Geolocator.isLocationServiceEnabled();
    if (!enabled) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(deniedLabel)));
      }
      return null;
    }
    return Geolocator.getCurrentPosition(locationSettings: const LocationSettings(
      accuracy: LocationAccuracy.high,
      timeLimit: Duration(seconds: 12),
    ));
  } catch (_) {
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(deniedLabel)));
    }
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
