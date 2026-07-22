import 'dart:math' as math;

import 'package:flutter/widgets.dart';

/// System gesture / nav bar inset at the bottom of the screen.
double systemBottomInset(BuildContext context) {
  final mq = MediaQuery.of(context);
  return math.max(mq.viewPadding.bottom, mq.padding.bottom);
}

/// Scroll padding with an extra system bottom inset (for push screens without SafeArea).
EdgeInsets scrollPaddingWithSystemBottom(
  BuildContext context, {
  double all = 16,
  double? left,
  double? top,
  double? right,
  double? bottom,
}) {
  final l = left ?? all;
  final t = top ?? all;
  final r = right ?? all;
  final b = (bottom ?? all) + systemBottomInset(context);
  return EdgeInsets.fromLTRB(l, t, r, b);
}
