import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Shared timeline type → icon / label / color (home activity + pet timeline).
IconData timelineTypeIcon(String type) {
  switch (type) {
    case 'heartrate':
      return Icons.favorite_outline;
    case 'weight':
      return Icons.monitor_weight_outlined;
    case 'blood_pressure':
      return Icons.monitor_heart_outlined;
    case 'lab_panel':
      return Icons.science_outlined;
    case 'message':
      return Icons.chat_bubble_outline;
    case 'care':
      return Icons.medical_services_outlined;
    case 'visit':
      return Icons.event_available;
    case 'event':
      return Icons.flag_outlined;
    default:
      return Icons.circle_outlined;
  }
}

String timelineTypeLabel(AppLocalizations l10n, String type) {
  switch (type) {
    case 'heartrate':
      return l10n.timelineTypeHeartrate;
    case 'weight':
      return l10n.timelineTypeWeight;
    case 'blood_pressure':
      return l10n.bloodPressureShort;
    case 'lab_panel':
      return l10n.labsTitle;
    case 'message':
      return l10n.timelineTypeMessage;
    case 'care':
      return l10n.timelineTypeCare;
    case 'visit':
      return l10n.timelineTypeVisit;
    case 'event':
      return l10n.timelineTypeEvent;
    default:
      return type;
  }
}

Color timelineTypeColor(String type, Color textMuted) {
  switch (type) {
    case 'heartrate':
      return AppColors.alert;
    case 'weight':
      return AppColors.brandTeal;
    case 'blood_pressure':
      return AppColors.primary;
    case 'lab_panel':
      return AppColors.gold;
    case 'message':
      return AppColors.primary;
    case 'care':
      return AppColors.gold;
    case 'visit':
      return AppColors.primary;
    default:
      return textMuted;
  }
}
