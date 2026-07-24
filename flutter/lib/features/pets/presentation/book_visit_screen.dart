import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/api/open_url.dart';
import 'package:petsfollow_mobile/core/models/practice_availability.dart';
import 'package:petsfollow_mobile/core/models/vet_link.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class BookVisitScreen extends StatefulWidget {
  const BookVisitScreen({
    super.key,
    required this.petId,
    required this.petName,
    this.practiceIdFilter,
    this.initialVets,
    this.rescheduleVisitId,
    /// When set (e.g. reschedule), skip vet picker and load this practice.
    this.practiceId,
    /// Test-only: skip network and use this availability payload.
    @visibleForTesting this.availabilityOverride,
  });

  final String petId;
  final String petName;
  /// Restrict vet list / booking to this practice (usually the pet's practice).
  final String? practiceIdFilter;
  final List<VetLink>? initialVets;
  /// When set, picking a slot proposes a reschedule instead of creating a visit.
  final String? rescheduleVisitId;
  /// Pre-selected practice (reschedule / single-vet shortcut).
  final String? practiceId;
  final PracticeAvailability? availabilityOverride;

  bool get isReschedule => rescheduleVisitId != null && rescheduleVisitId!.isNotEmpty;

  @override
  State<BookVisitScreen> createState() => _BookVisitScreenState();
}

class _BookVisitScreenState extends State<BookVisitScreen> {
  bool _loadingVets = true;
  bool _loadingSlots = false;
  bool _booking = false;
  List<VetLink> _vets = [];
  VetLink? _selectedVet;
  bool _enabled = false;
  List<DateTime> _slots = [];
  String? _practicePhone;
  String? _practiceName;
  String? _error;

  String? get _practiceId => _selectedVet?.practiceId ?? widget.practiceId;

  String? get _requiredPracticeId {
    final f = widget.practiceIdFilter?.trim();
    if (f != null && f.isNotEmpty) return f;
    return null;
  }

  bool _practiceMatches(String? practiceId) {
    final required = _requiredPracticeId;
    if (required == null) return true;
    return practiceId != null && practiceId == required;
  }

  @override
  void initState() {
    super.initState();
    _bootstrap();
  }

  Future<void> _bootstrap() async {
    if (widget.isReschedule && (widget.practiceId ?? '').isNotEmpty) {
      setState(() {
        _loadingVets = false;
        _selectedVet = VetLink(
          practiceId: widget.practiceId!,
          vetEmail: '',
          vetFullName: '',
          practiceName: '',
        );
      });
      await _loadAvailability();
      return;
    }

    setState(() {
      _loadingVets = true;
      _error = null;
    });
    try {
      final filter = _requiredPracticeId;
      var vets = widget.initialVets ?? await ApiClient.instance.getMyVets(primaryPracticeId: filter);
      if (filter != null) {
        vets = vets.where((v) => v.practiceId == filter).toList();
      }
      if (!mounted) return;
      if (vets.isEmpty) {
        setState(() {
          _vets = [];
          _loadingVets = false;
        });
        return;
      }
      if (vets.length == 1) {
        setState(() {
          _vets = vets!;
          _selectedVet = vets.first;
          _loadingVets = false;
        });
        await _loadAvailability();
        return;
      }
      setState(() {
        _vets = vets!;
        _loadingVets = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loadingVets = false;
        _error = mapApiError(e, AppLocalizations.of(context)!);
      });
    }
  }

