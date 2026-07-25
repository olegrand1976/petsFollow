import 'package:flutter/foundation.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';

class FeatureModulesController extends ChangeNotifier {
  FeatureModulesController._();
  static final instance = FeatureModulesController._();

  bool carePlus = false;
  bool horse = false;
  bool kennel = false;
  bool family = false;
  bool loaded = false;

  Future<void> load() async {
    if (ApiClient.instance.token == null) {
      loaded = true;
      notifyListeners();
      return;
    }
    try {
      final m = await ApiClient.instance.getFeatureModules();
      carePlus = m['moduleCarePlus'] == true;
      horse = m['moduleHorse'] == true;
      kennel = m['moduleKennel'] == true;
      family = m['moduleFamily'] == true;
    } catch (_) {
      // defaults stay false
    }
    loaded = true;
    notifyListeners();
  }

  Future<void> save({
    bool? carePlus,
    bool? horse,
    bool? kennel,
    bool? family,
  }) async {
    final next = {
      'moduleCarePlus': carePlus ?? this.carePlus,
      'moduleHorse': horse ?? this.horse,
      'moduleKennel': kennel ?? this.kennel,
      'moduleFamily': family ?? this.family,
    };
    final m = await ApiClient.instance.updateFeatureModules(next);
    this.carePlus = m['moduleCarePlus'] == true;
    this.horse = m['moduleHorse'] == true;
    this.kennel = m['moduleKennel'] == true;
    this.family = m['moduleFamily'] == true;
    notifyListeners();
  }
}
