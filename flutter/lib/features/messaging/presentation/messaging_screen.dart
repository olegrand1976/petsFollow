import 'dart:async';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:intl/intl.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/api/open_url.dart';
import 'package:petsfollow_mobile/core/models/message_thread.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/core/models/vet_link.dart';
import 'package:petsfollow_mobile/core/notifications/push_navigation.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/features/messaging/message_media_upload.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_create_flow.dart';
import 'package:petsfollow_mobile/features/vets/presentation/my_vets_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:video_compress/video_compress.dart';

class MessagingScreen extends StatefulWidget {
  const MessagingScreen({
    super.key,
    this.embedded = false,
    this.active = true,
    this.initialPetId,
    this.staffMode = false,
    this.staffClients = const [],
    this.staffPets = const [],
    this.onUnreadTotalChanged,
  });

  final bool embedded;

  /// When embedded in IndexedStack, true while the Messages tab is selected.
  final bool active;

  /// Prefill pet when composing from a pet fiche.
  final String? initialPetId;

  /// Practice staff (vet / assistant / secretary) — no client vet-link gate.
  final bool staffMode;

  /// Optional preloaded clients for staff compose (`id` / `fullName`).
  final List<Map<String, dynamic>> staffClients;

  /// Optional preloaded pets for staff compose (`id` / `name` / `ownerUserId`).
  final List<Map<String, dynamic>> staffPets;

  /// Total unread across threads (for shell nav badge).
  final ValueChanged<int>? onUnreadTotalChanged;

  @override
  State<MessagingScreen> createState() => _MessagingScreenState();
}

