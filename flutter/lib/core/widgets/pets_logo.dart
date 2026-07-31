import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';

enum PetsLogoVariant {
  /// Emblème + wordmark (optionnellement « Pro »).
  horizontal,

  /// Emblème seul (patte + pulse).
  emblem,

  /// Wordmark seul (sous un emblème hero déjà affiché).
  wordmark,
}

class PetsLogo extends StatelessWidget {
  const PetsLogo({
    super.key,
    this.variant = PetsLogoVariant.horizontal,
    this.height = 32,
    this.showPro = false,
    this.excludeSemantics = false,
  });

  final PetsLogoVariant variant;
  final double height;

  /// Affiche le suffixe « Pro » quand le logo est le seul signal de face.
  final bool showPro;

  /// Ignore le nœud sémantique (ex. wordmark sous un emblème déjà labellisé).
  final bool excludeSemantics;

  static const _emblemAsset = 'assets/brand/petsfollow-emblem.svg';

  @override
  Widget build(BuildContext context) {
    switch (variant) {
      case PetsLogoVariant.emblem:
        return SvgPicture.asset(
          _emblemAsset,
          height: height,
          semanticsLabel: excludeSemantics ? null : 'petsFollow',
          excludeFromSemantics: excludeSemantics,
        );
      case PetsLogoVariant.wordmark:
        return _Wordmark(
          height: height,
          showPro: showPro,
          excludeSemantics: excludeSemantics,
        );
      case PetsLogoVariant.horizontal:
        return Semantics(
          label: showPro ? 'petsFollow Pro' : 'petsFollow',
          excludeSemantics: excludeSemantics,
          child: Row(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              SvgPicture.asset(
                _emblemAsset,
                height: height,
                excludeFromSemantics: true,
              ),
              SizedBox(width: height * 0.28),
              _Wordmark(
                height: height,
                showPro: showPro,
                excludeSemantics: true,
              ),
            ],
          ),
        );
    }
  }
}

class _Wordmark extends StatelessWidget {
  const _Wordmark({
    required this.height,
    required this.showPro,
    this.excludeSemantics = false,
  });

  final double height;
  final bool showPro;
  final bool excludeSemantics;

  @override
  Widget build(BuildContext context) {
    final p = PetsPalette.of(context);
    final nameSize = height * 0.72;
    final proSize = height * 0.48;
    final text = Text.rich(
      TextSpan(
        children: [
          TextSpan(
            text: 'petsFollow',
            style: TextStyle(
              fontSize: nameSize,
              fontWeight: FontWeight.w700,
              color: p.text,
              height: 1,
            ),
          ),
          if (showPro)
            TextSpan(
              text: ' Pro',
              style: TextStyle(
                fontSize: proSize,
                fontWeight: FontWeight.w600,
                color: AppColors.primary,
                height: 1,
              ),
            ),
        ],
      ),
      maxLines: 1,
      softWrap: false,
    );
    if (excludeSemantics) {
      return ExcludeSemantics(child: text);
    }
    return text;
  }
}
