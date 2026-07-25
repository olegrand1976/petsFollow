import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Shows the durable app-invite QR (vet / care_pro / commercial).
class AppInviteQrScreen extends StatefulWidget {
  const AppInviteQrScreen({super.key});

  @override
  State<AppInviteQrScreen> createState() => _AppInviteQrScreenState();
}

class _AppInviteQrScreenState extends State<AppInviteQrScreen> {
  bool _loading = true;
  String? _error;
  Map<String, dynamic>? _invite;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final data = await ApiClient.instance.getMyAppInvite();
      if (!mounted) return;
      setState(() {
        _invite = data;
        _loading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _error = AppLocalizations.of(context)!.appInviteLoadError;
        _loading = false;
      });
    }
  }

  Uint8List? _qrBytes() {
    final raw = _invite?['qrCodeDataUrl'] as String?;
    if (raw == null || !raw.startsWith('data:image')) return null;
    final comma = raw.indexOf(',');
    if (comma < 0) return null;
    try {
      return base64Decode(raw.substring(comma + 1));
    } catch (_) {
      return null;
    }
  }

  Future<void> _copyLink() async {
    final url = _invite?['inviteUrl'] as String? ?? '';
    if (url.isEmpty) return;
    await Clipboard.setData(ClipboardData(text: url));
    if (!mounted) return;
    final l10n = AppLocalizations.of(context)!;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(l10n.appInviteCopied)),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final bytes = _qrBytes();
    final practice = _invite?['practiceName'] as String? ?? '';
    final name = (_invite?['displayName'] as String?) ??
        (_invite?['vetFullName'] as String?) ??
        '';
    final code = _invite?['code'] as String? ?? '';

    return Scaffold(
      appBar: AppBar(title: Text(l10n.appInviteTitle)),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(
                  child: Padding(
                    padding: const EdgeInsets.all(24),
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Text(_error!, textAlign: TextAlign.center),
                        const SizedBox(height: 12),
                        FilledButton(onPressed: _load, child: Text(l10n.appInviteRetry)),
                      ],
                    ),
                  ),
                )
              : ListView(
                  padding: const EdgeInsets.all(24),
                  children: [
                    Text(l10n.appInviteHint, style: Theme.of(context).textTheme.bodyMedium),
                    const SizedBox(height: 20),
                    if (bytes != null)
                      Center(
                        child: DecoratedBox(
                          decoration: BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(color: AppColors.textMuted),
                          ),
                          child: Padding(
                            padding: const EdgeInsets.all(12),
                            child: Image.memory(bytes, width: 220, height: 220),
                          ),
                        ),
                      ),
                    const SizedBox(height: 16),
                    if (name.isNotEmpty || practice.isNotEmpty)
                      Text(
                        [name, practice].where((s) => s.isNotEmpty).join(' — '),
                        textAlign: TextAlign.center,
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                    if (code.isNotEmpty) ...[
                      const SizedBox(height: 8),
                      Text(
                        '${l10n.appInviteCodeLabel} $code',
                        textAlign: TextAlign.center,
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ],
                    const SizedBox(height: 24),
                    FilledButton.icon(
                      onPressed: _copyLink,
                      icon: const Icon(Icons.copy),
                      label: Text(l10n.appInviteCopy),
                    ),
                  ],
                ),
    );
  }
}
