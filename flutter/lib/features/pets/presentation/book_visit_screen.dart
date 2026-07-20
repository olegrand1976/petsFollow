import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class BookVisitScreen extends StatefulWidget {
  const BookVisitScreen({
    super.key,
    required this.petId,
    required this.petName,
    required this.practiceId,
  });

  final String petId;
  final String petName;
  final String practiceId;

  @override
  State<BookVisitScreen> createState() => _BookVisitScreenState();
}

class _BookVisitScreenState extends State<BookVisitScreen> {
  bool _loading = true;
  bool _booking = false;
  bool _enabled = false;
  List<DateTime> _slots = [];
  String? _error;

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
      final from = DateTime.now();
      final to = from.add(const Duration(days: 14));
      final data = await ApiClient.instance.getPracticeAvailability(
        widget.practiceId,
        from: from,
        to: to,
      );
      final enabled = data['enabled'] == true;
      final raw = data['slots'] as List<dynamic>? ?? [];
      final slots = raw
          .map((e) => DateTime.tryParse((e as Map)['start']?.toString() ?? ''))
          .whereType<DateTime>()
          .toList();
      if (!mounted) return;
      setState(() {
        _enabled = enabled;
        _slots = slots;
        _loading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = AppLocalizations.of(context)!.errorGeneric('calendar');
      });
    }
  }

  Future<void> _book(DateTime slot) async {
    final l10n = AppLocalizations.of(context)!;
    setState(() => _booking = true);
    try {
      await ApiClient.instance.createVisit(widget.petId, scheduledAt: slot);
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(l10n.visitRequested)));
      Navigator.pop(context, true);
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.errorGeneric('visit'))),
      );
    } finally {
      if (mounted) setState(() => _booking = false);
    }
  }

  Future<void> _legacyRequest() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() => _booking = true);
    try {
      await ApiClient.instance.createVisit(widget.petId);
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(l10n.visitRequested)));
      Navigator.pop(context, true);
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.errorGeneric('visit'))),
      );
    } finally {
      if (mounted) setState(() => _booking = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.requestVisit)),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(20),
              children: [
                Text(
                  widget.petName,
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(color: AppColors.gold),
                ),
                const SizedBox(height: 12),
                if (_error != null) Text(_error!, style: TextStyle(color: Theme.of(context).colorScheme.error)),
                if (!_enabled) ...[
                  Text(l10n.calendarBookingDisabled),
                  const SizedBox(height: 16),
                  FilledButton(
                    onPressed: _booking ? null : _legacyRequest,
                    child: Text(l10n.requestVisit),
                  ),
                ] else if (_slots.isEmpty) ...[
                  Text(l10n.calendarNoSlots),
                ] else ...[
                  Text(l10n.calendarPickSlot),
                  const SizedBox(height: 12),
                  ..._slots.map((slot) {
                    final local = slot.toLocal();
                    final label =
                        '${local.day.toString().padLeft(2, '0')}/${local.month.toString().padLeft(2, '0')} '
                        '${local.hour.toString().padLeft(2, '0')}:${local.minute.toString().padLeft(2, '0')}';
                    return Card(
                      child: ListTile(
                        title: Text(label),
                        trailing: const Icon(Icons.chevron_right),
                        onTap: _booking ? null : () => _book(slot),
                      ),
                    );
                  }),
                ],
              ],
            ),
    );
  }
}
