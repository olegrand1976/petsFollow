import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
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
  late final TextEditingController microchip;
  late final TextEditingController healthBookNumber;
  late String selectedSpecies;
  List<XFile> healthBookPages = [];
  bool saving = false;
  bool clearingHealthBook = false;
  String? error;
  bool healthBookPdfAttached = false;

  @override
  void initState() {
    super.initState();
    name = TextEditingController(text: widget.pet.name);
    breed = TextEditingController(text: widget.pet.breed);
    microchip = TextEditingController(text: widget.pet.microchipNumber ?? '');
    healthBookNumber =
        TextEditingController(text: widget.pet.healthBookNumber ?? '');
    healthBookPdfAttached = widget.pet.healthBookPdfAttached;
    const known = ['dog', 'cat', 'horse', 'other'];
    selectedSpecies =
        known.contains(widget.pet.species) ? widget.pet.species : 'other';
  }

  @override
  void dispose() {
    name.dispose();
    breed.dispose();
    microchip.dispose();
    healthBookNumber.dispose();
    super.dispose();
  }

  Future<void> _pickHealthBookPages() async {
    final picker = ImagePicker();
    final files = await picker.pickMultiImage(
      maxWidth: 1600,
      maxHeight: 1600,
      imageQuality: 80,
    );
    if (files.isEmpty) return;
    setState(() {
      healthBookPages = [...healthBookPages, ...files].take(10).toList();
    });
  }

  Future<void> _clearHealthBookPdf() async {
    setState(() {
      clearingHealthBook = true;
      error = null;
    });
    try {
      final updated =
          await ApiClient.instance.deletePetHealthBook(widget.pet.id);
      if (!mounted) return;
      setState(() {
        healthBookPdfAttached = updated['healthBookPdfAttached'] == true;
        healthBookPages = [];
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => error = 'save');
    } finally {
      if (mounted) setState(() => clearingHealthBook = false);
    }
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
        'microchipNumber': microchip.text.trim(),
        'healthBookNumber': healthBookNumber.text.trim(),
      });
      if (healthBookPages.isNotEmpty) {
        try {
          await ApiClient.instance.uploadPetHealthBook(
            widget.pet.id,
            healthBookPages.map((f) => f.path).toList(),
          );
        } catch (_) {
          if (!mounted) return;
          setState(() => error = 'health_book');
          return;
        }
      }
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
    final hasPdf = healthBookPdfAttached;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.editPet)),
      body: SingleChildScrollView(
        padding: scrollPaddingWithSystemBottom(context, all: 16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            TextField(
              key: const Key('pet_edit_name'),
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
            const SizedBox(height: 12),
            TextField(
              key: const Key('pet_edit_microchip'),
              controller: microchip,
              decoration: InputDecoration(labelText: l10n.petMicrochipOptional),
            ),
            const SizedBox(height: 12),
            TextField(
              key: const Key('pet_edit_health_book_number'),
              controller: healthBookNumber,
              decoration:
                  InputDecoration(labelText: l10n.petHealthBookNumberOptional),
            ),
            const SizedBox(height: 12),
            if (hasPdf)
              ListTile(
                contentPadding: EdgeInsets.zero,
                leading: const Icon(Icons.picture_as_pdf_outlined),
                title: Text(l10n.petHealthBookPdfAttached),
                trailing: TextButton(
                  key: const Key('pet_edit_health_book_clear'),
                  onPressed: clearingHealthBook || saving
                      ? null
                      : _clearHealthBookPdf,
                  child: Text(l10n.petHealthBookRemovePdf),
                ),
              ),
            OutlinedButton.icon(
              key: const Key('pet_edit_health_book_pick'),
              onPressed: saving ? null : _pickHealthBookPages,
              icon: const Icon(Icons.photo_library_outlined),
              label: Text(
                healthBookPages.isEmpty
                    ? (hasPdf
                        ? l10n.petHealthBookReplacePages
                        : l10n.petHealthBookAddPages)
                    : l10n.petHealthBookPagesCount(healthBookPages.length),
              ),
            ),
            if (healthBookPages.isNotEmpty) ...[
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  for (var i = 0; i < healthBookPages.length; i++)
                    InputChip(
                      label: Text('${i + 1}'),
                      onDeleted: () => setState(() {
                        healthBookPages = List.of(healthBookPages)..removeAt(i);
                      }),
                    ),
                ],
              ),
            ],
            if (error != null) ...[
              const SizedBox(height: 12),
              Text(
                error == 'health_book'
                    ? l10n.errorHealthBookUploadFailed
                    : l10n.errorGeneric(error!),
                style: const TextStyle(color: AppColors.alert),
              ),
            ],
            const SizedBox(height: 24),
            FilledButton(
              key: const Key('pet_edit_save'),
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
