import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/models/preconsult.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class PreconsultScreen extends StatefulWidget {
  const PreconsultScreen({
    super.key,
    required this.visitId,
    this.petName,
  });

  final String visitId;
  final String? petName;

  @override
  State<PreconsultScreen> createState() => _PreconsultScreenState();
}

class _PreconsultScreenState extends State<PreconsultScreen> {
  bool _loading = true;
  bool _saving = false;
  String? _error;
  String? _petName;
  bool _submitted = false;

  final _complaintCtrl = TextEditingController();
  final _commentCtrl = TextEditingController();
  String _duration = 'unknown';
  String _behavior = 'unknown';
  String _appetite = 'unknown';
  String _thirst = 'unknown';
  String _elimination = 'unknown';
  String _urgency = 'medium';

  @override
  void initState() {
    super.initState();
    _petName = widget.petName;
    WidgetsBinding.instance.addPostFrameCallback((_) => _load());
  }

  @override
  void dispose() {
    _complaintCtrl.dispose();
    _commentCtrl.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final raw = await ApiClient.instance.getPreconsult(widget.visitId);
      final intake = PreconsultIntake.fromJson(raw);
      if (!mounted) return;
      setState(() {
        _petName = intake.petName ?? _petName;
        _submitted = intake.isSubmitted;
        _complaintCtrl.text = intake.answers.chiefComplaint;
        _commentCtrl.text = intake.answers.comment;
        _duration = intake.answers.duration;
        _behavior = intake.answers.behavior;
        _appetite = intake.answers.appetite;
        _thirst = intake.answers.thirst;
        _elimination = intake.answers.elimination;
        _urgency = intake.answers.urgency;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = mapApiError(e, l10n);
      });
    }
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context)!;
    if (_complaintCtrl.text.trim().isEmpty) {
      setState(() => _error = l10n.preconsultComplaintRequired);
      return;
    }
    setState(() {
      _saving = true;
      _error = null;
    });
    try {
      final answers = PreconsultAnswers(
        chiefComplaint: _complaintCtrl.text,
        duration: _duration,
        behavior: _behavior,
        appetite: _appetite,
        thirst: _thirst,
        elimination: _elimination,
        urgency: _urgency,
        comment: _commentCtrl.text,
      );
      await ApiClient.instance.submitPreconsult(widget.visitId, answers.toJson());
      if (!mounted) return;
      setState(() {
        _saving = false;
        _submitted = true;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.preconsultSubmitted)),
      );
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _saving = false;
        _error = mapApiError(e, l10n);
      });
    }
  }

  Widget _dropdown({
    required String label,
    required String value,
    required List<({String value, String label})> items,
    required ValueChanged<String> onChanged,
  }) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: InputDecorator(
        decoration: InputDecoration(
          labelText: label,
          border: const OutlineInputBorder(),
        ),
        child: DropdownButtonHideUnderline(
          child: DropdownButton<String>(
            isExpanded: true,
            value: value,
            items: [
              for (final i in items)
                DropdownMenuItem(value: i.value, child: Text(i.label)),
            ],
            onChanged: _submitted || _saving
                ? null
                : (v) {
                    if (v != null) onChanged(v);
                  },
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final title = _petName == null || _petName!.isEmpty
        ? l10n.preconsultTitle
        : l10n.preconsultTitlePet(_petName!);

    final durationItems = [
      (value: 'today', label: l10n.preconsultDurationToday),
      (value: 'few_days', label: l10n.preconsultDurationFewDays),
      (value: 'week', label: l10n.preconsultDurationWeek),
      (value: 'weeks', label: l10n.preconsultDurationWeeks),
      (value: 'months', label: l10n.preconsultDurationMonths),
      (value: 'unknown', label: l10n.preconsultUnknown),
    ];
    final behaviorItems = [
      (value: 'normal', label: l10n.preconsultBehaviorNormal),
      (value: 'lethargic', label: l10n.preconsultBehaviorLethargic),
      (value: 'restless', label: l10n.preconsultBehaviorRestless),
      (value: 'aggressive', label: l10n.preconsultBehaviorAggressive),
      (value: 'anxious', label: l10n.preconsultBehaviorAnxious),
      (value: 'other', label: l10n.preconsultBehaviorOther),
      (value: 'unknown', label: l10n.preconsultUnknown),
    ];
    final scaleItems = [
      (value: 'normal', label: l10n.preconsultScaleNormal),
      (value: 'decreased', label: l10n.preconsultScaleDecreased),
      (value: 'increased', label: l10n.preconsultScaleIncreased),
      (value: 'unknown', label: l10n.preconsultUnknown),
    ];
    final urgencyItems = [
      (value: 'low', label: l10n.preconsultUrgencyLow),
      (value: 'medium', label: l10n.preconsultUrgencyMedium),
      (value: 'high', label: l10n.preconsultUrgencyHigh),
    ];

    return Scaffold(
      appBar: AppBar(title: Text(title)),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: EdgeInsets.fromLTRB(16, 16, 16, systemBottomInset(context) + 24),
              children: [
                Text(l10n.preconsultIntro, style: Theme.of(context).textTheme.bodyMedium),
                const SizedBox(height: 16),
                if (_error != null) ...[
                  Text(_error!, style: const TextStyle(color: AppColors.alert)),
                  const SizedBox(height: 12),
                ],
                if (_submitted)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: Text(
                      l10n.preconsultAlreadySubmitted,
                      style: TextStyle(color: AppColors.accent, fontWeight: FontWeight.w600),
                    ),
                  ),
                TextField(
                  controller: _complaintCtrl,
                  enabled: !_submitted && !_saving,
                  maxLines: 3,
                  decoration: InputDecoration(
                    labelText: l10n.preconsultComplaint,
                    border: const OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 12),
                _dropdown(
                  label: l10n.preconsultDuration,
                  value: _duration,
                  items: durationItems,
                  onChanged: (v) => setState(() => _duration = v),
                ),
                _dropdown(
                  label: l10n.preconsultBehavior,
                  value: _behavior,
                  items: behaviorItems,
                  onChanged: (v) => setState(() => _behavior = v),
                ),
                _dropdown(
                  label: l10n.preconsultAppetite,
                  value: _appetite,
                  items: scaleItems,
                  onChanged: (v) => setState(() => _appetite = v),
                ),
                _dropdown(
                  label: l10n.preconsultThirst,
                  value: _thirst,
                  items: scaleItems,
                  onChanged: (v) => setState(() => _thirst = v),
                ),
                _dropdown(
                  label: l10n.preconsultElimination,
                  value: _elimination,
                  items: scaleItems,
                  onChanged: (v) => setState(() => _elimination = v),
                ),
                _dropdown(
                  label: l10n.preconsultUrgency,
                  value: _urgency,
                  items: urgencyItems,
                  onChanged: (v) => setState(() => _urgency = v),
                ),
                TextField(
                  controller: _commentCtrl,
                  enabled: !_submitted && !_saving,
                  maxLines: 3,
                  decoration: InputDecoration(
                    labelText: l10n.preconsultComment,
                    border: const OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 20),
                if (!_submitted)
                  FilledButton(
                    key: const Key('preconsult_submit'),
                    onPressed: _saving ? null : _submit,
                    child: _saving
                        ? const SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : Text(l10n.preconsultSubmit),
                  ),
              ],
            ),
    );
  }
}
