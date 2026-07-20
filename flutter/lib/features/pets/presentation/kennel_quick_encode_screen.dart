import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/billing_addon.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class KennelQuickEncodeScreen extends StatefulWidget {
  const KennelQuickEncodeScreen({super.key});

  @override
  State<KennelQuickEncodeScreen> createState() => _KennelQuickEncodeScreenState();
}

class _KennelRow {
  _KennelRow()
      : name = TextEditingController(),
        birth = TextEditingController(),
        litterTag = TextEditingController();

  final TextEditingController name;
  final TextEditingController birth;
  final TextEditingController litterTag;
  String species = 'dog';

  void dispose() {
    name.dispose();
    birth.dispose();
    litterTag.dispose();
  }
}

class _KennelQuickEncodeScreenState extends State<KennelQuickEncodeScreen> {
  final List<_KennelRow> _rows = [_KennelRow()];
  bool _loading = true;
  bool _hasKennel = false;
  bool _submitting = false;

  @override
  void initState() {
    super.initState();
    _gate();
  }

  @override
  void dispose() {
    for (final r in _rows) {
      r.dispose();
    }
    super.dispose();
  }

  Future<void> _gate() async {
    final ents = await AddonEntitlements.load();
    if (!mounted) return;
    setState(() {
      _hasKennel = ents?.hasKennel ?? false;
      _loading = false;
    });
  }

  void _addRow() {
    setState(() => _rows.add(_KennelRow()));
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context)!;
    if (!_hasKennel) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(l10n.kennelRequired)));
      return;
    }
    final pets = <Map<String, dynamic>>[];
    for (final r in _rows) {
      final name = r.name.text.trim();
      if (name.isEmpty) continue;
      final birth = r.birth.text.trim();
      pets.add({
        'name': name,
        'species': r.species,
        if (birth.isNotEmpty) 'birthDate': birth,
        'litterTag': r.litterTag.text.trim(),
      });
    }
    if (pets.isEmpty) return;
    setState(() => _submitting = true);
    try {
      await ApiClient.instance.createPetsBatch(pets);
      if (!mounted) return;
      Navigator.pop(context, true);
    } catch (e) {
      if (!mounted) return;
      final raw = e.toString();
      final msg = raw.contains('vet_link_required')
          ? l10n.noVets
          : raw.contains('kennel_required')
              ? l10n.kennelRequired
              : raw;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  String _speciesLabel(AppLocalizations l10n, String species) {
    switch (species) {
      case 'dog':
        return l10n.speciesDog;
      case 'cat':
        return l10n.speciesCat;
      case 'horse':
        return l10n.speciesHorse;
      default:
        return l10n.speciesOther;
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.kennelQuickEncodeTitle)),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : !_hasKennel
              ? Center(
                  child: Padding(
                    padding: const EdgeInsets.all(24),
                    child: Text(l10n.kennelRequired, textAlign: TextAlign.center),
                  ),
                )
              : ListView(
                  padding: const EdgeInsets.all(16),
                  children: [
                    for (var i = 0; i < _rows.length; i++) ...[
                      if (i > 0) const SizedBox(height: 16),
                      Text(
                        '${l10n.newPet} ${i + 1}',
                        style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
                      ),
                      const SizedBox(height: 8),
                      TextField(
                        controller: _rows[i].name,
                        decoration: InputDecoration(labelText: l10n.petName),
                      ),
                      const SizedBox(height: 8),
                      DropdownButtonFormField<String>(
                        key: ValueKey('kennel-species-$i'),
                        initialValue: _rows[i].species,
                        decoration: InputDecoration(labelText: l10n.species),
                        items: [
                          for (final s in const ['dog', 'cat', 'horse', 'other'])
                            DropdownMenuItem(value: s, child: Text(_speciesLabel(l10n, s))),
                        ],
                        onChanged: (v) {
                          if (v == null) return;
                          setState(() => _rows[i].species = v);
                        },
                      ),
                      const SizedBox(height: 8),
                      TextField(
                        controller: _rows[i].birth,
                        decoration: InputDecoration(labelText: l10n.horseCompetitionDate),
                      ),
                      const SizedBox(height: 8),
                      TextField(
                        controller: _rows[i].litterTag,
                        decoration: InputDecoration(labelText: l10n.litterTag),
                      ),
                    ],
                    const SizedBox(height: 16),
                    OutlinedButton.icon(
                      onPressed: _submitting ? null : _addRow,
                      icon: const Icon(Icons.add),
                      label: Text(l10n.newPet),
                    ),
                    const SizedBox(height: 12),
                    FilledButton(
                      onPressed: _submitting ? null : _submit,
                      child: _submitting
                          ? const SizedBox(
                              width: 20,
                              height: 20,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : Text(l10n.save),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      l10n.kennelPackHint,
                      style: TextStyle(color: AppColors.textMuted, fontSize: 12),
                    ),
                  ],
                ),
    );
  }
}
