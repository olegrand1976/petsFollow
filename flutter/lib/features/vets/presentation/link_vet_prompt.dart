import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/features/vets/presentation/my_vets_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Optional post-create invite after a pet was saved without practice.
///
/// Skipped when [hasLinkedVets] is false: Home already shows the soft
/// `_AddFirstVetCard` banner — avoid stacking dialog + banner.
Future<void> promptLinkVetIfNeeded(
  BuildContext context, {
  required bool promptLinkVet,
  bool? hasLinkedVets,
}) async {
  if (!promptLinkVet || !context.mounted) return;
  if (hasLinkedVets == false) return;

  final l10n = AppLocalizations.of(context)!;
  final linkNow = await showDialog<bool>(
        context: context,
        builder: (ctx) => AlertDialog(
          key: const Key('home_link_vet_dialog'),
          title: Text(l10n.linkVetHomeTitle),
          content: Text(l10n.linkVetHomeBody),
          actions: [
            TextButton(
              key: const Key('home_link_vet_later'),
              onPressed: () => Navigator.pop(ctx, false),
              child: Text(l10n.linkVetLater),
            ),
            FilledButton(
              key: const Key('home_link_vet'),
              onPressed: () => Navigator.pop(ctx, true),
              child: Text(l10n.addVetByEmail),
            ),
          ],
        ),
      ) ??
      false;
  if (!linkNow || !context.mounted) return;
  await Navigator.push<void>(
    context,
    MaterialPageRoute<void>(builder: (_) => const MyVetsScreen()),
  );
}
