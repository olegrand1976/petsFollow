import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/models/manager_overview.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/core/ui/load_error_view.dart';
import 'package:petsfollow_mobile/features/commercial/presentation/manager_member_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Team KPIs + member list for [commercial_manager].
class ManagerTeamResultsScreen extends StatefulWidget {
  const ManagerTeamResultsScreen({super.key});

  @override
  State<ManagerTeamResultsScreen> createState() =>
      _ManagerTeamResultsScreenState();
}

class _ManagerTeamResultsScreenState extends State<ManagerTeamResultsScreen> {
  ManagerOverview? _overview;
  String? _error;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final ov = await ApiClient.instance.getCommercialManagerOverview();
      if (!mounted) return;
      setState(() {
        _overview = ov;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      setState(() {
        _error = mapApiError(e, l10n);
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final p = PetsPalette.of(context);
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.managerTeamTitle),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? LoadErrorView(message: _error!, onRetry: _load)
              : RefreshIndicator(
                  onRefresh: _load,
                  child: ListView(
                    padding: const EdgeInsets.all(16),
                    children: [
                      Text(
                        l10n.managerTeamSection,
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      const SizedBox(height: 12),
                      _KpiGrid(
                        items: [
                          _Kpi(l10n.managerKpiProspects, '${_overview!.teamProspectsTotal}'),
                          _Kpi(l10n.managerKpiConverted, '${_overview!.teamProspectsConverted}'),
                          _Kpi(l10n.managerKpiConversion, _overview!.conversionRateLabel),
                          _Kpi(l10n.managerKpiAppointments, '${_overview!.teamAppointmentsUpcoming}'),
                          _Kpi(l10n.managerKpiStale, '${_overview!.teamStaleInPipeline}'),
                          _Kpi(
                            l10n.managerKpiMonthEarned,
                            formatEuroCents(_overview!.teamMonthEarnedCents),
                          ),
                        ],
                      ),
                      const SizedBox(height: 24),
                      Text(
                        l10n.managerSelfSection,
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      const SizedBox(height: 12),
                      _KpiGrid(
                        items: [
                          _Kpi(l10n.managerKpiVets, '${_overview!.self.assignedVets}'),
                          _Kpi(l10n.managerKpiProspects, '${_overview!.self.prospectsTotal}'),
                          _Kpi(l10n.managerKpiConverted, '${_overview!.self.prospectsConverted}'),
                          _Kpi(
                            l10n.managerKpiMonthEarned,
                            formatEuroCents(_overview!.self.monthEarnedCents),
                          ),
                        ],
                      ),
                      const SizedBox(height: 24),
                      Text(
                        l10n.managerMembersSection,
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      const SizedBox(height: 8),
                      if (_overview!.team.isEmpty)
                        Padding(
                          padding: const EdgeInsets.symmetric(vertical: 24),
                          child: Text(
                            l10n.managerTeamEmpty,
                            textAlign: TextAlign.center,
                            style: TextStyle(color: p.textMuted),
                          ),
                        )
                      else
                        ..._overview!.team.map((m) {
                          return Card(
                            key: Key('manager_team_row_${m.userId}'),
                            margin: const EdgeInsets.only(bottom: 8),
                            child: ListTile(
                              title: Text(m.fullName),
                              subtitle: Text(
                                '${m.email}\n'
                                '${l10n.managerKpiConverted}: ${m.prospectsConverted} · '
                                '${l10n.managerKpiStale}: ${m.staleInPipeline}',
                              ),
                              isThreeLine: true,
                              trailing: Text(
                                formatEuroCents(m.monthEarnedCents),
                                style: Theme.of(context).textTheme.titleSmall,
                              ),
                              onTap: () {
                                Navigator.of(context).push(
                                  MaterialPageRoute<void>(
                                    builder: (_) => ManagerMemberScreen(
                                      memberUserId: m.userId,
                                      memberName: m.fullName,
                                    ),
                                  ),
                                );
                              },
                            ),
                          );
                        }),
                    ],
                  ),
                ),
    );
  }
}

class _Kpi {
  const _Kpi(this.label, this.value);
  final String label;
  final String value;
}

class _KpiGrid extends StatelessWidget {
  const _KpiGrid({required this.items});
  final List<_Kpi> items;

  @override
  Widget build(BuildContext context) {
    final p = PetsPalette.of(context);
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: items.map((k) {
        return SizedBox(
          width: (MediaQuery.sizeOf(context).width - 40) / 2,
          child: Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: p.surfaceElevated,
              borderRadius: BorderRadius.circular(12),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  k.value,
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                const SizedBox(height: 4),
                Text(
                  k.label,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: p.textMuted,
                      ),
                ),
              ],
            ),
          ),
        );
      }).toList(),
    );
  }
}
