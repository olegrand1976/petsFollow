/**
 * Large InternalApi (200+ BFF routes) blows vue-tsc MatchedRoutes
 * (TS2589 / TS2345 NitroFetchOptions). Index signature keeps lookups shallow.
 * Casts at call sites remain OK; this avoids a flood of identical errors.
 */
export {}

declare module 'nitropack/types' {
  interface InternalApi {
    [path: string]: {
      [method: string]: any
    }
  }
}
