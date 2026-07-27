import 'package:petsfollow_mobile/core/api/media_url.dart';

class MessageThread {
  const MessageThread({
    required this.id,
    required this.practiceId,
    this.clientUserId,
    this.vetUserId,
    this.petId,
    this.petName,
    this.practiceName,
    this.vetName,
    this.clientName,
    this.lastMessagePreview,
    this.unreadCount = 0,
  });

  final String id;
  final String practiceId;
  final String? clientUserId;
  final String? vetUserId;
  final String? petId;
  final String? petName;
  final String? practiceName;
  final String? vetName;
  final String? clientName;
  final String? lastMessagePreview;
  final int unreadCount;

  /// Client shell: practice · pet. Staff shell: client · pet (or client alone).
  String get displayLabel {
    final client = (clientName != null && clientName!.isNotEmpty) ? clientName! : '';
    final pro = (practiceName != null && practiceName!.isNotEmpty)
        ? practiceName!
        : (vetName != null && vetName!.isNotEmpty ? vetName! : '');
    final pet = (petName != null && petName!.isNotEmpty) ? petName! : '';
    final primary = client.isNotEmpty ? client : pro;
    if (primary.isNotEmpty && pet.isNotEmpty) return '$primary · $pet';
    if (primary.isNotEmpty) return primary;
    if (pet.isNotEmpty) return pet;
    return id.length >= 8 ? id.substring(0, 8) : id;
  }

  MessageThread copyWith({
    String? id,
    String? practiceId,
    String? clientUserId,
    String? vetUserId,
    String? petId,
    String? petName,
    String? practiceName,
    String? vetName,
    String? clientName,
    String? lastMessagePreview,
    int? unreadCount,
  }) {
    return MessageThread(
      id: id ?? this.id,
      practiceId: practiceId ?? this.practiceId,
      clientUserId: clientUserId ?? this.clientUserId,
      vetUserId: vetUserId ?? this.vetUserId,
      petId: petId ?? this.petId,
      petName: petName ?? this.petName,
      practiceName: practiceName ?? this.practiceName,
      vetName: vetName ?? this.vetName,
      clientName: clientName ?? this.clientName,
      lastMessagePreview: lastMessagePreview ?? this.lastMessagePreview,
      unreadCount: unreadCount ?? this.unreadCount,
    );
  }

  factory MessageThread.fromJson(Map<String, dynamic> json) {
    return MessageThread(
      id: json['id'] as String? ?? '',
      practiceId: json['practiceId'] as String? ?? '',
      clientUserId: json['clientUserId'] as String?,
      vetUserId: json['vetUserId'] as String?,
      petId: json['petId'] as String?,
      petName: json['petName'] as String?,
      practiceName: json['practiceName'] as String?,
      vetName: json['vetFullName'] as String?,
      clientName: json['clientName'] as String?,
      lastMessagePreview: json['lastMessagePreview'] as String?,
      unreadCount: json['unreadCount'] as int? ?? 0,
    );
  }
}

class ChatMessage {
  const ChatMessage({
    required this.id,
    required this.threadId,
    required this.senderUserId,
    required this.body,
    required this.createdAt,
    this.readAt,
    this.mediaUrl,
    this.mediaType,
  });

  final String id;
  final String threadId;
  final String senderUserId;
  final String body;
  final DateTime createdAt;
  final DateTime? readAt;
  final String? mediaUrl;
  final String? mediaType;

  bool get hasMedia => mediaUrl != null && mediaUrl!.isNotEmpty;
  bool get isVideo => mediaType == 'video';
  bool get isImage => mediaType == 'image' || (hasMedia && !isVideo);

  factory ChatMessage.fromJson(Map<String, dynamic> json) {
    return ChatMessage(
      id: json['id'] as String? ?? '',
      threadId: json['threadId'] as String? ?? '',
      senderUserId: json['senderUserId'] as String? ?? '',
      body: json['body'] as String? ?? '',
      createdAt: DateTime.tryParse(json['createdAt'] as String? ?? '') ?? DateTime.now(),
      readAt: json['readAt'] != null ? DateTime.tryParse(json['readAt'] as String) : null,
      mediaUrl: resolveMediaUrl(json['mediaUrl'] as String?),
      mediaType: json['mediaType'] as String?,
    );
  }
}
