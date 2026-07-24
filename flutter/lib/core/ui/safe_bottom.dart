import 'dart:math' as math;

import 'package:flutter/widgets.dart';

/// System gesture / nav bar inset at the bottom of the screen.
double systemBottomInset(BuildContext context) {
  final mq = MediaQuery.of(context);
  return math.max(mq.viewPadding.bottom, mq.padding.bottom);
}

/// Soft-keyboard inset (`viewInsets.bottom`).
double keyboardBottomInset(BuildContext context) {
  return MediaQuery.of(context).viewInsets.bottom;
}

/// Scroll padding with an extra system bottom inset.
///
/// Use on push screens whose primary CTA sits at the end of a scroll view.
/// Prefer this over `SafeArea(bottom: true)` so padding is not applied twice.
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

/// Bottom padding for a sticky composer / footer bar.
///
/// - [embedded]: when true (e.g. above a shell `NavigationBar`), skip the
///   system inset — the nav bar already consumes it.
/// - Always adds [keyboardBottomInset] so the bar rises with the keyboard.
double composerBottomPadding(
  BuildContext context, {
  required bool embedded,
  double base = 8,
}) {
  final system = embedded ? 0.0 : systemBottomInset(context);
  return base + system + keyboardBottomInset(context);
}
