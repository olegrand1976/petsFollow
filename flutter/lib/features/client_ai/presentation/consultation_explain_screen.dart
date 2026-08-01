import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/core/ui/load_error_view.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Owner-facing AI explanation of a finalized visit report (additive, not a diagnosis).
class ConsultationExplainScreen extends StatefulWidget {
  const ConsultationExplainScreen({
    super.key,
    required this.visitId,
    this.petName,
  });

  final String visitId;
  final String? petName;

  @override
  State<ConsultationExplainScreen> createState() =>
      _ConsultationExplainScreenState();
}

class _ConsultationExplainScreenState extends State<ConsultationExplainScreen> {
  bool loading = true;
  String? loadError;
  Map<String, dynamic>? data;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) _load();
    });
  }

  Future<void> _load() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() {
      loading = true;
      loadError = null;
    });
    try {
      final raw =
          await ApiClient.instance.getClientConsultationExplain(widget.visitId);
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

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final p = PetsPalette.of(context);
    final disclaimer = (data?['disclaimer'] as String?)?.trim().isNotEmpty == true
        ? data!['disclaimer'] as String
        : l10n.clientAiExplainDisclaimer;
    final cards = _cards(data);

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.clientAiExplainTitle),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 12),
            child: Chip(
              key: const Key('client_ai_explain_dev_badge'),
              label: Text(l10n.clientAiDevBadge),
              visualDensity: VisualDensity.compact,
              materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
            ),
          ),
        ],
      ),
      body: loading
          ? Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const CircularProgressIndicator(),
                  const SizedBox(height: 16),
                  Text(l10n.clientAiExplainLoading),
                ],
              ),
            )
          : loadError != null
              ? LoadErrorView(message: loadError!, onRetry: _load)
              : Column(
                  children: [
                    Expanded(
                      child: ListView(
                        padding: scrollPaddingWithSystemBottom(context, all: 16),
                        children: [
                          if ((widget.petName ?? '').trim().isNotEmpty)
                            Padding(
                              padding: const EdgeInsets.only(bottom: 12),
                              child: Text(
                                widget.petName!,
                                style: Theme.of(context).textTheme.titleMedium,
                              ),
                            ),
                          ...cards.map((card) {
                            final title = (card['title'] as String?)?.trim() ?? '';
                            final body = (card['body'] as String?)?.trim() ?? '';
                            final kind = (card['kind'] as String?)?.trim() ?? 'general';
                            return Card(
                              key: Key('client_ai_explain_card_$kind'),
                              margin: const EdgeInsets.only(bottom: 12),
                              child: Padding(
                                padding: const EdgeInsets.all(16),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Row(
                                      children: [
                                        Icon(
                                          _iconForKind(kind),
                                          color: AppColors.primary,
                                          size: 20,
                                        ),
                                        const SizedBox(width: 8),
                                        Expanded(
                                          child: Text(
                                            title,
                                            style: Theme.of(context)
                                                .textTheme
                                                .titleMedium,
                                          ),
                                        ),
                                      ],
                                    ),
                                    const SizedBox(height: 8),
                                    Text(body),
                                  ],
                                ),
                              ),
                            );
                          }),
                        ],
                      ),
                    ),
                    Material(
                      color: p.surfaceElevated,
                      child: SafeArea(
                        top: false,
                        child: Padding(
                          padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
                          child: Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Icon(Icons.info_outline,
                                  size: 18, color: p.textMuted),
                              const SizedBox(width: 8),
                              Expanded(
                                child: Text(
                                  disclaimer,
                                  key: const Key('client_ai_explain_disclaimer'),
                                  style: TextStyle(
                                    color: p.textMuted,
                                    fontSize: 12,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
    );
  }

  IconData _iconForKind(String kind) {
    switch (kind) {
      case 'term':
        return Icons.menu_book_outlined;
      case 'medication':
        return Icons.medication_outlined;
      default:
        return Icons.lightbulb_outline;
    }
  }

  List<Map<String, dynamic>> _cards(Map<String, dynamic>? raw) {
    final list = raw?['cards'];
    if (list is! List) return const [];
    return list
        .whereType<Map>()
        .map((e) => Map<String, dynamic>.from(e))
        .toList();
  }
}
