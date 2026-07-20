import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/notifications/push_navigation.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
import 'package:petsfollow_mobile/features/care/presentation/care_tab.dart';
import 'package:petsfollow_mobile/features/home/presentation/home_tab.dart';
import 'package:petsfollow_mobile/features/messaging/presentation/messaging_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_timeline_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pets_tab.dart';
import 'package:petsfollow_mobile/features/settings/presentation/settings_menu_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class MainShellScreen extends StatefulWidget {
  const MainShellScreen({
    super.key,
    required this.onLogout,
    this.billingRefreshTick = 0,
  });

  final VoidCallback onLogout;
  /// Bumps Home/Pets tabs after Stripe return without remounting the whole shell.
  final int billingRefreshTick;

  @override
  State<MainShellScreen> createState() => _MainShellScreenState();
}

class _MainShellScreenState extends State<MainShellScreen> {
  int _index = 0;

  @override
  void initState() {
    super.initState();
    final nav = PushNavigation.instance;
    nav.onSelectTab = (i) {
      if (!mounted) return;
      setState(() => _index = i);
    };
    nav.onOpenPetTimeline = (petId) {
      final navigator = nav.navigatorKey.currentState;
      if (navigator == null) return;
      navigator.push(
        MaterialPageRoute<void>(
          builder: (_) => PetTimelineScreen(petId: petId),
        ),
      );
    };
  }

  @override
  void dispose() {
    final nav = PushNavigation.instance;
    nav.onSelectTab = null;
    nav.onOpenPetTimeline = null;
    super.dispose();
  }

  void _onTabSelected(int i) {
    setState(() => _index = i);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Container(
      decoration: const BoxDecoration(gradient: AppTheme.gradientBg),
      child: Scaffold(
        backgroundColor: Colors.transparent,
        body: IndexedStack(
          index: _index,
          children: [
            HomeTab(
              key: ValueKey('home-${widget.billingRefreshTick}'),
              onNavigateToPets: () => setState(() => _index = 1),
            ),
            PetsTab(key: ValueKey('pets-${widget.billingRefreshTick}')),
            const CareTab(),
            MessagingScreen(embedded: true, active: _index == PushNavigation.tabMessages),
            SettingsMenuScreen(onLogout: widget.onLogout, embedded: true),
          ],
        ),
        bottomNavigationBar: NavigationBar(
          selectedIndex: _index,
          onDestinationSelected: _onTabSelected,
          backgroundColor: AppColors.surface,
          indicatorColor: AppColors.primary.withValues(alpha: 0.2),
          destinations: [
            NavigationDestination(
              icon: const Icon(Icons.home_outlined),
              selectedIcon: const Icon(Icons.home),
              label: l10n.navHome,
            ),
            NavigationDestination(
              icon: const Icon(Icons.pets_outlined),
              selectedIcon: const Icon(Icons.pets),
              label: l10n.navPets,
            ),
            NavigationDestination(
              icon: const Icon(Icons.medical_services_outlined),
              selectedIcon: const Icon(Icons.medical_services),
              label: l10n.navCare,
            ),
            NavigationDestination(
              icon: const Icon(Icons.chat_outlined),
              selectedIcon: const Icon(Icons.chat),
              label: l10n.navMessages,
            ),
            NavigationDestination(
              icon: const Icon(Icons.person_outline),
              selectedIcon: const Icon(Icons.person),
              label: l10n.navProfile,
            ),
          ],
        ),
      ),
    );
  }
}

/// Gradient scaffold wrapper used by tab screens.
class PetsTabScaffold extends StatelessWidget {
  const PetsTabScaffold({
    super.key,
    required this.title,
    this.actions,
    required this.body,
    this.floatingActionButton,
  });

  final Widget? title;
  final List<Widget>? actions;
  final Widget body;
  final Widget? floatingActionButton;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(gradient: AppTheme.gradientBg),
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          title: title,
          actions: actions,
        ),
        body: body,
        floatingActionButton: floatingActionButton,
      ),
    );
  }
}

/// Compact logo for app bars.
class PetsAppBarLogo extends StatelessWidget {
  const PetsAppBarLogo({super.key});

  @override
  Widget build(BuildContext context) {
    return const PetsLogo(variant: PetsLogoVariant.horizontal, height: 28);
  }
}
