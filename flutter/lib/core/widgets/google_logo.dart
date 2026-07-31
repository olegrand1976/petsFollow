import 'package:flutter/material.dart';

/// Official-style Google « G » mark (4 brand colors), for Sign-In buttons.
class GoogleLogo extends StatelessWidget {
  const GoogleLogo({super.key, this.size = 20});

  final double size;

  @override
  Widget build(BuildContext context) {
    return Semantics(
      label: 'Google',
      excludeSemantics: true,
      child: SizedBox.square(
        dimension: size,
        child: const CustomPaint(painter: _GoogleGPainter()),
      ),
    );
  }
}

/// Stroke-based « G » so the hollow center stays transparent on any button bg.
class _GoogleGPainter extends CustomPainter {
  const _GoogleGPainter();

  static const _blue = Color(0xFF4285F4);
  static const _green = Color(0xFF34A853);
  static const _yellow = Color(0xFFFBBC05);
  static const _red = Color(0xFFEA4335);

  @override
  void paint(Canvas canvas, Size size) {
    final s = size.shortestSide;
    final c = Offset(s / 2, s / 2);
    final stroke = s * 0.175;
    final r = (s - stroke) / 2;
    final rect = Rect.fromCircle(center: c, radius: r);

    final paint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = stroke
      ..strokeCap = StrokeCap.butt;

    // Red (top → left).
    paint.color = _red;
    canvas.drawArc(rect, -2.35, 1.75, false, paint);

    // Yellow (bottom-left).
    paint.color = _yellow;
    canvas.drawArc(rect, -0.6, 1.0, false, paint);

    // Green (bottom).
    paint.color = _green;
    canvas.drawArc(rect, 0.4, 1.15, false, paint);

    // Blue arc (right) + horizontal bar of the G.
    paint.color = _blue;
    canvas.drawArc(rect, 1.55, 1.0, false, paint);

    paint
      ..style = PaintingStyle.fill
      ..strokeWidth = 0;
    canvas.drawRect(
      Rect.fromLTWH(c.dx - stroke * 0.1, c.dy - stroke / 2, r + stroke * 0.55, stroke),
      paint,
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
