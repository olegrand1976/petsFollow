import 'dart:async';
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
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/features/messaging/message_media_upload.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:video_compress/video_compress.dart';

class MessagingScreen extends StatefulWidget {
  const MessagingScreen({super.key, this.embedded = false, this.active = true});

  final bool embedded;
  /// When embedded in IndexedStack, true while the Messages tab is selected.
  final bool active;

  @override
  State<MessagingScreen> createState() => _MessagingScreenState();
}

class _MessagingScreenState extends State<MessagingScreen> with WidgetsBindingObserver {
  static const _pollInterval = Duration(seconds: 4);

  List<MessageThread> threads = [];
  List<ChatMessage> messages = [];
  String? threadId;
  String? currentUserId;
  final draft = TextEditingController();
  bool loading = true;
  bool sending = false;
  bool _refreshing = false;
  Timer? _pollTimer;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    PushNavigation.instance.messageRefreshTick.addListener(_onPushRefresh);
    if (widget.embedded) {
      PushNavigation.instance.onOpenMessageThread = _openThreadFromPush;
    }
    initThreads().then((_) {
      if (widget.active) _startPolling();
    });
  }

  @override
  void didUpdateWidget(MessagingScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.active && !oldWidget.active) {
      _silentRefresh();
      _startPolling();
    } else if (!widget.active && oldWidget.active) {
      _stopPolling();
    }
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed && widget.active) {
      _silentRefresh();
      _startPolling();
    } else if (state == AppLifecycleState.paused) {
      _stopPolling();
    }
  }

  @override
  void dispose() {
    _stopPolling();
    WidgetsBinding.instance.removeObserver(this);
    PushNavigation.instance.messageRefreshTick.removeListener(_onPushRefresh);
    if (widget.embedded &&
        PushNavigation.instance.onOpenMessageThread == _openThreadFromPush) {
      PushNavigation.instance.onOpenMessageThread = null;
    }
    draft.dispose();
    super.dispose();
  }

  void _startPolling() {
    _pollTimer?.cancel();
    if (!widget.active) return;
    _pollTimer = Timer.periodic(_pollInterval, (_) {
      if (!mounted || !widget.active || sending) return;
      _silentRefresh();
    });
  }

  void _stopPolling() {
    _pollTimer?.cancel();
    _pollTimer = null;
  }

  void _onPushRefresh() {
    if (!mounted) return;
    _silentRefresh();
  }

  void _openThreadFromPush(String id) {
    if (!mounted || id.isEmpty) return;
    selectThread(id);
  }

  Future<void> initThreads() async {
    if (mounted) setState(() => loading = true);
    await _fetchThreadsAndMessages(showErrors: true);
    if (threadId != null) await _markThreadReadIfNeeded(threadId!);
    if (mounted) setState(() => loading = false);
  }

  /// Refresh threads + open conversation without blanking the UI.
  Future<void> _silentRefresh() async {
    if (_refreshing || !mounted) return;
    _refreshing = true;
    try {
      await _fetchThreadsAndMessages(showErrors: false);
    } finally {
      _refreshing = false;
    }
  }

  Future<void> _fetchThreadsAndMessages({required bool showErrors}) async {
    try {
      if (currentUserId == null) {
        final me = await ApiClient.instance.getMe();
        currentUserId = me['userId'] as String?;
      }
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
      var nextThreadId = threadId;
      if (enriched.isNotEmpty &&
          (nextThreadId == null || !enriched.any((t) => t.id == nextThreadId))) {
        nextThreadId = enriched.first.id;
      }
      List<ChatMessage> nextMessages = messages;
      if (nextThreadId != null) {
        nextMessages = await ApiClient.instance.getChatMessages(nextThreadId);
      }
      if (!mounted) return;
      setState(() {
        threads = enriched;
        threadId = nextThreadId;
        messages = nextMessages;
      });
    } catch (e) {
      if (showErrors && mounted) {
        final l10n = AppLocalizations.of(context)!;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(mapApiError(e, l10n))),
        );
      }
    }
  }

  Future<void> selectThread(String id) async {
    setState(() => threadId = id);
    await loadMessages();
    await _markThreadReadIfNeeded(id);
  }

  Future<void> _markThreadReadIfNeeded(String id) async {
    final thread = threads.where((t) => t.id == id).firstOrNull;
    if (thread == null || thread.unreadCount <= 0) return;
    try {
      await ApiClient.instance.markThreadRead(id);
      _clearLocalUnread(id);
    } catch (_) {}
  }

  void _clearLocalUnread(String id) {
    if (!mounted) return;
    setState(() {
      threads = threads
          .map((t) => t.id == id
              ? MessageThread(
                  id: t.id,
                  practiceId: t.practiceId,
                  clientUserId: t.clientUserId,
                  vetUserId: t.vetUserId,
                  petId: t.petId,
                  practiceName: t.practiceName,
                  vetName: t.vetName,
                  lastMessagePreview: t.lastMessagePreview,
                  unreadCount: 0,
                )
              : t)
          .toList();
    });
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

  Future<void> _pickAndSendMedia({
    required bool video,
    required ImageSource source,
  }) async {
    if (threadId == null || sending) return;
    final l10n = AppLocalizations.of(context)!;
    final picker = ImagePicker();
    final XFile? picked = video
        ? await picker.pickVideo(
            source: source,
            maxDuration: kMaxMessageVideoDuration,
          )
        : await picker.pickImage(
            source: source,
            maxWidth: 1920,
            imageQuality: 85,
            preferredCameraDevice: CameraDevice.rear,
          );
    if (picked == null) return;

    setState(() => sending = true);
    try {
      var uploadPath = picked.path;
      var filename = picked.name;
      if (video) {
        final prepared = await _prepareVideoForUpload(picked.path, l10n: l10n);
        if (prepared == null) {
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(l10n.errorMediaTooLarge)),
            );
          }
          return;
        }
        uploadPath = prepared.path;
        filename = prepared.filename;
      } else {
        final size = await File(uploadPath).length();
        if (size > kMaxMessageMediaBytes) {
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(l10n.errorMediaTooLarge)),
            );
          }
          return;
        }
      }

      final caption = draft.text.trim();
      await ApiClient.instance.sendMessageMedia(
        threadId!,
        uploadPath,
        body: caption.isEmpty ? null : caption,
        filename: filename,
      );
      draft.clear();
      await loadMessages();
      unawaited(VideoCompress.deleteAllCache());
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

  /// Compresses large videos so GCS upload stays under 25 MiB.
  Future<({String path, String filename})?> _prepareVideoForUpload(
    String path, {
    required AppLocalizations l10n,
  }) async {
    final originalSize = await File(path).length();
    if (!shouldCompressMessageVideo(originalSize)) {
      return (
        path: path,
        filename: messageMediaUploadBasename(path, fromCompressedOutput: false),
      );
    }

    try {
      final info = await VideoCompress.getMediaInfo(path);
      final durationMs = info.duration;
      if (durationMs != null &&
          durationMs > kMaxMessageVideoDuration.inMilliseconds) {
        return null;
      }
    } catch (_) {
      // getMediaInfo est best-effort ; on tente quand même la compression.
    }

    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.compressingMedia)),
      );
    }

    String? bestPath;
    var bestSize = originalSize;

    Future<void> consider(MediaInfo? info) async {
      final out = info?.path;
      if (out == null || out.isEmpty) return;
      final size = await File(out).length();
      if (size <= kMaxMessageMediaBytes && size < bestSize) {
        bestPath = out;
        bestSize = size;
      }
    }

    await consider(
      await VideoCompress.compressVideo(
        path,
        quality: VideoQuality.MediumQuality,
        deleteOrigin: false,
        includeAudio: true,
        frameRate: 30,
      ),
    );
    if (bestPath == null || bestSize > (kMaxMessageMediaBytes ~/ 2)) {
      await consider(
        await VideoCompress.compressVideo(
          path,
          quality: VideoQuality.LowQuality,
          deleteOrigin: false,
          includeAudio: true,
          frameRate: 24,
        ),
      );
    }

    if (bestPath != null) {
      return (
        path: bestPath!,
        filename: messageMediaUploadBasename(bestPath!, fromCompressedOutput: true),
      );
    }
    if (originalSize <= kMaxMessageMediaBytes) {
      return (
        path: path,
        filename: messageMediaUploadBasename(path, fromCompressedOutput: false),
      );
    }
    return null;
  }

  Future<void> _showAttachSheet() async {
    final l10n = AppLocalizations.of(context)!;
    final kind = await showModalBottomSheet<bool>(
      context: context,
      builder: (ctx) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              key: const Key('message_attach_photo'),
              leading: const Icon(Icons.photo_outlined),
              title: Text(l10n.attachPhoto),
              onTap: () => Navigator.pop(ctx, false),
            ),
            ListTile(
              key: const Key('message_attach_video'),
              leading: const Icon(Icons.videocam_outlined),
              title: Text(l10n.attachVideo),
              onTap: () => Navigator.pop(ctx, true),
            ),
          ],
        ),
      ),
    );
    if (kind == null || !mounted) return;

    final source = await showModalBottomSheet<ImageSource>(
      context: context,
      builder: (ctx) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              key: const Key('message_attach_camera'),
              leading: Icon(
                kind ? Icons.videocam_outlined : Icons.photo_camera_outlined,
              ),
              title: Text(kind ? l10n.takeVideo : l10n.takePhoto),
              onTap: () => Navigator.pop(ctx, ImageSource.camera),
            ),
            ListTile(
              key: const Key('message_attach_gallery'),
              leading: const Icon(Icons.photo_library_outlined),
              title: Text(l10n.chooseFromGallery),
              onTap: () => Navigator.pop(ctx, ImageSource.gallery),
            ),
          ],
        ),
      ),
    );
    if (source == null || !mounted) return;
    await _pickAndSendMedia(video: kind, source: source);
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
    final p = PetsPalette.of(context);
    final timeFmt = DateFormat.Hm(Localizations.localeOf(context).toString());

    final chatArea = threadId == null
        ? Center(child: Text(l10n.noThreads, style: TextStyle(color: p.textMuted)))
        : Column(
            children: [
              Expanded(
                child: messages.isEmpty
                    ? Center(child: Text(l10n.vetMessaging, style: TextStyle(color: p.textMuted)))
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
                                    : p.surfaceElevated,
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
                                          : p.textMuted,
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
                padding: EdgeInsets.fromLTRB(
                  8,
                  8,
                  8,
                  composerBottomPadding(context, embedded: widget.embedded),
                ),
                child: Row(
                  children: [
                    IconButton(
                      key: const Key('message_attach_btn'),
                      tooltip: l10n.attachMedia,
                      onPressed: sending ? null : _showAttachSheet,
                      icon: const Icon(Icons.attach_file),
                    ),
                    Expanded(
                      child: TextField(
                        key: const Key('message_draft'),
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
                      key: const Key('message_send_btn'),
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
