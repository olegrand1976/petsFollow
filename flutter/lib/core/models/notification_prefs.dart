class NotificationPrefs {
  const NotificationPrefs({
    required this.userId,
    this.hr = true,
    this.care = true,
    this.visits = true,
    this.messages = true,
    this.discovery = true,
    this.billing = true,
  });

  final String userId;
  final bool hr;
  final bool care;
  final bool visits;
  final bool messages;
  final bool discovery;
  final bool billing;

  factory NotificationPrefs.fromJson(Map<String, dynamic> json) {
    return NotificationPrefs(
      userId: json['userId'] as String? ?? '',
      hr: json['hr'] as bool? ?? true,
      care: json['care'] as bool? ?? true,
      visits: json['visits'] as bool? ?? true,
      messages: json['messages'] as bool? ?? true,
      discovery: json['discovery'] as bool? ?? true,
      billing: json['billing'] as bool? ?? true,
    );
  }

  Map<String, dynamic> toJson() => {
        'hr': hr,
        'care': care,
        'visits': visits,
        'messages': messages,
        'discovery': discovery,
        'billing': billing,
      };

  NotificationPrefs copyWith({
    bool? hr,
    bool? care,
    bool? visits,
    bool? messages,
    bool? discovery,
    bool? billing,
  }) {
    return NotificationPrefs(
      userId: userId,
      hr: hr ?? this.hr,
      care: care ?? this.care,
      visits: visits ?? this.visits,
      messages: messages ?? this.messages,
      discovery: discovery ?? this.discovery,
      billing: billing ?? this.billing,
    );
  }
}
