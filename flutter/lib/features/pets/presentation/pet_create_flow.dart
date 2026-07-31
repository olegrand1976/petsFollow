import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/features/pets/presentation/kennel_quick_encode_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_form_screen.dart';
import 'package:petsfollow_mobile/features/vets/presentation/link_vet_prompt.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Outcome of a successful pet create (form or kennel). `null` from [Navigator.push] = cancelled.
class PetCreateResult {
  const PetCreateResult({
    required this.promptLinkVet,
    this.snack = PetCreateSnack.savedPendingPayment,
  });

  /// True when created pet(s) have no practice — Home may invite linking a vet.
  final bool promptLinkVet;
  final PetCreateSnack snack;
}

enum PetCreateSnack {
  savedPendingPayment,
  paymentPending,
  couldNotOpenLink,
}

void showPetCreateSnack(BuildContext context, PetCreateSnack snack) {
  final l10n = AppLocalizations.of(context)!;
  final text = switch (snack) {
    PetCreateSnack.savedPendingPayment => l10n.petSavedPendingPayment,
    PetCreateSnack.paymentPending => l10n.paymentPending,
    PetCreateSnack.couldNotOpenLink => l10n.errorCouldNotOpenLink,
  };
  final key = snack == PetCreateSnack.savedPendingPayment
      ? const Key('pet_form_saved')
      : null;
  ScaffoldMessenger.of(context).showSnackBar(
    SnackBar(key: key, content: Text(text)),
  );
}

Future<void> openPetFormAndFollowUp(
  BuildContext context, {
  required Future<void> Function() onReload,
  bool? hasLinkedVets,
}) async {
  final result = await Navigator.push<PetCreateResult>(
    context,
    MaterialPageRoute(builder: (_) => const PetFormScreen()),
  );
  if (!context.mounted || result == null) return;
  showPetCreateSnack(context, result.snack);
  await promptLinkVetIfNeeded(
    context,
    promptLinkVet: result.promptLinkVet,
    hasLinkedVets: hasLinkedVets,
  );
  if (context.mounted) await onReload();
}

Future<void> openKennelEncodeAndFollowUp(
  BuildContext context, {
  required Future<void> Function() onReload,
  bool? hasLinkedVets,
}) async {
  final result = await Navigator.push<PetCreateResult>(
    context,
    MaterialPageRoute(builder: (_) => const KennelQuickEncodeScreen()),
  );
  if (!context.mounted || result == null) return;
  showPetCreateSnack(context, result.snack);
  await promptLinkVetIfNeeded(
    context,
    promptLinkVet: result.promptLinkVet,
    hasLinkedVets: hasLinkedVets,
  );
  if (context.mounted) await onReload();
}
