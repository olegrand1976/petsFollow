import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/models/discovery_card.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class DiscoveryCardWidget extends StatelessWidget {
  const DiscoveryCardWidget({
    super.key,
    required this.card,
    this.mission = false,
    this.onComplete,
  });

  final DiscoveryCard card;
  final bool mission;
  final VoidCallback? onComplete;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final opacity = card.locked ? 0.45 : 1.0;

    return Opacity(
      opacity: opacity,
      child: Card(
        margin: EdgeInsets.only(bottom: mission ? 16 : 8),
        color: mission ? AppColors.gold.withValues(alpha: 0.12) : AppColors.surfaceElevated,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(20),
          side: mission
              ? BorderSide(color: AppColors.gold.withValues(alpha: 0.6))
              : BorderSide.none,
        ),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  CircleAvatar(
                    backgroundColor: card.completed
                        ? AppColors.primary.withValues(alpha: 0.2)
                        : AppColors.gold.withValues(alpha: 0.2),
                    child: card.completed
                        ? Icon(Icons.check, color: AppColors.primary, size: 20)
                        : Text(
                            l10n.discoveryDayBadge(card.dayIndex),
                            style: TextStyle(color: AppColors.gold, fontWeight: FontWeight.bold),
                          ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        if (mission)
                          Text(
                            l10n.discoveryMission,
                            style: TextStyle(
                              color: AppColors.gold,
                              fontWeight: FontWeight.w600,
                              fontSize: 12,
                            ),
                          ),
                        Text(
                          card.title,
                          style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 15),
                        ),
                      ],
                    ),
                  ),
                  if (card.locked)
                    Icon(Icons.lock_outline, size: 18, color: AppColors.textMuted)
                  else if (card.completed)
                    Icon(Icons.check_circle, color: AppColors.primary, size: 22),
                ],
              ),
              const SizedBox(height: 10),
              Text(
                card.body,
                style: TextStyle(color: AppColors.textMuted, height: 1.35),
              ),
              if (!card.completed && !card.locked && onComplete != null) ...[
                const SizedBox(height: 14),
                FilledButton(
                  onPressed: onComplete,
                  child: Text(l10n.discoveryMarkDone),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
