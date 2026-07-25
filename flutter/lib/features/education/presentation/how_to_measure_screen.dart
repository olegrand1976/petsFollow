import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class HowToMeasureScreen extends StatelessWidget {
  const HowToMeasureScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.howToMeasure)),
      body: SingleChildScrollView(
        padding: scrollPaddingWithSystemBottom(context, all: 20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(l10n.howToMeasureIntro, style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 16),
            Text(l10n.howToMeasureStep1, style: const TextStyle(height: 1.5)),
            const SizedBox(height: 12),
            Text(l10n.howToMeasureStep2, style: const TextStyle(height: 1.5)),
            const SizedBox(height: 12),
            Text(l10n.howToMeasureStep3, style: const TextStyle(height: 1.5)),
            const SizedBox(height: 24),
            Text(l10n.howToMeasureWhyTitle, style: Theme.of(context).textTheme.titleSmall),
            const SizedBox(height: 8),
            Text(l10n.howToMeasureWhyBody, style: const TextStyle(height: 1.5)),
          ],
        ),
      ),
    );
  }
}