class _MessagingScreenState extends State<MessagingScreen>
    with WidgetsBindingObserver {
  static const _pollInterval = Duration(seconds: 4);

  List<MessageThread> threads = [];
  List<ChatMessage> messages = [];
  String? threadId;
  String? currentUserId;
  final draft = TextEditingController();
  bool loading = true;
  bool sending = false;
  bool _refreshing = false;
  bool _ensuringThread = false;
  /// Bumped on each selection apply so stale ensure/load responses are ignored.
  int _selectionEpoch = 0;
  bool _hasLinkedVets = true;
  List<Pet> _clientPets = [];
  List<VetLink> _vets = [];
  List<Map<String, dynamic>> _staffClients = [];
  List<Map<String, dynamic>> _staffPets = [];
  String? _selectedPracticeId;
  String? _selectedClientId;
  String? _selectedPetId;
  Timer? _pollTimer;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    PushNavigation.instance.messageRefreshTick.addListener(_onPushRefresh);
    if (widget.embedded) {
      PushNavigation.instance.onOpenMessageThread = _openThreadFromPush;
    }
    _staffClients = List<Map<String, dynamic>>.from(widget.staffClients);
    _staffPets = List<Map<String, dynamic>>.from(widget.staffPets);
    initThreads().then((_) {
      if (!mounted) return;
      _consumePendingPushThread();
      if (widget.active) _startPolling();
    });
  }

  void _consumePendingPushThread() {
    final pending = PushNavigation.instance.pendingMessageThreadId;
    if (pending == null || pending.isEmpty) return;
    PushNavigation.instance.pendingMessageThreadId = null;
    selectThread(pending);
  }

  @override
  void didUpdateWidget(MessagingScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (!identical(widget.staffClients, oldWidget.staffClients)) {
      _staffClients = List<Map<String, dynamic>>.from(widget.staffClients);
    }
    if (!identical(widget.staffPets, oldWidget.staffPets)) {
      _staffPets = List<Map<String, dynamic>>.from(widget.staffPets);
    }
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
    // Load lists/threads without blocking the shell on ensure POST.
    await _fetchThreadsAndMessages(showErrors: true, ensureMissing: false);
    if (threadId != null) await _markThreadReadIfNeeded(threadId!);
    if (mounted) setState(() => loading = false);
    await _applySelection(ensureIfMissing: true);
  }

  /// Refresh threads + open conversation without blanking the UI.
  Future<void> _silentRefresh() async {
    if (_refreshing || !mounted) return;
    _refreshing = true;
    try {
      await _fetchThreadsAndMessages(showErrors: false, ensureMissing: false);
    } finally {
      _refreshing = false;
    }
  }

  String _clientIdOf(Map<String, dynamic> c) =>
      (c['userId'] as String?) ?? (c['id'] as String?) ?? '';

  List<Pet> _petsForSelectedVet() {
    final practiceId = _selectedPracticeId;
    if (practiceId == null || practiceId.isEmpty) return _clientPets;
    return _clientPets.where((p) {
      final pid = p.practiceId?.trim() ?? '';
      return pid.isEmpty || pid == practiceId;
    }).toList();
  }

  List<Map<String, dynamic>> _petsForSelectedClient() {
    final clientId = _selectedClientId;
    if (clientId == null || clientId.isEmpty) return const [];
    return _staffPets.where((p) {
      final owner = (p['ownerUserId'] as String?) ??
          (p['clientUserId'] as String?) ??
          '';
      return owner == clientId;
    }).toList();
  }

  bool get _selectionComplete {
    if (widget.staffMode) {
      return _selectedClientId != null && _selectedClientId!.isNotEmpty;
    }
    return _selectedPracticeId != null &&
        _selectedPracticeId!.isNotEmpty &&
        _selectedPetId != null &&
        _selectedPetId!.isNotEmpty;
  }

  String _incompleteSelectionHint(AppLocalizations l10n) {
    if (widget.staffMode) {
      if (_selectedClientId == null || _selectedClientId!.isEmpty) {
        return l10n.messageChooseClient;
      }
      return l10n.messageChoosePetOptional;
    }
    if (_selectedPracticeId == null || _selectedPracticeId!.isEmpty) {
      return l10n.messageChoosePro;
    }
    return l10n.messageChoosePet;
  }

  int _unreadSum({
    String? practiceId,
    String? clientId,
    String? petId,
    bool generalPetOnly = false,
  }) {
    var total = 0;
    for (final t in threads) {
      if (practiceId != null && t.practiceId != practiceId) continue;
      if (clientId != null && t.clientUserId != clientId) continue;
      if (generalPetOnly) {
        if ((t.petId ?? '').trim().isNotEmpty) continue;
      } else if (petId != null && t.petId != petId) {
        continue;
      }
      total += t.unreadCount;
    }
    return total;
  }

  String _labelWithUnread(String label, int unread) {
    if (unread <= 0) return label;
    return '$label ($unread)';
  }

  MessageThread? _findThreadForSelection() {
    if (!_selectionComplete) return null;
    if (widget.staffMode) {
      final clientId = _selectedClientId!;
      final petId = _selectedPetId;
      return threads.where((t) {
        if (t.clientUserId != clientId) return false;
        final tp = t.petId?.trim() ?? '';
        final want = petId?.trim() ?? '';
        return tp == want;
      }).firstOrNull;
    }
    final practiceId = _selectedPracticeId!;
    final petId = _selectedPetId!;
    return threads
        .where((t) => t.practiceId == practiceId && t.petId == petId)
        .firstOrNull;
  }

  void _syncSelectionFromThread(MessageThread t) {
    if (widget.staffMode) {
      _selectedClientId = t.clientUserId;
      _selectedPetId = t.petId;
    } else {
      _selectedPracticeId = t.practiceId;
      _selectedPetId = t.petId;
    }
  }

  void _bootstrapSelection() {
    final preferredPet = widget.initialPetId?.trim();

    if (threadId != null) {
      final current = threads.where((t) => t.id == threadId).firstOrNull;
      if (current != null) {
        _syncSelectionFromThread(current);
        return;
      }
    }

    if (preferredPet != null && preferredPet.isNotEmpty) {
      final byPet = threads.where((t) => t.petId == preferredPet).firstOrNull;
      if (byPet != null) {
        threadId = byPet.id;
        _syncSelectionFromThread(byPet);
        return;
      }
      _selectedPetId = preferredPet;
    }

    if (widget.staffMode) {
      if (_selectedClientId != null &&
          _staffClients.any((c) => _clientIdOf(c) == _selectedClientId)) {
        final pets = _petsForSelectedClient();
        if (_selectedPetId != null &&
            !pets.any((p) => p['id'] == _selectedPetId)) {
          _selectedPetId = null;
        }
        return;
      }
      if (threads.isNotEmpty) {
        threadId = threads.first.id;
        _syncSelectionFromThread(threads.first);
        return;
      }
      if (_staffClients.length == 1) {
        _selectedClientId = _clientIdOf(_staffClients.first);
        final pets = _petsForSelectedClient();
        if (pets.length == 1) {
          _selectedPetId = pets.first['id'] as String?;
        }
      }
      return;
    }

    if (_selectedPracticeId != null &&
        _vets.any((v) => v.practiceId == _selectedPracticeId)) {
      final pets = _petsForSelectedVet();
      if (_selectedPetId != null && !pets.any((p) => p.id == _selectedPetId)) {
        _selectedPetId = pets.length == 1 ? pets.first.id : null;
      }
      return;
    }

    if (threads.isNotEmpty) {
      threadId = threads.first.id;
      _syncSelectionFromThread(threads.first);
      return;
    }

    if (_vets.length == 1) {
      _selectedPracticeId = _vets.first.practiceId;
    } else if (_vets.isNotEmpty && _selectedPracticeId == null) {
      _selectedPracticeId = _vets.first.practiceId;
    }
    final pets = _petsForSelectedVet();
    if (_selectedPetId == null && pets.length == 1) {
      _selectedPetId = pets.first.id;
    } else if (_selectedPetId == null &&
        preferredPet != null &&
        pets.any((p) => p.id == preferredPet)) {
      _selectedPetId = preferredPet;
    } else if (_selectedPetId == null && pets.isNotEmpty) {
      _selectedPetId = pets.first.id;
    }
  }

  Future<void> _applySelection({required bool ensureIfMissing}) async {
    final epoch = ++_selectionEpoch;

    if (!_selectionComplete) {
      if (!mounted || epoch != _selectionEpoch) return;
      setState(() {
        threadId = null;
        messages = const [];
        _ensuringThread = false;
      });
      return;
    }

    final existing = _findThreadForSelection();
    if (existing != null) {
      if (epoch != _selectionEpoch) return;
      if (existing.id != threadId) {
        await selectThread(existing.id);
      } else {
        await loadMessages();
        if (epoch != _selectionEpoch) return;
        await _markThreadReadIfNeeded(existing.id);
      }
      if (mounted && epoch == _selectionEpoch && _ensuringThread) {
        setState(() => _ensuringThread = false);
      }
      return;
    }

    if (!ensureIfMissing) {
      if (!mounted || epoch != _selectionEpoch) return;
      setState(() {
        threadId = null;
        messages = const [];
        _ensuringThread = false;
      });
      return;
    }

    if (mounted) setState(() => _ensuringThread = true);
    final l10n = AppLocalizations.of(context)!;
    try {
      final MessageThread thread;
      if (widget.staffMode) {
        thread = await ApiClient.instance.ensureMessageThread(
          clientUserId: _selectedClientId,
          petId: _selectedPetId,
        );
      } else {
        thread = await ApiClient.instance.ensureMessageThread(
          practiceId: _selectedPracticeId,
          petId: _selectedPetId,
        );
      }
      if (!mounted || epoch != _selectionEpoch) return;
      setState(() {
        if (!threads.any((t) => t.id == thread.id)) {
          threads = [thread, ...threads];
        }
        threadId = thread.id;
        _ensuringThread = false;
      });
      await loadMessages();
      if (epoch != _selectionEpoch) return;
      await _markThreadReadIfNeeded(thread.id);
    } catch (e) {
      if (mounted && epoch == _selectionEpoch) {
        setState(() => _ensuringThread = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(mapApiError(e, l10n))),
        );
      }
    } finally {
      if (mounted && epoch == _selectionEpoch && _ensuringThread) {
        setState(() => _ensuringThread = false);
      }
    }
  }

  Future<void> _onPrimaryChanged(String? id) async {
    setState(() {
      if (widget.staffMode) {
        _selectedClientId = id;
        final pets = _petsForSelectedClient();
        if (_selectedPetId == null ||
            !pets.any((p) => p['id'] == _selectedPetId)) {
          _selectedPetId = pets.length == 1 ? pets.first['id'] as String? : null;
        }
      } else {
        _selectedPracticeId = id;
        final pets = _petsForSelectedVet();
        if (_selectedPetId == null ||
            !pets.any((p) => p.id == _selectedPetId)) {
          _selectedPetId = pets.length == 1 ? pets.first.id : null;
        }
      }
      threadId = null;
      messages = const [];
      draft.clear();
    });
    await _applySelection(ensureIfMissing: true);
  }

  Future<void> _onPetChanged(String? id) async {
    setState(() {
      _selectedPetId = (id == null || id.isEmpty) ? null : id;
      threadId = null;
      messages = const [];
      draft.clear();
    });
    await _applySelection(ensureIfMissing: true);
  }

  Future<void> _fetchThreadsAndMessages({
    required bool showErrors,
    required bool ensureMissing,
  }) async {
    try {
      if (currentUserId == null) {
        final me = await ApiClient.instance.getMe();
        currentUserId = me['userId'] as String? ?? me['id'] as String?;
      }
      List<VetLink> vets = _vets;
      List<Pet> clientPets = _clientPets;
      var staffClients = _staffClients;
      var staffPets = _staffPets;

      if (widget.staffMode) {
        if (staffClients.isEmpty) {
          try {
            final raw = await ApiClient.instance.listVetClients();
            staffClients = raw
                .whereType<Map>()
                .map((e) => Map<String, dynamic>.from(e))
                .toList();
          } catch (_) {}
        }
        if (staffPets.isEmpty) {
          try {
            final raw = await ApiClient.instance.listVetPets();
            staffPets = raw
                .whereType<Map>()
                .map((e) => Map<String, dynamic>.from(e))
                .toList();
          } catch (_) {}
        }
      } else {
        try {
          vets = await ApiClient.instance.getMyVets();
        } catch (_) {}
        try {
          final rawPets = await ApiClient.instance.getPets();
          clientPets = rawPets
              .whereType<Map>()
              .map((e) => Pet.fromJson(Map<String, dynamic>.from(e)))
              .where((p) => p.isActive)
              .toList();
        } catch (_) {}
      }

      final rawThreads = await ApiClient.instance.getMessageThreads();
      final enriched = rawThreads.map((t) {
        if (widget.staffMode) return t;
        final vet = vets.where((v) => v.practiceId == t.practiceId).firstOrNull;
        return t.copyWith(
          practiceName: (t.practiceName != null && t.practiceName!.isNotEmpty)
              ? t.practiceName
              : vet?.practiceName,
          vetName: (t.vetName != null && t.vetName!.isNotEmpty)
              ? t.vetName
              : vet?.vetFullName,
        );
      }).toList();

      if (!mounted) return;
      setState(() {
        _hasLinkedVets = widget.staffMode || vets.isNotEmpty;
        _clientPets = clientPets;
        _vets = vets;
        _staffClients = staffClients;
        _staffPets = staffPets;
        threads = enriched;
        if (threadId != null && !enriched.any((t) => t.id == threadId)) {
          threadId = null;
          messages = const [];
        }
        _bootstrapSelection();
      });
      _emitUnreadTotal();
      await _applySelection(ensureIfMissing: ensureMissing);
    } catch (e) {
      if (showErrors && mounted) {
        final l10n = AppLocalizations.of(context)!;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(mapApiError(e, l10n))),
        );
      }
    }
  }

  void _emitUnreadTotal() {
    final cb = widget.onUnreadTotalChanged;
    if (cb == null) return;
    var total = 0;
    for (final t in threads) {
      total += t.unreadCount;
    }
    cb(total);
  }

  Future<void> selectThread(String id) async {
    final t = threads.where((x) => x.id == id).firstOrNull;
    setState(() {
      threadId = id;
      if (t != null) _syncSelectionFromThread(t);
    });
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
          .map((t) => t.id == id ? t.copyWith(unreadCount: 0) : t)
          .toList();
    });
    _emitUnreadTotal();
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
        filename:
            messageMediaUploadBasename(bestPath!, fromCompressedOutput: true),
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

  Widget _buildSelectionBar(AppLocalizations l10n) {
    if (widget.staffMode) {
      final clientValue = _selectedClientId != null &&
              _staffClients.any((c) => _clientIdOf(c) == _selectedClientId)
          ? _selectedClientId
          : null;
      final petItems = _petsForSelectedClient();
      final petValue = _selectedPetId != null &&
              petItems.any((pet) => pet['id'] == _selectedPetId)
          ? _selectedPetId
          : '';

      return Padding(
        padding: const EdgeInsets.fromLTRB(12, 8, 12, 4),
        child: Column(
          children: [
            KeyedSubtree(
              key: const Key('message_select_client'),
              child: DropdownButtonFormField<String>(
                key: ValueKey('staff_client_${clientValue ?? ''}'),
                initialValue: clientValue,
                isExpanded: true,
                decoration: InputDecoration(
                  labelText: l10n.messageChooseClient,
                  border: const OutlineInputBorder(),
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                ),
                items: _staffClients.map((c) {
                  final id = _clientIdOf(c);
                  final name = (c['fullName'] as String?) ??
                      (c['email'] as String?) ??
                      id;
                  final unread = _unreadSum(clientId: id);
                  return DropdownMenuItem(
                    key: Key('message_select_client_item_$id'),
                    value: id,
                    child: Text(
                      _labelWithUnread(name, unread),
                      overflow: TextOverflow.ellipsis,
                    ),
                  );
                }).toList(),
                onChanged: _onPrimaryChanged,
              ),
            ),
            const SizedBox(height: 8),
            KeyedSubtree(
              key: const Key('message_select_pet'),
              child: DropdownButtonFormField<String>(
                key: ValueKey('staff_pet_${clientValue ?? ''}_$petValue'),
                initialValue: petValue,
                isExpanded: true,
                decoration: InputDecoration(
                  labelText: l10n.messageChoosePetOptional,
                  border: const OutlineInputBorder(),
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                ),
                items: [
                  DropdownMenuItem(
                    key: const Key('message_select_pet_none'),
                    value: '',
                    child: Text(
                      _labelWithUnread(
                        l10n.messageGeneralThread,
                        clientValue == null
                            ? 0
                            : _unreadSum(
                                clientId: clientValue,
                                generalPetOnly: true,
                              ),
                      ),
                    ),
                  ),
                  ...petItems.map((pet) {
                    final pid = pet['id'] as String? ?? '';
                    final pname = pet['name'] as String? ?? pid;
                    final unread = clientValue == null
                        ? 0
                        : _unreadSum(clientId: clientValue, petId: pid);
                    return DropdownMenuItem(
                      key: Key('message_select_pet_item_$pid'),
                      value: pid,
                      child: Text(
                        _labelWithUnread(pname, unread),
                        overflow: TextOverflow.ellipsis,
                      ),
                    );
                  }),
                ],
                onChanged: clientValue == null ? null : _onPetChanged,
              ),
            ),
          ],
        ),
      );
    }

    final practiceValue = _selectedPracticeId != null &&
            _vets.any((v) => v.practiceId == _selectedPracticeId)
        ? _selectedPracticeId
        : null;
    final petItems = _petsForSelectedVet();
    final petValue = _selectedPetId != null &&
            petItems.any((pet) => pet.id == _selectedPetId)
        ? _selectedPetId
        : null;

    return Padding(
      padding: const EdgeInsets.fromLTRB(12, 8, 12, 4),
      child: Column(
        children: [
          KeyedSubtree(
            key: const Key('message_select_vet'),
            child: DropdownButtonFormField<String>(
              key: ValueKey('client_vet_${practiceValue ?? ''}'),
              initialValue: practiceValue,
              isExpanded: true,
              decoration: InputDecoration(
                labelText: l10n.messageChoosePro,
                border: const OutlineInputBorder(),
                contentPadding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              ),
              items: _vets.map((v) {
                final label = v.practiceName.isNotEmpty
                    ? v.practiceName
                    : v.vetFullName;
                final base = v.vetFullName.isNotEmpty && v.practiceName.isNotEmpty
                    ? '$label · ${v.vetFullName}'
                    : label;
                final unread = _unreadSum(practiceId: v.practiceId);
                return DropdownMenuItem(
                  key: Key('message_select_vet_item_${v.practiceId}'),
                  value: v.practiceId,
                  child: Text(
                    _labelWithUnread(base, unread),
                    overflow: TextOverflow.ellipsis,
                  ),
                );
              }).toList(),
              onChanged: _onPrimaryChanged,
            ),
          ),
          const SizedBox(height: 8),
          KeyedSubtree(
            key: const Key('message_select_pet'),
            child: DropdownButtonFormField<String>(
              key: ValueKey(
                  'client_pet_${practiceValue ?? ''}_${petValue ?? ''}'),
              initialValue: petValue,
              isExpanded: true,
              decoration: InputDecoration(
                labelText: l10n.messageChoosePet,
                border: const OutlineInputBorder(),
                contentPadding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                hintText: petItems.isEmpty ? l10n.emptyPetsTitle : null,
              ),
              items: petItems
                  .map(
                    (pet) => DropdownMenuItem(
                      key: Key('message_select_pet_item_${pet.id}'),
                      value: pet.id,
                      child: Text(
                        _labelWithUnread(
                          pet.name,
                          practiceValue == null
                              ? 0
                              : _unreadSum(
                                  practiceId: practiceValue,
                                  petId: pet.id,
                                ),
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  )
                  .toList(),
              onChanged: practiceValue == null || petItems.isEmpty
                  ? null
                  : _onPetChanged,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildComposer(AppLocalizations l10n) {
    final enabled = threadId != null && !sending && !_ensuringThread;
    return Padding(
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
            onPressed: enabled ? _showAttachSheet : null,
            icon: const Icon(Icons.attach_file),
          ),
          Expanded(
            child: TextField(
              key: const Key('message_draft'),
              controller: draft,
              enabled: enabled,
              decoration: InputDecoration(
                hintText: l10n.vetMessaging,
                border: const OutlineInputBorder(),
                contentPadding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              ),
              onSubmitted: enabled ? (_) => send() : null,
            ),
          ),
          IconButton(
            key: const Key('message_send_btn'),
            onPressed: enabled ? send : null,
            icon: sending
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.send),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final p = PetsPalette.of(context);
    final timeFmt = DateFormat.Hm(Localizations.localeOf(context).toString());

    final noPetsClient =
        !widget.staffMode && _hasLinkedVets && _clientPets.isEmpty;
    final noStaffClients = widget.staffMode && _staffClients.isEmpty;

    final Widget chatArea;
    if (!_hasLinkedVets) {
      chatArea = _MessagingLocked(
        onLinkVet: () async {
          await Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const MyVetsScreen()),
          );
          await ApiClient.instance.ensureFreshSession();
          if (mounted) initThreads();
        },
      );
    } else if (noPetsClient) {
      chatArea = _MessagingNoPets(onAdd: _openPetForm);
    } else if (noStaffClients) {
      chatArea = Center(
        child: Text(l10n.proLightNoClients, style: TextStyle(color: p.textMuted)),
      );
    } else {
      chatArea = Column(
        children: [
          _buildSelectionBar(l10n),
          Expanded(
            child: _ensuringThread
                ? const Center(child: CircularProgressIndicator())
                : !_selectionComplete
                    ? Center(
                        child: Text(
                          _incompleteSelectionHint(l10n),
                          style: TextStyle(color: p.textMuted),
                        ),
                      )
                    : threadId == null
                        ? Center(
                            child: Text(
                              l10n.messageNoMessagesYet,
                              style: TextStyle(color: p.textMuted),
                            ),
                          )
                        : messages.isEmpty
                            ? Center(
                                child: Text(
                                  l10n.messageNoMessagesYet,
                                  style: TextStyle(color: p.textMuted),
                                ),
                              )
                            : ListView.builder(
                                reverse: true,
                                padding: const EdgeInsets.symmetric(
                                    horizontal: 12, vertical: 8),
                                itemCount: messages.length,
                                itemBuilder: (_, i) {
                                  final m = messages[messages.length - 1 - i];
                                  final isMine =
                                      m.senderUserId == currentUserId;
                                  return Align(
                                    alignment: isMine
                                        ? Alignment.centerRight
                                        : Alignment.centerLeft,
                                    child: Container(
                                      margin: const EdgeInsets.symmetric(
                                          vertical: 4),
                                      padding: const EdgeInsets.symmetric(
                                          horizontal: 14, vertical: 10),
                                      constraints: BoxConstraints(
                                          maxWidth:
                                              MediaQuery.of(context).size.width *
                                                  0.75),
                                      decoration: BoxDecoration(
                                        color: isMine
                                            ? AppColors.primary
                                                .withValues(alpha: 0.85)
                                            : p.surfaceElevated,
                                        borderRadius: BorderRadius.only(
                                          topLeft: const Radius.circular(16),
                                          topRight: const Radius.circular(16),
                                          bottomLeft:
                                              Radius.circular(isMine ? 16 : 4),
                                          bottomRight:
                                              Radius.circular(isMine ? 4 : 16),
                                        ),
                                      ),
                                      child: Column(
                                        crossAxisAlignment:
                                            CrossAxisAlignment.start,
                                        children: [
                                          if (m.hasMedia) ...[
                                            _MessageMedia(
                                              message: m,
                                              isMine: isMine,
                                              onOpen: () =>
                                                  _openMedia(m.mediaUrl!),
                                              l10n: l10n,
                                            ),
                                            if (m.body.isNotEmpty)
                                              const SizedBox(height: 8),
                                          ],
                                          if (m.body.isNotEmpty)
                                            Text(
                                              m.body,
                                              style: TextStyle(
                                                color:
                                                    isMine ? AppColors.bg : null,
                                                height: 1.3,
                                              ),
                                            ),
                                          const SizedBox(height: 4),
                                          Text(
                                            timeFmt.format(m.createdAt),
                                            style: TextStyle(
                                              fontSize: 10,
                                              color: isMine
                                                  ? AppColors.bg
                                                      .withValues(alpha: 0.7)
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
          _buildComposer(l10n),
        ],
      );
    }

    final content = loading
        ? const Center(child: CircularProgressIndicator())
        : chatArea;

    if (widget.embedded) {
      return content;
    }
    return Scaffold(
      appBar: AppBar(title: Text(l10n.vetMessaging)),
      body: content,
    );
  }

  Future<void> _openPetForm() => openPetFormAndFollowUp(
        context,
        hasLinkedVets: _hasLinkedVets,
        onReload: () async {
          await initThreads();
        },
      );
}

class _MessagingLocked extends StatelessWidget {
  const _MessagingLocked({required this.onLinkVet});

  final VoidCallback onLinkVet;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final p = PetsPalette.of(context);
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.lock_outline, size: 48, color: p.textMuted),
            const SizedBox(height: 12),
            Text(
              l10n.messageLockedTitle,
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Text(
              l10n.messageLockedBody,
              textAlign: TextAlign.center,
              style: TextStyle(color: p.textMuted),
            ),
            const SizedBox(height: 16),
            FilledButton(
              key: const Key('message_link_vet_cta'),
              onPressed: onLinkVet,
              child: Text(l10n.linkVetAfterSaveTitle),
            ),
          ],
        ),
      ),
    );
  }
}

class _MessagingNoPets extends StatelessWidget {
  const _MessagingNoPets({required this.onAdd});

  final VoidCallback onAdd;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final p = PetsPalette.of(context);
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.pets,
                size: 48, color: AppColors.gold.withValues(alpha: 0.8)),
            const SizedBox(height: 16),
            Text(l10n.emptyPetsTitle,
                style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 8),
            Text(
              l10n.emptyPetsBody,
              textAlign: TextAlign.center,
              style: TextStyle(color: p.textMuted, height: 1.35),
            ),
            const SizedBox(height: 16),
            FilledButton.icon(
              key: const Key('message_add_pet_cta'),
              onPressed: onAdd,
              icon: const Icon(Icons.add),
              label: Text(l10n.newPet),
            ),
          ],
        ),
      ),
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
            color: (isMine ? AppColors.bg : AppColors.primary)
                .withValues(alpha: 0.12),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Column(
            children: [
              Icon(Icons.play_circle_outline,
                  size: 40, color: isMine ? AppColors.bg : AppColors.primary),
              const SizedBox(height: 4),
              Text(
                l10n.mediaVideoLabel,
                style: TextStyle(
                    color: isMine ? AppColors.bg : null,
                    fontWeight: FontWeight.w600),
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
            child: Center(
                child: Text(l10n.openMedia,
                    style: TextStyle(color: isMine ? AppColors.bg : null))),
          ),
        ),
      ),
    );
  }
}
