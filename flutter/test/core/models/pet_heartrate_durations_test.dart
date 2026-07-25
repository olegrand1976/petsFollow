import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';

void main() {
  test('Pet.fromJson keeps practice heartrate durations', () {
    final pet = Pet.fromJson({
      'id': 'p1',
      'name': 'Rex',
      'species': 'dog',
      'breed': 'lab',
      'heartrateDurationsSec': [15, 30],
    });
    expect(pet.heartrateDurationsSec, [15, 30]);
  });

  test('Pet.fromJson defaults to 60 when durations absent', () {
    final pet = Pet.fromJson({
      'id': 'p1',
      'name': 'Rex',
      'species': 'dog',
      'breed': 'lab',
    });
    expect(pet.heartrateDurationsSec, [60]);
  });

  test('Pet.fromJson defaults to 60 when durations empty', () {
    final pet = Pet.fromJson({
      'id': 'p1',
      'name': 'Rex',
      'species': 'dog',
      'breed': 'lab',
      'heartrateDurationsSec': <int>[],
    });
    expect(pet.heartrateDurationsSec, [60]);
  });
}
