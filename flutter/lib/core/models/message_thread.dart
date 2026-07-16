class MessageThread {
  const MessageThread({
    required this.id,
    required this.practiceId,
    this.clientUserId,
    this.vetUserId,
    this.petId,
    this.practiceName,
    this.vetName,
    this.lastMessagePreview,
    this.unreadCount = 0,
  });

  final String id;
  final String practiceId;
  final String? clientUserId;
  final String? vetUserId;
  final String? petId;
  final String? practiceName;
  final String? vetName;
  final String? lastMessagePreview;
  final int unreadCount;

  String get displayLabel {
    if (practiceName != null && practiceName!.isNotEmpty) {
      if (vetName != null && vetName!.isNotEmpty) {
        return '$practiceName · $vetName';
      }
      return practiceName!;
    }
    if (vetName != null && vetName!.isNotEmpty) return vetName!;
    return id.substring(0, 8);
  }

  factory MessageThread.fromJson(Map<String, dynamic> json) {
    return MessageThread(
      id: json['id'] as String? ?? '',
      practiceId: json['practiceId'] as String? ?? '',
      clientUserId: json['clientUserId'] as String?,
      vetUserId: json['vetUserId'] as String?,
      petId: json['petId'] as String?,
      practiceName: json['practiceName'] as String?,
      vetName: json['vetFullName'] as String? ?? json['clientName'] as String?,
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
  });

  final String id;
  final String threadId;
  final String senderUserId;
  final String body;
  final DateTime createdAt;
  final DateTime? readAt;

  factory ChatMessage.fromJson(Map<String, dynamic> json) {
    return ChatMessage(
      id: json['id'] as String? ?? '',
      threadId: json['threadId'] as String? ?? '',
      senderUserId: json['senderUserId'] as String? ?? '',
      body: json['body'] as String? ?? '',
      createdAt: DateTime.tryParse(json['createdAt'] as String? ?? '') ?? DateTime.now(),
      readAt: json['readAt'] != null ? DateTime.tryParse(json['readAt'] as String) : null,
    );
  }
}
