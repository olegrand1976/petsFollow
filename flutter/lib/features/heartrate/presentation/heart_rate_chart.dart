import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';

class HeartRateChart extends StatelessWidget {
  const HeartRateChart({
    super.key,
    required this.points,
    this.height = 160,
  });

  final List<({DateTime date, int bpm, bool isAlert})> points;
  final double height;

  @override
  Widget build(BuildContext context) {
    if (points.isEmpty) {
      return SizedBox(height: height);
    }
    final sorted = List.of(points)..sort((a, b) => a.date.compareTo(b.date));
    final spots = <FlSpot>[];
    for (var i = 0; i < sorted.length; i++) {
      spots.add(FlSpot(i.toDouble(), sorted[i].bpm.toDouble()));
    }
    final maxY = (sorted.map((p) => p.bpm).reduce((a, b) => a > b ? a : b) + 20).toDouble();
    return SizedBox(
      height: height,
      child: BarChart(
        BarChartData(
          maxY: maxY < 80 ? 140 : maxY,
          minY: 0,
          gridData: const FlGridData(show: false),
          borderData: FlBorderData(show: false),
          titlesData: FlTitlesData(
            leftTitles: const AxisTitles(sideTitles: SideTitles(showTitles: true, reservedSize: 32)),
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
          barGroups: [
            for (var i = 0; i < sorted.length; i++)
              BarChartGroupData(
                x: i,
                barRods: [
                  BarChartRodData(
                    toY: sorted[i].bpm.toDouble(),
                    color: sorted[i].isAlert ? Colors.redAccent : Theme.of(context).colorScheme.primary,
                    width: 14,
                    borderRadius: const BorderRadius.vertical(top: Radius.circular(4)),
                  ),
                ],
              ),
          ],
        ),
      ),
    );
  }
}
