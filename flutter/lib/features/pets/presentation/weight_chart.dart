import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';

class WeightChart extends StatelessWidget {
  const WeightChart({
    super.key,
    required this.points,
    this.height = 160,
  });

  final List<({DateTime date, double kg})> points;
  final double height;

  @override
  Widget build(BuildContext context) {
    if (points.isEmpty) {
      return SizedBox(height: height);
    }
    final sorted = List.of(points)..sort((a, b) => a.date.compareTo(b.date));
    final spots = <FlSpot>[
      for (var i = 0; i < sorted.length; i++)
        FlSpot(i.toDouble(), sorted[i].kg),
    ];
    final maxY = sorted.map((p) => p.kg).reduce((a, b) => a > b ? a : b) * 1.15;
    final minY = sorted.map((p) => p.kg).reduce((a, b) => a < b ? a : b) * 0.85;
    return SizedBox(
      height: height,
      child: LineChart(
        LineChartData(
          minY: minY < 0 ? 0 : minY,
          maxY: maxY <= 0 ? 10 : maxY,
          gridData: const FlGridData(show: false),
          borderData: FlBorderData(show: false),
          titlesData: FlTitlesData(
            leftTitles: const AxisTitles(sideTitles: SideTitles(showTitles: true, reservedSize: 36)),
            rightTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
            topTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
            bottomTitles: AxisTitles(
              sideTitles: SideTitles(
                showTitles: true,
                getTitlesWidget: (value, meta) {
                  final i = value.toInt();
                  if (i < 0 || i >= sorted.length) return const SizedBox.shrink();
                  final d = sorted[i].date;
                  return Padding(
                    padding: const EdgeInsets.only(top: 6),
                    child: Text('${d.day}/${d.month}', style: const TextStyle(fontSize: 10)),
                  );
                },
              ),
            ),
          ),
          lineBarsData: [
            LineChartBarData(
              spots: spots,
              isCurved: true,
              color: AppColors.primary,
              barWidth: 3,
              dotData: const FlDotData(show: true),
              belowBarData: BarAreaData(
                show: true,
                color: AppColors.primary.withValues(alpha: 0.12),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
