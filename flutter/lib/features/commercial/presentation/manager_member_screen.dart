import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/models/manager_overview.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/core/ui/load_error_view.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Personal KPIs for one commercial in the manager's team.
class ManagerMemberScreen extends StatefulWidget {
  const ManagerMemberScreen({
    super.key,
    required this.memberUserId,
    required this.memberName,
  });

  final String memberUserId;
  final String memberName;

  @override
  State<ManagerMemberScreen> createState() => _ManagerMemberScreenState();
}

class _ManagerMemberScreenState extends State<ManagerMemberScreen> {
  CommercialSelfStats? _stats;
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
      final s = await ApiClient.instance
          .getCommercialManagerMemberOverview(widget.memberUserId);
      if (!mounted) return;
      setState(() {
        _stats = s;
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
        title: Text(widget.memberName),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? LoadErrorView(message: _error!, onRetry: _load)
              : ListView(
                  key: Key('manager_member_${widget.memberUserId}'),
                  padding: const EdgeInsets.all(16),
                  children: [
                    _row(l10n.managerKpiVets, '${_stats!.assignedVets}', p),
                    _row(l10n.managerKpiProspects, '${_stats!.prospectsTotal}', p),
                    _row(l10n.managerKpiConverted, '${_stats!.prospectsConverted}', p),
                    _row(
                      l10n.managerKpiAppointments,
                      '${_stats!.appointmentsUpcoming}',
                      p,
                    ),
                    _row(l10n.managerKpiStale, '${_stats!.staleInPipeline}', p),
                    _row(
                      l10n.managerKpiMonthEarned,
                      formatEuroCents(_stats!.monthEarnedCents),
                      p,
                    ),
                    _row(
                      l10n.managerKpiLifetimeEarned,
                      formatEuroCents(_stats!.lifetimeEarnedCents),
                      p,
                    ),
                  ],
                ),
    );
  }

  Widget _row(String label, String value, PetsPalette p) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          Expanded(
            child: Text(label, style: TextStyle(color: p.textMuted)),
          ),
          Text(value, style: Theme.of(context).textTheme.titleMedium),
        ],
      ),
    );
  }
}
