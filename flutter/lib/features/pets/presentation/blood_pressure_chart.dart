import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';

class BloodPressureChart extends StatelessWidget {
  const BloodPressureChart({
    super.key,
    required this.points,
    this.height = 160,
  });

  final List<({DateTime date, int sys, int dia})> points;
  final double height;

  @override
  Widget build(BuildContext context) {
    if (points.isEmpty) {
      return SizedBox(height: height);
    }
    final sorted = List.of(points)..sort((a, b) => a.date.compareTo(b.date));
    final sysSpots = <FlSpot>[
      for (var i = 0; i < sorted.length; i++)
        FlSpot(i.toDouble(), sorted[i].sys.toDouble()),
    ];
    final diaSpots = <FlSpot>[
      for (var i = 0; i < sorted.length; i++)
        FlSpot(i.toDouble(), sorted[i].dia.toDouble()),
    ];
    final all = [
      ...sorted.map((p) => p.sys.toDouble()),
      ...sorted.map((p) => p.dia.toDouble()),
    ];
    final maxY = all.reduce((a, b) => a > b ? a : b) * 1.1;
    final minY = all.reduce((a, b) => a < b ? a : b) * 0.9;
    return SizedBox(
      height: height,
      child: LineChart(
        LineChartData(
          minY: minY < 0 ? 0 : minY,
          maxY: maxY <= 0 ? 200 : maxY,
          gridData: const FlGridData(show: false),
          borderData: FlBorderData(show: false),
          titlesData: FlTitlesData(
            leftTitles: const AxisTitles(
              sideTitles: SideTitles(showTitles: true, reservedSize: 36),
            ),
            rightTitles: const AxisTitles(
              sideTitles: SideTitles(showTitles: false),
            ),
            topTitles: const AxisTitles(
              sideTitles: SideTitles(showTitles: false),
            ),
            bottomTitles: AxisTitles(
              sideTitles: SideTitles(
                showTitles: true,
                getTitlesWidget: (value, meta) {
                  final i = value.toInt();
                  if (i < 0 || i >= sorted.length) {
                    return const SizedBox.shrink();
                  }
                  final d = sorted[i].date;
                  return Padding(
                    padding: const EdgeInsets.only(top: 6),
                    child: Text(
                      '${d.day}/${d.month}',
                      style: const TextStyle(fontSize: 10),
                    ),
                  );
                },
              ),
            ),
          ),
          lineBarsData: [
            LineChartBarData(
              spots: sysSpots,
              isCurved: true,
              color: AppColors.accent,
              barWidth: 3,
              dotData: const FlDotData(show: true),
            ),
            LineChartBarData(
              spots: diaSpots,
              isCurved: true,
              color: AppColors.primary,
              barWidth: 2,
              dashArray: const [4, 3],
              dotData: const FlDotData(show: true),
            ),
          ],
        ),
      ),
    );
  }
}
