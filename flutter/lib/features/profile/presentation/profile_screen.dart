import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key});

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  final fullName = TextEditingController();
  final email = TextEditingController();
  final currentPassword = TextEditingController();
  final newPassword = TextEditingController();
  bool loading = true;
  bool saving = false;
  String? error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final me = await ApiClient.instance.getMe();
      fullName.text = me['fullName'] as String? ?? '';
      email.text = me['email'] as String? ?? '';
    } catch (_) {
      error = 'load';
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> _save() async {
    setState(() {
      saving = true;
      error = null;
    });
    try {
      await ApiClient.instance.updateMe(fullName.text.trim());
      if (mounted) {
        final l10n = AppLocalizations.of(context)!;
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(l10n.profileSaved)));
      }
    } catch (_) {
      setState(() => error = 'save');
    } finally {
      if (mounted) setState(() => saving = false);
    }
  }

  Future<void> _changePassword() async {
    if (currentPassword.text.isEmpty || newPassword.text.length < 8) return;
    setState(() {
      saving = true;
      error = null;
    });
    try {
      await ApiClient.instance.changePassword(currentPassword.text, newPassword.text);
      currentPassword.clear();
      newPassword.clear();
      if (mounted) {
        final l10n = AppLocalizations.of(context)!;
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(l10n.passwordChanged)));
      }
    } catch (_) {
      setState(() => error = 'password');
    } finally {
      if (mounted) setState(() => saving = false);
    }
  }

  Future<void> _deleteAccount() async {
    final l10n = AppLocalizations.of(context)!;
    final ok = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: Text(l10n.deleteAccount),
        content: Text(l10n.deleteAccountConfirm),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: Text(l10n.cancel)),
          FilledButton(onPressed: () => Navigator.pop(context, true), child: Text(l10n.deleteAccount)),
        ],
      ),
    );
    if (ok != true) return;
    try {
      await ApiClient.instance.deleteAccount();
      await ApiClient.instance.logout();
      if (mounted) Navigator.of(context).popUntil((r) => r.isFirst);
    } catch (_) {
      setState(() => error = 'delete');
    }
  }

  @override
  void dispose() {
    fullName.dispose();
    email.dispose();
    currentPassword.dispose();
    newPassword.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    if (loading) {
      return Scaffold(appBar: AppBar(title: Text(l10n.myData)), body: const Center(child: CircularProgressIndicator()));
    }
    return Scaffold(
      appBar: AppBar(title: Text(l10n.myData)),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          TextField(
            controller: fullName,
            decoration: InputDecoration(labelText: l10n.firstName),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: email,
            readOnly: true,
            decoration: InputDecoration(labelText: l10n.email),
          ),
          const SizedBox(height: 24),
          Text(l10n.changePassword, style: Theme.of(context).textTheme.titleSmall),
          const SizedBox(height: 8),
          TextField(
            controller: currentPassword,
            obscureText: true,
            decoration: InputDecoration(labelText: l10n.currentPassword),
          ),
          const SizedBox(height: 8),
          TextField(
            controller: newPassword,
            obscureText: true,
            decoration: InputDecoration(labelText: l10n.newPassword),
          ),
          const SizedBox(height: 8),
          OutlinedButton(onPressed: saving ? null : _changePassword, child: Text(l10n.changePassword)),
          if (error != null) ...[
            const SizedBox(height: 12),
            Text(l10n.errorGeneric(error!), style: const TextStyle(color: Colors.redAccent)),
          ],
          const SizedBox(height: 24),
          FilledButton(onPressed: saving ? null : _save, child: Text(l10n.save)),
          const SizedBox(height: 32),
          Center(
            child: TextButton(
              onPressed: _deleteAccount,
              child: Text(l10n.deleteAccount, style: const TextStyle(color: Colors.redAccent)),
            ),
          ),
        ],
      ),
    );
  }
}
