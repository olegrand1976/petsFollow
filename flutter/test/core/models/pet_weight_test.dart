import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';

void main() {
  test('Pet.fromJson parses weightKg', () {
    final pet = Pet.fromJson({
      'id': 'p1',
      'name': 'Rex',
      'species': 'dog',
      'breed': 'Lab',
      'weightKg': 12.35,
      'heartrateDurationsSec': [15, 30],
    });
    expect(pet.weightKg, 12.35);
    expect(pet.heartrateDurationsSec, [15, 30]);
  });

  test('Pet.fromJson allows null weightKg', () {
    final pet = Pet.fromJson({
      'id': 'p1',
      'name': 'Rex',
      'species': 'dog',
      'breed': 'Lab',
    });
    expect(pet.weightKg, isNull);
  });
}
