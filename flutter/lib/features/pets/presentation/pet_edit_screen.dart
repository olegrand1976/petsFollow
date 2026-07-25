import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Édition des infos d'un animal par son propriétaire (PUT /pets/{id}).
class PetEditScreen extends StatefulWidget {
  const PetEditScreen({super.key, required this.pet});

  final Pet pet;

  @override
  State<PetEditScreen> createState() => _PetEditScreenState();
}

class _PetEditScreenState extends State<PetEditScreen> {
  late final TextEditingController name;
  late final TextEditingController breed;
  late String selectedSpecies;
  bool saving = false;
  String? error;

  @override
  void initState() {
    super.initState();
    name = TextEditingController(text: widget.pet.name);
    breed = TextEditingController(text: widget.pet.breed);
    const known = ['dog', 'cat', 'horse', 'other'];
    selectedSpecies =
        known.contains(widget.pet.species) ? widget.pet.species : 'other';
  }

  @override
  void dispose() {
    name.dispose();
    breed.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    if (name.text.trim().isEmpty) return;
    setState(() {
      saving = true;
      error = null;
    });
    try {
      await ApiClient.instance.updatePet(widget.pet.id, {
        'name': name.text.trim(),
        'species': selectedSpecies,
        'breed': breed.text.trim(),
      });
      if (mounted) Navigator.pop(context, true);
    } catch (_) {
      if (!mounted) return;
      setState(() => error = 'save');
    } finally {
      if (mounted) setState(() => saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.editPet)),
      body: SingleChildScrollView(
        padding: scrollPaddingWithSystemBottom(context, all: 16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            TextField(
              controller: name,
              decoration: InputDecoration(labelText: l10n.petName),
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<String>(
              initialValue: selectedSpecies,
              decoration: InputDecoration(labelText: l10n.species),
              items: [
                DropdownMenuItem(value: 'dog', child: Text(l10n.speciesDog)),
                DropdownMenuItem(value: 'cat', child: Text(l10n.speciesCat)),
                DropdownMenuItem(value: 'horse', child: Text(l10n.speciesHorse)),
                DropdownMenuItem(value: 'other', child: Text(l10n.speciesOther)),
              ],
              onChanged: (v) => setState(() => selectedSpecies = v ?? 'other'),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: breed,
              decoration: InputDecoration(labelText: l10n.breed),
            ),
            if (error != null) ...[
              const SizedBox(height: 12),
              Text(l10n.errorGeneric(error!),
                  style: const TextStyle(color: AppColors.alert)),
            ],
            const SizedBox(height: 24),
            FilledButton(
              onPressed: saving ? null : _save,
              child: saving
                  ? const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(strokeWidth: 2))
                  : Text(l10n.save),
            ),
          ],
        ),
      ),
    );
  }
}
