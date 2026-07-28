import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/core/ui/load_error_view.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_create_flow.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_detail_screen.dart';
import 'package:petsfollow_mobile/features/settings/presentation/feature_modules_controller.dart';
import 'package:petsfollow_mobile/features/shell/presentation/main_shell_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class PetsTab extends StatefulWidget {
  const PetsTab({super.key});

  @override
  State<PetsTab> createState() => _PetsTabState();
}

class _PetsTabState extends State<PetsTab> {
  List<Pet> pets = [];
  bool loading = true;
  String? loadError;
  bool _hasLoadedOnce = false;

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    final l10n = AppLocalizations.of(context)!;
    final keepStale = _hasLoadedOnce;
    if (!keepStale && mounted) {
      setState(() {
        loading = true;
        loadError = null;
      });
    }
    await FeatureModulesController.instance.load();
    try {
      final data = await ApiClient.instance.getPets();
      if (mounted) {
        setState(() {
          pets = data
              .map((p) => Pet.fromJson(Map<String, dynamic>.from(p as Map)))
              .toList();
          loading = false;
          loadError = null;
          _hasLoadedOnce = true;
        });
      }
    } catch (e) {
      if (!mounted) return;
      final msg = mapApiError(e, l10n);
      if (keepStale) {
        setState(() => loading = false);
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text(msg)));
      } else {
        setState(() {
          loading = false;
          loadError = msg;
        });
      }
    }
  }

  Future<void> _openPetForm() =>
      openPetFormAndFollowUp(context, onReload: load);

  Future<void> _openKennelEncode() {
    if (!FeatureModulesController.instance.kennel) return Future.value();
    return openKennelEncodeAndFollowUp(context, onReload: load);
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
    final p = PetsPalette.of(context);

    return PetsTabScaffold(
      title: Text(l10n.navPets),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : loadError != null
              ? LoadErrorView(message: loadError!, onRetry: load)
              : pets.isEmpty
                  ? Center(
                      child: Padding(
                        padding: const EdgeInsets.all(24),
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.pets,
                                size: 48,
                                color: AppColors.gold.withValues(alpha: 0.8)),
                            const SizedBox(height: 16),
                            Text(l10n.emptyPetsTitle,
                                style: Theme.of(context).textTheme.titleLarge),
                            const SizedBox(height: 8),
                            Text(l10n.emptyPetsBody,
                                textAlign: TextAlign.center),
                          ],
                        ),
                      ),
                    )
                  : RefreshIndicator(
                      onRefresh: load,
                      child: ListView.builder(
                        padding: const EdgeInsets.all(16),
                        itemCount: pets.length,
                        itemBuilder: (_, i) {
                          final pet = pets[i];
                          return Card(
                            margin: const EdgeInsets.only(bottom: 8),
                            child: ListTile(
                              leading: CircleAvatar(
                                backgroundImage:
                                    pet.photoUrl?.isNotEmpty == true
                                        ? NetworkImage(pet.photoUrl!)
                                        : null,
                                child: pet.photoUrl?.isNotEmpty == true
                                    ? null
                                    : Text(pet.name.isNotEmpty
                                        ? pet.name[0].toUpperCase()
                                        : '?'),
                              ),
                              title: Text(pet.name),
                              subtitle: Text(
                                pet.isSharedAccess
                                    ? '${_speciesLabel(l10n, pet.species)} · ${pet.sharedAccessLabel(l10n)}'
                                    : '${_speciesLabel(l10n, pet.species)} · ${pet.breed}',
                              ),
                              trailing: pet.isSharedAccess
                                  ? Icon(
                                      pet.canWriteNotes
                                          ? Icons.edit_note_outlined
                                          : Icons.visibility_outlined,
                                      color: p.textMuted,
                                    )
                                  : pet.isActive
                                      ? Icon(Icons.check_circle,
                                          color: AppColors.primary)
                                      : Icon(Icons.schedule,
                                          color: AppColors.alert),
                              onTap: () async {
                                await Navigator.push(
                                  context,
                                  MaterialPageRoute(
                                    builder: (_) => PetDetailScreen(
                                        pet: pet, onUpdated: load),
                                  ),
                                );
                                load();
                              },
                            ),
                          );
                        },
                      ),
                    ),
      floatingActionButton: ListenableBuilder(
        listenable: FeatureModulesController.instance,
        builder: (context, _) {
          final showKennelFab = FeatureModulesController.instance.kennel &&
              pets.any((p) => p.isOwner);
          return Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              if (showKennelFab) ...[
                FloatingActionButton.extended(
                  heroTag: 'kennel-encode',
                  onPressed: _openKennelEncode,
                  icon: const Icon(Icons.pets_outlined),
                  label: Text(l10n.kennelQuickEncodeTitle),
                ),
                const SizedBox(height: 12),
              ],
              FloatingActionButton(
                heroTag: 'add-pet',
                onPressed: _openPetForm,
                child: const Icon(Icons.add),
              ),
            ],
          );
        },
      ),
    );
  }
}
