import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Canonical pet species codes (aligned with Pro / Go kernel).
const kPetSpeciesCodes = <String>[
  'dog',
  'cat',
  'horse',
  'donkey',
  'cattle',
  'sheep',
  'goat',
  'pig',
  'poultry',
  'rabbit',
  'alpaca',
  'llama',
  'other',
];

/// Species that show domicile / food-chain fields (équidés, rente, camélidés, lapin).
bool isFoodChainSpecies(String? species) {
  switch (species) {
    case 'horse':
    case 'donkey':
    case 'cattle':
    case 'sheep':
    case 'goat':
    case 'pig':
    case 'poultry':
    case 'rabbit':
    case 'alpaca':
    case 'llama':
      return true;
    default:
      return false;
  }
}

bool isKnownPetSpecies(String species) => kPetSpeciesCodes.contains(species);

String speciesLabel(AppLocalizations l10n, String species) {
  switch (species) {
    case 'dog':
      return l10n.speciesDog;
    case 'cat':
      return l10n.speciesCat;
    case 'horse':
      return l10n.speciesHorse;
    case 'donkey':
      return l10n.speciesDonkey;
    case 'cattle':
      return l10n.speciesCattle;
    case 'sheep':
      return l10n.speciesSheep;
    case 'goat':
      return l10n.speciesGoat;
    case 'pig':
      return l10n.speciesPig;
    case 'poultry':
      return l10n.speciesPoultry;
    case 'rabbit':
      return l10n.speciesRabbit;
    case 'alpaca':
      return l10n.speciesAlpaca;
    case 'llama':
      return l10n.speciesLlama;
    case 'other':
      return l10n.speciesOther;
    default:
      return l10n.speciesOther;
  }
}
