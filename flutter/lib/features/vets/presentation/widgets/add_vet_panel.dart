import 'dart:async';

import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/vet_lookup_hit.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Autocomplete lookup + invite, with fallback « nouveau véto » (email + téléphone).
class AddVetPanel extends StatefulWidget {
  const AddVetPanel({
    super.key,
    this.compact = false,
    this.onLinked,
    this.onSuggested,
  });

  /// Home card: slightly denser copy / filled search field.
  final bool compact;
  final VoidCallback? onLinked;
  final VoidCallback? onSuggested;

  @override
  State<AddVetPanel> createState() => _AddVetPanelState();
}

class _AddVetPanelState extends State<AddVetPanel> {
  final _searchCtrl = TextEditingController();
  final _emailCtrl = TextEditingController();
  final _phoneCtrl = TextEditingController();
  final _nameCtrl = TextEditingController();

  Timer? _debounce;
  List<VetLookupHit> _hits = [];
  bool _searching = false;
  bool _inviting = false;
  bool _suggesting = false;
  bool _showSuggestForm = false;
  String? _lastQuery;

  @override
  void dispose() {
    _debounce?.cancel();
    _searchCtrl.dispose();
    _emailCtrl.dispose();
    _phoneCtrl.dispose();
    _nameCtrl.dispose();
    super.dispose();
  }

  void _onQueryChanged(String raw) {
    _debounce?.cancel();
    final q = raw.trim();
    if (q.length < 2) {
      setState(() {
        _hits = [];
        _searching = false;
        _lastQuery = q;
      });
      return;
    }
    setState(() {
      _searching = true;
      _lastQuery = q;
    });
    _debounce = Timer(const Duration(milliseconds: 320), () => _runLookup(q));
  }

  Future<void> _runLookup(String q) async {
    try {
      final hits = await ApiClient.instance.lookupVets(q);
      if (!mounted || _lastQuery != q) return;
      setState(() {
        _hits = hits;
        _searching = false;
      });
    } catch (_) {
      if (!mounted || _lastQuery != q) return;
      setState(() {
        _hits = [];
        _searching = false;
      });
    }
  }

