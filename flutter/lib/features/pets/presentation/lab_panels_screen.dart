import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/api/open_url.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Read-only lab panels list + detail for the pet owner.
class LabPanelsScreen extends StatefulWidget {
  const LabPanelsScreen({
    super.key,
    required this.petId,
    this.initialPanelId,
  });

  final String petId;
  /// When set (e.g. from timeline), open this panel after the list loads.
  final String? initialPanelId;

  @override
  State<LabPanelsScreen> createState() => _LabPanelsScreenState();
}

class _LabPanelsScreenState extends State<LabPanelsScreen> {
  bool loading = true;
  String? error;
  List<Map<String, dynamic>> panels = [];
  Map<String, dynamic>? detail;
  String? detailDocumentUrl;
  String? detailDocumentLabel;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      loading = true;
      error = null;
    });
    try {
      final raw = await ApiClient.instance.getLabPanels(widget.petId);
      if (!mounted) return;
      setState(() {
        panels = raw
            .whereType<Map>()
            .map((e) => Map<String, dynamic>.from(e))
            .toList();
        loading = false;
      });
      final initial = widget.initialPanelId?.trim();
      if (initial != null && initial.isNotEmpty) {
        await _open(initial);
      }
    } catch (e) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      setState(() {
        loading = false;
        error = mapApiError(e, l10n);
      });
    }
  }

  String _flagLabel(AppLocalizations l10n, String flag) {
    switch (flag) {
      case 'low':
        return l10n.labsFlagLow;
      case 'high':
        return l10n.labsFlagHigh;
      case 'normal':
        return l10n.labsFlagNormal;
      default:
        return '';
    }
  }

  Future<void> _open(String panelId) async {
    try {
      final data = await ApiClient.instance.getLabPanel(widget.petId, panelId);
      String? docUrl;
      String? docLabel;
      final docId = data['documentId']?.toString();
      if (docId != null && docId.isNotEmpty) {
        final docs = await ApiClient.instance.listPetDocuments(widget.petId);
        for (final raw in docs) {
          if (raw is! Map) continue;
          if ('${raw['id']}' != docId) continue;
          docUrl = (raw['fileUrl'] as String?)?.trim();
          final title = (raw['title'] as String?)?.trim();
          final fileName = (raw['fileName'] as String?)?.trim();
          docLabel = (title != null && title.isNotEmpty)
              ? title
              : (fileName ?? docId);
          break;
        }
      }
      if (!mounted) return;
      setState(() {
        detail = data;
        detailDocumentUrl = docUrl;
        detailDocumentLabel = docLabel;
      });
    } catch (e) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(mapApiError(e, l10n))),
      );
    }
  }

  Future<void> _openDocument() async {
    final url = detailDocumentUrl?.trim() ?? '';
    if (url.isEmpty) return;
    final l10n = AppLocalizations.of(context)!;
    final opened = await openExternalUrl(url);
    if (!opened && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.errorCouldNotOpenLink)),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.labsTitle)),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : error != null
              ? Center(child: Text(error!))
              : panels.isEmpty
                  ? Center(child: Text(l10n.labsEmpty))
                  : ListView(
                      padding: const EdgeInsets.all(16),
                      children: [
                        for (final p in panels)
                          Card(
                            key: Key('lab_panel_${p['id']}'),
                            child: ListTile(
                              title: Text(
                                (p['labName'] as String?)?.trim().isNotEmpty ==
                                        true
                                    ? p['labName'] as String
                                    : l10n.labsTitle,
                              ),
                              subtitle: Text(
                                [
                                  if (p['collectedAt'] != null)
                                    '${p['collectedAt']}'.split('T').first,
                                  if ((p['abnormalCount'] as num?) != null &&
                                      (p['abnormalCount'] as num) > 0)
                                    l10n.labsAbnormalCount(
                                      (p['abnormalCount'] as num).toInt(),
                                    ),
                                ].join(' · '),
                              ),
                              trailing: const Icon(Icons.chevron_right),
                              onTap: () => _open('${p['id']}'),
                            ),
                          ),
                        if (detail != null) ...[
                          const SizedBox(height: 16),
                          Text(
                            l10n.labsResults,
                            style: Theme.of(context).textTheme.titleMedium,
                          ),
                          if (detailDocumentUrl != null &&
                              detailDocumentUrl!.isNotEmpty) ...[
                            const SizedBox(height: 8),
                            OutlinedButton.icon(
                              key: const Key('lab_panel_document_open'),
                              onPressed: _openDocument,
                              icon: const Icon(Icons.picture_as_pdf_outlined),
                              label: Text(
                                detailDocumentLabel?.isNotEmpty == true
                                    ? detailDocumentLabel!
                                    : l10n.labsResults,
                              ),
                            ),
                          ],
                          for (final r in (detail!['results'] as List? ?? []))
                            if (r is Map)
                              ListTile(
                                dense: true,
                                title: Text('${r['analyteCode']}'.toUpperCase()),
                                subtitle: Text(
                                  [
                                    if (r['valueNum'] != null ||
                                        r['valueText'] != null)
                                      '${r['valueNum'] ?? r['valueText']}${r['unit'] != null && '${r['unit']}'.isNotEmpty ? ' ${r['unit']}' : ''}',
                                    _flagLabel(l10n, '${r['flag'] ?? ''}'),
                                  ].where((s) => s.isNotEmpty).join(' · '),
                                ),
                              ),
                        ],
                      ],
                    ),
    );
  }
}
