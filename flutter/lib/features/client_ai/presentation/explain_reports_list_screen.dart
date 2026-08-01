import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/core/models/visit.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/core/ui/load_error_view.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/features/client_ai/presentation/consultation_explain_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class _ExplainVisitRow {
  const _ExplainVisitRow({
    required this.visit,
    required this.petName,
  });

  final Visit visit;
  final String petName;
}

/// Lists finalized visit reports the owner can open in [ConsultationExplainScreen].
class ExplainReportsListScreen extends StatefulWidget {
  const ExplainReportsListScreen({super.key});

  @override
  State<ExplainReportsListScreen> createState() =>
      _ExplainReportsListScreenState();
}

class _ExplainReportsListScreenState extends State<ExplainReportsListScreen> {
  bool loading = true;
  String? loadError;
  List<_ExplainVisitRow> rows = const [];

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
      final pets = await ApiClient.instance.getPets();
      final petList = pets
          .whereType<Map>()
          .map((e) => Pet.fromJson(Map<String, dynamic>.from(e)))
          .where((p) => p.id.isNotEmpty)
          .toList();
      final collected = <_ExplainVisitRow>[];
      for (final pet in petList) {
        final visits = await ApiClient.instance.getVisits(pet.id);
        for (final v in visits) {
          if (!v.consultationAvailable) continue;
          collected.add(_ExplainVisitRow(visit: v, petName: pet.name));
        }
      }
      collected.sort(
        (a, b) => b.visit.displayDate.compareTo(a.visit.displayDate),
      );
      if (!mounted) return;
      setState(() {
        rows = collected;
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

  void _openExplain(_ExplainVisitRow row) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => ConsultationExplainScreen(
          visitId: row.visit.id,
          petName: row.petName.isEmpty ? null : row.petName,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final p = PetsPalette.of(context);
    final dateFmt =
        DateFormat.yMMMd(Localizations.localeOf(context).toString()).add_Hm();

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.clientAiExplainListTitle),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 12),
            child: Chip(
              key: const Key('client_ai_explain_list_dev_badge'),
              label: Text(l10n.clientAiDevBadge),
              visualDensity: VisualDensity.compact,
            ),
          ),
        ],
      ),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : loadError != null
              ? LoadErrorView(message: loadError!, onRetry: load)
              : RefreshIndicator(
                  onRefresh: load,
                  child: rows.isEmpty
                      ? ListView(
                          physics: const AlwaysScrollableScrollPhysics(),
                          padding: scrollPaddingWithSystemBottom(context, all: 24),
                          children: [
                            const SizedBox(height: 48),
                            Icon(
                              Icons.auto_awesome_outlined,
                              size: 48,
                              color: p.textMuted,
                            ),
                            const SizedBox(height: 16),
                            Text(
                              l10n.clientAiExplainListEmpty,
                              textAlign: TextAlign.center,
                              style: TextStyle(color: p.textMuted),
                            ),
                          ],
                        )
                      : ListView.separated(
                          padding: scrollPaddingWithSystemBottom(context, all: 16),
                          itemCount: rows.length,
                          separatorBuilder: (_, __) => const SizedBox(height: 8),
                          itemBuilder: (context, i) {
                            final row = rows[i];
                            return Card(
                              child: ListTile(
                                key: Key(
                                  'client_ai_explain_list_item_${row.visit.id}',
                                ),
                                leading: CircleAvatar(
                                  backgroundColor:
                                      AppColors.primary.withValues(alpha: 0.15),
                                  child: Icon(
                                    Icons.auto_awesome_outlined,
                                    color: AppColors.primary,
                                    size: 20,
                                  ),
                                ),
                                title: Text(
                                  row.petName.isEmpty
                                      ? l10n.consultationTitle
                                      : row.petName,
                                ),
                                subtitle: Text(
                                  dateFmt.format(row.visit.displayDate.toLocal()),
                                ),
                                trailing: const Icon(Icons.chevron_right),
                                onTap: () => _openExplain(row),
                              ),
                            );
                          },
                        ),
                ),
    );
  }
}
