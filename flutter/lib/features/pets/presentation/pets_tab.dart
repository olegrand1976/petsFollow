import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_detail_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_form_screen.dart';
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

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    try {
      final data = await ApiClient.instance.getPets();
      if (mounted) {
        setState(() {
          pets = data.map((p) => Pet.fromJson(Map<String, dynamic>.from(p as Map))).toList();
          loading = false;
        });
      }
    } catch (_) {
      if (mounted) setState(() => loading = false);
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

    return PetsTabScaffold(
      title: Text(l10n.navPets),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : pets.isEmpty
              ? Center(
                  child: Padding(
                    padding: const EdgeInsets.all(24),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.pets, size: 48, color: AppColors.gold.withValues(alpha: 0.8)),
                        const SizedBox(height: 16),
                        Text(l10n.emptyPetsTitle, style: Theme.of(context).textTheme.titleLarge),
                        const SizedBox(height: 8),
                        Text(l10n.emptyPetsBody, textAlign: TextAlign.center),
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
                            backgroundImage: pet.photoUrl?.isNotEmpty == true
                                ? NetworkImage(pet.photoUrl!)
                                : null,
                            child: pet.photoUrl?.isNotEmpty == true
                                ? null
                                : Text(pet.name.isNotEmpty ? pet.name[0].toUpperCase() : '?'),
                          ),
                          title: Text(pet.name),
                          subtitle: Text('${_speciesLabel(l10n, pet.species)} · ${pet.breed}'),
                          trailing: pet.isActive
                              ? Icon(Icons.check_circle, color: AppColors.primary)
                              : Icon(Icons.schedule, color: AppColors.alert),
                          onTap: () async {
                            await Navigator.push(
                              context,
                              MaterialPageRoute(
                                builder: (_) => PetDetailScreen(pet: pet, onUpdated: load),
                              ),
                            );
                            load();
                          },
                        ),
                      );
                    },
                  ),
                ),
      floatingActionButton: FloatingActionButton(
        onPressed: () async {
          await Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const PetFormScreen()),
          );
          load();
        },
        child: const Icon(Icons.add),
      ),
    );
  }
}
