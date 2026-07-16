import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/message_thread.dart';
import 'package:petsfollow_mobile/core/models/vet_link.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:intl/intl.dart';

class MessagingScreen extends StatefulWidget {
  const MessagingScreen({super.key, this.embedded = false});

  final bool embedded;

  @override
  State<MessagingScreen> createState() => _MessagingScreenState();
}

class _MessagingScreenState extends State<MessagingScreen> {
  List<MessageThread> threads = [];
  List<ChatMessage> messages = [];
  String? threadId;
  String? currentUserId;
  final draft = TextEditingController();
  bool loading = true;
  bool sending = false;

  @override
  void initState() {
    super.initState();
    initThreads();
  }

  @override
  void dispose() {
    draft.dispose();
    super.dispose();
  }

  Future<void> initThreads() async {
    setState(() => loading = true);
    try {
      final me = await ApiClient.instance.getMe();
      currentUserId = me['userId'] as String?;
      final rawThreads = await ApiClient.instance.getMessageThreads();
      List<VetLink> vets = [];
      try {
        vets = await ApiClient.instance.getMyVets();
      } catch (_) {}
      final enriched = rawThreads.map((t) {
        final vet = vets.where((v) => v.practiceId == t.practiceId).firstOrNull;
        return MessageThread(
          id: t.id,
          practiceId: t.practiceId,
          clientUserId: t.clientUserId,
          vetUserId: t.vetUserId,
          petId: t.petId,
          practiceName: vet?.practiceName,
          vetName: vet?.vetFullName,
          lastMessagePreview: t.lastMessagePreview,
          unreadCount: t.unreadCount,
        );
      }).toList();
      if (mounted) {
        setState(() {
          threads = enriched;
          if (threads.isNotEmpty && threadId == null) {
            threadId = threads.first.id;
          }
          loading = false;
        });
      }
      if (threadId != null) await loadMessages();
    } catch (_) {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> selectThread(String id) async {
    setState(() => threadId = id);
    await loadMessages();
  }

  Future<void> loadMessages() async {
    if (threadId == null) return;
    final data = await ApiClient.instance.getChatMessages(threadId!);
    if (mounted) setState(() => messages = data);
  }

  Future<void> send() async {
    if (threadId == null || draft.text.trim().isEmpty || sending) return;
    setState(() => sending = true);
    try {
      await ApiClient.instance.sendMessage(threadId!, draft.text.trim());
      draft.clear();
      await loadMessages();
    } finally {
      if (mounted) setState(() => sending = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final timeFmt = DateFormat.Hm(Localizations.localeOf(context).toString());

    final chatArea = threadId == null
        ? Center(child: Text(l10n.noThreads, style: TextStyle(color: AppColors.textMuted)))
        : Column(
            children: [
              Expanded(
                child: messages.isEmpty
                    ? Center(child: Text(l10n.vetMessaging, style: TextStyle(color: AppColors.textMuted)))
                    : ListView.builder(
                        reverse: true,
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                        itemCount: messages.length,
                        itemBuilder: (_, i) {
                          final m = messages[messages.length - 1 - i];
                          final isMine = m.senderUserId == currentUserId;
                          return Align(
                            alignment: isMine ? Alignment.centerRight : Alignment.centerLeft,
                            child: Container(
                              margin: const EdgeInsets.symmetric(vertical: 4),
                              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                              constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.75),
                              decoration: BoxDecoration(
                                color: isMine
                                    ? AppColors.primary.withValues(alpha: 0.85)
                                    : AppColors.surfaceElevated,
                                borderRadius: BorderRadius.only(
                                  topLeft: const Radius.circular(16),
                                  topRight: const Radius.circular(16),
                                  bottomLeft: Radius.circular(isMine ? 16 : 4),
                                  bottomRight: Radius.circular(isMine ? 4 : 16),
                                ),
                              ),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    m.body,
                                    style: TextStyle(
                                      color: isMine ? AppColors.bg : null,
                                      height: 1.3,
                                    ),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    timeFmt.format(m.createdAt),
                                    style: TextStyle(
                                      fontSize: 10,
                                      color: isMine
                                          ? AppColors.bg.withValues(alpha: 0.7)
                                          : AppColors.textMuted,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          );
                        },
                      ),
              ),
              Padding(
                padding: const EdgeInsets.all(8),
                child: Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: draft,
                        decoration: InputDecoration(
                          hintText: l10n.vetMessaging,
                          border: const OutlineInputBorder(),
                          contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                        ),
                        onSubmitted: (_) => send(),
                      ),
                    ),
                    IconButton(
                      onPressed: sending ? null : send,
                      icon: sending
                          ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
                          : const Icon(Icons.send),
                    ),
                  ],
                ),
              ),
            ],
          );

    final content = loading
        ? const Center(child: CircularProgressIndicator())
        : Row(
            children: [
              if (threads.length > 1 || widget.embedded)
                SizedBox(
                  width: threads.length > 1 ? 140 : 0,
                  child: threads.length > 1
                      ? ListView.builder(
                          itemCount: threads.length,
                          itemBuilder: (_, i) {
                            final t = threads[i];
                            final selected = t.id == threadId;
                            return Material(
                              color: selected ? AppColors.primary.withValues(alpha: 0.15) : Colors.transparent,
                              child: ListTile(
                                dense: true,
                                selected: selected,
                                title: Text(
                                  t.displayLabel,
                                  maxLines: 2,
                                  overflow: TextOverflow.ellipsis,
                                  style: const TextStyle(fontSize: 12),
                                ),
                                subtitle: t.lastMessagePreview != null && t.lastMessagePreview!.isNotEmpty
                                    ? Text(
                                        t.lastMessagePreview!,
                                        maxLines: 1,
                                        overflow: TextOverflow.ellipsis,
                                        style: const TextStyle(fontSize: 10),
                                      )
                                    : null,
                                onTap: () => selectThread(t.id),
                              ),
                            );
                          },
                        )
                      : null,
                ),
              Expanded(child: chatArea),
            ],
          );

    if (widget.embedded) {
      return Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          title: Text(l10n.navMessages),
          actions: [
            if (threads.isNotEmpty)
              Padding(
                padding: const EdgeInsets.only(right: 12),
                child: Center(
                  child: Text(
                    threads.where((t) => t.id == threadId).firstOrNull?.displayLabel ?? '',
                    style: TextStyle(fontSize: 12, color: AppColors.textMuted),
                  ),
                ),
              ),
          ],
        ),
        body: content,
      );
    }

    return Scaffold(
      appBar: AppBar(title: Text(l10n.vetMessaging)),
      body: content,
    );
  }
}
