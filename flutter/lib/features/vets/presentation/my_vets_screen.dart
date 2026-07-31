import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/models/vet_link.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/core/ui/load_error_view.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/features/vets/presentation/widgets/add_vet_panel.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class MyVetsScreen extends StatefulWidget {
  const MyVetsScreen({super.key});

  @override
  State<MyVetsScreen> createState() => _MyVetsScreenState();
}

class _MyVetsScreenState extends State<MyVetsScreen> {
  List<VetLink> vets = [];
  bool loading = true;
  String? loadError;
  bool _hasLoadedOnce = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) load();
    });
  }

  Future<void> load() async {
    if (!mounted) return;
    final l10n = AppLocalizations.of(context)!;
    final keepStale = _hasLoadedOnce;
    if (!keepStale && mounted) {
      setState(() {
        loading = true;
        loadError = null;
      });
    }
    try {
      final data = await ApiClient.instance.getMyVets();
      if (mounted) {
        setState(() {
          vets = data;
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
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
      } else {
        setState(() {
          loading = false;
          loadError = msg;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final p = PetsPalette.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.myVets)),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : loadError != null
              ? LoadErrorView(message: loadError!, onRetry: load)
              : RefreshIndicator(
                  onRefresh: load,
                  child: ListView(
                    padding: scrollPaddingWithSystemBottom(context, all: 16),
                    children: [
                      AddVetPanel(onLinked: () async {
                        await ApiClient.instance.ensureFreshSession();
                        if (mounted) load();
                      }),
                      const SizedBox(height: 20),
                      if (vets.isEmpty)
                        Center(
                          child: Padding(
                            padding: const EdgeInsets.all(32),
                            child: Text(l10n.noVets, style: TextStyle(color: p.textMuted)),
                          ),
                        )
                      else
                        ...vets.map(
                          (v) => Card(
                            margin: const EdgeInsets.only(bottom: 8),
                            child: ListTile(
                              leading: const Icon(Icons.local_hospital_outlined, color: AppColors.primary),
                              title: Text(v.practiceName),
                              subtitle: Text('${v.vetFullName}\n${v.vetEmail}'),
                              isThreeLine: true,
                              trailing: v.isPrimary
                                  ? Chip(
                                      label: Text(l10n.primaryVet, style: const TextStyle(fontSize: 11)),
                                      backgroundColor: AppColors.gold.withValues(alpha: 0.15),
                                    )
                                  : null,
                            ),
                          ),
                        ),
                    ],
                  ),
                ),
    );
  }
}
