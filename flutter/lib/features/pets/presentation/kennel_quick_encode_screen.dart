import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_create_flow.dart';
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
  bool _submitting = false;

  @override
  void dispose() {
    for (final r in _rows) {
      r.dispose();
    }
    super.dispose();
  }

  void _addRow() {
    setState(() => _rows.add(_KennelRow()));
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context)!;
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
        // Default steer plan (same as API defaultStr) — explicit for clarity.
        'plan': 'triennial',
        'billingMode': 'subscription',
      });
    }
    if (pets.isEmpty) return;
    setState(() => _submitting = true);
    try {
      final res = await ApiClient.instance.createPetsBatch(pets);
      if (!mounted) return;
      final rawPets = res['pets'];
      var needsVet = false;
      if (rawPets is List) {
        for (final item in rawPets) {
          if (item is! Map) continue;
          final pid = item['practiceId']?.toString().trim() ?? '';
          if (pid.isEmpty) {
            needsVet = true;
            break;
          }
        }
      } else {
        needsVet = true;
      }
      // Caller shows snack + optional link-vet prompt on Home/Pets.
      Navigator.pop(
        context,
        PetCreateResult(promptLinkVet: needsVet),
      );
    } catch (e) {
      if (!mounted) return;
      final msg = mapApiError(e, l10n);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          key: const Key('kennel_error'),
          content: Text(msg),
        ),
      );
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
      body: ListView(
        padding: scrollPaddingWithSystemBottom(context, all: 16),
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
              decoration: InputDecoration(
                labelText: l10n.petBirthDate,
                hintText: 'YYYY-MM-DD',
              ),
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
            key: const Key('kennel_submit'),
            onPressed: _submitting ? null : _submit,
            child: _submitting
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : Text(l10n.save),
          ),
        ],
      ),
    );
  }
}
