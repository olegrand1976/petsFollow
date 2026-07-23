import 'package:petsfollow_mobile/core/api/media_url.dart';

class PetEntitlement {
  const PetEntitlement({
    this.billingMode,
    this.status,
    this.validUntil,
    this.planCode,
  });

  final String? billingMode;
  final String? status;
  final DateTime? validUntil;
  final String? planCode;

  factory PetEntitlement.fromJson(Map<String, dynamic>? json) {
    if (json == null) return const PetEntitlement();
    return PetEntitlement(
      billingMode: json['billingMode'] as String?,
      status: json['status'] as String?,
      validUntil: json['validUntil'] != null
          ? DateTime.tryParse(json['validUntil'] as String)
          : null,
      planCode: json['planCode'] as String?,
    );
  }

  bool get isSubscription => billingMode == 'subscription';

  /// Aligns with Go `HasActiveEntitlement` / `AllowsAccess`.
  bool get allowsAccess {
    switch (status) {
      case 'active':
      case 'past_due':
      case 'cancelled':
        if (validUntil != null && DateTime.now().isAfter(validUntil!)) {
          return false;
        }
        return true;
      default:
        return false;
    }
  }
}

class Pet {
  const Pet({
    required this.id,
    required this.name,
    required this.species,
    required this.breed,
    this.photoUrl,
    this.paymentStatus = 'pending_payment',
    this.practiceId,
    this.entitlement,
    this.heartrateDurationsSec = const [60],
  });

  final String id;
  final String name;
  final String species;
  final String breed;
  final String? photoUrl;
  final String paymentStatus;
  final String? practiceId;
  final PetEntitlement? entitlement;
  final List<int> heartrateDurationsSec;

  /// Premium access aligned with Go `HasActiveEntitlement` when entitlement is present.
  bool get isActive {
    final ent = entitlement;
    if (ent?.status != null && ent!.status!.isNotEmpty) {
      return ent.allowsAccess;
    }
    return paymentStatus == 'active';
  }

  /// Resume checkout only when payment is pending and entitlement does not grant access.
  bool get needsResumePayment =>
      paymentStatus == 'pending_payment' && !isActive;

  factory Pet.fromJson(Map<String, dynamic> json) {
    final rawDurations = json['heartrateDurationsSec'] as List<dynamic>?;
    return Pet(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
      species: json['species'] as String? ?? '',
      breed: json['breed'] as String? ?? '',
      photoUrl: resolveMediaUrl(json['photoUrl'] as String?),
      paymentStatus: json['paymentStatus'] as String? ?? 'pending_payment',
      practiceId: json['practiceId'] as String?,
      entitlement: PetEntitlement.fromJson(
        json['entitlement'] as Map<String, dynamic>?,
      ),
      heartrateDurationsSec: rawDurations == null || rawDurations.isEmpty
          ? const [60]
          : rawDurations.map((e) => (e as num).toInt()).toList(),
    );
  }
}
