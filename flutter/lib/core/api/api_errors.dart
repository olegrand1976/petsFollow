import 'package:dio/dio.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

String mapApiError(Object e, AppLocalizations l10n) {
  if (e is DioException) {
    if (e.type == DioExceptionType.connectionTimeout ||
        e.type == DioExceptionType.sendTimeout ||
        e.type == DioExceptionType.receiveTimeout ||
        e.type == DioExceptionType.connectionError) {
      return l10n.errorNetwork;
    }
    final data = e.response?.data;
    if (data is Map) {
      final err = data['error'];
      if (err is Map) {
        final key = (err['msgKey'] ?? err['messageKey'] ?? err['code'])?.toString();
        switch (key) {
          case 'image_too_large':
          case 'payload_too_large':
            return l10n.errorMediaTooLarge;
          case 'invalid_image_type':
            return l10n.errorInvalidMediaType;
          case 'payment_required':
            return l10n.errorPaymentRequired;
          case 'file_required':
            return l10n.errorInvalidMediaType;
          default:
            final message = err['message']?.toString();
            if (message != null && message.isNotEmpty) return message;
        }
      }
    }
  }
  return l10n.errorGeneric(e.toString());
}

/// Extrait le code d'erreur métier brut d'une réponse API enveloppée
/// `{ error: { code | msgKey | message } }` (ou `{ code }` à plat),
/// pour les écrans qui mappent eux-mêmes les codes vers leurs messages.
String? apiErrorCode(Object e) {
  if (e is! DioException) return null;
  final data = e.response?.data;
  if (data is Map) {
    final err = data['error'];
    if (err is Map) {
      return (err['code'] ?? err['msgKey'] ?? err['message'])?.toString();
    }
    return data['code']?.toString();
  }
  return null;
}
