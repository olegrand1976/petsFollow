/// Shared JSON fixtures for Flutter widget / unit tests.
class Fixtures {
  static Map<String, dynamic> pet({
    String id = 'pet-1',
    String name = 'Rex',
    String species = 'dog',
    String breed = 'Labrador',
    String paymentStatus = 'active',
    String? ownerUserId = 'user-1',
    double? weightKg = 12.5,
    List<int> durations = const [15, 30, 60],
  }) =>
      {
        'id': id,
        'name': name,
        'species': species,
        'breed': breed,
        'paymentStatus': paymentStatus,
        'ownerUserId': ownerUserId,
        if (weightKg != null) 'weightKg': weightKg,
        'heartrateDurationsSec': durations,
        'entitlement': {
          'status': 'active',
          'billingMode': 'subscription',
          'planCode': 'annual',
        },
      };

  static Map<String, dynamic> me({
    String id = 'user-1',
    String email = 'client.demo@petsfollow.test',
    String role = 'client',
  }) =>
      {
        'id': id,
        'email': email,
        'role': role,
        'firstName': 'Demo',
        'lastName': 'Client',
      };

  static Map<String, dynamic> loginSuccess({
    String token = 'test-jwt',
    String userId = 'user-1',
  }) =>
      {
        'accessToken': token,
        'refreshToken': 'refresh-$token',
        'user': me(id: userId),
      };

  static Map<String, dynamic> visit({
    String id = 'visit-1',
    String petId = 'pet-1',
    String status = 'confirmed',
  }) =>
      {
        'id': id,
        'petId': petId,
        'status': status,
        'scheduledAt': DateTime.now().toUtc().toIso8601String(),
      };

  static Map<String, dynamic> messageThread({
    String id = 'thread-1',
    String petId = 'pet-1',
  }) =>
      {
        'id': id,
        'petId': petId,
        'petName': 'Rex',
        'unreadCount': 0,
      };

  static List<Map<String, dynamic>> billingPlans() => [
        {
          'code': 'monthly',
          'amountCents': 350,
          'currency': 'eur',
          'interval': 'month',
        },
        {
          'code': 'annual',
          'amountCents': 3500,
          'currency': 'eur',
          'interval': 'year',
        },
      ];
}