  Future<void> _selectVet(VetLink vet) async {
    if (!_practiceMatches(vet.practiceId)) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.noVets)),
      );
      return;
    }
    setState(() {
      _selectedVet = vet;
      _error = null;
      _slots = [];
      _practicePhone = null;
      _practiceName = null;
    });
    await _loadAvailability();
  }

  Future<void> _loadAvailability() async {
    final practiceId = _practiceId;
    if (practiceId == null || practiceId.isEmpty) return;
    if (!_practiceMatches(practiceId)) {
      if (!mounted) return;
      setState(() {
        _error = AppLocalizations.of(context)!.noVets;
        _loadingSlots = false;
      });
      return;
    }
    setState(() {
      _loadingSlots = true;
      _error = null;
    });
    try {
      final PracticeAvailability data;
      if (widget.availabilityOverride != null) {
        data = widget.availabilityOverride!;
      } else {
        final from = DateTime.now();
        final to = from.add(const Duration(days: 14));
        data = await ApiClient.instance.getPracticeAvailability(
          practiceId,
          from: from,
          to: to,
        );
      }
      if (!mounted) return;
      setState(() {
        _enabled = data.enabled;
        _slots = data.slots.map((s) => s.start).toList();
        _practicePhone = data.practicePhone.isEmpty ? null : data.practicePhone;
        _practiceName = data.practiceName.isEmpty ? null : data.practiceName;
        _loadingSlots = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loadingSlots = false;
        _error = mapApiError(e, AppLocalizations.of(context)!);
      });
    }
  }

  Future<void> _book(DateTime slot) async {
    final l10n = AppLocalizations.of(context)!;
    if (!_practiceMatches(_practiceId)) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(l10n.noVets)));
      return;
    }
    setState(() => _booking = true);
    try {
      if (widget.isReschedule) {
        await ApiClient.instance.visitAction(
          widget.rescheduleVisitId!,
          action: 'propose_reschedule',
          proposedScheduledAt: slot,
        );
      } else {
        await ApiClient.instance.createVisit(widget.petId, scheduledAt: slot);
      }
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            widget.isReschedule ? l10n.visitRescheduleProposed : l10n.visitRequested,
          ),
        ),
      );
      Navigator.pop(context, true);
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(mapApiError(e, l10n))),
      );
    } finally {
      if (mounted) setState(() => _booking = false);
    }
  }

  Future<void> _pickManualSlot() async {
    final now = DateTime.now();
    final date = await showDatePicker(
      context: context,
      initialDate: now.add(const Duration(days: 1)),
      firstDate: now,
      lastDate: now.add(const Duration(days: 60)),
    );
    if (date == null || !mounted) return;
    final time = await showTimePicker(
      context: context,
      initialTime: const TimeOfDay(hour: 10, minute: 0),
    );
    if (time == null || !mounted) return;
    await _book(DateTime(date.year, date.month, date.day, time.hour, time.minute));
  }

  Future<void> _callPractice() async {
    final phone = _practicePhone?.trim() ?? '';
    if (phone.isEmpty) return;
    final digits = phone.replaceAll(RegExp(r'[^\d+]'), '');
    if (digits.isEmpty) return;
    await openExternalUrl('tel:$digits');
  }

  void _clearVetSelection() {
    if (_vets.length <= 1) return;
    setState(() {
      _selectedVet = null;
      _slots = [];
      _enabled = false;
      _practicePhone = null;
      _practiceName = null;
      _error = null;
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final title = widget.isReschedule ? l10n.visitProposeReschedule : l10n.requestVisit;
    return Scaffold(
      appBar: AppBar(
        title: Text(title),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: _booking
              ? null
              : () {
                  if (_selectedVet != null && _vets.length > 1 && !widget.isReschedule) {
                    _clearVetSelection();
                    return;
                  }
                  Navigator.maybePop(context);
                },
        ),
      ),
      body: _buildBody(l10n),
    );
  }

  Widget _buildBody(AppLocalizations l10n) {
    if (_loadingVets || (_selectedVet != null && _loadingSlots)) {
      return const Center(child: CircularProgressIndicator());
    }

    return ListView(
      padding: scrollPaddingWithSystemBottom(context, all: 20),
      children: [
        Text(
          widget.petName,
          style: Theme.of(context).textTheme.titleMedium?.copyWith(color: AppColors.gold),
        ),
        const SizedBox(height: 12),
        if (_error != null) ...[
          Text(_error!, style: const TextStyle(color: AppColors.alert)),
          const SizedBox(height: 12),
          OutlinedButton(
            onPressed: _booking
                ? null
                : () {
                    if (_selectedVet == null) {
                      _bootstrap();
                    } else {
                      _loadAvailability();
                    }
                  },
            child: Text(l10n.retryAction),
          ),
          if (widget.isReschedule) ...[
            const SizedBox(height: 8),
            FilledButton(
              onPressed: _booking ? null : _pickManualSlot,
              child: Text(l10n.visitProposeReschedule),
            ),
          ],
        ] else if (_selectedVet == null) ...[
          if (_vets.isEmpty)
            Text(l10n.noVets)
          else ...[
            Text(l10n.calendarSelectVet),
            const SizedBox(height: 12),
            ..._vets.map((vet) {
              return Card(
                child: ListTile(
                  key: Key('book_visit_vet_${vet.practiceId}'),
                  title: Text(vet.vetFullName.isNotEmpty ? vet.vetFullName : vet.practiceName),
                  subtitle: Text(
                    vet.vetFullName.isNotEmpty && vet.practiceName.isNotEmpty
                        ? vet.practiceName
                        : vet.vetEmail,
                  ),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: _booking ? null : () => _selectVet(vet),
                ),
              );
            }),
          ],
        ] else if (!_enabled) ...[
          if ((_practiceName ?? _selectedVet?.practiceName ?? '').isNotEmpty) ...[
            Text(
              _practiceName ?? _selectedVet!.practiceName,
              style: Theme.of(context).textTheme.titleSmall,
            ),
            const SizedBox(height: 8),
          ],
          Text(
            widget.isReschedule
                ? l10n.calendarBookingDisabledReschedule
                : l10n.calendarBookingDisabled,
          ),
          const SizedBox(height: 16),
          if (widget.isReschedule)
            FilledButton(
              onPressed: _booking ? null : _pickManualSlot,
              child: Text(l10n.visitProposeReschedule),
            )
          else if ((_practicePhone ?? '').isNotEmpty)
            FilledButton.icon(
              key: const Key('book_visit_call_practice'),
              onPressed: _booking ? null : _callPractice,
              icon: const Icon(Icons.phone),
              label: Text('${l10n.calendarCallPractice} ($_practicePhone)'),
            )
          else
            Text(l10n.calendarNoPhone),
        ] else if (_slots.isEmpty) ...[
          Text(l10n.calendarNoSlots),
          if (widget.isReschedule) ...[
            const SizedBox(height: 16),
            OutlinedButton(
              onPressed: _booking ? null : _pickManualSlot,
              child: Text(l10n.visitProposeReschedule),
            ),
          ],
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
    );
  }
}
