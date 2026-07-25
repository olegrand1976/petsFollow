import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';

enum PetsLogoVariant { horizontal, emblem }

class PetsLogo extends StatelessWidget {
  const PetsLogo({
    super.key,
    this.variant = PetsLogoVariant.horizontal,
    this.height = 32,
  });

  final PetsLogoVariant variant;
  final double height;

  @override
  Widget build(BuildContext context) {
    final asset = variant == PetsLogoVariant.horizontal
        ? 'assets/brand/petsfollow-horizontal.svg'
        : 'assets/brand/petsfollow-emblem.svg';

    return SvgPicture.asset(
      asset,
      height: height,
      semanticsLabel: 'petsFollow',
    );
  }
}
