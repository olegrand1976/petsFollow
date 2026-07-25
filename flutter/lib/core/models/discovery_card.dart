class DiscoveryCard {
  const DiscoveryCard({
    required this.dayIndex,
    required this.title,
    required this.body,
    this.completed = false,
    this.locked = false,
  });

  /// Day index in the 7-day journey: 0, 2, 4, or 6.
  final int dayIndex;
  final String title;
  final String body;
  final bool completed;
  final bool locked;

  String get cardKey => 'day$dayIndex';

  static const journeyDays = [0, 2, 4, 6];
}
