import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/models/pet_species.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_en.dart';

void main() {
  final l10n = AppLocalizationsEn();

  test('known species include production animals', () {
    expect(kPetSpeciesCodes, containsAll(['cattle', 'pig', 'alpaca', 'llama', 'donkey']));
    expect(isKnownPetSpecies('sheep'), isTrue);
    expect(isKnownPetSpecies('dragon'), isFalse);
  });

  test('food-chain species', () {
    expect(isFoodChainSpecies('cattle'), isTrue);
    expect(isFoodChainSpecies('horse'), isTrue);
    expect(isFoodChainSpecies('alpaca'), isTrue);
    expect(isFoodChainSpecies('dog'), isFalse);
    expect(isFoodChainSpecies('other'), isFalse);
  });

  test('species labels', () {
    expect(speciesLabel(l10n, 'cattle'), 'Cattle');
    expect(speciesLabel(l10n, 'alpaca'), 'Alpaca');
    expect(speciesLabel(l10n, 'unknown'), 'Other');
  });
}
