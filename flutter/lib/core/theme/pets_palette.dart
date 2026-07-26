import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';

/// Structural pets colors that flip with [Brightness]. Accents stay on [AppColors].
@immutable
class PetsPalette extends ThemeExtension<PetsPalette> {
  const PetsPalette({
    required this.bg,
    required this.surface,
    required this.surfaceElevated,
    required this.text,
    required this.textMuted,
  });

  final Color bg;
  final Color surface;
  final Color surfaceElevated;
  final Color text;
  final Color textMuted;

  static const dark = PetsPalette(
    bg: AppColors.bg,
    surface: AppColors.surface,
    surfaceElevated: AppColors.surfaceElevated,
    text: AppColors.cream,
    textMuted: AppColors.textMuted,
  );

  static const light = PetsPalette(
    bg: AppColors.bgLight,
    surface: AppColors.surfaceLight,
    surfaceElevated: AppColors.surfaceElevatedLight,
    text: AppColors.creamLight,
    textMuted: AppColors.textMutedLight,
  );

  static PetsPalette of(BuildContext context) {
    return Theme.of(context).extension<PetsPalette>() ?? PetsPalette.dark;
  }

  @override
  PetsPalette copyWith({
    Color? bg,
    Color? surface,
    Color? surfaceElevated,
    Color? text,
    Color? textMuted,
  }) {
    return PetsPalette(
      bg: bg ?? this.bg,
      surface: surface ?? this.surface,
      surfaceElevated: surfaceElevated ?? this.surfaceElevated,
      text: text ?? this.text,
      textMuted: textMuted ?? this.textMuted,
    );
  }

  @override
  PetsPalette lerp(ThemeExtension<PetsPalette>? other, double t) {
    if (other is! PetsPalette) return this;
    return PetsPalette(
      bg: Color.lerp(bg, other.bg, t)!,
      surface: Color.lerp(surface, other.surface, t)!,
      surfaceElevated: Color.lerp(surfaceElevated, other.surfaceElevated, t)!,
      text: Color.lerp(text, other.text, t)!,
      textMuted: Color.lerp(textMuted, other.textMuted, t)!,
    );
  }
}
