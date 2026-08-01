import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/api/open_url.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/core/ui/load_error_view.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/features/messaging/presentation/messaging_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/book_visit_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Result of the pet picker bottom sheet.
enum _PetPickKind { pet, none, cancelled }

class _PetPick {
  const _PetPick.pet(this.pet) : kind = _PetPickKind.pet;
  const _PetPick.none() : pet = null, kind = _PetPickKind.none;
  const _PetPick.cancelled() : pet = null, kind = _PetPickKind.cancelled;

  final _PetPickKind kind;
  final Pet? pet;
}

/// Conversational 24/7 triage (green / orange / red) — separate from vet messaging.
class TriageChatScreen extends StatefulWidget {
  const TriageChatScreen({
    super.key,
    this.initialPetId,
    this.initialPetName,
    @visibleForTesting this.petsOverride,
  });

  final String? initialPetId;
  final String? initialPetName;

  /// Test-only pet list (skips network).
  final List<Pet>? petsOverride;

  @override
  State<TriageChatScreen> createState() => _TriageChatScreenState();
}

class _TriageChatScreenState extends State<TriageChatScreen> {
  final _draft = TextEditingController();
  final _scroll = ScrollController();

  bool bootLoading = true;
  bool sending = false;
  String? bootError;
  String? sessionId;
  String? petId;
  String? petName;
  Map<String, dynamic> escalation = {};
  List<Map<String, dynamic>> messages = [];
  String? lastLevel;
  List<String> lastWatchSigns = [];
  List<Pet> _pets = [];

