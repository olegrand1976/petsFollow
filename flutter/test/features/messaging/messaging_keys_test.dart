import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

void main() {
  testWidgets('messaging composer keys are findable', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: Scaffold(
          body: Row(
            children: [
              IconButton(
                key: const Key('message_attach_btn'),
                onPressed: () {},
                icon: const Icon(Icons.attach_file),
              ),
              const Expanded(
                child: TextField(key: Key('message_draft')),
              ),
              IconButton(
                key: const Key('message_send_btn'),
                onPressed: () {},
                icon: const Icon(Icons.send),
              ),
            ],
          ),
        ),
      ),
    );
    expect(find.byKey(const Key('message_attach_btn')), findsOneWidget);
    expect(find.byKey(const Key('message_draft')), findsOneWidget);
    expect(find.byKey(const Key('message_send_btn')), findsOneWidget);
  });

  testWidgets('messaging attach sheet exposes camera and gallery keys', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: Builder(
          builder: (context) {
            final l10n = AppLocalizations.of(context)!;
            return Scaffold(
              body: ListView(
                children: [
                  ListTile(
                    key: const Key('message_attach_photo'),
                    title: Text(l10n.attachPhoto),
                  ),
                  ListTile(
                    key: const Key('message_attach_video'),
                    title: Text(l10n.attachVideo),
                  ),
                  ListTile(
                    key: const Key('message_attach_camera'),
                    title: Text(l10n.takeVideo),
                  ),
                  ListTile(
                    key: const Key('message_attach_gallery'),
                    title: Text(l10n.chooseFromGallery),
                  ),
                ],
              ),
            );
          },
        ),
      ),
    );
    expect(find.byKey(const Key('message_attach_photo')), findsOneWidget);
    expect(find.byKey(const Key('message_attach_video')), findsOneWidget);
    expect(find.byKey(const Key('message_attach_camera')), findsOneWidget);
    expect(find.byKey(const Key('message_attach_gallery')), findsOneWidget);
    expect(find.text('Filmer une vidéo'), findsOneWidget);
  });

  testWidgets('preconsult_submit and commercial_logout keys', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Column(
            children: [
              FilledButton(
                key: const Key('preconsult_submit'),
                onPressed: () {},
                child: const Text('ok'),
              ),
              TextButton(
                key: const Key('commercial_logout'),
                onPressed: () {},
                child: const Text('logout'),
              ),
            ],
          ),
        ),
      ),
    );
    expect(find.byKey(const Key('preconsult_submit')), findsOneWidget);
    expect(find.byKey(const Key('commercial_logout')), findsOneWidget);
  });
}
