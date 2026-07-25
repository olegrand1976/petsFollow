import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/features/settings/presentation/feature_modules_controller.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class FeatureModulesScreen extends StatefulWidget {
  const FeatureModulesScreen({super.key, this.catalogOnly = false});

  final bool catalogOnly;

  @override
  State<FeatureModulesScreen> createState() => _FeatureModulesScreenState();
}

class _FeatureModulesScreenState extends State<FeatureModulesScreen> {
  @override
  void initState() {
    super.initState();
    FeatureModulesController.instance.load();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.catalogOnly ? l10n.featureModulesCatalog : l10n.featureModules),
      ),
      body: ListenableBuilder(
        listenable: FeatureModulesController.instance,
        builder: (context, _) {
          final c = FeatureModulesController.instance;
          if (!c.loaded) {
            return const Center(child: CircularProgressIndicator());
          }
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text(l10n.featureModulesSubtitle),
              const SizedBox(height: 16),
              _ModuleTile(
                title: l10n.moduleCarePlus,
                subtitle: l10n.moduleCarePlusDesc,
                active: c.carePlus,
                onActivate: () => c.save(carePlus: true),
                onToggle: widget.catalogOnly
                    ? null
                    : (v) => c.save(carePlus: v),
                l10n: l10n,
              ),
              _ModuleTile(
                title: l10n.moduleHorse,
                subtitle: l10n.moduleHorseDesc,
                active: c.horse,
                onActivate: () => c.save(horse: true),
                onToggle: widget.catalogOnly ? null : (v) => c.save(horse: v),
                l10n: l10n,
              ),
              _ModuleTile(
                title: l10n.moduleKennel,
                subtitle: l10n.moduleKennelDesc,
                active: c.kennel,
                onActivate: () => c.save(kennel: true),
                onToggle: widget.catalogOnly ? null : (v) => c.save(kennel: v),
                l10n: l10n,
              ),
              _ModuleTile(
                title: l10n.moduleFamily,
                subtitle: l10n.moduleFamilyDesc,
                active: c.family,
                onActivate: () => c.save(family: true),
                onToggle: widget.catalogOnly ? null : (v) => c.save(family: v),
                l10n: l10n,
              ),
            ],
          );
        },
      ),
    );
  }
}

class _ModuleTile extends StatelessWidget {
  const _ModuleTile({
    required this.title,
    required this.subtitle,
    required this.active,
    required this.onActivate,
    required this.l10n,
    this.onToggle,
  });

  final String title;
  final String subtitle;
  final bool active;
  final VoidCallback onActivate;
  final ValueChanged<bool>? onToggle;
  final AppLocalizations l10n;

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(child: Text(title, style: Theme.of(context).textTheme.titleMedium)),
                if (onToggle != null)
                  Switch(value: active, onChanged: onToggle)
                else if (active)
                  Chip(label: Text(l10n.moduleActive))
                else
                  FilledButton(onPressed: onActivate, child: Text(l10n.moduleActivate)),
              ],
            ),
            const SizedBox(height: 8),
            Text(subtitle),
          ],
        ),
      ),
    );
  }
}
