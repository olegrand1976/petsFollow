import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/core/ui/load_error_view.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class ConsultationViewScreen extends StatefulWidget {
  const ConsultationViewScreen({
    super.key,
    required this.visitId,
    this.petName,
  });

  final String visitId;
  final String? petName;

  @override
  State<ConsultationViewScreen> createState() => _ConsultationViewScreenState();
}

class _ConsultationViewScreenState extends State<ConsultationViewScreen> {
  bool loading = true;
  String? loadError;
  Map<String, dynamic>? data;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) load();
    });
  }

  Future<void> load() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() {
      loading = true;
      loadError = null;
    });
    try {
      final raw = await ApiClient.instance.getClientConsultation(widget.visitId);
      if (!mounted) return;
      setState(() {
        data = raw;
        loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        loading = false;
        loadError = mapApiError(e, l10n);
      });
    }
  }

  Future<void> _share() async {
    final l10n = AppLocalizations.of(context)!;
    final email = await showDialog<String>(
      context: context,
      builder: (ctx) => _SendConsultationEmailDialog(l10n: l10n),
    );
    if (email == null || email.isEmpty || !mounted) return;
    if (!email.contains('@')) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.sendConsultationInvalidEmail)),
      );
      return;
    }
    final messenger = ScaffoldMessenger.of(context);
    try {
      await ApiClient.instance.sendConsultationShare(widget.visitId, email);
      if (!mounted) return;
      messenger.showSnackBar(SnackBar(content: Text(l10n.sendConsultationSuccess)));
    } catch (e) {
      if (!mounted) return;
      messenger.showSnackBar(SnackBar(content: Text(mapApiError(e, l10n))));
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final p = PetsPalette.of(context);
    final dateFmt = DateFormat.yMMMd(Localizations.localeOf(context).toString()).add_Hm();
    final titleName = (data?['petName'] as String?)?.trim().isNotEmpty == true
        ? data!['petName'] as String
        : (widget.petName ?? '');
    final scheduledRaw = data?['scheduledAt'] as String?;
    final scheduledAt = scheduledRaw != null && scheduledRaw.isNotEmpty
        ? DateTime.tryParse(scheduledRaw)
        : null;
    final statusLabel = (data?['status'] as String?)?.trim() ?? '';
    final metaSubtitle = [
      if (scheduledAt != null) dateFmt.format(scheduledAt.toLocal()),
      if (statusLabel.isNotEmpty) statusLabel,
    ].join(' · ');

    return Scaffold(
      appBar: AppBar(
        title: Text(
          titleName.isEmpty
              ? l10n.consultationTitle
              : l10n.consultationTitleWithPet(titleName),
        ),
        actions: [
          if (data != null)
            IconButton(
              key: Key('consultation_share_btn_${widget.visitId}'),
              tooltip: l10n.sendConsultationToVet,
              onPressed: _share,
              icon: const Icon(Icons.share_outlined),
            ),
        ],
      ),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : loadError != null
              ? LoadErrorView(message: loadError!, onRetry: load)
              : RefreshIndicator(
                  onRefresh: load,
                  child: ListView(
                    padding: scrollPaddingWithSystemBottom(context, all: 16),
                    children: [
                      if ((data?['practiceName'] as String?)?.isNotEmpty == true ||
                          (data?['scheduledAt'] as String?)?.isNotEmpty == true)
                        Card(
                          child: ListTile(
                            leading: Icon(Icons.event, color: AppColors.primary),
                            title: Text(
                              (data?['practiceName'] as String?)?.isNotEmpty == true
                                  ? data!['practiceName'] as String
                                  : l10n.consultationVisitMeta,
                            ),
                            subtitle: metaSubtitle.isEmpty ? null : Text(metaSubtitle),
                          ),
                        ),
                      const SizedBox(height: 8),
                      ..._reports(data).map((r) {
                        final author = (r['authorName'] as String?)?.trim() ?? '';
                        final body = (r['bodyText'] as String?)?.trim() ?? '';
                        final finalized = r['finalizedAt'] != null
                            ? DateTime.tryParse(r['finalizedAt'].toString())
                            : null;
                        return Card(
                          margin: const EdgeInsets.only(bottom: 12),
                          child: Padding(
                            padding: const EdgeInsets.all(16),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  author.isEmpty
                                      ? l10n.consultationReportFallback
                                      : l10n.consultationReportBy(author),
                                  style: Theme.of(context).textTheme.titleMedium,
                                ),
                                if (finalized != null) ...[
                                  const SizedBox(height: 4),
                                  Text(
                                    dateFmt.format(finalized.toLocal()),
                                    style: TextStyle(color: p.textMuted, fontSize: 12),
                                  ),
                                ],
                                const SizedBox(height: 12),
                                SelectableText(
                                  body.isEmpty ? l10n.consultationReportEmpty : body,
                                  style: Theme.of(context).textTheme.bodyMedium,
                                ),
                              ],
                            ),
                          ),
                        );
                      }),
                      const SizedBox(height: 8),
                      FilledButton.icon(
                        key: Key('consultation_share_cta_${widget.visitId}'),
                        onPressed: _share,
                        icon: const Icon(Icons.mail_outline),
                        label: Text(l10n.sendConsultationToVet),
                      ),
                    ],
                  ),
                ),
    );
  }

  List<Map<String, dynamic>> _reports(Map<String, dynamic>? raw) {
    final list = raw?['reports'];
    if (list is! List) return const [];
    return list
        .whereType<Map>()
        .map((e) => Map<String, dynamic>.from(e))
        .toList();
  }
}

class _SendConsultationEmailDialog extends StatefulWidget {
  const _SendConsultationEmailDialog({required this.l10n});

  final AppLocalizations l10n;

  @override
  State<_SendConsultationEmailDialog> createState() => _SendConsultationEmailDialogState();
}

class _SendConsultationEmailDialogState extends State<_SendConsultationEmailDialog> {
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
      key: const Key('consultation_share_dialog'),
      scrollable: true,
      title: Text(l10n.sendConsultationToVet),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            l10n.sendConsultationPhiWarning,
            style: Theme.of(context).textTheme.bodySmall,
          ),
          const SizedBox(height: 12),
          TextField(
            key: const Key('consultation_share_email'),
            controller: _controller,
            keyboardType: TextInputType.emailAddress,
            autofocus: true,
            decoration: InputDecoration(
              labelText: l10n.sendConsultationEmailLabel,
              hintText: l10n.sendConsultationEmailHint,
            ),
          ),
          CheckboxListTile(
            key: const Key('consultation_share_consent'),
            value: _consented,
            onChanged: (v) => setState(() => _consented = v ?? false),
            controlAffinity: ListTileControlAffinity.leading,
            contentPadding: EdgeInsets.zero,
            title: Text(
              l10n.sendConsultationPhiConsent,
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
          key: const Key('consultation_share_confirm'),
          onPressed: _consented
              ? () => Navigator.pop(context, _controller.text.trim())
              : null,
          child: Text(l10n.sendConsultationConfirm),
        ),
      ],
    );
  }
}
