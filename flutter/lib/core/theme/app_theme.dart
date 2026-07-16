import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';

abstract final class AppTheme {
  static const gradientBg = LinearGradient(
    begin: Alignment.topCenter,
    end: Alignment.bottomCenter,
    colors: [AppColors.bg, AppColors.surface],
    stops: [0.0, 0.48],
  );

  static const loginGradient = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [AppColors.bg, Color(0xFF1B2838), AppColors.surfaceElevated],
  );

  static const double radiusLg = 28;
  static const double radiusMd = 20;
}

ThemeData buildAppTheme() {
  final base = ThemeData(useMaterial3: true, brightness: Brightness.dark);
  return base.copyWith(
    colorScheme: ColorScheme.dark(
      primary: AppColors.primary,
      secondary: AppColors.accent,
      tertiary: AppColors.gold,
      surface: AppColors.surface,
      error: AppColors.alert,
    ),
    scaffoldBackgroundColor: AppColors.bg,
    textTheme: GoogleFonts.dmSansTextTheme(base.textTheme).apply(
      bodyColor: AppColors.cream,
      displayColor: AppColors.cream,
    ),
    appBarTheme: const AppBarTheme(
      backgroundColor: Colors.transparent,
      elevation: 0,
      centerTitle: false,
    ),
    cardTheme: CardThemeData(
      color: AppColors.surfaceElevated,
      elevation: 2,
      shadowColor: Colors.black.withValues(alpha: 0.3),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppTheme.radiusLg),
        side: BorderSide(color: AppColors.gold.withValues(alpha: 0.08)),
      ),
    ),
    filledButtonTheme: FilledButtonThemeData(
      style: FilledButton.styleFrom(
        backgroundColor: AppColors.primary,
        foregroundColor: AppColors.bg,
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(AppTheme.radiusMd)),
      ),
    ),
    outlinedButtonTheme: OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(
        foregroundColor: AppColors.gold,
        side: BorderSide(color: AppColors.gold.withValues(alpha: 0.5)),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(AppTheme.radiusMd)),
      ),
    ),
    navigationBarTheme: NavigationBarThemeData(
      backgroundColor: AppColors.surface,
      indicatorColor: AppColors.primary.withValues(alpha: 0.2),
      labelTextStyle: WidgetStateProperty.resolveWith((states) {
        if (states.contains(WidgetState.selected)) {
          return TextStyle(color: AppColors.primary, fontSize: 12);
        }
        return TextStyle(color: AppColors.textMuted, fontSize: 12);
      }),
    ),
    chipTheme: ChipThemeData(
      backgroundColor: AppColors.gold.withValues(alpha: 0.15),
      labelStyle: TextStyle(color: AppColors.gold),
      side: BorderSide(color: AppColors.gold.withValues(alpha: 0.4)),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: AppColors.surfaceElevated,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(AppTheme.radiusMd),
        borderSide: BorderSide(color: AppColors.textMuted.withValues(alpha: 0.3)),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(AppTheme.radiusMd),
        borderSide: BorderSide(color: AppColors.textMuted.withValues(alpha: 0.3)),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(AppTheme.radiusMd),
        borderSide: const BorderSide(color: AppColors.gold, width: 1.5),
      ),
      labelStyle: TextStyle(color: AppColors.textMuted),
    ),
  );
}