  Future<void> _inviteHit(VetLookupHit hit) async {
    final l10n = AppLocalizations.of(context)!;
    if (_inviting) return;
    setState(() => _inviting = true);
    try {
      final result = await ApiClient.instance.inviteVet(vetUserId: hit.vetUserId);
      if (!mounted) return;
      if (result['found'] == true) {
        _searchCtrl.clear();
        setState(() {
          _hits = [];
          _showSuggestForm = false;
        });
        final practice = (result['practiceName'] as String?)?.trim() ?? hit.practiceName;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              practice.isEmpty ? l10n.vetInviteSent : l10n.vetInviteSentNamed(practice),
            ),
          ),
        );
        widget.onLinked?.call();
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
    if (mounted) setState(() => _inviting = false);
  }

  Future<void> _submitSuggest() async {
    final l10n = AppLocalizations.of(context)!;
    final email = _emailCtrl.text.trim();
    final phone = _phoneCtrl.text.trim();
    if (email.isEmpty || phone.isEmpty || _suggesting) return;
    setState(() => _suggesting = true);
    try {
      final result = await ApiClient.instance.suggestVet(
        email: email,
        phone: phone,
        fullName: _nameCtrl.text.trim().isEmpty ? null : _nameCtrl.text.trim(),
      );
      if (!mounted) return;
      if (result['found'] == true || result['status'] == 'invited') {
        final practice = (result['practiceName'] as String?)?.trim() ?? '';
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              practice.isEmpty ? l10n.vetInviteSent : l10n.vetInviteSentNamed(practice),
            ),
          ),
        );
        _emailCtrl.clear();
        _phoneCtrl.clear();
        _nameCtrl.clear();
        setState(() => _showSuggestForm = false);
        widget.onLinked?.call();
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.vetSuggestSent)),
        );
        _emailCtrl.clear();
        _phoneCtrl.clear();
        _nameCtrl.clear();
        setState(() => _showSuggestForm = false);
        widget.onSuggested?.call();
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorGeneric('suggest'))),
        );
      }
    }
    if (mounted) setState(() => _suggesting = false);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final p = PetsPalette.of(context);
    final busy = _inviting || _suggesting;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextField(
          key: const Key('add_vet_search'),
          controller: _searchCtrl,
          enabled: !busy,
          autocorrect: false,
          textInputAction: TextInputAction.search,
          decoration: InputDecoration(
            labelText: l10n.addVetSearchLabel,
            hintText: l10n.addVetSearchFieldHint,
            filled: widget.compact,
            fillColor: widget.compact ? p.surface : null,
            suffixIcon: _searching
                ? const Padding(
                    padding: EdgeInsets.all(12),
                    child: SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                  )
                : const Icon(Icons.search),
          ),
          onChanged: _onQueryChanged,
        ),
        const SizedBox(height: 8),
        Text(
          l10n.addVetSearchHint,
          style: TextStyle(
            color: p.textMuted,
            fontSize: widget.compact ? 12 : 13,
            height: 1.35,
          ),
        ),
        if (_hits.isNotEmpty) ...[
          const SizedBox(height: 12),
          ..._hits.map(
            (h) => Card(
              margin: const EdgeInsets.only(bottom: 6),
              child: ListTile(
                key: Key('add_vet_hit_${h.vetUserId}'),
                leading: const Icon(Icons.local_hospital_outlined, color: AppColors.primary),
                title: Text(h.practiceName.isEmpty ? h.vetFullName : h.practiceName),
                subtitle: Text(
                  [
                    if (h.vetFullName.isNotEmpty) h.vetFullName,
                    if (h.vetEmail.isNotEmpty) h.vetEmail,
                  ].join('\n'),
                ),
                isThreeLine: h.vetFullName.isNotEmpty && h.vetEmail.isNotEmpty,
                trailing: _inviting
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.person_add_alt_1),
                onTap: busy ? null : () => _inviteHit(h),
              ),
            ),
          ),
        ],
        if (!_showSuggestForm) ...[
          const SizedBox(height: 8),
          TextButton(
            key: const Key('add_vet_not_listed'),
            onPressed: busy
                ? null
                : () => setState(() {
                      _showSuggestForm = true;
                      if (_emailCtrl.text.isEmpty && _searchCtrl.text.contains('@')) {
                        _emailCtrl.text = _searchCtrl.text.trim();
                      }
                    }),
            child: Text(l10n.addVetNotListed),
          ),
        ] else ...[
          const SizedBox(height: 16),
          Text(
            l10n.addVetSuggestTitle,
            style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 4),
          Text(
            l10n.addVetSuggestBody,
            style: TextStyle(color: p.textMuted, fontSize: 13, height: 1.35),
          ),
          const SizedBox(height: 12),
          TextField(
            key: const Key('add_vet_suggest_email'),
            controller: _emailCtrl,
            enabled: !busy,
            keyboardType: TextInputType.emailAddress,
            autocorrect: false,
            decoration: InputDecoration(
              labelText: l10n.addVetSuggestEmail,
              hintText: l10n.vetEmailHint,
              filled: widget.compact,
              fillColor: widget.compact ? p.surface : null,
            ),
          ),
          const SizedBox(height: 10),
          TextField(
            key: const Key('add_vet_suggest_phone'),
            controller: _phoneCtrl,
            enabled: !busy,
            keyboardType: TextInputType.phone,
            decoration: InputDecoration(
              labelText: l10n.addVetSuggestPhone,
              filled: widget.compact,
              fillColor: widget.compact ? p.surface : null,
            ),
          ),
          const SizedBox(height: 10),
          TextField(
            key: const Key('add_vet_suggest_name'),
            controller: _nameCtrl,
            enabled: !busy,
            textCapitalization: TextCapitalization.words,
            decoration: InputDecoration(
              labelText: l10n.addVetSuggestNameOptional,
              filled: widget.compact,
              fillColor: widget.compact ? p.surface : null,
            ),
          ),
          const SizedBox(height: 14),
          FilledButton.icon(
            key: const Key('add_vet_suggest_submit'),
            onPressed: busy ? null : _submitSuggest,
            icon: _suggesting
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2, color: AppColors.bg),
                  )
                : const Icon(Icons.send_outlined),
            label: Text(l10n.addVetSuggestCta),
          ),
        ],
      ],
    );
  }
}
