/// Species that support heart-rate measurement (aligned with API kernel).
/// Empty/unknown species are not supported.
bool supportsHeartRateControl(String? species) {
  switch (species) {
    case 'dog':
    case 'cat':
    case 'horse':
      return true;
    default:
      return false;
  }
}
