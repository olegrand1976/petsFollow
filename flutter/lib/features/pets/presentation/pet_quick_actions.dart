import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Compact heart + weight + blood pressure actions for an owned active pet.
class PetQuickActions extends StatelessWidget {
  const PetQuickActions({
    super.key,
    required this.petId,
    this.onHeartRate,
    this.onWeightRecorded,
    this.onBloodPressureRecorded,
  });

  final String petId;
  final VoidCallback? onHeartRate;
  final VoidCallback? onWeightRecorded;
  final VoidCallback? onBloodPressureRecorded;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final showHr = onHeartRate != null;
    return Column(
      children: [
        Row(
          children: [
            if (showHr) ...[
              Expanded(
                child: OutlinedButton.icon(
                  key: Key('pet_action_heartrate_$petId'),
                  onPressed: onHeartRate,
                  icon: const Icon(Icons.favorite, size: 18),
                  label: Text(l10n.heartRateShort),
                ),
              ),
              const SizedBox(width: 8),
            ],
            Expanded(
              child: OutlinedButton.icon(
                key: Key('pet_action_weight_$petId'),
                onPressed: () => showRecordWeightSheet(
                  context,
                  petId: petId,
                  onSaved: onWeightRecorded,
                ),
                icon: const Icon(Icons.monitor_weight_outlined, size: 18),
                label: Text(l10n.weightShort),
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        SizedBox(
          width: double.infinity,
          child: OutlinedButton.icon(
            key: Key('pet_action_bp_$petId'),
            onPressed: () => showRecordBloodPressureSheet(
              context,
              petId: petId,
              onSaved: onBloodPressureRecorded,
            ),
            icon: const Icon(Icons.monitor_heart_outlined, size: 18),
            label: Text(l10n.bloodPressureShort),
          ),
        ),
      ],
    );
  }
}

Future<void> showRecordWeightSheet(
  BuildContext context, {
  required String petId,
  VoidCallback? onSaved,
}) async {
  final l10n = AppLocalizations.of(context)!;
  final kgController = TextEditingController();
  final commentController = TextEditingController();
  var saving = false;
  String? error;

  try {
    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (sheetContext) {
        return StatefulBuilder(
          builder: (ctx, setModal) {
            final bottom =
                keyboardBottomInset(ctx) + systemBottomInset(ctx);
            return Padding(
              key: const Key('weight_sheet'),
              padding: EdgeInsets.only(
                left: 20,
                right: 20,
                top: 20,
                bottom: bottom + 20,
              ),
              child: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(
                      l10n.recordWeightTitle,
                      style: Theme.of(ctx).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      key: const Key('weight_kg_field'),
                      controller: kgController,
                      enabled: !saving,
                      keyboardType: const TextInputType.numberWithOptions(
                        decimal: true,
                      ),
                      inputFormatters: [
                        FilteringTextInputFormatter.allow(RegExp(r'[0-9.,]')),
                      ],
                      decoration: InputDecoration(
                        labelText: l10n.weightKgLabel,
                        border: const OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      key: const Key('weight_comment_field'),
                      controller: commentController,
                      enabled: !saving,
                      maxLength: 500,
                      maxLines: 2,
                      decoration: InputDecoration(
                        labelText: l10n.weightCommentLabel,
                        hintText: l10n.weightCommentHint,
                        border: const OutlineInputBorder(),
                      ),
                    ),
                    if (error != null) ...[
                      const SizedBox(height: 8),
                      Text(
                        error!,
                        key: const Key('weight_error'),
                        style: const TextStyle(color: AppColors.alert),
                      ),
                    ],
                    const SizedBox(height: 12),
                    FilledButton(
                      key: const Key('weight_save_btn'),
                      onPressed: saving
                          ? null
                          : () async {
                              final raw = kgController.text
                                  .trim()
                                  .replaceAll(',', '.');
                              final kg = double.tryParse(raw);
                              if (kg == null ||
                                  kg < 0.01 ||
                                  kg > 999.99) {
                                setModal(() => error = l10n.weightInvalid);
                                return;
                              }
                              setModal(() {
                                saving = true;
                                error = null;
                              });
                              try {
                                final comment =
                                    commentController.text.trim();
                                await ApiClient.instance.createWeightReading(
                                  petId,
                                  weightKg: kg,
                                  comment:
                                      comment.isEmpty ? null : comment,
                                );
                                if (sheetContext.mounted) {
                                  Navigator.of(sheetContext).pop();
                                }
                                if (context.mounted) {
                                  ScaffoldMessenger.of(context).showSnackBar(
                                    SnackBar(
                                      content: Text(l10n.weightSentToVet),
                                    ),
                                  );
                                }
                                onSaved?.call();
                              } catch (e) {
                                if (!ctx.mounted) return;
                                setModal(() {
                                  saving = false;
                                  final code = apiErrorCode(e);
                                  error = code == 'invalid_weight'
                                      ? l10n.weightInvalid
                                      : mapApiError(e, l10n);
                                });
                              }
                            },
                      child: saving
                          ? const SizedBox(
                              width: 20,
                              height: 20,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                              ),
                            )
                          : Text(l10n.weightSave),
                    ),
                  ],
                ),
              ),
            );
          },
        );
      },
    );
  } finally {
    kgController.dispose();
    commentController.dispose();
  }
}

