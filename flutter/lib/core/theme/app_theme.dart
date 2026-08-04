import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';

abstract final class AppTheme {
  static const double radiusLg = 28;
  static const double radiusMd = 20;

  static LinearGradient gradientBgFor(PetsPalette p) => LinearGradient(
        begin: Alignment.topCenter,
        end: Alignment.bottomCenter,
        colors: [p.bg, p.surface],
        stops: const [0.0, 0.48],
      );

  static LinearGradient loginGradientFor(PetsPalette p) => LinearGradient(
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
        colors: [p.bg, p.surface, p.surfaceElevated],
      );

  static LinearGradient gradientBgOf(BuildContext context) =>
      gradientBgFor(PetsPalette.of(context));

  static LinearGradient loginGradientOf(BuildContext context) =>
      loginGradientFor(PetsPalette.of(context));
}

ThemeData buildAppDarkTheme() => _buildTheme(Brightness.dark, PetsPalette.dark);

ThemeData buildAppLightTheme() => _buildTheme(Brightness.light, PetsPalette.light);

/// Legacy alias — dark pets theme.
ThemeData buildAppTheme() => buildAppDarkTheme();

/// Familles de repli couvrant le cyrillique, dans l'ordre de préférence :
/// Roboto / Noto Sans côté Android, Helvetica / Arial côté iOS et macOS.
/// Une famille absente est simplement ignorée par le moteur de rendu.
const _cyrillicFallback = <String>['Roboto', 'Noto Sans', 'Helvetica', 'Arial'];

ThemeData _buildTheme(Brightness brightness, PetsPalette palette) {
  final isDark = brightness == Brightness.dark;
  final base = ThemeData(useMaterial3: true, brightness: brightness);
  final shadow = isDark ? Colors.black.withValues(alpha: 0.3) : AppColors.brandNavy.withValues(alpha: 0.08);
  return base.copyWith(
    extensions: [palette],
    colorScheme: ColorScheme(
      brightness: brightness,
      primary: AppColors.primary,
      onPrimary: AppColors.bg,
      secondary: AppColors.accent,
      onSecondary: AppColors.bg,
      tertiary: AppColors.gold,
      onTertiary: AppColors.bg,
      surface: palette.surface,
      onSurface: palette.text,
      error: AppColors.alert,
      onError: AppColors.cream,
    ),
    scaffoldBackgroundColor: palette.bg,
    textTheme: GoogleFonts.dmSansTextTheme(base.textTheme).apply(
      bodyColor: palette.text,
      displayColor: palette.text,
      // DM Sans n'a pas de glyphes cyrilliques : sans repli, l'ukrainien (et le
      // russe à venir) s'affiche en tofu sur les cibles sans substitution
      // automatique. On délègue aux familles système qui couvrent le cyrillique.
      fontFamilyFallback: _cyrillicFallback,
    ),
    appBarTheme: AppBarTheme(
      backgroundColor: Colors.transparent,
      elevation: 0,
      centerTitle: false,
      foregroundColor: palette.text,
      systemOverlayStyle: isDark ? SystemUiOverlayStyle.light : SystemUiOverlayStyle.dark,
    ),
    cardTheme: CardThemeData(
      color: palette.surfaceElevated,
      elevation: isDark ? 2 : 1,
      shadowColor: shadow,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppTheme.radiusLg),
        side: BorderSide(color: AppColors.gold.withValues(alpha: isDark ? 0.08 : 0.2)),
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
      backgroundColor: palette.surface,
      indicatorColor: AppColors.primary.withValues(alpha: 0.2),
      labelTextStyle: WidgetStateProperty.resolveWith((states) {
        if (states.contains(WidgetState.selected)) {
          return const TextStyle(color: AppColors.primary, fontSize: 12);
        }
        return TextStyle(color: palette.textMuted, fontSize: 12);
      }),
      iconTheme: WidgetStateProperty.resolveWith((states) {
        if (states.contains(WidgetState.selected)) {
          return const IconThemeData(color: AppColors.primary);
        }
        return IconThemeData(color: palette.textMuted);
      }),
    ),
    chipTheme: ChipThemeData(
      backgroundColor: AppColors.gold.withValues(alpha: 0.15),
      labelStyle: const TextStyle(color: AppColors.gold),
      side: BorderSide(color: AppColors.gold.withValues(alpha: 0.4)),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: palette.surfaceElevated,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(AppTheme.radiusMd),
        borderSide: BorderSide(color: palette.textMuted.withValues(alpha: 0.3)),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(AppTheme.radiusMd),
        borderSide: BorderSide(color: palette.textMuted.withValues(alpha: 0.3)),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(AppTheme.radiusMd),
        borderSide: const BorderSide(color: AppColors.gold, width: 1.5),
      ),
      labelStyle: TextStyle(color: palette.textMuted),
      hintStyle: TextStyle(color: palette.textMuted),
    ),
    listTileTheme: ListTileThemeData(
      iconColor: palette.textMuted,
      textColor: palette.text,
    ),
    dividerTheme: DividerThemeData(color: palette.textMuted.withValues(alpha: 0.25)),
    switchTheme: SwitchThemeData(
      thumbColor: WidgetStateProperty.resolveWith((states) {
        if (states.contains(WidgetState.selected)) return AppColors.primary;
        return palette.textMuted;
      }),
      trackColor: WidgetStateProperty.resolveWith((states) {
        if (states.contains(WidgetState.selected)) {
          return AppColors.primary.withValues(alpha: 0.35);
        }
        return palette.surfaceElevated;
      }),
    ),
    dropdownMenuTheme: DropdownMenuThemeData(
      textStyle: TextStyle(color: palette.text),
    ),
  );
}
