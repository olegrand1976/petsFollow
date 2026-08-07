import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Practical tips for the household pets (fail-soft; empty → shrink).
class PetTipsSection extends StatefulWidget {
  const PetTipsSection({super.key, required this.l10n, this.reloadEpoch = 0});

  final AppLocalizations l10n;
  final int reloadEpoch;

  @override
  State<PetTipsSection> createState() => _PetTipsSectionState();
}

class _PetTipsSectionState extends State<PetTipsSection> {
  List<Map<String, dynamic>>? _items;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void didUpdateWidget(covariant PetTipsSection oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.reloadEpoch != widget.reloadEpoch) {
      _load();
    }
  }

  Future<void> _load() async {
    try {
      final raw = await ApiClient.instance.getPetTips();
      if (!mounted) return;
      setState(() {
        _items = raw
            .map((e) => Map<String, dynamic>.from(e as Map))
            .where((m) => '${m['title'] ?? ''}'.trim().isNotEmpty)
            .toList();
      });
    } catch (_) {
      if (mounted) setState(() => _items = null);
    }
  }

  @override
  Widget build(BuildContext context) {
    final items = _items;
    if (items == null || items.isEmpty) return const SizedBox.shrink();
    final p = PetsPalette.of(context);
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Container(
        key: const Key('home_pet_tips'),
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: p.surfaceElevated,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: AppColors.primary.withValues(alpha: 0.25)),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              widget.l10n.homePetTipsTitle,
              key: const Key('home_pet_tips_title'),
              style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
            ),
            const SizedBox(height: 4),
            Text(
              widget.l10n.homePetTipsDisclaimer,
              style: TextStyle(color: p.textMuted, fontSize: 11, height: 1.3),
            ),
            const SizedBox(height: 10),
            ...items.map((item) {
              final id = '${item['id'] ?? ''}';
              final title = '${item['title'] ?? ''}';
              final body = '${item['body'] ?? ''}';
              return Padding(
                key: Key('home_pet_tips_item_$id'),
                padding: const EdgeInsets.only(bottom: 10),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      title,
                      style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      body,
                      style: TextStyle(color: p.textMuted, fontSize: 12, height: 1.35),
                    ),
                  ],
                ),
              );
            }),
          ],
        ),
      ),
    );
  }
}
