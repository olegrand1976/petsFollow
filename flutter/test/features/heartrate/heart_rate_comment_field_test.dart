import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

void main() {
  testWidgets('hr comment field is present with key', (tester) async {
    final controller = TextEditingController();
    addTearDown(controller.dispose);

    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: Scaffold(
          body: Builder(
            builder: (context) {
              final l10n = AppLocalizations.of(context)!;
              return TextField(
                key: const Key('hr_comment_field'),
                controller: controller,
                maxLength: 500,
                maxLines: 3,
                decoration: InputDecoration(
                  labelText: l10n.heartRateCommentLabel,
                  hintText: l10n.heartRateCommentHint,
                  border: const OutlineInputBorder(),
                ),
              );
            },
          ),
        ),
      ),
    );

    expect(find.byKey(const Key('hr_comment_field')), findsOneWidget);
    await tester.enterText(find.byKey(const Key('hr_comment_field')), 'ok');
    expect(controller.text, 'ok');
  });
}
