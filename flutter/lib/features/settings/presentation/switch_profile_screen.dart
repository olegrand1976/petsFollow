import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class SwitchProfileScreen extends StatefulWidget {
  const SwitchProfileScreen({super.key});

  @override
  State<SwitchProfileScreen> createState() => _SwitchProfileScreenState();
}

class _SwitchProfileScreenState extends State<SwitchProfileScreen> {
  List<Map<String, dynamic>> _profiles = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final list = await ApiClient.instance.getProfiles();
      if (mounted) {
        setState(() {
          _profiles = list;
          _loading = false;
        });
      }
    } catch (_) {
      if (mounted) setState(() => _loading = false);
    }
  }

  String _label(AppLocalizations l10n, Map<String, dynamic> p) {
    final role = p['role'] as String? ?? '';
    if (role == 'client') return l10n.profilePersonal;
    return '${l10n.profilePro} ($role)';
  }

  Future<void> _switchTo(String id) async {
    await ApiClient.instance.switchProfile(id);
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(AppLocalizations.of(context)!.profileSwitched)),
    );
    Navigator.of(context).popUntil((r) => r.isFirst);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.switchProfile)),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : ListView.builder(
              itemCount: _profiles.length,
              itemBuilder: (context, i) {
                final p = _profiles[i];
                final active = p['active'] == true;
                return ListTile(
                  title: Text(_label(l10n, p)),
                  subtitle: Text(p['role'] as String? ?? ''),
                  trailing: active
                      ? const Icon(Icons.check_circle, color: Colors.green)
                      : const Icon(Icons.swap_horiz),
                  onTap: active ? null : () => _switchTo(p['id'] as String),
                );
              },
            ),
    );
  }
}
