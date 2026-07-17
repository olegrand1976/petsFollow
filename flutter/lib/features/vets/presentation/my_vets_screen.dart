import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/vet_link.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class MyVetsScreen extends StatefulWidget {
  const MyVetsScreen({super.key});

  @override
  State<MyVetsScreen> createState() => _MyVetsScreenState();
}

class _MyVetsScreenState extends State<MyVetsScreen> {
  List<VetLink> vets = [];
  bool loading = true;
  final emailCtrl = TextEditingController();
  bool inviting = false;

  @override
  void initState() {
    super.initState();
    load();
  }

  @override
  void dispose() {
    emailCtrl.dispose();
    super.dispose();
  }

  Future<void> load() async {
    setState(() => loading = true);
    try {
      final data = await ApiClient.instance.getMyVets();
      if (mounted) setState(() => vets = data);
    } catch (_) {}
    if (mounted) setState(() => loading = false);
  }

  Future<void> invite() async {
    final l10n = AppLocalizations.of(context)!;
    final email = emailCtrl.text.trim();
    if (email.isEmpty) return;
    setState(() => inviting = true);
    try {
      final result = await ApiClient.instance.inviteVet(email);
      final found = result['found'] == true;
      if (!mounted) return;
      if (found) {
        emailCtrl.clear();
        await load();
        if (!mounted) return;
        final practice = (result['practiceName'] as String?)?.trim() ?? '';
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              practice.isEmpty ? l10n.vetInviteSent : l10n.vetInviteSentNamed(practice),
            ),
          ),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.vetNotFound)),
        );
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorGeneric('invite'))),
        );
      }
    }
    if (mounted) setState(() => inviting = false);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.myVets)),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: load,
              child: ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  TextField(
                    controller: emailCtrl,
                    keyboardType: TextInputType.emailAddress,
                    autocorrect: false,
                    decoration: InputDecoration(
                      labelText: l10n.addVetByEmail,
                      hintText: l10n.vetEmailHint,
                      suffixIcon: inviting
                          ? const Padding(
                              padding: EdgeInsets.all(12),
                              child: SizedBox(
                                width: 20,
                                height: 20,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              ),
                            )
                          : IconButton(
                              icon: const Icon(Icons.person_add),
                              onPressed: invite,
                              tooltip: l10n.homeAddFirstVetCta,
                            ),
                    ),
                    onSubmitted: (_) => invite(),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    l10n.addVetSearchHint,
                    style: TextStyle(color: AppColors.textMuted, fontSize: 13, height: 1.35),
                  ),
                  const SizedBox(height: 20),
                  if (vets.isEmpty)
                    Center(
                      child: Padding(
                        padding: const EdgeInsets.all(32),
                        child: Text(l10n.noVets, style: TextStyle(color: AppColors.textMuted)),
                      ),
                    )
                  else
                    ...vets.map(
                      (v) => Card(
                        margin: const EdgeInsets.only(bottom: 8),
                        child: ListTile(
                          leading: const Icon(Icons.local_hospital_outlined, color: AppColors.primary),
                          title: Text(v.practiceName),
                          subtitle: Text('${v.vetFullName}\n${v.vetEmail}'),
                          isThreeLine: true,
                          trailing: v.isPrimary
                              ? Chip(
                                  label: Text(l10n.primaryVet, style: const TextStyle(fontSize: 11)),
                                  backgroundColor: AppColors.gold.withValues(alpha: 0.15),
                                )
                              : null,
                        ),
                      ),
                    ),
                ],
              ),
            ),
    );
  }
}
