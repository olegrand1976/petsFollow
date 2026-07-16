import 'package:dio/dio.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

String mapApiError(Object e, AppLocalizations l10n) {
  if (e is DioException) {
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
    if (e.type == DioExceptionType.sendTimeout ||
        e.type == DioExceptionType.receiveTimeout ||
        e.type == DioExceptionType.connectionTimeout) {
      return l10n.errorMediaTooLarge;
    }
  }
  return l10n.errorGeneric(e.toString());
}
