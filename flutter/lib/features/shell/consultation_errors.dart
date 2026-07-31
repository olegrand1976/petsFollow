import 'package:dio/dio.dart';

/// True when cancel of a walk-in visit was rejected because a CR is already saved.
bool isConsultationHasReportError(Object e) {
  if (e is! DioException) return false;
  final data = e.response?.data;
  if (data is! Map) return false;
  final err = data['error'];
  if (err is! Map) return false;
  final msgKey = (err['msgKey'] ?? err['messageKey'])?.toString();
  return msgKey == 'consultation_has_report';
}
