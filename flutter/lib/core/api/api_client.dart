import 'package:dio/dio.dart';

class ApiClient {
  ApiClient._();
  static final instance = ApiClient._();

  String? token;
  final dio = Dio(BaseOptions(
    baseUrl: const String.fromEnvironment('API_BASE', defaultValue: 'http://10.0.2.2:8291'),
    headers: {'Content-Type': 'application/json'},
  ));

  void loadToken() {
    dio.interceptors.clear();
    dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) {
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
    ));
  }

  Future<Map<String, dynamic>> login(String email, String password) async {
    final res = await dio.post('/api/v1/auth/login', data: {
      'email': email,
      'password': password,
    });
    final data = res.data['data'] as Map<String, dynamic>;
    token = data['accessToken'] as String?;
    loadToken();
    return data;
  }

  Future<List<dynamic>> getPets() async {
    final res = await dio.get('/api/v1/pets');
    return res.data['data'] as List<dynamic>;
  }

  Future<Map<String, dynamic>> createPet(Map<String, dynamic> body) async {
    final res = await dio.post('/api/v1/pets', data: body);
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getPet(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId');
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<String> resumeCheckout(String petId) async {
    final res = await dio.post('/api/v1/pets/$petId/billing/checkout');
    return res.data['data']['checkoutUrl'] as String;
  }

  Future<String> billingPortal(String petId) async {
    final res = await dio.post('/api/v1/pets/$petId/billing/portal', data: {});
    return res.data['data']['url'] as String;
  }

  Future<List<dynamic>> getBillingPlans() async {
    final res = await dio.get('/api/v1/billing/plans');
    return res.data['data']['plans'] as List<dynamic>;
  }

  Future<Map<String, dynamic>> startHeartRate(String petId) async {
    final res = await dio.post('/api/v1/pets/$petId/heartrate/sessions');
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> completeHeartRate(String sessionId, int tapCount) async {
    final res = await dio.patch('/api/v1/heartrate/sessions/$sessionId', data: {'tapCount': tapCount});
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> validateHeartRate(String sessionId) async {
    final res = await dio.post('/api/v1/heartrate/sessions/$sessionId/validate');
    return res.data['data'] as Map<String, dynamic>;
  }

  Future<void> cancelHeartRate(String sessionId) async {
    await dio.post('/api/v1/heartrate/sessions/$sessionId/cancel');
  }

  Future<List<dynamic>> getTimeline(String petId) async {
    final res = await dio.get('/api/v1/pets/$petId/timeline');
    return res.data['data'] as List<dynamic>;
  }

  Future<List<dynamic>> getThreads() async {
    final res = await dio.get('/api/v1/messaging/threads');
    return res.data['data'] as List<dynamic>;
  }

  Future<List<dynamic>> getMessages(String threadId) async {
    final res = await dio.get('/api/v1/messaging/threads/$threadId/messages');
    return res.data['data'] as List<dynamic>;
  }

  Future<void> sendMessage(String threadId, String body) async {
    await dio.post('/api/v1/messaging/threads/$threadId/messages', data: {'body': body});
  }
}