Future<void> showRecordBloodPressureSheet(
  BuildContext context, {
  required String petId,
  VoidCallback? onSaved,
}) async {
  final l10n = AppLocalizations.of(context)!;
  final sysController = TextEditingController();
  final diaController = TextEditingController();
  final siteController = TextEditingController();
  final commentController = TextEditingController();
  var method = 'doppler';
  var saving = false;
  String? error;

  try {
    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (sheetContext) {
        return StatefulBuilder(
          builder: (ctx, setModal) {
            final bottom =
                keyboardBottomInset(ctx) + systemBottomInset(ctx);
            return Padding(
              key: const Key('bp_sheet'),
              padding: EdgeInsets.only(
                left: 20,
                right: 20,
                top: 20,
                bottom: bottom + 20,
              ),
              child: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(
                      l10n.recordBloodPressureTitle,
                      style: Theme.of(ctx).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      key: const Key('bp_sys_field'),
                      controller: sysController,
                      enabled: !saving,
                      keyboardType: TextInputType.number,
                      inputFormatters: [
                        FilteringTextInputFormatter.digitsOnly,
                      ],
                      decoration: InputDecoration(
                        labelText: l10n.bpSystolicLabel,
                        border: const OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      key: const Key('bp_dia_field'),
                      controller: diaController,
                      enabled: !saving,
                      keyboardType: TextInputType.number,
                      inputFormatters: [
                        FilteringTextInputFormatter.digitsOnly,
                      ],
                      decoration: InputDecoration(
                        labelText: l10n.bpDiastolicLabel,
                        border: const OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 12),
                    DropdownButtonFormField<String>(
                      key: const Key('bp_method_field'),
                      value: method,
                      decoration: InputDecoration(
                        labelText: l10n.bpMethodLabel,
                        border: const OutlineInputBorder(),
                      ),
                      items: [
                        DropdownMenuItem(
                          value: 'doppler',
                          child: Text(l10n.bpMethodDoppler),
                        ),
                        DropdownMenuItem(
                          value: 'oscillometric',
                          child: Text(l10n.bpMethodOscillometric),
                        ),
                        DropdownMenuItem(
                          value: 'invasive',
                          child: Text(l10n.bpMethodInvasive),
                        ),
                        DropdownMenuItem(
                          value: 'unknown',
                          child: Text(l10n.bpMethodUnknown),
                        ),
                      ],
                      onChanged: saving
                          ? null
                          : (v) => setModal(() => method = v ?? 'unknown'),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      key: const Key('bp_site_field'),
                      controller: siteController,
                      enabled: !saving,
                      maxLength: 80,
                      decoration: InputDecoration(
                        labelText: l10n.bpSiteLabel,
                        border: const OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      key: const Key('bp_comment_field'),
                      controller: commentController,
                      enabled: !saving,
                      maxLength: 500,
                      maxLines: 2,
                      decoration: InputDecoration(
                        labelText: l10n.bpCommentLabel,
                        border: const OutlineInputBorder(),
                      ),
                    ),
                    if (error != null) ...[
                      const SizedBox(height: 8),
                      Text(
                        error!,
                        key: const Key('bp_error'),
                        style: const TextStyle(color: AppColors.alert),
                      ),
                    ],
                    const SizedBox(height: 12),
                    FilledButton(
                      key: const Key('bp_save_btn'),
                      onPressed: saving
                          ? null
                          : () async {
                              final sys = int.tryParse(
                                sysController.text.trim(),
                              );
                              final dia = int.tryParse(
                                diaController.text.trim(),
                              );
                              if (sys == null ||
                                  dia == null ||
                                  sys < 20 ||
                                  sys > 400 ||
                                  dia < 10 ||
                                  dia > 300 ||
                                  dia > sys) {
                                setModal(
                                  () => error = l10n.bloodPressureInvalid,
                                );
                                return;
                              }
                              setModal(() {
                                saving = true;
                                error = null;
                              });
                              try {
                                final site = siteController.text.trim();
                                final comment =
                                    commentController.text.trim();
                                await ApiClient.instance
                                    .createBloodPressureReading(
                                  petId,
                                  systolicMmHg: sys,
                                  diastolicMmHg: dia,
                                  method: method,
                                  site: site.isEmpty ? null : site,
                                  comment:
                                      comment.isEmpty ? null : comment,
                                );
                                if (sheetContext.mounted) {
                                  Navigator.of(sheetContext).pop();
                                }
                                if (context.mounted) {
                                  ScaffoldMessenger.of(context).showSnackBar(
                                    SnackBar(
                                      content: Text(l10n.bloodPressureSaved),
                                    ),
                                  );
                                }
                                onSaved?.call();
                              } catch (e) {
                                if (!ctx.mounted) return;
                                setModal(() {
                                  saving = false;
                                  final code = apiErrorCode(e);
                                  error = code == 'invalid_blood_pressure'
                                      ? l10n.bloodPressureInvalid
                                      : mapApiError(e, l10n);
                                });
                              }
                            },
                      child: saving
                          ? const SizedBox(
                              width: 20,
                              height: 20,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                              ),
                            )
                          : Text(l10n.bloodPressureSave),
                    ),
                  ],
                ),
              ),
            );
          },
        );
      },
    );
  } finally {
    sysController.dispose();
    diaController.dispose();
    siteController.dispose();
    commentController.dispose();
  }
}
