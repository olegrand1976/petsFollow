import 'dart:io';

import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:intl/intl.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/api/open_url.dart';
import 'package:petsfollow_mobile/core/models/message_thread.dart';
import 'package:petsfollow_mobile/core/models/vet_link.dart';
import 'package:petsfollow_mobile/core/notifications/push_navigation.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class MessagingScreen extends StatefulWidget {
  const MessagingScreen({super.key, this.embedded = false, this.active = true});

  final bool embedded;
  /// When embedded in IndexedStack, true while the Messages tab is selected.
  final bool active;

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
    PushNavigation.instance.messageRefreshTick.addListener(_onPushRefresh);
    PushNavigation.instance.onOpenMessageThread = _openThreadFromPush;
    initThreads();
  }

  @override
  void didUpdateWidget(MessagingScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.active && !oldWidget.active) {
      initThreads();
    }
  }

  @override
  void dispose() {
    PushNavigation.instance.messageRefreshTick.removeListener(_onPushRefresh);
    if (PushNavigation.instance.onOpenMessageThread == _openThreadFromPush) {
      PushNavigation.instance.onOpenMessageThread = null;
    }
    draft.dispose();
    super.dispose();
  }

  void _onPushRefresh() {
    if (!mounted) return;
    initThreads();
  }

  void _openThreadFromPush(String id) {
    if (!mounted || id.isEmpty) return;
    selectThread(id);
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
    final l10n = AppLocalizations.of(context)!;
    try {
      final data = await ApiClient.instance.getChatMessages(threadId!);
      if (mounted) setState(() => messages = data);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(mapApiError(e, l10n))),
        );
      }
    }
  }

  Future<void> send() async {
    if (threadId == null || draft.text.trim().isEmpty || sending) return;
    final l10n = AppLocalizations.of(context)!;
    setState(() => sending = true);
    try {
      await ApiClient.instance.sendMessage(threadId!, draft.text.trim());
      draft.clear();
      await loadMessages();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(mapApiError(e, l10n))),
        );
      }
    } finally {
      if (mounted) setState(() => sending = false);
    }
  }

  Future<void> _pickAndSendMedia({required bool video}) async {
    if (threadId == null || sending) return;
    final l10n = AppLocalizations.of(context)!;
    final picker = ImagePicker();
    final XFile? file = video
        ? await picker.pickVideo(source: ImageSource.gallery, maxDuration: const Duration(minutes: 2))
        : await picker.pickImage(source: ImageSource.gallery, maxWidth: 1920, imageQuality: 85);
    if (file == null) return;
    final size = await File(file.path).length();
    if (size > 25 << 20) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorMediaTooLarge)),
        );
      }
      return;
    }
    setState(() => sending = true);
    try {
      final caption = draft.text.trim();
      await ApiClient.instance.sendMessageMedia(
        threadId!,
        file.path,
        body: caption.isEmpty ? null : caption,
        filename: file.name,
      );
      draft.clear();
      await loadMessages();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(mapApiError(e, l10n))),
        );
      }
    } finally {
      if (mounted) setState(() => sending = false);
    }
  }

  Future<void> _showAttachSheet() async {
    final l10n = AppLocalizations.of(context)!;
    await showModalBottomSheet<void>(
      context: context,
      builder: (ctx) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.photo_outlined),
              title: Text(l10n.attachPhoto),
              onTap: () {
                Navigator.pop(ctx);
                _pickAndSendMedia(video: false);
              },
            ),
            ListTile(
              leading: const Icon(Icons.videocam_outlined),
              title: Text(l10n.attachVideo),
              onTap: () {
                Navigator.pop(ctx);
                _pickAndSendMedia(video: true);
              },
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _openMedia(String url) async {
    final ok = await openExternalUrl(url);
    if (!ok && mounted) {
      final l10n = AppLocalizations.of(context)!;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.errorCouldNotOpenLink)),
      );
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
                                  if (m.hasMedia) ...[
                                    _MessageMedia(
                                      message: m,
                                      isMine: isMine,
                                      onOpen: () => _openMedia(m.mediaUrl!),
                                      l10n: l10n,
                                    ),
                                    if (m.body.isNotEmpty) const SizedBox(height: 8),
                                  ],
                                  if (m.body.isNotEmpty)
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
                    IconButton(
                      tooltip: l10n.attachMedia,
                      onPressed: sending ? null : _showAttachSheet,
                      icon: const Icon(Icons.attach_file),
                    ),
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
                            return ListTile(
                              dense: true,
                              selected: selected,
                              title: Text(t.displayLabel, maxLines: 2, overflow: TextOverflow.ellipsis),
                              onTap: () => selectThread(t.id),
                            );
                          },
                        )
                      : null,
                ),
              Expanded(child: chatArea),
            ],
          );

    if (widget.embedded) return content;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.vetMessaging)),
      body: content,
    );
  }
}

class _MessageMedia extends StatelessWidget {
  const _MessageMedia({
    required this.message,
    required this.isMine,
    required this.onOpen,
    required this.l10n,
  });

  final ChatMessage message;
  final bool isMine;
  final VoidCallback onOpen;
  final AppLocalizations l10n;

  @override
  Widget build(BuildContext context) {
    if (message.isVideo) {
      return InkWell(
        onTap: onOpen,
        child: Container(
          width: double.infinity,
          padding: const EdgeInsets.symmetric(vertical: 20, horizontal: 12),
          decoration: BoxDecoration(
            color: (isMine ? AppColors.bg : AppColors.primary).withValues(alpha: 0.12),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Column(
            children: [
              Icon(Icons.play_circle_outline, size: 40, color: isMine ? AppColors.bg : AppColors.primary),
              const SizedBox(height: 4),
              Text(
                l10n.mediaVideoLabel,
                style: TextStyle(color: isMine ? AppColors.bg : null, fontWeight: FontWeight.w600),
              ),
            ],
          ),
        ),
      );
    }
    return GestureDetector(
      onTap: onOpen,
      child: ClipRRect(
        borderRadius: BorderRadius.circular(12),
        child: Image.network(
          message.mediaUrl!,
          fit: BoxFit.cover,
          width: double.infinity,
          height: 180,
          errorBuilder: (_, __, ___) => SizedBox(
            height: 80,
            child: Center(child: Text(l10n.openMedia, style: TextStyle(color: isMine ? AppColors.bg : null))),
          ),
        ),
      ),
    );
  }
}
