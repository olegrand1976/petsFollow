import 'package:flutter/material.dart';

/// Coordinates shell tab switches and deep-links from FCM taps.
class PushNavigation {
  PushNavigation._();
  static final instance = PushNavigation._();

  static const tabMessages = 3;
  static const tabPets = 1;

  final GlobalKey<NavigatorState> navigatorKey = GlobalKey<NavigatorState>();

  void Function(int index)? onSelectTab;
  void Function(String threadId)? onOpenMessageThread;
  void Function(String petId)? onOpenPetTimeline;

  final ValueNotifier<int> messageRefreshTick = ValueNotifier(0);

  void selectTab(int index) => onSelectTab?.call(index);

  void openMessages({String? threadId}) {
    selectTab(tabMessages);
    if (threadId != null && threadId.isNotEmpty) {
      onOpenMessageThread?.call(threadId);
    }
    bumpMessageRefresh();
  }

  void openPetTimeline(String petId) {
    if (petId.isEmpty) {
      selectTab(tabPets);
      return;
    }
    onOpenPetTimeline?.call(petId);
  }

  void bumpMessageRefresh() {
    messageRefreshTick.value = messageRefreshTick.value + 1;
  }

  void handlePushData(Map<String, dynamic> data) {
    final type = data['type']?.toString() ?? '';
    switch (type) {
      case 'message':
        openMessages(threadId: data['threadId']?.toString());
      case 'visit_confirmed':
      case 'visit_proposed':
      case 'visit_reschedule':
        openPetTimeline(data['petId']?.toString() ?? '');
      default:
        return;
    }
  }
}
