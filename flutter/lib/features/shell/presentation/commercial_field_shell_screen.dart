import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
import 'package:petsfollow_mobile/features/invite/presentation/app_invite_qr_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:url_launcher/url_launcher.dart';

/// Minimal field shell for commercial / commercial_manager: QR + open Pro web.
class CommercialFieldShellScreen extends StatefulWidget {
  const CommercialFieldShellScreen({super.key, required this.onLogout});

  final VoidCallback onLogout;

  @override
  State<CommercialFieldShellScreen> createState() =>
      _CommercialFieldShellScreenState();
}

class _CommercialFieldShellScreenState extends State<CommercialFieldShellScreen> {
  static const _proSiteDefined = String.fromEnvironment('PRO_PUBLIC_SITE_URL');

  String _proSiteUrl = _proSiteDefined.isNotEmpty
      ? _proSiteDefined
      : 'http://localhost:3002';

  @override
  void initState() {
    super.initState();
    _resolveProSiteUrl();
  }

  Future<void> _resolveProSiteUrl() async {
    if (_proSiteDefined.isNotEmpty) return;
    try {
      final data = await ApiClient.instance.getMyAppInvite();
      final fromApi = (data['proSiteUrl'] as String?)?.trim() ?? '';
      if (fromApi.isEmpty || !mounted) return;
      setState(() => _proSiteUrl = fromApi);
    } catch (_) {
      // Keep compile-time / localhost fallback.
    }
  }

  Future<void> _openProWeb() async {
    final uri = Uri.parse(_proSiteUrl);
    await launchUrl(uri, mode: LaunchMode.externalApplication);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Container(
      decoration: const BoxDecoration(gradient: AppTheme.loginGradient),
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          title: Row(
            children: [
              const PetsLogo(height: 28),
              const SizedBox(width: 8),
              Text(l10n.commercialFieldTitle),
            ],
          ),
        ),
        body: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                l10n.commercialFieldSubtitle,
                style: Theme.of(context).textTheme.bodyLarge,
              ),
              const SizedBox(height: 24),
              FilledButton.icon(
                onPressed: () {
                  Navigator.of(context).push(
                    MaterialPageRoute<void>(
                      builder: (_) => const AppInviteQrScreen(),
                    ),
                  );
                },
                icon: const Icon(Icons.qr_code_2),
                label: Text(l10n.appInviteTitle),
              ),
              const SizedBox(height: 12),
              OutlinedButton.icon(
                onPressed: _openProWeb,
                icon: const Icon(Icons.open_in_browser),
                label: Text(l10n.commercialOpenProWeb),
              ),
              const Spacer(),
              TextButton(
                onPressed: () async {
                  await ApiClient.instance.logout();
                  widget.onLogout();
                },
                child: Text(l10n.logout),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
