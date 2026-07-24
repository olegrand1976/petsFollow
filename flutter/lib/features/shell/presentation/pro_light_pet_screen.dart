import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/care_reminder.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:url_launcher/url_launcher.dart';

/// Fiche animal pour care_pro (infos, timeline, documents).
class ProLightPetScreen extends StatefulWidget {
  const ProLightPetScreen({super.key, required this.petId, this.petName});

  final String petId;
  final String? petName;

  @override
  State<ProLightPetScreen> createState() => _ProLightPetScreenState();
}

class _ProLightPetScreenState extends State<ProLightPetScreen> {
  bool _loading = true;
  String? _error;
  Map<String, dynamic>? _pet;
  List<dynamic> _timeline = [];
  List<dynamic> _documents = [];
  List<CareReminder> _reminders = [];
  int _loadGen = 0;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final gen = ++_loadGen;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final pet = await ApiClient.instance.getPet(widget.petId);
      final timeline = await ApiClient.instance.getTimeline(widget.petId);
      final docs = await ApiClient.instance.listPetDocuments(widget.petId);
      final reminders = await ApiClient.instance.getCareReminders(widget.petId);
      if (!mounted || gen != _loadGen) return;
      setState(() {
        _pet = pet;
        _timeline = timeline;
        _documents = docs;
        _reminders = reminders;
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

  Future<void> _openDoc(Map<String, dynamic> doc) async {
    final url = (doc['fileUrl'] as String?)?.trim() ?? '';
    if (url.isEmpty) return;
    final uri = Uri.tryParse(url);
    if (uri != null) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final title = _pet?['name'] as String? ?? widget.petName ?? l10n.proLightPets;
    return Scaffold(
      appBar: AppBar(
        title: Text(title),
        actions: [
          IconButton(onPressed: _load, icon: const Icon(Icons.refresh)),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(child: Text(_error!, style: const TextStyle(color: AppColors.alert)))
              : RefreshIndicator(
                  onRefresh: _load,
                  child: ListView(
                    padding: const EdgeInsets.all(16),
                    children: [
                      if (_pet != null) ...[
                        Text(
                          [
                            _pet!['species'],
                            _pet!['breed'],
                          ].whereType<String>().where((s) => s.isNotEmpty).join(' · '),
                          style: Theme.of(context).textTheme.titleMedium,
                        ),
                        if ((_pet!['litterTag'] as String?)?.isNotEmpty == true)
                          Text('${l10n.proLightLitterTag}: ${_pet!['litterTag']}'),
                        const SizedBox(height: 16),
                      ],
                      Text(l10n.proLightReminders, style: Theme.of(context).textTheme.titleSmall),
                      const SizedBox(height: 8),
                      if (_reminders.isEmpty)
                        Text(l10n.proLightNoReminders, style: Theme.of(context).textTheme.bodySmall)
                      else
                        ..._reminders.take(8).map((r) {
                          return ListTile(
                            dense: true,
                            contentPadding: EdgeInsets.zero,
                            title: Text(r.title.isNotEmpty ? r.title : (r.type ?? '')),
                            subtitle: Text('${r.dueAt.toIso8601String()} · ${r.status}'),
                          );
                        }),
                      const SizedBox(height: 16),
                      Text(l10n.proLightDocuments, style: Theme.of(context).textTheme.titleSmall),
                      const SizedBox(height: 8),
                      if (_documents.isEmpty)
                        Text(l10n.proLightNoDocuments, style: Theme.of(context).textTheme.bodySmall)
                      else
                        ..._documents.map((raw) {
                          final d = Map<String, dynamic>.from(raw as Map);
                          final label = (d['title'] as String?)?.isNotEmpty == true
                              ? d['title'] as String
                              : (d['fileName'] as String? ?? '—');
                          return ListTile(
                            dense: true,
                            contentPadding: EdgeInsets.zero,
                            leading: const Icon(Icons.description_outlined),
                            title: Text(label),
                            trailing: const Icon(Icons.open_in_new, size: 18),
                            onTap: () => _openDoc(d),
                          );
                        }),
                      const SizedBox(height: 16),
                      Text(l10n.proLightTimeline, style: Theme.of(context).textTheme.titleSmall),
                      const SizedBox(height: 8),
                      if (_timeline.isEmpty)
                        Text(l10n.proLightNoTimeline, style: Theme.of(context).textTheme.bodySmall)
                      else
                        ..._timeline.take(20).map((raw) {
                          final t = Map<String, dynamic>.from(raw as Map);
                          final when = t['createdAt']?.toString() ?? '';
                          final label = t['title'] ?? t['type'] ?? '';
                          final body = (t['body'] as String?)?.trim() ?? '';
                          return ListTile(
                            dense: true,
                            contentPadding: EdgeInsets.zero,
                            title: Text('$label'),
                            subtitle: Text(body.isEmpty ? when : '$when · $body'),
                          );
                        }),
                    ],
                  ),
                ),
    );
  }
}