  @override
  void initState() {
    super.initState();
    petId = widget.initialPetId;
    petName = widget.initialPetName;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) _bootstrap();
    });
  }

  @override
  void dispose() {
    _draft.dispose();
    _scroll.dispose();
    super.dispose();
  }

  Future<void> _bootstrap() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() {
      bootLoading = true;
      bootError = null;
    });
    try {
      if (widget.petsOverride != null) {
        _pets = widget.petsOverride!;
      } else {
        final raw = await ApiClient.instance.getPets();
        _pets = raw
            .whereType<Map>()
            .map((e) => Pet.fromJson(Map<String, dynamic>.from(e)))
            .toList();
      }
      if (!mounted) return;

      if (petId == null || petId!.isEmpty) {
        if (_pets.length == 1) {
          petId = _pets.first.id;
          petName = _pets.first.name;
        } else if (_pets.isNotEmpty) {
          setState(() => bootLoading = false);
          final pick = await _pickPet();
          if (!mounted) return;
          switch (pick.kind) {
            case _PetPickKind.cancelled:
              if (mounted) Navigator.of(context).maybePop();
              return;
            case _PetPickKind.none:
              break;
            case _PetPickKind.pet:
              petId = pick.pet!.id;
              petName = pick.pet!.name;
          }
          setState(() => bootLoading = true);
        }
      }

      final raw = await ApiClient.instance.createClientAiTriageSession(petId: petId);
      if (!mounted) return;
      final sess = raw['session'];
      final sid = sess is Map ? sess['id']?.toString() : null;
      if (sid == null || sid.isEmpty) {
        setState(() {
          bootLoading = false;
          bootError = l10n.errorNetwork;
        });
        return;
      }
      sessionId = sid;
      escalation = Map<String, dynamic>.from(
        raw['escalation'] is Map ? raw['escalation'] as Map : {},
      );
      if ((raw['petName'] as String?)?.trim().isNotEmpty == true) {
        petName = raw['petName'] as String;
      }
      setState(() => bootLoading = false);
    } catch (e) {
      if (!mounted) return;
      setState(() {
        bootLoading = false;
        bootError = mapApiError(e, l10n);
      });
    }
  }

  Future<_PetPick> _pickPet() async {
    final l10n = AppLocalizations.of(context)!;
    final result = await showModalBottomSheet<_PetPick>(
      context: context,
      builder: (ctx) {
        return SafeArea(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              ListTile(title: Text(l10n.clientAiTriageSelectPet)),
              ..._pets.map(
                (p) => ListTile(
                  key: Key('client_ai_triage_pet_${p.id}'),
                  title: Text(p.name),
                  onTap: () => Navigator.pop(ctx, _PetPick.pet(p)),
                ),
              ),
              ListTile(
                key: const Key('client_ai_triage_no_pet'),
                title: Text(l10n.clientAiTriageNoPet),
                onTap: () => Navigator.pop(ctx, const _PetPick.none()),
              ),
            ],
          ),
        );
      },
    );
    return result ?? const _PetPick.cancelled();
  }

  Future<void> _send() async {
    final l10n = AppLocalizations.of(context)!;
    final text = _draft.text.trim();
    if (text.isEmpty || sessionId == null || sending) return;
    setState(() => sending = true);
    try {
      final raw =
          await ApiClient.instance.postClientAiTriageMessage(sessionId!, text);
      if (!mounted) return;
      final userMsg = Map<String, dynamic>.from(
        raw['userMessage'] is Map ? raw['userMessage'] as Map : {},
      );
      final asstMsg = Map<String, dynamic>.from(
        raw['assistantMessage'] is Map ? raw['assistantMessage'] as Map : {},
      );
      escalation = Map<String, dynamic>.from(
        raw['escalation'] is Map ? raw['escalation'] as Map : escalation,
      );
      lastLevel = (raw['level'] as String?)?.toLowerCase();
      final signs = raw['watchSigns'];
      lastWatchSigns = signs is List
          ? signs.map((e) => e.toString()).where((s) => s.trim().isNotEmpty).toList()
          : <String>[];
      _draft.clear();
      setState(() {
        if (userMsg.isNotEmpty) messages = [...messages, userMsg];
        if (asstMsg.isNotEmpty) messages = [...messages, asstMsg];
        sending = false;
      });
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (_scroll.hasClients) {
          _scroll.animateTo(
            _scroll.position.maxScrollExtent,
            duration: const Duration(milliseconds: 250),
            curve: Curves.easeOut,
          );
        }
      });
    } catch (e) {
      if (!mounted) return;
      // Keep draft text so the owner can retry after a Gemini/network failure.
      if (_draft.text.trim().isEmpty) {
        _draft.text = text;
        _draft.selection = TextSelection.collapsed(offset: text.length);
      }
      setState(() => sending = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(mapApiError(e, l10n))),
      );
    }
  }

  Future<void> _callPractice() async {
    final phone = (escalation['practicePhone'] as String?)?.trim() ?? '';
    if (phone.isEmpty) return;
    final digits = phone.replaceAll(RegExp(r'[^\d+]'), '');
    if (digits.isEmpty) return;
    await openExternalUrl('tel:$digits');
  }

  void _openMessaging() {
    final id = petId ?? (escalation['petId'] as String?)?.toString();
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => MessagingScreen(
          initialPetId: (id != null && id.isNotEmpty) ? id : null,
        ),
      ),
    );
  }

  void _openBookVisit() {
    final id = petId ?? (escalation['petId'] as String?)?.toString() ?? '';
    if (id.isEmpty) return;
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => BookVisitScreen(
          petId: id,
          petName: petName ?? '',
          practiceIdFilter: (escalation['practiceId'] as String?)?.toString(),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final p = PetsPalette.of(context);

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.clientAiTriageTitle),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 12),
            child: Chip(
              key: const Key('client_ai_triage_dev_badge'),
              label: Text(l10n.clientAiDevBadge),
              visualDensity: VisualDensity.compact,
              materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
            ),
          ),
        ],
      ),
      body: bootLoading
          ? const Center(child: CircularProgressIndicator())
          : bootError != null
              ? LoadErrorView(message: bootError!, onRetry: _bootstrap)
              : Column(
                  children: [
                    Padding(
                      padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                      child: Text(
                        l10n.clientAiTriageSubtitle,
                        style: TextStyle(color: p.textMuted),
                      ),
                    ),
                    if ((petName ?? '').isNotEmpty)
                      Padding(
                        padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
                        child: Align(
                          alignment: Alignment.centerLeft,
                          child: Chip(label: Text(petName!)),
                        ),
                      ),
                    Expanded(
                      child: ListView.builder(
                        controller: _scroll,
                        padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
                        itemCount: messages.length,
                        itemBuilder: (context, i) {
                          final m = messages[i];
                          final mine = m['role'] == 'user';
                          final body = (m['body'] as String?)?.trim() ?? '';
                          final level = (m['level'] as String?)?.toLowerCase();
                          return Align(
                            alignment:
                                mine ? Alignment.centerRight : Alignment.centerLeft,
                            child: Container(
                              key: Key('client_ai_triage_msg_$i'),
                              margin: const EdgeInsets.only(bottom: 8),
                              padding: const EdgeInsets.symmetric(
                                horizontal: 14,
                                vertical: 10,
                              ),
                              constraints: BoxConstraints(
                                maxWidth:
                                    MediaQuery.sizeOf(context).width * 0.82,
                              ),
                              decoration: BoxDecoration(
                                color: mine
                                    ? AppColors.primary.withValues(alpha: 0.15)
                                    : p.surfaceElevated,
                                borderRadius: BorderRadius.circular(14),
                              ),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  if (!mine && level != null && level.isNotEmpty)
                                    Padding(
                                      padding: const EdgeInsets.only(bottom: 6),
                                      child: _LevelBadge(level: level, l10n: l10n),
                                    ),
                                  Text(body),
                                ],
                              ),
                            ),
                          );
                        },
                      ),
                    ),
                    if (lastLevel != null)
                      _EscalationBar(
                        level: lastLevel!,
                        watchSigns: lastWatchSigns,
                        escalation: escalation,
                        l10n: l10n,
                        onCall: _callPractice,
                        onBook: _openBookVisit,
                        onMessage: _openMessaging,
                      ),
                    SafeArea(
                      top: false,
                      child: Padding(
                        padding: EdgeInsets.fromLTRB(
                          12,
                          8,
                          12,
                          composerBottomPadding(context, embedded: false),
                        ),
                        child: Row(
                          children: [
                            Expanded(
                              child: TextField(
                                key: const Key('client_ai_triage_composer'),
                                controller: _draft,
                                minLines: 1,
                                maxLines: 4,
                                textInputAction: TextInputAction.send,
                                onSubmitted: (_) => _send(),
                                decoration: InputDecoration(
                                  hintText: l10n.clientAiTriageHint,
                                  border: const OutlineInputBorder(),
                                  isDense: true,
                                ),
                              ),
                            ),
                            const SizedBox(width: 8),
                            IconButton.filled(
                              key: const Key('client_ai_triage_send'),
                              onPressed: sending ? null : _send,
                              icon: sending
                                  ? const SizedBox(
                                      width: 18,
                                      height: 18,
                                      child: CircularProgressIndicator(
                                        strokeWidth: 2,
                                      ),
                                    )
                                  : const Icon(Icons.send),
                              tooltip: l10n.clientAiTriageSend,
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
    );
  }
}

class _LevelBadge extends StatelessWidget {
  const _LevelBadge({required this.level, required this.l10n});

  final String level;
  final AppLocalizations l10n;

  @override
  Widget build(BuildContext context) {
    final (label, color) = switch (level) {
      'red' => (l10n.clientAiTriageLevelRed, Colors.red.shade700),
      'orange' => (l10n.clientAiTriageLevelOrange, Colors.orange.shade800),
      _ => (l10n.clientAiTriageLevelGreen, Colors.green.shade700),
    };
    return Text(
      label,
      key: Key('client_ai_triage_level_$level'),
      style: TextStyle(
        color: color,
        fontWeight: FontWeight.w600,
        fontSize: 12,
      ),
    );
  }
}

class _EscalationBar extends StatelessWidget {
  const _EscalationBar({
    required this.level,
    required this.watchSigns,
    required this.escalation,
    required this.l10n,
    required this.onCall,
    required this.onBook,
    required this.onMessage,
  });

  final String level;
  final List<String> watchSigns;
  final Map<String, dynamic> escalation;
  final AppLocalizations l10n;
  final VoidCallback onCall;
  final VoidCallback onBook;
  final VoidCallback onMessage;

  @override
  Widget build(BuildContext context) {
    final phone = (escalation['practicePhone'] as String?)?.trim() ?? '';
    final canBook = escalation['canBookVisit'] == true;
    final p = PetsPalette.of(context);

    return Material(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(12, 8, 12, 4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (watchSigns.isNotEmpty) ...[
              Text(
                l10n.clientAiTriageWatchSigns,
                style: Theme.of(context).textTheme.labelLarge,
              ),
              const SizedBox(height: 4),
              ...watchSigns.map((s) => Text('• $s')),
              const SizedBox(height: 8),
            ],
            if (level == 'red' && phone.isNotEmpty)
              FilledButton.icon(
                key: const Key('client_ai_triage_call'),
                onPressed: onCall,
                icon: const Icon(Icons.phone),
                label: Text(l10n.clientAiTriageCallPractice),
              ),
            if (level == 'red' && phone.isEmpty)
              Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: Text(
                  l10n.clientAiTriageEmergencyFallback,
                  key: const Key('client_ai_triage_emergency_fallback'),
                  style: TextStyle(color: p.textMuted, fontSize: 13),
                ),
              ),
            if (level == 'orange' || level == 'red') ...[
              if (canBook)
                Padding(
                  padding: const EdgeInsets.only(top: 6),
                  child: OutlinedButton.icon(
                    key: const Key('client_ai_triage_book'),
                    onPressed: onBook,
                    icon: const Icon(Icons.event),
                    label: Text(l10n.clientAiTriageBookVisit),
                  ),
                ),
            ],
            TextButton.icon(
              key: const Key('client_ai_triage_message'),
              onPressed: onMessage,
              icon: const Icon(Icons.chat_outlined),
              label: Text(
                canBook || (escalation['petId'] as String?)?.isNotEmpty == true
                    ? l10n.clientAiTriageMessageVet
                    : l10n.clientAiTriageOpenMessages,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
